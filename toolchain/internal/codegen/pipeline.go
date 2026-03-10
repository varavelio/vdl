package codegen

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/varavelio/vdl/toolchain/internal/codegen/plugintypes"
	"github.com/varavelio/vdl/toolchain/internal/core/analysis"
	"github.com/varavelio/vdl/toolchain/internal/core/ir"
	"github.com/varavelio/vdl/toolchain/internal/core/ir/irtypes"
	"github.com/varavelio/vdl/toolchain/internal/core/vfs"
	"github.com/varavelio/vdl/toolchain/internal/version"
)

// preparePlugins loads plugin scripts from disk and builds the input payload for
// each configured plugin.
func preparePlugins(plugins []runtimePlugin) ([]preparedPlugin, error) {
	prepared := make([]preparedPlugin, 0, len(plugins))
	for _, plugin := range plugins {
		scriptPath := plugin.Source.LocalPath
		if plugin.Source.Kind == pluginSourceKindRemote {
			scriptPath = plugin.Source.CachePath
		}

		scriptBytes, err := os.ReadFile(scriptPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read plugin script %q: %w", plugin.Source.DisplayName, err)
		}
		plugin.Source.ContentHash = sha256Digest(scriptBytes)

		input, err := buildPluginInput(plugin)
		if err != nil {
			return nil, err
		}

		prepared = append(prepared, preparedPlugin{
			Plugin: plugin,
			Script: string(scriptBytes),
			Input:  input,
		})
	}

	return prepared, nil
}

// buildPluginInput analyzes the plugin schema and converts its IR into the
// generated plugin input types.
func buildPluginInput(plugin runtimePlugin) (plugintypes.PluginInput, error) {
	fs := vfs.New()
	program, diagnostics := analysis.Analyze(fs, plugin.SchemaPath)
	if len(diagnostics) > 0 {
		return plugintypes.PluginInput{}, diagnosticsToError(diagnostics)
	}

	schema := ir.FromProgram(program)
	pluginIR, err := convertIRSchema(schema)
	if err != nil {
		return plugintypes.PluginInput{}, fmt.Errorf("failed to build plugin IR for %q: %w", plugin.Source.DisplayName, err)
	}

	return plugintypes.PluginInput{
		Version: version.Version,
		Ir:      pluginIR,
		Options: cloneStringMap(plugin.Options),
	}, nil
}

// convertIRSchema converts the core IR representation into the generated plugin
// IR types through JSON compatibility.
func convertIRSchema(schema *irtypes.IrSchema) (plugintypes.IrSchema, error) {
	data, err := json.Marshal(schema)
	if err != nil {
		return plugintypes.IrSchema{}, fmt.Errorf("failed to marshal IR schema: %w", err)
	}

	var pluginIR plugintypes.IrSchema
	if err := json.Unmarshal(data, &pluginIR); err != nil {
		return plugintypes.IrSchema{}, fmt.Errorf("failed to unmarshal IR into plugin input schema: %w", err)
	}

	return pluginIR, nil
}

// executePlugins runs all prepared plugins concurrently and fails fast once any
// plugin returns an execution or semantic error.
func executePlugins(prepared []preparedPlugin) ([]executedPlugin, error) {
	results := make([]executedPlugin, len(prepared))
	errCh := make(chan error, len(prepared))
	var wg sync.WaitGroup

	for i := range prepared {
		wg.Go(func() {
			plugin := prepared[i]
			output, err := runPlugin(plugin.Plugin.Source.DisplayName, plugin.Script, plugin.Input)
			if err != nil {
				errCh <- fmt.Errorf("plugin %q failed: %w", plugin.Plugin.Source.DisplayName, err)
				return
			}

			if pluginErrors := output.GetErrors(); len(pluginErrors) > 0 {
				errCh <- fmt.Errorf(
					"plugin %q reported errors:\n%s",
					plugin.Plugin.Source.DisplayName,
					formatPluginOutputErrors(pluginErrors),
				)
				return
			}

			results[i] = executedPlugin{Plugin: plugin.Plugin, Output: output}
		})
	}

	wg.Wait()
	close(errCh)

	if err, ok := <-errCh; ok {
		return nil, err
	}

	return results, nil
}

// formatPluginOutputErrors renders plugin-reported errors into the multi-line
// message format used by the CLI.
func formatPluginOutputErrors(errors []plugintypes.PluginOutputError) string {
	lines := make([]string, 0, len(errors))
	for _, pluginErr := range errors {
		line := pluginErr.GetMessage()
		if pluginErr.Position != nil {
			line = fmt.Sprintf(
				"%s:%d:%d: %s",
				pluginErr.Position.GetFile(),
				pluginErr.Position.GetLine(),
				pluginErr.Position.GetColumn(),
				pluginErr.GetMessage(),
			)
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

// planOutputWrites validates all generated files and builds the final set of
// writes to perform atomically.
func planOutputWrites(results []executedPlugin) (outputPlan, error) {
	plan := outputPlan{
		OutDirs: make([]string, 0, len(results)),
		Writes:  make([]outputWrite, 0),
	}
	seenOutDirs := make(map[string]bool, len(results))
	seenPaths := make(map[string]string)

	for _, result := range results {
		if !seenOutDirs[result.Plugin.OutDir] {
			seenOutDirs[result.Plugin.OutDir] = true
			plan.OutDirs = append(plan.OutDirs, result.Plugin.OutDir)
		}

		for _, file := range result.Output.GetFiles() {
			relativePath, absolutePath, err := resolveOutputWritePath(result.Plugin.OutDir, file.GetPath())
			if err != nil {
				return outputPlan{}, fmt.Errorf("plugin %q: %w", result.Plugin.Source.DisplayName, err)
			}

			if previousPlugin, exists := seenPaths[absolutePath]; exists {
				return outputPlan{}, fmt.Errorf(
					"plugins %q and %q both write to %q",
					previousPlugin,
					result.Plugin.Source.DisplayName,
					absolutePath,
				)
			}
			seenPaths[absolutePath] = result.Plugin.Source.DisplayName

			plan.Writes = append(plan.Writes, outputWrite{
				PluginName:   result.Plugin.Source.DisplayName,
				OutDir:       result.Plugin.OutDir,
				RelativePath: relativePath,
				AbsolutePath: absolutePath,
				Content:      file.GetContent(),
			})
		}
	}

	sortOutputDirs(plan.OutDirs)
	sort.Slice(plan.Writes, func(i, j int) bool {
		return plan.Writes[i].AbsolutePath < plan.Writes[j].AbsolutePath
	})

	return plan, nil
}

// resolveOutputWritePath validates a plugin-emitted relative path and resolves
// its absolute destination under outDir.
func resolveOutputWritePath(outDir, rawRelativePath string) (string, string, error) {
	relativePath := strings.TrimSpace(rawRelativePath)
	if relativePath == "" {
		return "", "", fmt.Errorf("plugin output path cannot be empty")
	}

	normalized := filepath.Clean(filepath.FromSlash(relativePath))
	if normalized == "." || normalized == ".." || filepath.IsAbs(normalized) {
		return "", "", fmt.Errorf("plugin output path %q is invalid: it must stay inside the outDir", rawRelativePath)
	}
	if strings.HasPrefix(normalized, ".."+string(filepath.Separator)) {
		return "", "", fmt.Errorf("plugin output path %q is invalid: it must stay inside the outDir", rawRelativePath)
	}

	absolutePath := filepath.Join(outDir, normalized)
	absolutePath, err := filepath.Abs(absolutePath)
	if err != nil {
		return "", "", fmt.Errorf("failed to resolve output path %q: %w", rawRelativePath, err)
	}

	relToOutDir, err := filepath.Rel(outDir, absolutePath)
	if err != nil {
		return "", "", fmt.Errorf("failed to validate output path %q: %w", rawRelativePath, err)
	}
	if relToOutDir == ".." || strings.HasPrefix(relToOutDir, ".."+string(filepath.Separator)) {
		return "", "", fmt.Errorf("plugin output path %q is invalid: it escapes the outDir", rawRelativePath)
	}

	return normalized, absolutePath, nil
}

// applyOutputWrites stages every generated file and then applies the final
// filesystem changes only after all validation has succeeded.
func applyOutputWrites(config runtimeConfig, plan outputPlan) error {
	stagingRoot, err := os.MkdirTemp(config.Dir, ".vdl-generate-")
	if err != nil {
		return fmt.Errorf("failed to create staging directory: %w", err)
	}
	defer os.RemoveAll(stagingRoot)

	stagingDirs := make(map[string]string, len(plan.OutDirs))
	for i, outDir := range plan.OutDirs {
		stageDir := filepath.Join(stagingRoot, fmt.Sprintf("outdir-%03d", i))
		if err := os.MkdirAll(stageDir, generatedDirMode); err != nil {
			return fmt.Errorf("failed to create staging directory %q: %w", stageDir, err)
		}
		stagingDirs[outDir] = stageDir
	}

	for _, write := range plan.Writes {
		stagePath := filepath.Join(stagingDirs[write.OutDir], write.RelativePath)
		if err := os.MkdirAll(filepath.Dir(stagePath), generatedDirMode); err != nil {
			return fmt.Errorf("failed to create staging parent for %q: %w", write.AbsolutePath, err)
		}
		if err := os.WriteFile(stagePath, []byte(write.Content), generatedFileMode); err != nil {
			return fmt.Errorf("failed to stage output file %q: %w", write.AbsolutePath, err)
		}
	}

	if config.Config.GetCleanOutDirOr(true) {
		return replaceOutputDirs(plan.OutDirs, stagingDirs)
	}

	return mergeOutputDirs(plan.OutDirs, plan.Writes, stagingDirs)
}

// replaceOutputDirs swaps each output directory with its staged replacement.
func replaceOutputDirs(outDirs []string, stagingDirs map[string]string) error {
	for i := len(outDirs) - 1; i >= 0; i-- {
		if err := os.RemoveAll(outDirs[i]); err != nil {
			return fmt.Errorf("failed to clean output directory %q: %w", outDirs[i], err)
		}
	}

	for _, outDir := range outDirs {
		if err := os.MkdirAll(filepath.Dir(outDir), generatedDirMode); err != nil {
			return fmt.Errorf("failed to create parent directory for %q: %w", outDir, err)
		}
		if err := os.Rename(stagingDirs[outDir], outDir); err != nil {
			return fmt.Errorf("failed to move staged output into %q: %w", outDir, err)
		}
	}

	return nil
}

// mergeOutputDirs writes staged files into existing output directories without
// removing pre-existing files.
func mergeOutputDirs(outDirs []string, writes []outputWrite, stagingDirs map[string]string) error {
	for _, outDir := range outDirs {
		if err := os.MkdirAll(outDir, generatedDirMode); err != nil {
			return fmt.Errorf("failed to create output directory %q: %w", outDir, err)
		}
	}

	for _, write := range writes {
		stagePath := filepath.Join(stagingDirs[write.OutDir], write.RelativePath)
		if err := os.MkdirAll(filepath.Dir(write.AbsolutePath), generatedDirMode); err != nil {
			return fmt.Errorf("failed to create parent directory for %q: %w", write.AbsolutePath, err)
		}
		data, err := os.ReadFile(stagePath)
		if err != nil {
			return fmt.Errorf("failed to read staged output for %q: %w", write.AbsolutePath, err)
		}
		if err := os.WriteFile(write.AbsolutePath, data, generatedFileMode); err != nil {
			return fmt.Errorf("failed to write output file %q: %w", write.AbsolutePath, err)
		}
	}

	return nil
}

// sortOutputDirs orders output directories so parent directories are processed
// before nested directories.
func sortOutputDirs(outDirs []string) {
	sort.Slice(outDirs, func(i, j int) bool {
		if len(outDirs[i]) == len(outDirs[j]) {
			return outDirs[i] < outDirs[j]
		}
		return len(outDirs[i]) < len(outDirs[j])
	})
}

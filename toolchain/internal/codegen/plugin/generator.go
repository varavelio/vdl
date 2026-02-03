package plugin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/varavelio/vdl/toolchain/internal/codegen/config/configtypes"
	"github.com/varavelio/vdl/toolchain/internal/codegen/plugin/plugintypes"
	"github.com/varavelio/vdl/toolchain/internal/core/ir/irtypes"
	"github.com/varavelio/vdl/toolchain/internal/version"
)

// Generator implements codegen.Generator for external plugins.
type Generator struct {
	config *configtypes.PluginTargetConfig
}

// New creates a new PluginGenerator.
func New(config *configtypes.PluginTargetConfig) *Generator {
	return &Generator{
		config: config,
	}
}

// Name returns the name of the generator.
func (g *Generator) Name() string {
	return "plugin"
}

// Generate executes the plugin and returns the generated files.
func (g *Generator) Generate(ctx context.Context, ir *irtypes.IrSchema, formattedSchema string) ([]plugintypes.PluginFile, error) {
	if len(g.config.Command) == 0 {
		return nil, fmt.Errorf("plugin command is empty")
	}

	cmdName := g.config.Command[0]
	cmdArgs := g.config.Command[1:]

	cmd := exec.CommandContext(ctx, cmdName, cmdArgs...)
	cmd.Stderr = os.Stderr // Stream stderr to user

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start plugin command: %w", err)
	}

	irJson, err := json.Marshal(ir)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal ir: %w", err)
	}

	typedIrSchema := plugintypes.IrSchema{}
	if err := json.Unmarshal(irJson, &typedIrSchema); err != nil {
		return nil, fmt.Errorf("failed to marshal ir: %w", err)
	}

	// Prepare input
	input := plugintypes.PluginInput{
		Version: version.Version,
		Ir:      typedIrSchema,
		Schema:  formattedSchema,
		Options: g.config.GetOptionsOr(map[string]string{}),
	}

	// Write input to stdin in a goroutine to avoid deadlock if plugin reads slowly
	go func() {
		defer stdin.Close()
		encoder := json.NewEncoder(stdin)
		if err := encoder.Encode(input); err != nil {
			// We can't easily propagate this error to the main thread,
			// but if writing fails, the plugin will likely fail or exit,
			// which we catch in cmd.Wait().
			// Ideally we could log this to stderr as well.
			fmt.Fprintf(os.Stderr, "vdl: failed to write to plugin stdin: %v\n", err)
		}
	}()

	// Read output from stdout
	outputBytes, err := io.ReadAll(stdout)
	if err != nil {
		return nil, fmt.Errorf("failed to read plugin stdout: %w", err)
	}

	if err := cmd.Wait(); err != nil {
		return nil, fmt.Errorf("plugin command failed: %w", err)
	}

	if len(bytes.TrimSpace(outputBytes)) == 0 {
		return nil, nil
	}

	var output plugintypes.PluginOutput
	if err := json.Unmarshal(outputBytes, &output); err != nil {
		return nil, nil
	}

	return output.GetFilesOr([]plugintypes.PluginFile{}), nil
}

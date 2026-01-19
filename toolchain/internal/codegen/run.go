package codegen

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/varavelio/vdl/toolchain/internal/codegen/dart"
	"github.com/varavelio/vdl/toolchain/internal/codegen/golang"
	"github.com/varavelio/vdl/toolchain/internal/codegen/openapi"
	"github.com/varavelio/vdl/toolchain/internal/codegen/playground"
	"github.com/varavelio/vdl/toolchain/internal/codegen/typescript"
	"github.com/varavelio/vdl/toolchain/internal/core/analysis"
	"github.com/varavelio/vdl/toolchain/internal/core/ir"
	"github.com/varavelio/vdl/toolchain/internal/core/vfs"
	"github.com/varavelio/vdl/toolchain/internal/formatter"
	"github.com/varavelio/vdl/toolchain/internal/util/filepathutil"
)

// Run runs the code generator and returns an error if one occurred.
func Run(configPath string) error {
	configBytes, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read %s config file: %s", configPath, err)
	}

	config := Config{}
	if err := config.UnmarshalAndValidate(configBytes); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	///////////////////////////////////////
	// PARSE AND ANALYZE THE VDL SCHEMA  //
	///////////////////////////////////////

	absConfigPath, err := filepathutil.NormalizeFromWD(configPath)
	if err != nil {
		return fmt.Errorf("failed to normalize config path: %w", err)
	}

	absConfigDir := filepath.Dir(absConfigPath)
	absSchemaPath := filepath.Join(absConfigDir, config.Schema)

	// Create virtual filesystem and analyze
	fs := vfs.New()
	program, diagnostics := analysis.Analyze(fs, absSchemaPath)
	if len(diagnostics) > 0 {
		// Collect all error messages
		var errMsgs []string
		for _, d := range diagnostics {
			errMsgs = append(errMsgs, d.String())
		}
		return fmt.Errorf("schema validation failed:\n%s", joinErrors(errMsgs))
	}

	// Convert analysis.Program to IR Schema
	schema := ir.FromProgram(program)

	/////////////////////////
	// RUN CODE GENERATORS //
	/////////////////////////

	ctx := context.Background()

	if config.HasOpenAPI() {
		if err := runOpenAPI(ctx, absConfigDir, config.OpenAPI, schema); err != nil {
			return fmt.Errorf("failed to run openapi code generator: %w", err)
		}
	}

	if config.HasPlayground() {
		// Get formatted schema for playground
		formattedSchema := getFormattedSchema(fs, absSchemaPath)
		if err := runPlayground(ctx, absConfigDir, config.Playground, config.OpenAPI, schema, formattedSchema); err != nil {
			return fmt.Errorf("failed to run playground code generator: %w", err)
		}
	}

	if config.HasGolangServer() {
		cfg := *config.GolangServer
		cfg.IncludeServer = true
		cfg.IncludeClient = false
		if err := runGolang(ctx, absConfigDir, &cfg, schema); err != nil {
			return fmt.Errorf("failed to run golang-server code generator: %w", err)
		}
	}

	if config.HasGolangClient() {
		cfg := *config.GolangClient
		cfg.IncludeServer = false
		cfg.IncludeClient = true
		if err := runGolang(ctx, absConfigDir, &cfg, schema); err != nil {
			return fmt.Errorf("failed to run golang-client code generator: %w", err)
		}
	}

	if config.HasTypescriptClient() {
		cfg := *config.TypescriptClient
		cfg.IncludeServer = false
		cfg.IncludeClient = true
		if err := runTypescript(ctx, absConfigDir, &cfg, schema); err != nil {
			return fmt.Errorf("failed to run typescript-client code generator: %w", err)
		}
	}

	if config.HasDartClient() {
		if err := runDart(ctx, absConfigDir, config.DartClient, schema); err != nil {
			return fmt.Errorf("failed to run dart-client code generator: %w", err)
		}
	}

	return nil
}

// joinErrors joins multiple error messages with newlines.
func joinErrors(errs []string) string {
	result := ""
	for i, e := range errs {
		if i > 0 {
			result += "\n"
		}
		result += e
	}
	return result
}

// getFormattedSchema reads and formats the schema file.
func getFormattedSchema(fs *vfs.FileSystem, absSchemaPath string) string {
	content, err := fs.ReadFile(absSchemaPath)
	if err != nil {
		return ""
	}
	formatted, err := formatter.Format(absSchemaPath, string(content))
	if err != nil {
		return string(content) // Return original if formatting fails
	}
	return formatted
}

func runOpenAPI(ctx context.Context, absConfigDir string, config openapi.Config, schema *ir.Schema) error {
	outputFile := filepath.Join(absConfigDir, config.OutputFile)
	outputDir := filepath.Dir(outputFile)

	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Generate the code using new Generator interface
	gen := openapi.New(config)
	files, err := gen.Generate(ctx, schema)
	if err != nil {
		return fmt.Errorf("failed to generate code: %w", err)
	}

	// Write the generated files
	for _, file := range files {
		outPath := filepath.Join(outputDir, file.RelativePath)
		if err := os.WriteFile(outPath, file.Content, 0644); err != nil {
			return fmt.Errorf("failed to write generated code to file: %w", err)
		}
	}

	return nil
}

func runPlayground(ctx context.Context, absConfigDir string, config *playground.Config, openAPIConfig openapi.Config, schema *ir.Schema, formattedSchema string) error {
	outputDir := filepath.Join(absConfigDir, config.OutputDir)

	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Set the formatted schema in the config
	config.FormattedSchema = formattedSchema

	// Generate the playground using new Generator interface
	gen := playground.New(*config)
	files, err := gen.Generate(ctx, schema)
	if err != nil {
		return fmt.Errorf("failed to generate playground: %w", err)
	}

	// Write all generated files
	for _, file := range files {
		outPath := filepath.Join(outputDir, file.RelativePath)
		outDir := filepath.Dir(outPath)
		if err := os.MkdirAll(outDir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}
		if err := os.WriteFile(outPath, file.Content, 0644); err != nil {
			return fmt.Errorf("failed to write generated file %s: %w", outPath, err)
		}
	}

	// Generate the openapi.yaml file for the playground
	openAPIOutputFile := filepath.Join(outputDir, "openapi.yaml")
	openAPIConfig.OutputFile = "openapi.yaml"
	if openAPIConfig.BaseURL == "" {
		openAPIConfig.BaseURL = config.DefaultBaseURL
	}

	openAPIGen := openapi.New(openAPIConfig)
	openAPIFiles, err := openAPIGen.Generate(ctx, schema)
	if err != nil {
		return fmt.Errorf("failed to generate openapi.yaml code: %w", err)
	}

	for _, file := range openAPIFiles {
		outPath := filepath.Join(outputDir, file.RelativePath)
		if err := os.WriteFile(outPath, file.Content, 0644); err != nil {
			return fmt.Errorf("failed to write generated openapi.yaml code to file: %w", err)
		}
	}

	_ = openAPIOutputFile // Silence unused variable warning

	return nil
}

func runGolang(ctx context.Context, absConfigDir string, config *golang.Config, schema *ir.Schema) error {
	outputFile := filepath.Join(absConfigDir, config.OutputFile)
	outputDir := filepath.Dir(outputFile)

	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Generate the code using new Generator interface
	gen := golang.New(*config)
	files, err := gen.Generate(ctx, schema)
	if err != nil {
		return fmt.Errorf("failed to generate code: %w", err)
	}

	// Write the generated files
	for _, file := range files {
		outPath := filepath.Join(outputDir, file.RelativePath)
		outDir := filepath.Dir(outPath)
		if err := os.MkdirAll(outDir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}
		if err := os.WriteFile(outPath, file.Content, 0644); err != nil {
			return fmt.Errorf("failed to write generated code to file: %w", err)
		}
	}

	return nil
}

func runTypescript(ctx context.Context, absConfigDir string, config *typescript.Config, schema *ir.Schema) error {
	outputFile := filepath.Join(absConfigDir, config.OutputFile)
	outputDir := filepath.Dir(outputFile)

	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Generate the code using new Generator interface
	gen := typescript.New(*config)
	files, err := gen.Generate(ctx, schema)
	if err != nil {
		return fmt.Errorf("failed to generate code: %w", err)
	}

	// Write the generated files
	for _, file := range files {
		outPath := filepath.Join(outputDir, file.RelativePath)
		outDir := filepath.Dir(outPath)
		if err := os.MkdirAll(outDir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}
		if err := os.WriteFile(outPath, file.Content, 0644); err != nil {
			return fmt.Errorf("failed to write generated code to file: %w", err)
		}
	}

	return nil
}

func runDart(ctx context.Context, absConfigDir string, config *dart.Config, schema *ir.Schema) error {
	outputDir := filepath.Join(absConfigDir, config.OutputDir)

	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Generate the code using new Generator interface
	gen := dart.New(*config)
	files, err := gen.Generate(ctx, schema)
	if err != nil {
		return fmt.Errorf("failed to generate code: %w", err)
	}

	// Write the generated files
	for _, file := range files {
		outputFile := filepath.Join(outputDir, file.RelativePath)
		outputFileDir := filepath.Dir(outputFile)
		if err := os.MkdirAll(outputFileDir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}

		if err := os.WriteFile(outputFile, file.Content, 0644); err != nil {
			return fmt.Errorf("failed to write generated code to file %s: %w", outputFile, err)
		}
	}

	return nil
}

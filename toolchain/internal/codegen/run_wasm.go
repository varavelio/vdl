package codegen

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/varavelio/vdl/toolchain/internal/codegen/config"
	"github.com/varavelio/vdl/toolchain/internal/codegen/dart"
	"github.com/varavelio/vdl/toolchain/internal/codegen/golang"
	"github.com/varavelio/vdl/toolchain/internal/codegen/jsonschema"
	"github.com/varavelio/vdl/toolchain/internal/codegen/python"
	"github.com/varavelio/vdl/toolchain/internal/codegen/typescript"
	"github.com/varavelio/vdl/toolchain/internal/core/analysis"
	"github.com/varavelio/vdl/toolchain/internal/core/ir"
	"github.com/varavelio/vdl/toolchain/internal/core/parser"
)

// RunWasmOptions contains options for running code generators in WASM mode
// without writing to files.
type RunWasmOptions struct {
	// Generator must be one of: "go", "typescript", "dart", "jsonschema".
	Generator string `json:"generator"`
	// SchemaInput is the schema content as a string (VDL schema only).
	SchemaInput string `json:"schemaInput"`
	// PackageName is required for "go" and "dart" generators.
	PackageName string `json:"packageName"`
	// GoGenClient determines if the Go client should be generated (default: false).
	GoGenClient bool `json:"goGenClient"`
	// GoGenServer determines if the Go server should be generated (default: false).
	GoGenServer bool `json:"goGenServer"`
}

// RunWasmOutputFile is a single generated file.
type RunWasmOutputFile struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

// RunWasmOutput is the output of the RunWasm function.
type RunWasmOutput struct {
	Files []RunWasmOutputFile `json:"files"`
}

// RunWasmString is a wrapper around RunWasm that returns the generated code as a string.
func RunWasmString(opts RunWasmOptions) (string, error) {
	output, err := runWasm(opts)
	if err != nil {
		return "", err
	}

	jsonOutput, err := json.Marshal(output)
	if err != nil {
		return "", fmt.Errorf("failed to marshal output: %w", err)
	}

	return string(jsonOutput), nil
}

// runWasm executes a single generator and returns the generated code as a string.
func runWasm(opts RunWasmOptions) (RunWasmOutput, error) {
	if opts.Generator == "" {
		return RunWasmOutput{}, fmt.Errorf("missing generator")
	}
	if opts.SchemaInput == "" {
		return RunWasmOutput{}, fmt.Errorf("missing schema input")
	}

	// Parse input into AST
	astSchema, err := parser.ParserInstance.ParseString("schema.vdl", opts.SchemaInput)
	if err != nil {
		return RunWasmOutput{}, fmt.Errorf("failed to parse VDL schema: %s", err)
	}

	// Run semantic analysis on the parsed schema
	program, diagnostics := analysis.AnalyzeSchema(astSchema, "/virtual/schema.vdl")
	if len(diagnostics) > 0 {
		// Collect all error messages
		var errMsgs string
		for i, d := range diagnostics {
			if i > 0 {
				errMsgs += "\n"
			}
			errMsgs += d.String()
		}
		return RunWasmOutput{}, fmt.Errorf("schema validation failed:\n%s", errMsgs)
	}

	// Convert to IR Schema
	schema := ir.FromProgram(program)

	ctx := context.Background()

	switch opts.Generator {
	case "go":
		if opts.PackageName == "" {
			return RunWasmOutput{}, fmt.Errorf("go generator requires 'packageName'")
		}
		cfg := &config.GoConfig{
			Package: opts.PackageName,
			CommonConfig: config.CommonConfig{
				Output: "vdl.go",
			},
			ServerConfig: config.ServerConfig{GenServer: opts.GoGenServer},
			ClientConfig: config.ClientConfig{GenClient: opts.GoGenClient},
		}
		gen := golang.New(cfg)
		files, err := gen.Generate(ctx, schema)
		if err != nil {
			return RunWasmOutput{}, fmt.Errorf("failed to generate go code: %s", err)
		}
		return convertGolangFiles(files), nil

	case "typescript":
		cfg := &config.TypeScriptConfig{
			CommonConfig: config.CommonConfig{
				Output: "src",
			},
			ClientConfig: config.ClientConfig{GenClient: true},
			ServerConfig: config.ServerConfig{GenServer: false},
		}
		gen := typescript.New(cfg)
		files, err := gen.Generate(ctx, schema)
		if err != nil {
			return RunWasmOutput{}, fmt.Errorf("failed to generate typescript code: %s", err)
		}
		return convertTypescriptFiles(files), nil

	case "dart":
		if opts.PackageName == "" {
			return RunWasmOutput{}, fmt.Errorf("dart generator requires 'packageName'")
		}
		cfg := &config.DartConfig{
			CommonConfig: config.CommonConfig{
				Output: "lib",
			},
		}
		gen := dart.New(cfg)
		files, err := gen.Generate(ctx, schema)
		if err != nil {
			return RunWasmOutput{}, fmt.Errorf("failed to generate dart code: %s", err)
		}
		return convertDartFiles(files), nil

	case "python":
		cfg := &config.PythonConfig{
			CommonConfig: config.CommonConfig{
				Output: ".",
			},
		}
		gen := python.New(cfg)
		files, err := gen.Generate(ctx, schema)
		if err != nil {
			return RunWasmOutput{}, fmt.Errorf("failed to generate python code: %s", err)
		}
		return convertPythonFiles(files), nil

	case "jsonschema":
		cfg := &config.JSONSchemaConfig{
			CommonConfig: config.CommonConfig{
				Output: ".",
			},
			Filename: "schema.json",
		}
		gen := jsonschema.New(cfg)
		files, err := gen.Generate(ctx, schema)
		if err != nil {
			return RunWasmOutput{}, fmt.Errorf("failed to generate jsonschema: %s", err)
		}
		return convertJSONSchemaFiles(files), nil

	default:
		return RunWasmOutput{}, fmt.Errorf("unsupported generator: %s", opts.Generator)
	}
}

// convertGolangFiles converts golang.File slice to RunWasmOutput.
func convertGolangFiles(files []golang.File) RunWasmOutput {
	outputFiles := make([]RunWasmOutputFile, len(files))
	for i, file := range files {
		outputFiles[i] = RunWasmOutputFile{
			Path:    file.RelativePath,
			Content: string(file.Content),
		}
	}
	return RunWasmOutput{Files: outputFiles}
}

// convertTypescriptFiles converts typescript.File slice to RunWasmOutput.
func convertTypescriptFiles(files []typescript.File) RunWasmOutput {
	outputFiles := make([]RunWasmOutputFile, len(files))
	for i, file := range files {
		outputFiles[i] = RunWasmOutputFile{
			Path:    file.RelativePath,
			Content: string(file.Content),
		}
	}
	return RunWasmOutput{Files: outputFiles}
}

// convertPythonFiles converts python.File slice to RunWasmOutput.
func convertPythonFiles(files []python.File) RunWasmOutput {
	outputFiles := make([]RunWasmOutputFile, len(files))
	for i, file := range files {
		outputFiles[i] = RunWasmOutputFile{
			Path:    file.RelativePath,
			Content: string(file.Content),
		}
	}
	return RunWasmOutput{Files: outputFiles}
}

// convertDartFiles converts dart.File slice to RunWasmOutput.
func convertDartFiles(files []dart.File) RunWasmOutput {
	outputFiles := make([]RunWasmOutputFile, len(files))
	for i, file := range files {
		outputFiles[i] = RunWasmOutputFile{
			Path:    file.RelativePath,
			Content: string(file.Content),
		}
	}
	return RunWasmOutput{Files: outputFiles}
}

// convertJSONSchemaFiles converts jsonschema.File slice to RunWasmOutput.
func convertJSONSchemaFiles(files []jsonschema.File) RunWasmOutput {
	outputFiles := make([]RunWasmOutputFile, len(files))
	for i, file := range files {
		outputFiles[i] = RunWasmOutputFile{
			Path:    file.RelativePath,
			Content: string(file.Content),
		}
	}
	return RunWasmOutput{Files: outputFiles}
}

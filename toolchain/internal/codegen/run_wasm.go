package codegen

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/varavelio/vdl/toolchain/internal/codegen/dart"
	"github.com/varavelio/vdl/toolchain/internal/codegen/golang"
	"github.com/varavelio/vdl/toolchain/internal/codegen/typescript"
	"github.com/varavelio/vdl/toolchain/internal/core/analysis"
	"github.com/varavelio/vdl/toolchain/internal/core/ir"
	"github.com/varavelio/vdl/toolchain/internal/core/parser"
)

// RunWasmOptions contains options for running code generators in WASM mode
// without writing to files.
type RunWasmOptions struct {
	// Generator must be one of: "golang-server", "golang-client", "typescript-client", "dart-client".
	Generator string `json:"generator"`
	// SchemaInput is the schema content as a string (VDL schema only).
	SchemaInput string `json:"schemaInput"`
	// GolangPackageName is required when Generator is golang-server or golang-client.
	GolangPackageName string `json:"golangPackageName"`
	// DartPackageName is required when Generator is dart-client.
	DartPackageName string `json:"dartPackageName"`
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

	if opts.Generator == "golang-server" {
		if opts.GolangPackageName == "" {
			return RunWasmOutput{}, fmt.Errorf("golang-server requires 'GolangPackageName'")
		}
		cfg := golang.Config{PackageName: opts.GolangPackageName, IncludeServer: true, IncludeClient: false}
		gen := golang.New(cfg)
		files, err := gen.Generate(ctx, schema)
		if err != nil {
			return RunWasmOutput{}, fmt.Errorf("failed to generate golang server: %s", err)
		}
		return convertGolangFiles(files), nil
	}

	if opts.Generator == "golang-client" {
		if opts.GolangPackageName == "" {
			return RunWasmOutput{}, fmt.Errorf("golang-client requires 'GolangPackageName'")
		}
		cfg := golang.Config{PackageName: opts.GolangPackageName, IncludeServer: false, IncludeClient: true}
		gen := golang.New(cfg)
		files, err := gen.Generate(ctx, schema)
		if err != nil {
			return RunWasmOutput{}, fmt.Errorf("failed to generate golang client: %s", err)
		}
		return convertGolangFiles(files), nil
	}

	if opts.Generator == "typescript-client" {
		cfg := typescript.Config{IncludeServer: false, IncludeClient: true}
		gen := typescript.New(cfg)
		files, err := gen.Generate(ctx, schema)
		if err != nil {
			return RunWasmOutput{}, fmt.Errorf("failed to generate typescript client: %s", err)
		}
		return convertTypescriptFiles(files), nil
	}

	if opts.Generator == "dart-client" {
		cfg := dart.Config{PackageName: opts.DartPackageName}
		gen := dart.New(cfg)
		files, err := gen.Generate(ctx, schema)
		if err != nil {
			return RunWasmOutput{}, fmt.Errorf("failed to generate dart client: %s", err)
		}
		return convertDartFiles(files), nil
	}

	return RunWasmOutput{}, fmt.Errorf("unsupported generator: %s", opts.Generator)
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

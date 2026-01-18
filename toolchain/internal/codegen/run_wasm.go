package codegen

import (
	"encoding/json"
	"fmt"

	"github.com/uforg/uforpc/urpc/internal/codegen/dart"
	"github.com/uforg/uforpc/urpc/internal/codegen/golang"
	"github.com/uforg/uforpc/urpc/internal/codegen/typescript"
	"github.com/uforg/uforpc/urpc/internal/transpile"
	"github.com/uforg/uforpc/urpc/internal/urpc/parser"
)

// RunWasmOptions contains options for running code generators in WASM mode
// without writing to files.
type RunWasmOptions struct {
	// Generator must be one of: "golang-server", "golang-client", "typescript-client", "dart-client".
	Generator string `json:"generator"`
	// SchemaInput is the schema content as a string (URPC schema only).
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

	// Parse input into JSON schema
	astSchema, err := parser.ParserInstance.ParseString("schema.urpc", opts.SchemaInput)
	if err != nil {
		return RunWasmOutput{}, fmt.Errorf("failed to parse URPC schema: %s", err)
	}
	jsonSchema, err := transpile.ToJSON(*astSchema)
	if err != nil {
		return RunWasmOutput{}, fmt.Errorf("failed to transpile URPC to JSON: %s", err)
	}

	if opts.Generator == "golang-server" {
		if opts.GolangPackageName == "" {
			return RunWasmOutput{}, fmt.Errorf("golang-server requires 'GolangPackageName'")
		}
		cfg := golang.Config{PackageName: opts.GolangPackageName, IncludeServer: true, IncludeClient: false}
		fileContent, err := golang.Generate(jsonSchema, cfg)
		if err != nil {
			return RunWasmOutput{}, fmt.Errorf("failed to generate golang server: %s", err)
		}
		return RunWasmOutput{Files: []RunWasmOutputFile{{Path: "server.go", Content: fileContent}}}, nil
	}

	if opts.Generator == "golang-client" {
		if opts.GolangPackageName == "" {
			return RunWasmOutput{}, fmt.Errorf("golang-client requires 'GolangPackageName'")
		}
		cfg := golang.Config{PackageName: opts.GolangPackageName, IncludeServer: false, IncludeClient: true}
		fileContent, err := golang.Generate(jsonSchema, cfg)
		if err != nil {
			return RunWasmOutput{}, fmt.Errorf("failed to generate golang client: %s", err)
		}
		return RunWasmOutput{Files: []RunWasmOutputFile{{Path: "client.go", Content: fileContent}}}, nil
	}

	if opts.Generator == "typescript-client" {
		cfg := typescript.Config{IncludeServer: false, IncludeClient: true}
		fileContent, err := typescript.Generate(jsonSchema, cfg)
		if err != nil {
			return RunWasmOutput{}, fmt.Errorf("failed to generate typescript client: %s", err)
		}
		return RunWasmOutput{Files: []RunWasmOutputFile{{Path: "client.ts", Content: fileContent}}}, nil
	}

	if opts.Generator == "dart-client" {
		cfg := dart.Config{PackageName: opts.DartPackageName}
		output, err := dart.Generate(jsonSchema, cfg)
		if err != nil {
			return RunWasmOutput{}, fmt.Errorf("failed to generate dart client: %s", err)
		}
		files := make([]RunWasmOutputFile, len(output.Files))
		for i, file := range output.Files {
			files[i] = RunWasmOutputFile{Path: file.Path, Content: file.Content}
		}
		return RunWasmOutput{Files: files}, nil
	}

	return RunWasmOutput{}, fmt.Errorf("unsupported generator: %s", opts.Generator)
}

package codegen

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/uforg/uforpc/urpc/internal/codegen/dart"
	"github.com/uforg/uforpc/urpc/internal/codegen/golang"
	"github.com/uforg/uforpc/urpc/internal/codegen/openapi"
	"github.com/uforg/uforpc/urpc/internal/codegen/playground"
	"github.com/uforg/uforpc/urpc/internal/codegen/typescript"
	"github.com/uforg/uforpc/urpc/internal/schema"
	"github.com/uforg/uforpc/urpc/internal/transpile"
	"github.com/uforg/uforpc/urpc/internal/urpc/analyzer"
	"github.com/uforg/uforpc/urpc/internal/urpc/ast"
	"github.com/uforg/uforpc/urpc/internal/urpc/docstore"
	"github.com/uforg/uforpc/urpc/internal/util/filepathutil"
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
	// PARSE AND ANALYZE THE URPC SCHEMA //
	///////////////////////////////////////

	absConfigPath, err := filepathutil.NormalizeFromWD(configPath)
	if err != nil {
		return fmt.Errorf("failed to normalize config path: %w", err)
	}

	absConfigDir := filepath.Dir(absConfigPath)
	absSchemaPath := filepath.Join(absConfigDir, config.Schema)

	an, err := analyzer.NewAnalyzer(docstore.NewDocstore())
	if err != nil {
		return fmt.Errorf("failed to create URPC analyzer: %w", err)
	}

	astSchema, _, err := an.Analyze(absSchemaPath)
	if err != nil {
		return fmt.Errorf("invalid schema: %w", err)
	}

	///////////////////////
	// TRANSPILE TO JSON //
	///////////////////////

	jsonSchema, err := transpile.ToJSON(*astSchema)
	if err != nil {
		return fmt.Errorf("failed to transpile schema to its JSON representation: %w", err)
	}

	/////////////////////////
	// RUN CODE GENERATORS //
	/////////////////////////

	if config.HasOpenAPI() {
		if err := runOpenAPI(absConfigDir, config.OpenAPI, jsonSchema); err != nil {
			return fmt.Errorf("failed to run openapi code generator: %w", err)
		}
	}

	if config.HasPlayground() {
		if err := runPlayground(absConfigDir, config.Playground, config.OpenAPI, astSchema, jsonSchema); err != nil {
			return fmt.Errorf("failed to run playground code generator: %w", err)
		}
	}

	if config.HasGolangServer() {
		cfg := *config.GolangServer
		cfg.IncludeServer = true
		cfg.IncludeClient = false
		if err := runGolang(absConfigDir, &cfg, jsonSchema); err != nil {
			return fmt.Errorf("failed to run golang-server code generator: %w", err)
		}
	}

	if config.HasGolangClient() {
		cfg := *config.GolangClient
		cfg.IncludeServer = false
		cfg.IncludeClient = true
		if err := runGolang(absConfigDir, &cfg, jsonSchema); err != nil {
			return fmt.Errorf("failed to run golang-client code generator: %w", err)
		}
	}

	if config.HasTypescriptClient() {
		cfg := *config.TypescriptClient
		cfg.IncludeServer = false
		cfg.IncludeClient = true
		if err := runTypescript(absConfigDir, &cfg, jsonSchema); err != nil {
			return fmt.Errorf("failed to run typescript-client code generator: %w", err)
		}
	}

	if config.HasDartClient() {
		if err := runDart(absConfigDir, config.DartClient, jsonSchema); err != nil {
			return fmt.Errorf("failed to run dart-client code generator: %w", err)
		}
	}

	return nil
}

func runOpenAPI(absConfigDir string, config openapi.Config, schema schema.Schema) error {
	outputFile := filepath.Join(absConfigDir, config.OutputFile)
	outputDir := filepath.Dir(outputFile)

	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Generate the code
	code, err := openapi.Generate(schema, config)
	if err != nil {
		return fmt.Errorf("failed to generate code: %w", err)
	}

	if err := os.WriteFile(outputFile, []byte(code), 0644); err != nil {
		return fmt.Errorf("failed to write generated code to file: %w", err)
	}

	return nil
}

func runPlayground(absConfigDir string, config *playground.Config, openAPIConfig openapi.Config, astSchema *ast.Schema, jsonSchema schema.Schema) error {
	outputDir := filepath.Join(absConfigDir, config.OutputDir)
	openAPIOutputFile := filepath.Join(outputDir, "openapi.yaml")

	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Generate the playground
	err := playground.Generate(absConfigDir, astSchema, *config)
	if err != nil {
		return fmt.Errorf("failed to generate playground: %w", err)
	}

	// Generate the openapi.yaml file
	openAPIConfig.OutputFile = openAPIOutputFile
	if openAPIConfig.BaseURL == "" {
		openAPIConfig.BaseURL = config.DefaultBaseURL
	}

	code, err := openapi.Generate(jsonSchema, openAPIConfig)
	if err != nil {
		return fmt.Errorf("failed to generate openapi.yaml code: %w", err)
	}

	if err := os.WriteFile(openAPIOutputFile, []byte(code), 0644); err != nil {
		return fmt.Errorf("failed to write generated openapi.yaml code to file: %w", err)
	}

	return nil
}

func runGolang(absConfigDir string, config *golang.Config, schema schema.Schema) error {
	outputFile := filepath.Join(absConfigDir, config.OutputFile)
	outputDir := filepath.Dir(outputFile)

	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Generate the code
	code, err := golang.Generate(schema, *config)
	if err != nil {
		return fmt.Errorf("failed to generate code: %w", err)
	}

	// Write the code to the output file
	if err := os.WriteFile(outputFile, []byte(code), 0644); err != nil {
		return fmt.Errorf("failed to write generated code to file: %w", err)
	}

	return nil
}

func runTypescript(absConfigDir string, config *typescript.Config, schema schema.Schema) error {
	outputFile := filepath.Join(absConfigDir, config.OutputFile)
	outputDir := filepath.Dir(outputFile)

	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Generate the code
	code, err := typescript.Generate(schema, *config)
	if err != nil {
		return fmt.Errorf("failed to generate code: %w", err)
	}

	// Write the code to the output file
	if err := os.WriteFile(outputFile, []byte(code), 0644); err != nil {
		return fmt.Errorf("failed to write generated code to file: %w", err)
	}

	return nil
}

func runDart(absConfigDir string, config *dart.Config, schema schema.Schema) error {
	outputDir := filepath.Join(absConfigDir, config.OutputDir)

	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Generate the code
	output, err := dart.Generate(schema, *config)
	if err != nil {
		return fmt.Errorf("failed to generate code: %w", err)
	}

	for _, file := range output.Files {
		outputFile := filepath.Join(outputDir, file.Path)
		outputFileDir := filepath.Dir(outputFile)
		if err := os.MkdirAll(outputFileDir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}

		if err := os.WriteFile(outputFile, []byte(file.Content), 0644); err != nil {
			return fmt.Errorf("failed to write generated code to file %s: %w", outputFile, err)
		}
	}

	return nil
}

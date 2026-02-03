package wasm

import (
	"context"
	"fmt"
	"strings"

	"github.com/varavelio/vdl/toolchain/internal/codegen/config/configtypes"
	"github.com/varavelio/vdl/toolchain/internal/codegen/dart"
	"github.com/varavelio/vdl/toolchain/internal/codegen/golang"
	"github.com/varavelio/vdl/toolchain/internal/codegen/jsonschema"
	"github.com/varavelio/vdl/toolchain/internal/codegen/openapi"
	"github.com/varavelio/vdl/toolchain/internal/codegen/python"
	"github.com/varavelio/vdl/toolchain/internal/codegen/typescript"
	"github.com/varavelio/vdl/toolchain/internal/core/analysis"
	"github.com/varavelio/vdl/toolchain/internal/core/ir"
	"github.com/varavelio/vdl/toolchain/internal/core/parser"
	"github.com/varavelio/vdl/toolchain/internal/wasm/wasmtypes"
)

func runCodegen(input wasmtypes.CodegenInput) (*wasmtypes.CodegenOutput, error) {
	ctx := context.Background()

	// Parse input into AST
	astSchema, err := parser.ParserInstance.ParseString("schema.vdl", input.VdlSchema)
	if err != nil {
		return nil, fmt.Errorf("failed to parse VDL schema: %s", err)
	}

	// Run semantic analysis on the parsed schema
	program, diagnostics := analysis.AnalyzeSchema(astSchema, "/virtual/schema.vdl")
	if len(diagnostics) > 0 {
		var errMsgs strings.Builder
		for i, d := range diagnostics {
			if i > 0 {
				errMsgs.WriteString("\n")
			}
			errMsgs.WriteString(d.String())
		}
		return nil, fmt.Errorf("schema validation failed:\n%s", errMsgs.String())
	}

	// Convert to IR Schema
	schema := ir.FromProgram(program)

	switch input.Target {
	case wasmtypes.CodegenTargetGo:
		cfg := input.GetGoConfigOr(wasmtypes.CodegenInputGoConfig{
			Package:     "vdl",
			GenPatterns: true,
			GenConsts:   true,
			GenClient:   true,
			GenServer:   true,
		})

		gen := golang.New(&configtypes.GoTargetConfig{
			Package:     cfg.Package,
			GenServer:   configtypes.Ptr(cfg.GenServer),
			GenClient:   configtypes.Ptr(cfg.GenClient),
			GenPatterns: configtypes.Ptr(cfg.GenPatterns),
			GenConsts:   configtypes.Ptr(cfg.GenConsts),
		})

		genFiles, err := gen.Generate(ctx, schema)
		if err != nil {
			return nil, fmt.Errorf("failed to generate go code: %w", err)
		}

		outFiles := make([]wasmtypes.CodegenFile, len(genFiles))
		for i, genFile := range genFiles {
			outFiles[i] = wasmtypes.CodegenFile{Path: genFile.RelativePath, Content: string(genFile.Content)}
		}

		return &wasmtypes.CodegenOutput{Files: outFiles}, nil

	case wasmtypes.CodegenTargetTypescript:
		cfg := input.GetTypescriptConfigOr(wasmtypes.CodegenInputTypescriptConfig{
			ImportExtension: "none",
			GenPatterns:     true,
			GenConsts:       true,
			GenClient:       true,
			GenServer:       true,
		})

		importExt := configtypes.TypescriptTargetImportExtension(cfg.ImportExtension.String())
		gen := typescript.New(&configtypes.TypeScriptTargetConfig{
			ImportExtension: &importExt,
			GenServer:       configtypes.Ptr(cfg.GenServer),
			GenClient:       configtypes.Ptr(cfg.GenClient),
			GenPatterns:     configtypes.Ptr(cfg.GenPatterns),
			GenConsts:       configtypes.Ptr(cfg.GenConsts),
		})

		genFiles, err := gen.Generate(ctx, schema)
		if err != nil {
			return nil, fmt.Errorf("failed to generate typescript code: %w", err)
		}

		outFiles := make([]wasmtypes.CodegenFile, len(genFiles))
		for i, genFile := range genFiles {
			outFiles[i] = wasmtypes.CodegenFile{Path: genFile.RelativePath, Content: string(genFile.Content)}
		}

		return &wasmtypes.CodegenOutput{Files: outFiles}, nil

	case wasmtypes.CodegenTargetDart:
		cfg := input.GetDartConfigOr(wasmtypes.CodegenInputDartConfig{
			GenPatterns: true,
			GenConsts:   true,
		})

		gen := dart.New(&configtypes.DartTargetConfig{
			GenPatterns: configtypes.Ptr(cfg.GenPatterns),
			GenConsts:   configtypes.Ptr(cfg.GenConsts),
		})

		genFiles, err := gen.Generate(ctx, schema)
		if err != nil {
			return nil, fmt.Errorf("failed to generate dart code: %w", err)
		}

		outFiles := make([]wasmtypes.CodegenFile, len(genFiles))
		for i, genFile := range genFiles {
			outFiles[i] = wasmtypes.CodegenFile{Path: genFile.RelativePath, Content: string(genFile.Content)}
		}

		return &wasmtypes.CodegenOutput{Files: outFiles}, nil

	case wasmtypes.CodegenTargetPython:
		cfg := input.GetPythonConfigOr(wasmtypes.CodegenInputPythonConfig{
			GenPatterns: true,
			GenConsts:   true,
		})

		gen := python.New(&configtypes.PythonTargetConfig{
			GenPatterns: configtypes.Ptr(cfg.GenPatterns),
			GenConsts:   configtypes.Ptr(cfg.GenConsts),
		})

		genFiles, err := gen.Generate(ctx, schema)
		if err != nil {
			return nil, fmt.Errorf("failed to generate python code: %w", err)
		}

		outFiles := make([]wasmtypes.CodegenFile, len(genFiles))
		for i, genFile := range genFiles {
			outFiles[i] = wasmtypes.CodegenFile{Path: genFile.RelativePath, Content: string(genFile.Content)}
		}

		return &wasmtypes.CodegenOutput{Files: outFiles}, nil

	case wasmtypes.CodegenTargetOpenApi:
		cfg := input.GetOpenApiConfigOr(wasmtypes.CodegenInputOpenApiConfig{
			Title:   "VDL RPC API",
			Version: "1.0.0",
		})

		gen := openapi.New(&configtypes.OpenApiTargetConfig{
			Title:        cfg.Title,
			Version:      cfg.Version,
			Description:  configtypes.Ptr(cfg.GetDescription()),
			BaseUrl:      configtypes.Ptr(cfg.GetBaseUrl()),
			ContactName:  configtypes.Ptr(cfg.GetContactName()),
			ContactEmail: configtypes.Ptr(cfg.GetContactEmail()),
			LicenseName:  configtypes.Ptr(cfg.GetLicenseName()),
			Filename:     configtypes.Ptr("openapi.yaml"),
		})

		genFiles, err := gen.Generate(ctx, schema)
		if err != nil {
			return nil, fmt.Errorf("failed to generate openapi code: %w", err)
		}

		outFiles := make([]wasmtypes.CodegenFile, len(genFiles))
		for i, genFile := range genFiles {
			outFiles[i] = wasmtypes.CodegenFile{Path: genFile.RelativePath, Content: string(genFile.Content)}
		}

		return &wasmtypes.CodegenOutput{Files: outFiles}, nil

	case wasmtypes.CodegenTargetJsonSchema:
		cfg := input.GetJsonSchemaConfigOr(wasmtypes.CodegenInputJsonSchemaConfig{})

		gen := jsonschema.New(&configtypes.JsonSchemaTargetConfig{
			Id:       configtypes.Ptr(cfg.SchemaId),
			Filename: configtypes.Ptr("schema.json"),
		})

		genFiles, err := gen.Generate(ctx, schema)
		if err != nil {
			return nil, fmt.Errorf("failed to generate jsonschema code: %w", err)
		}

		outFiles := make([]wasmtypes.CodegenFile, len(genFiles))
		for i, genFile := range genFiles {
			outFiles[i] = wasmtypes.CodegenFile{Path: genFile.RelativePath, Content: string(genFile.Content)}
		}

		return &wasmtypes.CodegenOutput{Files: outFiles}, nil
	}

	return nil, fmt.Errorf("target %s is not supported in WASM", input.Target)
}

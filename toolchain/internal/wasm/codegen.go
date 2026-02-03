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

	switch {
	case input.Target.Go != nil:
		gen := golang.New(&configtypes.GoTargetConfig{
			Package:     input.Target.Go.Package,
			GenServer:   input.Target.Go.GenServer,
			GenClient:   input.Target.Go.GenClient,
			GenPatterns: input.Target.Go.GenPatterns,
			GenConsts:   input.Target.Go.GenConsts,
		})

		genFiles, err := gen.Generate(ctx, schema)
		if err != nil {
			return nil, fmt.Errorf("failed to generate go code: %w", err)
		}

		outFiles := make([]wasmtypes.CodegenOutputFile, len(genFiles))
		for i, genFile := range genFiles {
			outFiles[i] = wasmtypes.CodegenOutputFile{Path: genFile.RelativePath, Content: string(genFile.Content)}
		}

		return &wasmtypes.CodegenOutput{Files: outFiles}, nil

	case input.Target.Typescript != nil:
		importExt := configtypes.TypescriptTargetImportExtension(input.Target.Typescript.ImportExtension.String())
		gen := typescript.New(&configtypes.TypeScriptTargetConfig{
			ImportExtension: &importExt,
			GenServer:       input.Target.Typescript.GenServer,
			GenClient:       input.Target.Typescript.GenClient,
			GenPatterns:     input.Target.Typescript.GenPatterns,
			GenConsts:       input.Target.Typescript.GenConsts,
		})

		genFiles, err := gen.Generate(ctx, schema)
		if err != nil {
			return nil, fmt.Errorf("failed to generate typescript code: %w", err)
		}

		outFiles := make([]wasmtypes.CodegenOutputFile, len(genFiles))
		for i, genFile := range genFiles {
			outFiles[i] = wasmtypes.CodegenOutputFile{Path: genFile.RelativePath, Content: string(genFile.Content)}
		}

		return &wasmtypes.CodegenOutput{Files: outFiles}, nil

	case input.Target.Dart != nil:
		gen := dart.New(&configtypes.DartTargetConfig{
			GenPatterns: input.Target.Dart.GenPatterns,
			GenConsts:   input.Target.Dart.GenConsts,
		})

		genFiles, err := gen.Generate(ctx, schema)
		if err != nil {
			return nil, fmt.Errorf("failed to generate dart code: %w", err)
		}

		outFiles := make([]wasmtypes.CodegenOutputFile, len(genFiles))
		for i, genFile := range genFiles {
			outFiles[i] = wasmtypes.CodegenOutputFile{Path: genFile.RelativePath, Content: string(genFile.Content)}
		}

		return &wasmtypes.CodegenOutput{Files: outFiles}, nil

	case input.Target.Python != nil:
		gen := python.New(&configtypes.PythonTargetConfig{
			GenPatterns: input.Target.Python.GenPatterns,
			GenConsts:   input.Target.Python.GenConsts,
		})

		genFiles, err := gen.Generate(ctx, schema)
		if err != nil {
			return nil, fmt.Errorf("failed to generate python code: %w", err)
		}

		outFiles := make([]wasmtypes.CodegenOutputFile, len(genFiles))
		for i, genFile := range genFiles {
			outFiles[i] = wasmtypes.CodegenOutputFile{Path: genFile.RelativePath, Content: string(genFile.Content)}
		}

		return &wasmtypes.CodegenOutput{Files: outFiles}, nil

	case input.Target.Openapi != nil:
		gen := openapi.New(&configtypes.OpenApiTargetConfig{
			Title:        input.Target.Openapi.Title,
			Version:      input.Target.Openapi.Version,
			Description:  input.Target.Openapi.Description,
			BaseUrl:      input.Target.Openapi.BaseUrl,
			ContactName:  input.Target.Openapi.ContactName,
			ContactEmail: input.Target.Openapi.ContactEmail,
			LicenseName:  input.Target.Openapi.LicenseName,
			Filename:     input.Target.Openapi.Filename,
		})

		genFiles, err := gen.Generate(ctx, schema)
		if err != nil {
			return nil, fmt.Errorf("failed to generate openapi code: %w", err)
		}

		outFiles := make([]wasmtypes.CodegenOutputFile, len(genFiles))
		for i, genFile := range genFiles {
			outFiles[i] = wasmtypes.CodegenOutputFile{Path: genFile.RelativePath, Content: string(genFile.Content)}
		}

		return &wasmtypes.CodegenOutput{Files: outFiles}, nil

	case input.Target.Jsonschema != nil:
		gen := jsonschema.New(&configtypes.JsonSchemaTargetConfig{
			Filename: input.Target.Jsonschema.Filename,
			Id:       input.Target.Jsonschema.Id,
			Root:     input.Target.Jsonschema.Root,
		})

		genFiles, err := gen.Generate(ctx, schema)
		if err != nil {
			return nil, fmt.Errorf("failed to generate jsonschema code: %w", err)
		}

		outFiles := make([]wasmtypes.CodegenOutputFile, len(genFiles))
		for i, genFile := range genFiles {
			outFiles[i] = wasmtypes.CodegenOutputFile{Path: genFile.RelativePath, Content: string(genFile.Content)}
		}

		return &wasmtypes.CodegenOutput{Files: outFiles}, nil
	}

	return nil, fmt.Errorf("target is not supported in WASM")
}

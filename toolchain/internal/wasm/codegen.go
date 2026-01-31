package wasm

import (
	"context"
	"fmt"
	"strings"

	"github.com/varavelio/vdl/toolchain/internal/codegen/config"
	"github.com/varavelio/vdl/toolchain/internal/codegen/golang"
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
		cfg := input.GoConfig.Or(wasmtypes.CodegenInputGoConfig{
			Package:     "vdl",
			GenPatterns: true,
			GenConsts:   true,
			GenClient:   true,
			GenServer:   true,
		})

		gen := golang.New(&config.GoConfig{
			Package:        cfg.Package,
			ServerConfig:   config.ServerConfig{GenServer: cfg.GenServer},
			ClientConfig:   config.ClientConfig{GenClient: cfg.GenClient},
			PatternsConfig: config.PatternsConfig{GenPatterns: &cfg.GenPatterns},
			ConstsConfig:   config.ConstsConfig{GenConsts: &cfg.GenConsts},
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
	case wasmtypes.CodegenTargetDart:
	case wasmtypes.CodegenTargetPython:
	case wasmtypes.CodegenTargetOpenApi:
	case wasmtypes.CodegenTargetJsonSchema:
	}

	return nil, fmt.Errorf("target %s is not supported in WASM", input.Target)
}

package wasm

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/varavelio/vdl/toolchain/internal/core/analysis"
	"github.com/varavelio/vdl/toolchain/internal/core/ir"
	"github.com/varavelio/vdl/toolchain/internal/core/parser"
	"github.com/varavelio/vdl/toolchain/internal/wasm/wasmtypes"
)

func runIrgen(input wasmtypes.IrgenInput) (*wasmtypes.IrgenOutput, error) {
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

	// Convert irtypes.IrSchema to wasmtypes.IrgenOutput via JSON round-trip
	// Both types share the same JSON structure (IrgenOutput spreads IrSchema)
	jsonBytes, err := json.Marshal(schema)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal IR schema: %w", err)
	}

	var output wasmtypes.IrgenOutput
	if err := json.Unmarshal(jsonBytes, &output); err != nil {
		return nil, fmt.Errorf("failed to unmarshal IR schema to output: %w", err)
	}

	return &output, nil
}

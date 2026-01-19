package typescript

import (
	"strings"

	"github.com/varavelio/gen"
	"github.com/varavelio/vdl/toolchain/internal/core/ir"
)

// generateConstants generates TypeScript constant definitions.
func generateConstants(schema *ir.Schema, _ *flatSchema, _ Config) (string, error) {
	if len(schema.Constants) == 0 {
		return "", nil
	}

	g := gen.New().WithSpaces(2)

	g.Line("// -----------------------------------------------------------------------------")
	g.Line("// Constants")
	g.Line("// -----------------------------------------------------------------------------")
	g.Break()

	for _, constant := range schema.Constants {
		renderConstant(g, constant)
	}

	return g.String(), nil
}

// renderConstant renders a single constant definition.
func renderConstant(g *gen.Generator, constant ir.Constant) {
	// Generate doc comment
	if constant.Doc != "" {
		g.Linef("/**")
		renderPartialMultilineComment(g, strings.TrimSpace(constant.Doc))
		if constant.Deprecated != nil {
			renderDeprecated(g, constant.Deprecated)
		}
		g.Linef(" */")
	} else if constant.Deprecated != nil {
		g.Linef("/**")
		renderDeprecated(g, constant.Deprecated)
		g.Linef(" */")
	}

	// Determine the TypeScript type and value format
	var tsType, tsValue string
	switch constant.ValueType {
	case ir.ConstValueTypeString:
		tsType = "string"
		tsValue = `"` + constant.Value + `"`
	case ir.ConstValueTypeInt:
		tsType = "number"
		tsValue = constant.Value
	case ir.ConstValueTypeFloat:
		tsType = "number"
		tsValue = constant.Value
	case ir.ConstValueTypeBool:
		tsType = "boolean"
		tsValue = constant.Value
	default:
		tsType = "unknown"
		tsValue = constant.Value
	}

	g.Linef("export const %s: %s = %s;", constant.Name, tsType, tsValue)
	g.Break()
}

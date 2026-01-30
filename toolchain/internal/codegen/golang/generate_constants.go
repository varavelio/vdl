package golang

import (
	"github.com/varavelio/gen"
	"github.com/varavelio/vdl/toolchain/internal/codegen/config"
	"github.com/varavelio/vdl/toolchain/internal/core/ir"
)

func generateConstants(schema *ir.Schema, config *config.GoConfig) (string, error) {
	if !config.ShouldGenConsts() {
		return "", nil
	}

	if len(schema.Constants) == 0 {
		return "", nil
	}

	g := gen.New().WithTabs()

	g.Line("// -----------------------------------------------------------------------------")
	g.Line("// Constants")
	g.Line("// -----------------------------------------------------------------------------")
	g.Break()

	for _, constant := range schema.Constants {
		generateConstant(g, constant)
	}

	return g.String(), nil
}

// generateConstant generates Go code for a single constant.
func generateConstant(g *gen.Generator, constant ir.Constant) {
	// Documentation
	if constant.Doc != "" {
		renderMultilineComment(g, constant.Doc)
	} else {
		g.Linef("// %s represents the constant %q.", constant.Name, constant.Value)
	}

	// Deprecation
	renderDeprecated(g, constant.Deprecated)

	// Constant definition
	switch constant.ValueType {
	case ir.ConstValueTypeString:
		g.Linef("const %s = %q", constant.Name, constant.Value)
	case ir.ConstValueTypeInt:
		g.Linef("const %s = %s", constant.Name, constant.Value)
	case ir.ConstValueTypeFloat:
		g.Linef("const %s = %s", constant.Name, constant.Value)
	case ir.ConstValueTypeBool:
		g.Linef("const %s = %s", constant.Name, constant.Value)
	default:
		// Fallback: treat as string
		g.Linef("const %s = %q", constant.Name, constant.Value)
	}
	g.Break()
}

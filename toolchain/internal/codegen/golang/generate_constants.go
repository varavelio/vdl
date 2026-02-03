package golang

import (
	"github.com/varavelio/gen"
	"github.com/varavelio/vdl/toolchain/internal/codegen/config"
	"github.com/varavelio/vdl/toolchain/internal/codegen/config/configtypes"
	"github.com/varavelio/vdl/toolchain/internal/core/ir/irtypes"
)

func generateConstants(schema *irtypes.IrSchema, cfg *configtypes.GoTargetConfig) (string, error) {
	if !config.ShouldGenConsts(cfg.GenConsts) {
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
func generateConstant(g *gen.Generator, constant irtypes.ConstantDef) {
	// Documentation
	if constant.GetDoc() != "" {
		renderMultilineComment(g, constant.GetDoc())
	} else {
		g.Linef("// %s represents the constant %q.", constant.Name, constant.Value)
	}

	// Deprecation
	renderDeprecated(g, constant.Deprecated)

	// Constant definition
	switch constant.ConstType {
	case irtypes.ConstTypeString:
		g.Linef("const %s = %q", constant.Name, constant.Value)
	case irtypes.ConstTypeInt:
		g.Linef("const %s = %s", constant.Name, constant.Value)
	case irtypes.ConstTypeFloat:
		g.Linef("const %s = %s", constant.Name, constant.Value)
	case irtypes.ConstTypeBool:
		g.Linef("const %s = %s", constant.Name, constant.Value)
	default:
		// Fallback: treat as string
		g.Linef("const %s = %q", constant.Name, constant.Value)
	}
	g.Break()
}

package typescript

import (
	"github.com/varavelio/gen"
	"github.com/varavelio/vdl/toolchain/internal/codegen/config"
	"github.com/varavelio/vdl/toolchain/internal/codegen/config/configtypes"
	"github.com/varavelio/vdl/toolchain/internal/core/ir/irtypes"
)

func generateConstants(schema *irtypes.IrSchema, cfg *configtypes.TypeScriptTargetConfig) (string, error) {
	if !config.ShouldGenConsts(cfg.GenConsts) {
		return "", nil
	}

	if len(schema.Constants) == 0 {
		return "", nil
	}

	g := gen.New().WithSpaces(2)

	g.Line("// -----------------------------------------------------------------------------")
	g.Line("// Constants")
	g.Line("// -----------------------------------------------------------------------------")
	g.Break()

	for _, constant := range schema.Constants {
		generateConstant(g, constant)
	}

	return g.String(), nil
}

// generateConstant generates TypeScript code for a single constant.
func generateConstant(g *gen.Generator, constant irtypes.ConstantDef) {
	// Documentation
	if constant.GetDoc() != "" {
		renderMultilineComment(g, constant.GetDoc())
	}

	// Deprecation
	renderDeprecated(g, constant.Deprecated)

	// Constant definition
	switch constant.ConstType {
	case irtypes.ConstTypeString:
		g.Linef("export const %s: string = %q;", constant.Name, constant.Value)
	case irtypes.ConstTypeInt:
		g.Linef("export const %s: number = %s;", constant.Name, constant.Value)
	case irtypes.ConstTypeFloat:
		g.Linef("export const %s: number = %s;", constant.Name, constant.Value)
	case irtypes.ConstTypeBool:
		g.Linef("export const %s: boolean = %s;", constant.Name, constant.Value)
	default:
		// Fallback: treat as string
		g.Linef("export const %s: string = %q;", constant.Name, constant.Value)
	}
	g.Break()
}

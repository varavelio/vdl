package dart

import (
	"fmt"
	"strings"

	"github.com/varavelio/gen"
	"github.com/varavelio/vdl/toolchain/internal/codegen/config"
	"github.com/varavelio/vdl/toolchain/internal/codegen/config/configtypes"
	"github.com/varavelio/vdl/toolchain/internal/core/ir/irtypes"
)

// generateConstants generates Dart constant definitions.
func generateConstants(schema *irtypes.IrSchema, cfg *configtypes.DartConfig) (string, error) {
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
		renderDartConstant(g, constant)
	}

	return g.String(), nil
}

// renderDartConstant renders a single Dart constant definition.
func renderDartConstant(g *gen.Generator, constant irtypes.ConstantDef) {
	// Generate doc comment
	if constant.GetDoc() != "" {
		doc := strings.TrimSpace(constant.GetDoc())
		renderMultilineCommentDart(g, doc)
	}
	if constant.Deprecated != nil {
		renderDeprecatedDart(g, constant.Deprecated)
	}

	// Determine the Dart type and value format
	var dartType, dartValue string
	switch constant.ConstType {
	case irtypes.ConstTypeString:
		dartType = "String"
		dartValue = fmt.Sprintf("'%s'", constant.Value)
	case irtypes.ConstTypeInt:
		dartType = "int"
		dartValue = constant.Value
	case irtypes.ConstTypeFloat:
		dartType = "double"
		dartValue = constant.Value
	case irtypes.ConstTypeBool:
		dartType = "bool"
		dartValue = constant.Value
	default:
		dartType = "dynamic"
		dartValue = constant.Value
	}

	g.Linef("const %s %s = %s;", dartType, constant.Name, dartValue)
	g.Break()
}

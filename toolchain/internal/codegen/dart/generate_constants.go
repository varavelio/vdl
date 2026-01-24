package dart

import (
	"fmt"
	"strings"

	"github.com/varavelio/gen"
	"github.com/varavelio/vdl/toolchain/internal/codegen/config"
	"github.com/varavelio/vdl/toolchain/internal/core/ir"
)

// generateConstants generates Dart constant definitions.
func generateConstants(schema *ir.Schema, config *config.DartConfig) (string, error) {
	if !config.ShouldGenConsts() {
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
func renderDartConstant(g *gen.Generator, constant ir.Constant) {
	// Generate doc comment
	if constant.Doc != "" {
		doc := strings.TrimSpace(constant.Doc)
		renderMultilineCommentDart(g, doc)
	}
	if constant.Deprecated != nil {
		renderDeprecatedDart(g, constant.Deprecated)
	}

	// Determine the Dart type and value format
	var dartType, dartValue string
	switch constant.ValueType {
	case ir.ConstValueTypeString:
		dartType = "String"
		dartValue = fmt.Sprintf("'%s'", constant.Value)
	case ir.ConstValueTypeInt:
		dartType = "int"
		dartValue = constant.Value
	case ir.ConstValueTypeFloat:
		dartType = "double"
		dartValue = constant.Value
	case ir.ConstValueTypeBool:
		dartType = "bool"
		dartValue = constant.Value
	default:
		dartType = "dynamic"
		dartValue = constant.Value
	}

	g.Linef("const %s %s = %s;", dartType, constant.Name, dartValue)
	g.Break()
}

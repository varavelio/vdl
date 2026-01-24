package dart

import (
	"fmt"
	"strings"

	"github.com/varavelio/gen"
	"github.com/varavelio/vdl/toolchain/internal/codegen/config"
	"github.com/varavelio/vdl/toolchain/internal/core/ir"
	"github.com/varavelio/vdl/toolchain/internal/util/strutil"
)

// generatePatterns generates Dart pattern template functions.
func generatePatterns(schema *ir.Schema, config *config.DartConfig) (string, error) {
	if !config.ShouldGenPatterns() {
		return "", nil
	}

	if len(schema.Patterns) == 0 {
		return "", nil
	}

	g := gen.New().WithSpaces(2)

	g.Line("// -----------------------------------------------------------------------------")
	g.Line("// Patterns")
	g.Line("// -----------------------------------------------------------------------------")
	g.Break()

	for _, pattern := range schema.Patterns {
		renderDartPattern(g, pattern)
	}

	return g.String(), nil
}

// renderDartPattern renders a single Dart pattern template function.
func renderDartPattern(g *gen.Generator, pattern ir.Pattern) {
	// Generate doc comment
	if pattern.Doc != "" {
		doc := strings.TrimSpace(strutil.NormalizeIndent(pattern.Doc))
		renderMultilineCommentDart(g, doc)
	}
	if pattern.Deprecated != nil {
		renderDeprecatedDart(g, pattern.Deprecated)
	}

	// Generate function signature with parameters
	params := make([]string, len(pattern.Placeholders))
	for i, placeholder := range pattern.Placeholders {
		params[i] = fmt.Sprintf("String %s", placeholder)
	}

	g.Linef("String %s(%s) {", pattern.Name, strings.Join(params, ", "))
	g.Block(func() {
		// Convert template to Dart string interpolation
		templateLiteral := convertPatternToDartInterpolation(pattern.Template, pattern.Placeholders)
		g.Linef("return %s;", templateLiteral)
	})
	g.Linef("}")
	g.Break()
}

// convertPatternToDartInterpolation converts a VDL pattern template to a Dart string interpolation.
// Pattern format: "Hello, {name}!" -> 'Hello, $name!'
func convertPatternToDartInterpolation(template string, placeholders []string) string {
	result := template

	// Replace each {placeholder} with $placeholder
	for _, placeholder := range placeholders {
		result = strings.ReplaceAll(result, "{"+placeholder+"}", "$"+placeholder)
	}

	// Wrap in single quotes for Dart string
	return "'" + result + "'"
}

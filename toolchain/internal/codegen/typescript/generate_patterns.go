package typescript

import (
	"fmt"
	"strings"

	"github.com/varavelio/gen"
	"github.com/varavelio/vdl/toolchain/internal/codegen/config"
	"github.com/varavelio/vdl/toolchain/internal/core/ir"
	"github.com/varavelio/vdl/toolchain/internal/util/strutil"
)

// generatePatterns generates TypeScript pattern template functions.
func generatePatterns(schema *ir.Schema, _ *flatSchema, config *config.TypeScriptConfig) (string, error) {
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
		renderPattern(g, pattern)
	}

	return g.String(), nil
}

// renderPattern renders a single pattern template function.
func renderPattern(g *gen.Generator, pattern ir.Pattern) {
	// Generate doc comment
	if pattern.Doc != "" {
		doc := strings.TrimSpace(strutil.NormalizeIndent(pattern.Doc))
		renderMultilineComment(g, doc)
	} else if pattern.Deprecated != nil {
		g.Linef("/**")
		renderDeprecated(g, pattern.Deprecated)
		g.Linef(" */")
	}

	// Generate function signature with parameters
	params := make([]string, len(pattern.Placeholders))
	for i, placeholder := range pattern.Placeholders {
		params[i] = fmt.Sprintf("%s: string", placeholder)
	}

	g.Linef("export function %s(%s): string {", pattern.Name, strings.Join(params, ", "))
	g.Block(func() {
		// Convert template to TypeScript template literal
		templateLiteral := convertPatternToTemplateLiteral(pattern.Template, pattern.Placeholders)
		g.Linef("return %s;", templateLiteral)
	})
	g.Line("}")
	g.Break()
}

// convertPatternToTemplateLiteral converts a VDL pattern template to a TypeScript template literal.
// Pattern format: "Hello, {name}!" -> `Hello, ${name}!`
func convertPatternToTemplateLiteral(template string, placeholders []string) string {
	result := template

	// Replace each {placeholder} with ${placeholder}
	for _, placeholder := range placeholders {
		// Use strings.ReplaceAll instead of regexp to avoid $ interpretation issues
		result = strings.ReplaceAll(result, "{"+placeholder+"}", "${"+placeholder+"}")
	}

	// Wrap in backticks for template literal
	return "`" + result + "`"
}

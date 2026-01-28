package typescript

import (
	"fmt"
	"strings"

	"github.com/varavelio/gen"
	"github.com/varavelio/vdl/toolchain/internal/codegen/config"
	"github.com/varavelio/vdl/toolchain/internal/core/ir"
)

// generatePatterns generates TypeScript pattern template functions.
func generatePatterns(schema *ir.Schema, config *config.TypeScriptConfig) (string, error) {
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
		renderMultilineComment(g, pattern.Doc)
	} else if pattern.Deprecated != nil {
		g.Linef("/**")
		renderDeprecated(g, pattern.Deprecated)
		g.Linef(" */")
	}

	// Generate function signature with parameters
	// Deduplicate placeholders while preserving order
	seen := make(map[string]bool)
	var params []string
	var uniquePlaceholders []string

	for _, placeholder := range pattern.Placeholders {
		if !seen[placeholder] {
			seen[placeholder] = true
			uniquePlaceholders = append(uniquePlaceholders, placeholder)
			params = append(params, fmt.Sprintf("%s: string", placeholder))
		}
	}

	g.Linef("export function %s(%s): string {", pattern.Name, strings.Join(params, ", "))
	g.Block(func() {
		// Convert template to TypeScript template literal
		// We pass uniquePlaceholders because strings.ReplaceAll replaces all occurrences anyway
		templateLiteral := convertPatternToTemplateLiteral(pattern.Template, uniquePlaceholders)
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

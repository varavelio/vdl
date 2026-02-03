package golang

import (
	"strings"

	"github.com/varavelio/gen"
	"github.com/varavelio/vdl/toolchain/internal/codegen/config"
	"github.com/varavelio/vdl/toolchain/internal/codegen/config/configtypes"
	"github.com/varavelio/vdl/toolchain/internal/core/ir/irtypes"
)

func generatePatterns(schema *irtypes.IrSchema, cfg *configtypes.GoConfig) (string, error) {
	if !config.ShouldGenPatterns(cfg.GenPatterns) {
		return "", nil
	}

	if len(schema.Patterns) == 0 {
		return "", nil
	}

	g := gen.New().WithTabs()

	g.Line("// -----------------------------------------------------------------------------")
	g.Line("// Patterns")
	g.Line("// -----------------------------------------------------------------------------")
	g.Break()

	for _, pattern := range schema.Patterns {
		generatePattern(g, pattern)
	}

	return g.String(), nil
}

// generatePattern generates a Go function for a pattern template.
func generatePattern(g *gen.Generator, pattern irtypes.PatternDef) {
	// Documentation
	if pattern.GetDoc() != "" {
		doc := pattern.GetDoc()
		if pattern.Template != "" {
			doc += "\n\nTemplate: " + pattern.Template
		}
		renderMultilineComment(g, doc)
	} else {
		g.Linef("// %s generates a string from the pattern template.", pattern.Name)
		if pattern.Template != "" {
			g.Linef("//")
			g.Linef("// Template: %s", pattern.Template)
		}
	}

	// Deprecation
	renderDeprecated(g, pattern.Deprecated)

	// Generate function signature
	var params []string
	seen := make(map[string]bool)
	for _, placeholder := range pattern.Placeholders {
		if !seen[placeholder] {
			params = append(params, placeholder+" string")
			seen[placeholder] = true
		}
	}

	g.Linef("func %s(%s) string {", pattern.Name, strings.Join(params, ", "))
	g.Block(func() {
		// Build the return expression using string concatenation
		// Template: "events.users.{userId}.{eventType}"
		// Result: "events.users." + userId + "." + eventType

		parts := parsePatternTemplate(pattern.Template, pattern.Placeholders)
		if len(parts) == 0 {
			g.Linef("return %q", pattern.Template)
		} else {
			g.Linef("return %s", strings.Join(parts, " + "))
		}
	})
	g.Line("}")
	g.Break()
}

// parsePatternTemplate parses a pattern template and returns Go code parts.
// Each part is either a quoted string literal or a variable name.
func parsePatternTemplate(template string, placeholders []string) []string {
	if len(placeholders) == 0 {
		return []string{`"` + template + `"`}
	}

	var parts []string
	remaining := template

	for remaining != "" {
		// Find the next placeholder
		nextPlaceholderIdx := -1
		nextPlaceholder := ""

		for _, ph := range placeholders {
			pattern := "{" + ph + "}"
			idx := strings.Index(remaining, pattern)
			if idx != -1 && (nextPlaceholderIdx == -1 || idx < nextPlaceholderIdx) {
				nextPlaceholderIdx = idx
				nextPlaceholder = ph
			}
		}

		if nextPlaceholderIdx == -1 {
			// No more placeholders, add the rest as a string literal
			if remaining != "" {
				parts = append(parts, `"`+remaining+`"`)
			}
			break
		}

		// Add the part before the placeholder as a string literal
		if nextPlaceholderIdx > 0 {
			parts = append(parts, `"`+remaining[:nextPlaceholderIdx]+`"`)
		}

		// Add the placeholder as a variable
		parts = append(parts, nextPlaceholder)

		// Move past the placeholder
		pattern := "{" + nextPlaceholder + "}"
		remaining = remaining[nextPlaceholderIdx+len(pattern):]
	}

	return parts
}

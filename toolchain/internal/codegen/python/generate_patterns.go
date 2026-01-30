package python

import (
	"fmt"
	"strings"

	"github.com/varavelio/gen"
	"github.com/varavelio/vdl/toolchain/internal/codegen/config"
	"github.com/varavelio/vdl/toolchain/internal/core/ir"
	"github.com/varavelio/vdl/toolchain/internal/util/strutil"
)

func generatePatterns(schema *ir.Schema, cfg *config.PythonConfig) (string, error) {
	g := gen.New()
	g.Break()

	for _, p := range schema.Patterns {
		funcName := strutil.ToSnakeCase(p.Name)

		// Generate arguments
		seen := make(map[string]bool)
		var args []string
		for _, ph := range p.Placeholders {
			argName := strutil.ToSnakeCase(ph)
			argName = sanitizeIdentifier(argName)
			if seen[argName] {
				continue
			}
			seen[argName] = true
			args = append(args, fmt.Sprintf("%s: str", argName))
		}

		g.Linef("def %s(%s) -> str:", funcName, strings.Join(args, ", "))

		// Generate docstring
		doc := p.Doc
		if doc == "" {
			doc = fmt.Sprintf("%s generates a string from the pattern template.", p.Name)
		}
		if p.Template != "" {
			doc = fmt.Sprintf("%s\n\nTemplate: `%s`", doc, p.Template)
		}
		renderDocstringPython(g, doc)
		renderDeprecatedPython(g, p.Deprecated)

		// Convert template to python f-string
		// VDL template uses {placeholder}. Python f-string also uses {placeholder}.
		// But we renamed placeholders to snake_case.
		template := p.Template
		for _, ph := range p.Placeholders {
			argName := strutil.ToSnakeCase(ph)
			argName = sanitizeIdentifier(argName)
			template = strings.ReplaceAll(template, "{"+ph+"}", "{"+argName+"}")
		}

		g.Linef("    return f%q", template)
		g.Break()
	}

	return g.String(), nil
}

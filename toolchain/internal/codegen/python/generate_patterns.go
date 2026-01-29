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
	g.Line("from typing import Any")
	g.Break()

	for _, p := range schema.Patterns {
		if p.Doc != "" {
			g.Linef("# %s", strings.ReplaceAll(p.Doc, "\n", "\n# "))
		}
		funcName := strutil.ToSnakeCase(p.Name)
		var args []string
		for _, ph := range p.Placeholders {
			argName := strutil.ToSnakeCase(ph)
			argName = sanitizeIdentifier(argName)
			args = append(args, fmt.Sprintf("%s: Any", argName))
		}

		g.Linef("def %s(%s) -> str:", funcName, strings.Join(args, ", "))

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

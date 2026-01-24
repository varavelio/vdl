package formatter

import (
	"strings"

	"github.com/varavelio/gen"
	"github.com/varavelio/vdl/toolchain/internal/core/ast"
	"github.com/varavelio/vdl/toolchain/internal/util/strutil"
)

// FormatComment formats a comment.
func FormatComment(g *gen.Generator, comment *ast.Comment) {
	if comment.Simple != nil {
		g.Line(*comment.Simple)
	}
	if comment.Block != nil {
		g.Line(*comment.Block)
	}
}

// FormatInlineComment formats an inline comment.
func FormatInlineComment(g *gen.Generator, comment *ast.Comment) {
	if comment.Simple != nil {
		g.Inlinef(" %s", *comment.Simple)
	}
	if comment.Block != nil {
		g.Inlinef(" %s", *comment.Block)
	}
}

func normalizeDocstring(s string) string {
	if !strings.Contains(s, "\n") {
		return s
	}
	return strutil.NormalizeIndent(s)
}

// FormatDocstring formats a docstring with proper indentation.
// If the docstring is multi-line, it writes it to the generator and returns true.
// If it is single-line, it returns the normalized string and false, allowing the caller
// to handle printing (e.g. for inline comments).
func FormatDocstring(g *gen.Generator, raw string) (string, bool) {
	normalized := normalizeDocstring(raw)

	if strings.Contains(normalized, "\n") {
		lines := strings.Split(normalized, "\n")

		g.Inline(`"""`)
		if lines[0] != "" {
			g.Raw(lines[0])
		}
		g.Break()

		for i := 1; i < len(lines)-1; i++ {
			g.Line(lines[i])
		}

		last := lines[len(lines)-1]
		if last != "" {
			g.Line(last)
		}
		g.Line(`"""`)
		return "", true
	}

	trimmed := strings.TrimSpace(normalized)
	if trimmed == "" {
		return "", false
	}

	return " " + trimmed + " ", false
}

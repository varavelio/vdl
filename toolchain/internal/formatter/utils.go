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
//
// Formatting rules:
//   - Single line: """ content """
//   - Multi-line:
//     """
//     line 1
//     line 2
//     """
func FormatDocstring(g *gen.Generator, raw string) (string, bool) {
	normalized := normalizeDocstring(raw)

	if strings.Contains(normalized, "\n") {
		lines := strings.Split(normalized, "\n")

		// Always start multiline docstrings with """ on its own line
		g.Line(`"""`)

		// Skip leading empty line if present (content started with \n)
		startIdx := 0
		if lines[0] == "" {
			startIdx = 1
		}

		// Skip trailing empty line if present (content ended with \n)
		endIdx := len(lines)
		if endIdx > 0 && lines[endIdx-1] == "" {
			endIdx--
		}

		for i := startIdx; i < endIdx; i++ {
			g.Line(lines[i])
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

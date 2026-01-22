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

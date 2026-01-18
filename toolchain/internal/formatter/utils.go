package formatter

import (
	"regexp"
	"strings"

	"github.com/varavelio/gen"
	"github.com/varavelio/vdl/toolchain/internal/core/ast"
	"github.com/varavelio/vdl/toolchain/internal/util/strutil"
)

var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

func toSnakeCase(str string) string {
	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}

func toUpperSnakeCase(str string) string {
	return strings.ToUpper(toSnakeCase(str))
}

// limitConsecutiveNewlines limits the number of consecutive newlines in a string.
// This is a wrapper around strutil.LimitConsecutiveNewlines if it exists, or local impl.
// The original formatter used strutil.LimitConsecutiveNewlines.
// I'll assume it exists as per previous code reading.

func escapeQuotes(s string) string {
	return strutil.EscapeQuotes(s)
}

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

package formatter

import (
	"github.com/varavelio/gen"
	"github.com/varavelio/vdl/toolchain/internal/core/ast"
	"github.com/varavelio/vdl/toolchain/internal/util/strutil"
)

type includeFormatter struct {
	g           *gen.Generator
	includeDecl *ast.Include
}

func newIncludeFormatter(g *gen.Generator, includeDecl *ast.Include) *includeFormatter {
	if includeDecl == nil {
		includeDecl = &ast.Include{}
	}

	return &includeFormatter{
		g:           g,
		includeDecl: includeDecl,
	}
}

func (f *includeFormatter) format() *gen.Generator {
	f.g.Inlinef("include \"%s\"", strutil.EscapeQuotes(string(f.includeDecl.Path)))
	return f.g
}

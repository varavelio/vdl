package formatter

import (
	"github.com/uforg/ufogenkit"
	"github.com/varavelio/vdl/toolchain/internal/urpc/ast"
	"github.com/varavelio/vdl/toolchain/internal/util/strutil"
)

type includeFormatter struct {
	g           *ufogenkit.GenKit
	includeDecl *ast.Include
}

func newIncludeFormatter(g *ufogenkit.GenKit, includeDecl *ast.Include) *includeFormatter {
	if includeDecl == nil {
		includeDecl = &ast.Include{}
	}

	return &includeFormatter{
		g:           g,
		includeDecl: includeDecl,
	}
}

func (f *includeFormatter) format() *ufogenkit.GenKit {
	f.g.Inlinef("include \"%s\"", strutil.EscapeQuotes(string(f.includeDecl.Path)))
	return f.g
}

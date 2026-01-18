package formatter

import (
	"github.com/varavelio/gen"
	"github.com/varavelio/vdl/toolchain/internal/core/ast"
	"github.com/varavelio/vdl/toolchain/internal/util/strutil"
)

type patternFormatter struct {
	g           *gen.Generator
	patternDecl *ast.PatternDecl
}

func newPatternFormatter(g *gen.Generator, patternDecl *ast.PatternDecl) *patternFormatter {
	if patternDecl == nil {
		patternDecl = &ast.PatternDecl{}
	}

	return &patternFormatter{
		g:           g,
		patternDecl: patternDecl,
	}
}

func (f *patternFormatter) format() *gen.Generator {
	if f.patternDecl.Docstring != nil {
		f.g.Linef(`"""%s"""`, normalizeDocstring(string(f.patternDecl.Docstring.Value)))
	}

	if f.patternDecl.Deprecated != nil {
		if f.patternDecl.Deprecated.Message == nil {
			f.g.Inline("deprecated ")
		}
		if f.patternDecl.Deprecated.Message != nil {
			f.g.Linef("deprecated(\"%s\")", strutil.EscapeQuotes(string(*f.patternDecl.Deprecated.Message)))
		}
	}

	// Force strict PascalCase
	f.g.Inlinef("pattern %s = ", strutil.ToPascalCase(f.patternDecl.Name))
	f.g.Inlinef(`"%s"`, strutil.EscapeQuotes(string(f.patternDecl.Value)))

	return f.g
}

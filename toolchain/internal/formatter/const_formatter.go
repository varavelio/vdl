package formatter

import (
	"github.com/varavelio/gen"
	"github.com/varavelio/vdl/toolchain/internal/core/ast"
	"github.com/varavelio/vdl/toolchain/internal/util/strutil"
)

type constFormatter struct {
	g         *gen.Generator
	constDecl *ast.ConstDecl
}

func newConstFormatter(g *gen.Generator, constDecl *ast.ConstDecl) *constFormatter {
	if constDecl == nil {
		constDecl = &ast.ConstDecl{}
	}

	return &constFormatter{
		g:         g,
		constDecl: constDecl,
	}
}

func (f *constFormatter) format() *gen.Generator {
	if f.constDecl.Docstring != nil {
		normalized, printed := FormatDocstring(f.g, string(f.constDecl.Docstring.Value))
		if !printed {
			f.g.Linef(`"""%s"""`, normalized)
		}
	}

	if f.constDecl.Deprecated != nil {
		if f.constDecl.Deprecated.Message == nil {
			f.g.Inline("deprecated ")
		}
		if f.constDecl.Deprecated.Message != nil {
			f.g.Linef("deprecated(\"%s\")", strutil.EscapeQuotes(string(*f.constDecl.Deprecated.Message)))
		}
	}

	// Force strict UPPER_SNAKE_CASE
	f.g.Inlinef("const %s = ", strutil.ToUpperSnakeCase(f.constDecl.Name))

	if f.constDecl.Value != nil {
		f.g.Inline(f.constDecl.Value.String())
	}

	return f.g
}

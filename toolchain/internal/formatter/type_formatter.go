package formatter

import (
	"github.com/varavelio/gen"
	"github.com/varavelio/vdl/toolchain/internal/core/ast"
	"github.com/varavelio/vdl/toolchain/internal/util/strutil"
)

type typeFormatter struct {
	g        *gen.Generator
	typeDecl *ast.TypeDecl
}

func newTypeFormatter(g *gen.Generator, typeDecl *ast.TypeDecl) *typeFormatter {
	if typeDecl == nil {
		typeDecl = &ast.TypeDecl{}
	}

	return &typeFormatter{
		g:        g,
		typeDecl: typeDecl,
	}
}

// format formats the entire typeDecl, handling spacing and EOL comments.
//
// Returns the formatted gen.Generator.
func (f *typeFormatter) format() *gen.Generator {
	if f.typeDecl.Docstring != nil {
		normalized, printed := FormatDocstring(f.g, string(f.typeDecl.Docstring.Value))
		if !printed {
			f.g.Linef(`"""%s"""`, normalized)
		}
	}

	if f.typeDecl.Deprecated != nil {
		if f.typeDecl.Deprecated.Message == nil {
			f.g.Inline("deprecated ")
		}
		if f.typeDecl.Deprecated.Message != nil {
			f.g.Linef("deprecated(\"%s\")", strutil.EscapeQuotes(string(*f.typeDecl.Deprecated.Message)))
		}
	}

	// Force strict PascalCase
	f.g.Inlinef(`type %s `, strutil.ToPascalCase(f.typeDecl.Name))

	// Use typeBodyFormatter
	bodyFormatter := newTypeBodyFormatter(f.g, f.typeDecl, f.typeDecl.Children)
	bodyFormatter.format()

	return f.g
}

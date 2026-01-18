package formatter

import (
	"github.com/uforg/ufogenkit"
	"github.com/varavelio/vdl/toolchain/internal/core/ast"
	"github.com/varavelio/vdl/toolchain/internal/util/strutil"
)

type typeFormatter struct {
	g        *ufogenkit.GenKit
	typeDecl *ast.TypeDecl
}

func newTypeFormatter(g *ufogenkit.GenKit, typeDecl *ast.TypeDecl) *typeFormatter {
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
// Returns the formatted genkit.GenKit.
func (f *typeFormatter) format() *ufogenkit.GenKit {
	if f.typeDecl.Docstring != nil {
		f.g.Linef(`"""%s"""`, normalizeDocstring(string(f.typeDecl.Docstring.Value)))
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

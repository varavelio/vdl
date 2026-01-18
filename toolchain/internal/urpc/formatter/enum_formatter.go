package formatter

import (
	"github.com/uforg/ufogenkit"
	"github.com/varavelio/vdl/urpc/internal/urpc/ast"
	"github.com/varavelio/vdl/urpc/internal/util/strutil"
)

type enumFormatter struct {
	g        *ufogenkit.GenKit
	enumDecl *ast.EnumDecl
}

func newEnumFormatter(g *ufogenkit.GenKit, enumDecl *ast.EnumDecl) *enumFormatter {
	if enumDecl == nil {
		enumDecl = &ast.EnumDecl{}
	}

	return &enumFormatter{
		g:        g,
		enumDecl: enumDecl,
	}
}

func (f *enumFormatter) format() *ufogenkit.GenKit {
	if f.enumDecl.Docstring != nil {
		f.g.Linef(`"""%s"""`, normalizeDocstring(string(f.enumDecl.Docstring.Value)))
	}

	if f.enumDecl.Deprecated != nil {
		if f.enumDecl.Deprecated.Message == nil {
			f.g.Inline("deprecated ")
		}
		if f.enumDecl.Deprecated.Message != nil {
			f.g.Linef("deprecated(\"%s\")", strutil.EscapeQuotes(string(*f.enumDecl.Deprecated.Message)))
		}
	}

	// Force strict PascalCase
	enumName := strutil.ToPascalCase(f.enumDecl.Name)
	if len(f.enumDecl.Members) == 0 {
		f.g.Linef("enum %s {}", enumName)
		return f.g
	}

	f.g.Linef("enum %s {", enumName)

	f.g.Block(func() {
		for i, member := range f.enumDecl.Members {
			// Handle comments for members if they exist
			// Note: The AST for EnumMember has a Comment field.
			if member.Comment != nil {
				if member.Comment.Simple != nil {
					f.g.Line(*member.Comment.Simple)
				}
				if member.Comment.Block != nil {
					f.g.Line(*member.Comment.Block)
				}
			}

			f.g.Inline(strutil.ToPascalCase(member.Name))

			if member.Value != nil {
				f.g.Inline(" = ")
				if member.Value.Str != nil {
					f.g.Inlinef(`"%s"`, strutil.EscapeQuotes(string(*member.Value.Str)))
				} else if member.Value.Int != nil {
					f.g.Inline(*member.Value.Int)
				}
			}
			f.g.Line("")

			_ = i
		}
	})

	f.g.Inline("}")

	return f.g
}

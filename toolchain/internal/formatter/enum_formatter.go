package formatter

import (
	"github.com/varavelio/gen"
	"github.com/varavelio/vdl/toolchain/internal/core/ast"
	"github.com/varavelio/vdl/toolchain/internal/util/strutil"
)

type enumFormatter struct {
	g                 *gen.Generator
	enumDecl          *ast.EnumDecl
	members           []*ast.EnumMember
	maxIndex          int
	currentIndex      int
	currentIndexEOF   bool
	currentIndexChild *ast.EnumMember
}

func newEnumFormatter(g *gen.Generator, enumDecl *ast.EnumDecl) *enumFormatter {
	if enumDecl == nil {
		enumDecl = &ast.EnumDecl{}
	}

	if enumDecl.Members == nil {
		enumDecl.Members = []*ast.EnumMember{}
	}

	maxIndex := max(len(enumDecl.Members)-1, 0)
	currentIndex := 0
	currentIndexEOF := len(enumDecl.Members) < 1
	var currentIndexChild *ast.EnumMember

	if !currentIndexEOF {
		currentIndexChild = enumDecl.Members[0]
	}

	return &enumFormatter{
		g:                 g,
		enumDecl:          enumDecl,
		members:           enumDecl.Members,
		maxIndex:          maxIndex,
		currentIndex:      currentIndex,
		currentIndexEOF:   currentIndexEOF,
		currentIndexChild: currentIndexChild,
	}
}

func (f *enumFormatter) loadNextChild() {
	currentIndex := f.currentIndex + 1
	currentIndexEOF := currentIndex > f.maxIndex
	var currentIndexChild *ast.EnumMember

	if !currentIndexEOF {
		currentIndexChild = f.members[currentIndex]
	}

	f.currentIndex = currentIndex
	f.currentIndexEOF = currentIndexEOF
	f.currentIndexChild = currentIndexChild
}

func (f *enumFormatter) peekChild(offset int) (*ast.EnumMember, ast.LineDiff, bool) {
	peekIndex := f.currentIndex + offset
	peekIndexEOF := peekIndex < 0 || peekIndex > f.maxIndex
	var peekIndexChild *ast.EnumMember
	lineDiff := ast.LineDiff{}

	if !peekIndexEOF {
		peekIndexChild = f.members[peekIndex]
		lineDiff = ast.GetLineDiff(peekIndexChild, f.currentIndexChild)
	}

	return peekIndexChild, lineDiff, peekIndexEOF
}

// lineAndComment writes a line of content, checking if the next member is an EOL comment.
func (f *enumFormatter) lineAndComment(content string) {
	next, nextLineDiff, nextEOF := f.peekChild(1)

	// Check if next member is an inline comment (on the same line)
	if !nextEOF && next.Comment != nil && nextLineDiff.StartToEnd == 0 {
		f.g.Inline(content)

		if next.Comment.Simple != nil {
			f.g.Inlinef(" %s", *next.Comment.Simple)
		}
		if next.Comment.Block != nil {
			f.g.Inlinef(" %s", *next.Comment.Block)
		}
		f.g.Break()

		// Skip the inline comment because it's already written
		f.loadNextChild()
		return
	}

	f.g.Line(content)
}

func (f *enumFormatter) format() *gen.Generator {
	if f.enumDecl.Docstring != nil {
		normalized, printed := FormatDocstring(f.g, string(f.enumDecl.Docstring.Value))
		if !printed {
			f.g.Linef(`"""%s"""`, normalized)
		}
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
	if f.currentIndexEOF {
		f.g.Linef("enum %s {}", enumName)
		return f.g
	}

	f.g.Linef("enum %s {", enumName)

	f.g.Block(func() {
		for !f.currentIndexEOF {
			f.formatMember()
			f.loadNextChild()
		}
	})

	f.g.Inline("}")

	return f.g
}

func (f *enumFormatter) formatMember() {
	member := f.currentIndexChild

	// Handle standalone comments (not EOL comments)
	if member.Comment != nil {
		f.formatComment()
		return
	}

	// Format the member name and value
	memberLine := strutil.ToPascalCase(member.Name)

	if member.Value != nil {
		if member.Value.Str != nil {
			memberLine += ` = "` + strutil.EscapeQuotes(string(*member.Value.Str)) + `"`
		} else if member.Value.Int != nil {
			memberLine += " = " + *member.Value.Int
		}
	}

	f.lineAndComment(memberLine)
}

func (f *enumFormatter) formatComment() {
	_, prevLineDiff, prevEOF := f.peekChild(-1)

	shouldBreakBefore := !prevEOF && prevLineDiff.EndToStart > 1

	if shouldBreakBefore {
		f.g.Break()
	}

	if f.currentIndexChild.Comment.Simple != nil {
		f.g.Line(*f.currentIndexChild.Comment.Simple)
	}

	if f.currentIndexChild.Comment.Block != nil {
		f.g.Line(*f.currentIndexChild.Comment.Block)
	}
}

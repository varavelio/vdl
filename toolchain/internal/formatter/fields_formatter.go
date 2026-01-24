package formatter

import (
	"github.com/varavelio/gen"
	"github.com/varavelio/vdl/toolchain/internal/core/ast"
	"github.com/varavelio/vdl/toolchain/internal/util/strutil"
)

// Common helpers

func formatComment(g *gen.Generator, comment *ast.Comment, breakBefore bool) {
	if breakBefore {
		g.Break()
	}
	if comment.Simple != nil {
		g.Line(*comment.Simple)
	}
	if comment.Block != nil {
		g.Line(*comment.Block)
	}
}

func formatSpread(g *gen.Generator, spread *ast.Spread, breakBefore bool) {
	if breakBefore {
		g.Break()
	}
	// Force strict PascalCase
	g.Inlinef("...%s", strutil.ToPascalCase(spread.TypeName))
}

func formatField(g *gen.Generator, field *ast.Field, breakBefore bool, _ any) {
	if breakBefore {
		g.Line("")
	}

	if field.Docstring != nil {
		normalized, printed := FormatDocstring(g, string(field.Docstring.Value))
		if !printed {
			g.Linef(`"""%s"""`, normalized)
		}
	}

	// Force strict camelCase
	name := strutil.ToCamelCase(field.Name)
	if field.Optional {
		g.Inlinef("%s?: ", name)
	} else {
		g.Inlinef("%s: ", name)
	}

	formatFieldType(g, field.Type)
}

func formatFieldType(g *gen.Generator, ft ast.FieldType) {
	if ft.Base.Named != nil {
		typeLiteral := *ft.Base.Named
		// Force strict pascal case for non primitive types
		if !ast.IsPrimitiveType(typeLiteral) {
			typeLiteral = strutil.ToPascalCase(typeLiteral)
		}
		g.Inline(typeLiteral)
	} else if ft.Base.Map != nil {
		g.Inline("map<")
		formatFieldType(g, *ft.Base.Map.ValueType)
		g.Inline(">")
	} else if ft.Base.Object != nil {
		formatter := newTypeBodyFormatter(g, ft.Base.Object, ft.Base.Object.Children)
		formatter.format()
	}

	// Array dimensions
	for i := 0; i < int(ft.Dimensions); i++ {
		g.Inline("[]")
	}
}

// ----------------------------------------------------------------------------
// TypeBodyFormatter (for TypeDecl and inline objects)

type typeBodyFormatter struct {
	g                 *gen.Generator
	parent            ast.WithPositions
	children          []*ast.TypeDeclChild
	maxIndex          int
	currentIndex      int
	currentIndexEOF   bool
	currentIndexChild ast.TypeDeclChild
}

func newTypeBodyFormatter(g *gen.Generator, parent ast.WithPositions, children []*ast.TypeDeclChild) *typeBodyFormatter {
	if children == nil {
		children = []*ast.TypeDeclChild{}
	}

	maxIndex := max(len(children)-1, 0)
	currentIndex := 0
	currentIndexEOF := len(children) < 1
	currentIndexChild := ast.TypeDeclChild{}

	if !currentIndexEOF {
		currentIndexChild = *children[0]
	}

	return &typeBodyFormatter{
		g:                 g,
		parent:            parent,
		children:          children,
		maxIndex:          maxIndex,
		currentIndex:      currentIndex,
		currentIndexEOF:   currentIndexEOF,
		currentIndexChild: currentIndexChild,
	}
}

func (f *typeBodyFormatter) loadNextChild() {
	currentIndex := f.currentIndex + 1
	currentIndexEOF := currentIndex > f.maxIndex
	currentIndexChild := ast.TypeDeclChild{}

	if !currentIndexEOF {
		currentIndexChild = *f.children[currentIndex]
	}

	f.currentIndex = currentIndex
	f.currentIndexEOF = currentIndexEOF
	f.currentIndexChild = currentIndexChild
}

func (f *typeBodyFormatter) peekChild(offset int) (ast.TypeDeclChild, ast.LineDiff, bool) {
	peekIndex := f.currentIndex + offset
	peekIndexEOF := peekIndex < 0 || peekIndex > f.maxIndex
	peekIndexChild := ast.TypeDeclChild{}
	lineDiff := ast.LineDiff{}

	if !peekIndexEOF {
		peekIndexChild = *f.children[peekIndex]
		lineDiff = ast.GetLineDiff(peekIndexChild, f.currentIndexChild)
	}

	return peekIndexChild, lineDiff, peekIndexEOF
}

func (f *typeBodyFormatter) LineAndComment(content string) {
	// Check for inline comment on the next child
	next, nextLineDiff, nextEOF := f.peekChild(1)

	if !nextEOF && next.Comment != nil && nextLineDiff.StartToEnd == 0 {
		f.g.Inline(content)

		if next.Comment.Simple != nil {
			f.g.Inlinef(" %s", *next.Comment.Simple)
		}
		if next.Comment.Block != nil {
			f.g.Inlinef(" %s", *next.Comment.Block)
		}

		f.loadNextChild()
		return
	}

	f.g.Line(content)
}

func (f *typeBodyFormatter) format() *gen.Generator {
	if f.currentIndexEOF {
		f.g.Inline("{}")
		return f.g
	}

	hasInlineComment := false
	if f.currentIndexChild.Comment != nil && f.parent != nil {
		lineDiff := ast.GetLineDiff(f.currentIndexChild, f.parent)
		if lineDiff.StartToStart == 0 {
			hasInlineComment = true
		}
	}

	if hasInlineComment {
		// If the first child is an inline comment, print it on the same line as "{"
		f.g.Inline("{")
		FormatInlineComment(f.g, f.currentIndexChild.Comment)
		f.g.Break()
		f.loadNextChild() // Skip it
	} else {
		f.g.Line("{")
	}

	f.g.Block(func() {
		for !f.currentIndexEOF {
			f.formatChild()
			f.loadNextChild()
		}
	})

	f.g.Inline("}")
	return f.g
}

func (f *typeBodyFormatter) formatChild() {
	// Determine spacing
	_, prevLineDiff, prevEOF := f.peekChild(-1)
	shouldBreak := false

	if !prevEOF {
		// General rule: preserve blank lines
		if prevLineDiff.EndToStart > 1 {
			// fmt.Printf("DEBUG: Break because EndToStart %d > 1\n", prevLineDiff.EndToStart)
			shouldBreak = true
		}

		// If current is docstring/field with docstring, ensure break?
		// formatField handles docstring break internally usually, but we check here too.
		if f.currentIndexChild.Field != nil && f.currentIndexChild.Field.Docstring != nil {
			shouldBreak = true
		}
	}

	if f.currentIndexChild.Comment != nil {
		formatComment(f.g, f.currentIndexChild.Comment, shouldBreak)
	} else if f.currentIndexChild.Spread != nil {
		formatSpread(f.g, f.currentIndexChild.Spread, shouldBreak)
		f.LineAndComment("")
	} else if f.currentIndexChild.Field != nil {
		formatField(f.g, f.currentIndexChild.Field, shouldBreak, f)
		f.LineAndComment("")
	}
}

// If current is docstring/field with docstring, ensure break?
type ioBodyFormatter struct {
	g                 *gen.Generator
	parent            ast.WithPositions
	children          []*ast.InputOutputChild
	maxIndex          int
	currentIndex      int
	currentIndexEOF   bool
	currentIndexChild ast.InputOutputChild
}

func newIOBodyFormatter(g *gen.Generator, parent ast.WithPositions, children []*ast.InputOutputChild) *ioBodyFormatter {
	if children == nil {
		children = []*ast.InputOutputChild{}
	}

	maxIndex := max(len(children)-1, 0)
	currentIndex := 0
	currentIndexEOF := len(children) < 1
	currentIndexChild := ast.InputOutputChild{}

	if !currentIndexEOF {
		currentIndexChild = *children[0]
	}

	return &ioBodyFormatter{
		g:                 g,
		parent:            parent,
		children:          children,
		maxIndex:          maxIndex,
		currentIndex:      currentIndex,
		currentIndexEOF:   currentIndexEOF,
		currentIndexChild: currentIndexChild,
	}
}

func (f *ioBodyFormatter) loadNextChild() {
	currentIndex := f.currentIndex + 1
	currentIndexEOF := currentIndex > f.maxIndex
	currentIndexChild := ast.InputOutputChild{}

	if !currentIndexEOF {
		currentIndexChild = *f.children[currentIndex]
	}

	f.currentIndex = currentIndex
	f.currentIndexEOF = currentIndexEOF
	f.currentIndexChild = currentIndexChild
}

func (f *ioBodyFormatter) peekChild(offset int) (ast.InputOutputChild, ast.LineDiff, bool) {
	peekIndex := f.currentIndex + offset
	peekIndexEOF := peekIndex < 0 || peekIndex > f.maxIndex
	peekIndexChild := ast.InputOutputChild{}
	lineDiff := ast.LineDiff{}

	if !peekIndexEOF {
		peekIndexChild = *f.children[peekIndex]
		lineDiff = ast.GetLineDiff(peekIndexChild, f.currentIndexChild)
	}

	return peekIndexChild, lineDiff, peekIndexEOF
}

func (f *ioBodyFormatter) LineAndComment(content string) {
	next, nextLineDiff, nextEOF := f.peekChild(1)

	if !nextEOF && next.Comment != nil && nextLineDiff.StartToEnd == 0 {
		f.g.Inline(content)

		if next.Comment.Simple != nil {
			f.g.Inlinef(" %s", *next.Comment.Simple)
		}
		if next.Comment.Block != nil {
			f.g.Inlinef(" %s", *next.Comment.Block)
		}

		f.loadNextChild()
		return
	}

	f.g.Line(content)
}

func (f *ioBodyFormatter) format() *gen.Generator {
	if f.currentIndexEOF {
		f.g.Inline("{}")
		return f.g
	}

	hasInlineComment := false
	if f.currentIndexChild.Comment != nil && f.parent != nil {
		lineDiff := ast.GetLineDiff(f.currentIndexChild, f.parent)
		if lineDiff.StartToStart == 0 {
			hasInlineComment = true
		}
	}

	if hasInlineComment {
		f.g.Inline("{")
		FormatInlineComment(f.g, f.currentIndexChild.Comment)
		f.g.Break()
		f.loadNextChild()
	} else {
		f.g.Line("{")
	}

	f.g.Block(func() {
		for !f.currentIndexEOF {
			f.formatChild()
			f.loadNextChild()
		}
	})

	f.g.Inline("}")
	return f.g
}

func (f *ioBodyFormatter) formatChild() {
	_, prevLineDiff, prevEOF := f.peekChild(-1)
	shouldBreak := false

	if !prevEOF {
		if prevLineDiff.EndToStart > 1 {
			shouldBreak = true
		}

		if f.currentIndexChild.Field != nil && f.currentIndexChild.Field.Docstring != nil {
			shouldBreak = true
		}
	}

	if f.currentIndexChild.Comment != nil {
		formatComment(f.g, f.currentIndexChild.Comment, shouldBreak)
	} else if f.currentIndexChild.Spread != nil {
		formatSpread(f.g, f.currentIndexChild.Spread, shouldBreak)
		f.LineAndComment("")
	} else if f.currentIndexChild.Field != nil {
		formatField(f.g, f.currentIndexChild.Field, shouldBreak, f)
		f.LineAndComment("")
	}
}

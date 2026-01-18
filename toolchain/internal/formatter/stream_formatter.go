package formatter

import (
	"github.com/varavelio/gen"
	"github.com/varavelio/vdl/toolchain/internal/core/ast"
	"github.com/varavelio/vdl/toolchain/internal/util/strutil"
)

type streamFormatter struct {
	g                 *gen.Generator
	streamDecl        *ast.StreamDecl
	children          []*ast.ProcOrStreamDeclChild
	maxIndex          int
	currentIndex      int
	currentIndexEOF   bool
	currentIndexChild ast.ProcOrStreamDeclChild
}

func newStreamFormatter(g *gen.Generator, streamDecl *ast.StreamDecl) *streamFormatter {
	if streamDecl == nil {
		streamDecl = &ast.StreamDecl{}
	}

	if streamDecl.Children == nil {
		streamDecl.Children = []*ast.ProcOrStreamDeclChild{}
	}

	maxIndex := max(len(streamDecl.Children)-1, 0)
	currentIndex := 0
	currentIndexEOF := len(streamDecl.Children) < 1
	currentIndexChild := ast.ProcOrStreamDeclChild{}

	if !currentIndexEOF {
		currentIndexChild = *streamDecl.Children[0]
	}

	return &streamFormatter{
		g:                 g,
		streamDecl:        streamDecl,
		children:          streamDecl.Children,
		maxIndex:          maxIndex,
		currentIndex:      currentIndex,
		currentIndexEOF:   currentIndexEOF,
		currentIndexChild: currentIndexChild,
	}
}

// loadNextChild moves the current index to the next child.
func (f *streamFormatter) loadNextChild() {
	currentIndex := f.currentIndex + 1
	currentIndexEOF := currentIndex > f.maxIndex
	currentIndexChild := ast.ProcOrStreamDeclChild{}

	if !currentIndexEOF {
		currentIndexChild = *f.children[currentIndex]
	}

	f.currentIndex = currentIndex
	f.currentIndexEOF = currentIndexEOF
	f.currentIndexChild = currentIndexChild
}

// peekChild returns information about the child at the current index +- offset.
func (f *streamFormatter) peekChild(offset int) (ast.ProcOrStreamDeclChild, ast.LineDiff, bool) {
	peekIndex := f.currentIndex + offset
	peekIndexEOF := peekIndex < 0 || peekIndex > f.maxIndex
	peekIndexChild := ast.ProcOrStreamDeclChild{}
	lineDiff := ast.LineDiff{}

	if !peekIndexEOF {
		peekIndexChild = *f.children[peekIndex]
		lineDiff = ast.GetLineDiff(peekIndexChild, f.currentIndexChild)
	}

	return peekIndexChild, lineDiff, peekIndexEOF
}

// format formats the entire streamDecl, handling spacing and EOL comments.
func (f *streamFormatter) format() *gen.Generator {
	if f.streamDecl.Docstring != nil {
		f.g.Linef(`"""%s"""`, normalizeDocstring(string(f.streamDecl.Docstring.Value)))
	}

	if f.streamDecl.Deprecated != nil {
		if f.streamDecl.Deprecated.Message == nil {
			f.g.Inline("deprecated ")
		}
		if f.streamDecl.Deprecated.Message != nil {
			f.g.Linef("deprecated(\"%s\")", strutil.EscapeQuotes(string(*f.streamDecl.Deprecated.Message)))
		}
	}

	// Force strict PascalCase
	f.g.Inlinef(`stream %s `, strutil.ToPascalCase(f.streamDecl.Name))

	if len(f.streamDecl.Children) < 1 {
		f.g.Inline("{}")
		return f.g
	}

	hasInlineComment := false
	if f.currentIndexChild.Comment != nil {
		lineDiff := ast.GetLineDiff(f.currentIndexChild, f.streamDecl)
		if lineDiff.StartToStart == 0 {
			hasInlineComment = true
		}
	}

	if hasInlineComment {
		f.g.Inline("{ ")
	} else {
		f.g.Line("{")
	}

	f.g.Block(func() {
		for !f.currentIndexEOF {
			if f.currentIndexChild.Comment != nil {
				f.formatComment()
			}

			if f.currentIndexChild.Input != nil {
				f.formatInput()
			}

			if f.currentIndexChild.Output != nil {
				f.formatOutput()
			}

			f.loadNextChild()
		}
	})

	f.g.Inline("}")

	return f.g
}

func (f *streamFormatter) formatComment() {
	_, prevLineDiff, prevEOF := f.peekChild(-1)

	shouldBreakBefore := false
	if !prevEOF {
		if prevLineDiff.StartToStart > 1 {
			shouldBreakBefore = true
		}
	}

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

func (f *streamFormatter) breakBeforeBlock() {
	prev, prevLineDiff, prevEOF := f.peekChild(-1)
	prevWasComment := prev.Comment != nil

	if prevEOF {
		return
	}

	if prevWasComment {
		if prevLineDiff.StartToStart > 1 {
			f.g.Break()
			return
		}
		return
	}

	f.g.Break()
}

func (f *streamFormatter) formatInput() {
	f.breakBeforeBlock()
	f.g.Inline("input ")
	// Use ioBodyFormatter
	bodyFormatter := newIOBodyFormatter(f.g, f.currentIndexChild, f.currentIndexChild.Input.Children)
	bodyFormatter.format()
	f.g.Break()
}

func (f *streamFormatter) formatOutput() {
	f.breakBeforeBlock()
	f.g.Inline("output ")
	// Use ioBodyFormatter
	bodyFormatter := newIOBodyFormatter(f.g, f.currentIndexChild, f.currentIndexChild.Output.Children)
	bodyFormatter.format()
	f.g.Break()
}

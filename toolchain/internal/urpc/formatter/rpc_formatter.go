package formatter

import (
	"github.com/uforg/ufogenkit"
	"github.com/varavelio/vdl/urpc/internal/urpc/ast"
	"github.com/varavelio/vdl/urpc/internal/util/strutil"
)

type rpcFormatter struct {
	g                 *ufogenkit.GenKit
	rpcDecl           *ast.RPCDecl
	children          []*ast.RPCChild
	maxIndex          int
	currentIndex      int
	currentIndexEOF   bool
	currentIndexChild ast.RPCChild
}

func newRPCFormatter(g *ufogenkit.GenKit, rpcDecl *ast.RPCDecl) *rpcFormatter {
	if rpcDecl == nil {
		rpcDecl = &ast.RPCDecl{}
	}

	if rpcDecl.Children == nil {
		rpcDecl.Children = []*ast.RPCChild{}
	}

	maxIndex := max(len(rpcDecl.Children)-1, 0)
	currentIndex := 0
	currentIndexEOF := len(rpcDecl.Children) < 1
	currentIndexChild := ast.RPCChild{}

	if !currentIndexEOF {
		currentIndexChild = *rpcDecl.Children[0]
	}

	return &rpcFormatter{
		g:                 g,
		rpcDecl:           rpcDecl,
		children:          rpcDecl.Children,
		maxIndex:          maxIndex,
		currentIndex:      currentIndex,
		currentIndexEOF:   currentIndexEOF,
		currentIndexChild: currentIndexChild,
	}
}

// loadNextChild moves the current index to the next child.
func (f *rpcFormatter) loadNextChild() {
	currentIndex := f.currentIndex + 1
	currentIndexEOF := currentIndex > f.maxIndex
	currentIndexChild := ast.RPCChild{}

	if !currentIndexEOF {
		currentIndexChild = *f.children[currentIndex]
	}

	f.currentIndex = currentIndex
	f.currentIndexEOF = currentIndexEOF
	f.currentIndexChild = currentIndexChild
}

// peekChild returns information about the child at the current index +- offset.
func (f *rpcFormatter) peekChild(offset int) (ast.RPCChild, ast.LineDiff, bool) {
	peekIndex := f.currentIndex + offset
	peekIndexEOF := peekIndex < 0 || peekIndex > f.maxIndex
	peekIndexChild := ast.RPCChild{}
	lineDiff := ast.LineDiff{}

	if !peekIndexEOF {
		peekIndexChild = *f.children[peekIndex]
		lineDiff = ast.GetLineDiff(peekIndexChild, f.currentIndexChild)
	}

	return peekIndexChild, lineDiff, peekIndexEOF
}

func (f *rpcFormatter) format() *ufogenkit.GenKit {
	if f.rpcDecl.Docstring != nil {
		f.g.Linef(`"""%s"""`, f.rpcDecl.Docstring.Value)
	}

	if f.rpcDecl.Deprecated != nil {
		if f.rpcDecl.Deprecated.Message == nil {
			f.g.Inline("deprecated ")
		}
		if f.rpcDecl.Deprecated.Message != nil {
			f.g.Linef("deprecated(\"%s\")", strutil.EscapeQuotes(string(*f.rpcDecl.Deprecated.Message)))
		}
	}

	rpcName := strutil.ToPascalCase(f.rpcDecl.Name)
	if f.currentIndexEOF {
		f.g.Linef("rpc %s {}", rpcName)
		return f.g
	}

	f.g.Linef("rpc %s {", rpcName)

	f.g.Block(func() {
		for !f.currentIndexEOF {
			// Handle docstrings, procs, streams, comments
			if f.currentIndexChild.Comment != nil {
				f.formatComment()
			} else if f.currentIndexChild.Docstring != nil {
				f.formatStandaloneDocstring()
			} else if f.currentIndexChild.Proc != nil {
				f.formatProc()
			} else if f.currentIndexChild.Stream != nil {
				f.formatStream()
			}

			f.loadNextChild()
		}
	})

	f.g.Inline("}")

	return f.g
}

func (f *rpcFormatter) formatComment() {
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

func (f *rpcFormatter) formatStandaloneDocstring() {
	f.breakBeforeBlock()
	f.g.Linef(`"""%s"""`, normalizeDocstring(string(f.currentIndexChild.Docstring.Value)))
}

func (f *rpcFormatter) formatProc() {
	f.breakBeforeBlock()
	procFormatter := newProcFormatter(f.g, f.currentIndexChild.Proc)
	procFormatter.format()
	f.g.Line("") // Add newline after proc
}

func (f *rpcFormatter) formatStream() {
	f.breakBeforeBlock()
	streamFormatter := newStreamFormatter(f.g, f.currentIndexChild.Stream)
	streamFormatter.format()
	f.g.Line("") // Add newline after stream
}

func (f *rpcFormatter) breakBeforeBlock() {
	prev, prevLineDiff, prevEOF := f.peekChild(-1)

	// If first element, no break unless comment/docstring logic?

	// Spec says: "Contents inside non-empty blocks start on a new, indented line." handled by Block().
	// "Separate each endpoint with one blank line."

	if prevEOF {
		return
	}

	// Check if previous was comment
	if prev.Comment != nil {
		if prevLineDiff.EndToStart > 1 {
			f.g.Break()
			return
		}
		// If comment is adjacent, do NOT break
		return
	}

	// Always break between endpoints or docstrings inside RPC if it's not the first element
	// But we need to check if we already have a break due to comments or source blank lines.
	// Actually, strict formatter should enforce blank line.

	// If previous was comment, we might have already handled break.
	// But let's look at logic.

	if prevLineDiff.EndToStart > 1 {
		// Already spaced in source?
		f.g.Break()
		return
	}

	// Force break
	f.g.Break()
}

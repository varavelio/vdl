package formatter

import (
	"fmt"
	"strings"

	"github.com/uforg/ufogenkit"
	"github.com/uforg/uforpc/urpc/internal/urpc/ast"
	"github.com/uforg/uforpc/urpc/internal/urpc/parser"
	"github.com/uforg/uforpc/urpc/internal/util/strutil"
)

// Format formats URPC code according to the spec, using 2 spaces for indentation.
func Format(filename, content string) (string, error) {
	if strings.TrimSpace(content) == "" {
		return "", nil
	}

	schema, err := parser.ParserInstance.ParseString(filename, content)
	if err != nil {
		return "", fmt.Errorf("error parsing URPC: %w", err)
	}

	return FormatSchema(schema), nil
}

// FormatSchema formats an already parsed UFO RPC AST Schema.
func FormatSchema(sch *ast.Schema) string {
	g := ufogenkit.NewGenKit().WithSpaces(2)

	schFormatter := newSchemaFormatter(g, sch)
	formatted := schFormatter.format().String()

	// Ensure the formatted string does not have more than 2 consecutive newlines
	formatted = strutil.LimitConsecutiveNewlines(formatted, 2)

	// Ensure the formatted string ends with exactly one newline
	formatted = strings.TrimSpace(formatted)
	formatted += "\n"
	return formatted
}

// schemaFormatter is a formatter for a schema.
type schemaFormatter struct {
	g                 *ufogenkit.GenKit
	sch               *ast.Schema
	maxIndex          int
	currentIndex      int
	currentIndexEOF   bool
	currentIndexChild ast.SchemaChild
}

// newSchemaFormatter creates a new schema formatter and initializes all the necessary fields.
func newSchemaFormatter(g *ufogenkit.GenKit, sch *ast.Schema) *schemaFormatter {
	maxIndex := max(len(sch.Children)-1, 0)
	currentIndex := 0
	currentIndexEOF := len(sch.Children) < 1
	currentIndexChild := ast.SchemaChild{}

	if !currentIndexEOF {
		currentIndexChild = *sch.Children[0]
	}

	return &schemaFormatter{
		g:                 g,
		sch:               sch,
		maxIndex:          maxIndex,
		currentIndex:      currentIndex,
		currentIndexEOF:   currentIndexEOF,
		currentIndexChild: currentIndexChild,
	}
}

// loadNextChild moves the current index to the next child.
func (f *schemaFormatter) loadNextChild() {
	currentIndex := f.currentIndex + 1
	currentIndexEOF := currentIndex > f.maxIndex
	currentIndexChild := ast.SchemaChild{}

	if !currentIndexEOF {
		currentIndexChild = *f.sch.Children[currentIndex]
	}

	f.currentIndex = currentIndex
	f.currentIndexEOF = currentIndexEOF
	f.currentIndexChild = currentIndexChild
}

// peekChild returns information about the child at the current index +- offset.
//
// Returns:
//   - The child at the current index +- offset.
//   - The line diff between the peeked child and the current child.
//   - A bool indicating if the peeked child is out of bounds (EOL).
func (f *schemaFormatter) peekChild(offset int) (ast.SchemaChild, ast.LineDiff, bool) {
	peekIndex := f.currentIndex + offset
	peekIndexEOF := peekIndex < 0 || peekIndex > f.maxIndex
	peekIndexChild := ast.SchemaChild{}
	lineDiff := ast.LineDiff{}

	if !peekIndexEOF {
		peekIndexChild = *f.sch.Children[peekIndex]
		lineDiff = ast.GetLineDiff(peekIndexChild, f.currentIndexChild)
	}

	return peekIndexChild, lineDiff, peekIndexEOF
}

// format formats the entire schema, handling spacing and EOL comments.
//
// Returns the formatted genkit.GenKit.
func (f *schemaFormatter) format() *ufogenkit.GenKit {
	for !f.currentIndexEOF {
		switch f.currentIndexChild.Kind() {
		case ast.SchemaChildKindComment:
			f.formatComment()
		case ast.SchemaChildKindDocstring:
			f.formatStandaloneDocstring()
		case ast.SchemaChildKindInclude:
			f.formatInclude()
		case ast.SchemaChildKindConst:
			f.formatConst()
		case ast.SchemaChildKindEnum:
			f.formatEnum()
		case ast.SchemaChildKindPattern:
			f.formatPattern()
		case ast.SchemaChildKindType:
			f.formatType()
		case ast.SchemaChildKindRPC:
			f.formatRPC()
		}

		f.loadNextChild()
	}

	return f.g
}

// LineAndComment writes a line of content to the formatter. It also handles inline comments.
func (f *schemaFormatter) LineAndComment(content string) {
	next, nextLineDiff, nextEOF := f.peekChild(1)

	// If next is an inline comment
	if !nextEOF && next.Kind() == ast.SchemaChildKindComment && nextLineDiff.StartToEnd == 0 {
		f.g.Inline(content)

		if next.Comment.Simple != nil {
			f.g.Linef(" %s", *next.Comment.Simple)
		}

		if next.Comment.Block != nil {
			f.g.Linef(" %s", *next.Comment.Block)
		}

		// Skip the inline comment because it's already written
		f.loadNextChild()
		return
	}

	f.g.Line(content)
}

// LineAndCommentf is the same as Line but with a formatted string.
func (f *schemaFormatter) LineAndCommentf(format string, args ...any) {
	f.LineAndComment(fmt.Sprintf(format, args...))
}

func (f *schemaFormatter) formatComment() {
	_, prevLineDiff, prevEOF := f.peekChild(-1)

	shouldBreakBefore := false
	if !prevEOF {
		if prevLineDiff.EndToStart > 1 {
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

func (f *schemaFormatter) formatStandaloneDocstring() {
	prev, prevLineDiff, prevEOF := f.peekChild(-1)

	shouldBreakBefore := false
	if !prevEOF {
		if prev.Kind() != ast.SchemaChildKindDocstring && prev.Kind() != ast.SchemaChildKindComment {
			shouldBreakBefore = true
		}

		// Preserve blank lines from source
		if prevLineDiff.EndToStart > 1 {
			shouldBreakBefore = true
		}
	}

	if shouldBreakBefore {
		f.g.Line("")
	}

	f.LineAndCommentf(`"""%s"""`, normalizeDocstring(string(f.currentIndexChild.Docstring.Value)))
}

func (f *schemaFormatter) formatInclude() {
	// Group consecutive includes, but respect existing newlines if they are separate groups
	prev, prevLineDiff, prevEOF := f.peekChild(-1)

	shouldBreakBefore := false
	if !prevEOF {
		// If previous was NOT an include or comment, break
		if prev.Kind() != ast.SchemaChildKindInclude && prev.Kind() != ast.SchemaChildKindComment {
			shouldBreakBefore = true
		} else if prevLineDiff.EndToStart > 1 {
			// If there was a blank line in source, preserve it
			shouldBreakBefore = true
		}
	}

	if shouldBreakBefore {
		f.g.Break()
	}

	includeFormatter := newIncludeFormatter(f.g, f.currentIndexChild.Include)
	includeFormatter.format()
	f.LineAndComment("")
}

func (f *schemaFormatter) formatConst() {
	f.breakBeforeGroupable(ast.SchemaChildKindConst)
	constFormatter := newConstFormatter(f.g, f.currentIndexChild.Const)
	constFormatter.format()
	f.LineAndComment("")
}

func (f *schemaFormatter) formatEnum() {
	f.breakBeforeTopLevel()
	enumFormatter := newEnumFormatter(f.g, f.currentIndexChild.Enum)
	enumFormatter.format()
	f.LineAndComment("")
}

func (f *schemaFormatter) formatPattern() {
	f.breakBeforeGroupable(ast.SchemaChildKindPattern)
	patternFormatter := newPatternFormatter(f.g, f.currentIndexChild.Pattern)
	patternFormatter.format()
	f.LineAndComment("")
}

func (f *schemaFormatter) breakBeforeGroupable(kind ast.SchemaChildKind) {
	prev, prevLineDiff, prevEOF := f.peekChild(-1)

	shouldBreakBefore := false
	if !prevEOF {
		// If previous was NOT the same kind or comment, break (standard top-level separation)
		if prev.Kind() != kind && prev.Kind() != ast.SchemaChildKindComment {
			shouldBreakBefore = true
		} else if prev.Kind() == ast.SchemaChildKindComment {
			// If previous was comment, use standard spacing
			if prevLineDiff.EndToStart > 1 {
				shouldBreakBefore = true
			}
		} else {
			// Previous was SAME kind.
			// Only break if one of them has a docstring.
			// Note: If PREVIOUS had a docstring, it usually occupies multiple lines,
			// but we are looking at the gap between the end of previous and start of current.
			// Actually, if previous had a docstring, we DO want a blank line after it to separate it from the next item.
			// AND if CURRENT has a docstring, we want a blank line before it.

			// Check current docstring
			if kind == ast.SchemaChildKindConst && f.currentIndexChild.Const.Docstring != nil {
				shouldBreakBefore = true
			} else if kind == ast.SchemaChildKindPattern && f.currentIndexChild.Pattern.Docstring != nil {
				shouldBreakBefore = true
			}

			// Check previous docstring
			// We need to check if previous item (of same kind) had a docstring.
			if !shouldBreakBefore {
				if kind == ast.SchemaChildKindConst && prev.Const.Docstring != nil {
					shouldBreakBefore = true
				} else if kind == ast.SchemaChildKindPattern && prev.Pattern.Docstring != nil {
					shouldBreakBefore = true
				}
			}
		}
	}

	if shouldBreakBefore {
		f.g.Break()
	}
}

func (f *schemaFormatter) breakBeforeTopLevel() {
	prev, prevLineDiff, prevEOF := f.peekChild(-1)

	shouldBreakBefore := false
	if !prevEOF {
		if prev.Kind() != ast.SchemaChildKindComment {
			shouldBreakBefore = true
		}

		if prevLineDiff.EndToStart > 1 {
			shouldBreakBefore = true
		}
	}

	if shouldBreakBefore {
		f.g.Break()
	}
}

func (f *schemaFormatter) formatType() {
	f.breakBeforeTopLevel()

	typeFormatter := newTypeFormatter(f.g, f.currentIndexChild.Type)
	typeFormatter.format()
	f.LineAndComment("")
}

func (f *schemaFormatter) formatRPC() {
	f.breakBeforeTopLevel()

	rpcFormatter := newRPCFormatter(f.g, f.currentIndexChild.RPC)
	rpcFormatter.format()
	f.LineAndComment("")
}

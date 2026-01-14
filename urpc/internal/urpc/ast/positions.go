package ast

import "github.com/alecthomas/participle/v2/lexer"

// Any node in the AST containing a field Pos lexer.Position
// will be automatically populated from the nearest matching token.
//
// Any node in the AST containing a field EndPos lexer.Position
// will be automatically populated from the token at the end of the node.
//
// https://github.com/alecthomas/participle/blob/master/README.md#error-reporting

// Position is an alias for the participle.Position type.
type Position = lexer.Position

// Positions is a struct that contains the start and end positions of a node.
//
// Used to embed in structs that contain a start and end position and
// automatically populate the Pos field, EndPos field, and the
// GetPositions method.
type Positions struct {
	Pos    Position
	EndPos Position
}

// GetPositions returns the start and end positions of the node.
func (p Positions) GetPositions() Positions {
	return p
}

// WithPositions is an interface that can be implemented by any type
// that has a GetPositions method.
type WithPositions interface {
	GetPositions() Positions
}

type LineDiff struct {
	// The difference in lines between the start of the first position and the start of the second position.
	StartToStart int
	// The difference in lines between the start of the first position and the end of the second position.
	StartToEnd int
	// The difference in lines between the end of the first position and the start of the second position.
	EndToStart int
	// The difference in lines between the end of the first position and the end of the second position.
	EndToEnd int

	// The absolute difference in lines between the start of the first position and the start of the second position.
	AbsStartToStart int
	// The absolute difference in lines between the start of the first position and the end of the second position.
	AbsStartToEnd int
	// The absolute difference in lines between the end of the first position and the start of the second position.
	AbsEndToStart int
	// The absolute difference in lines between the end of the first position and the end of the second position.
	AbsEndToEnd int
}

// GetLineDiff returns the line diff between two positions.
func GetLineDiff(from, to WithPositions) LineDiff {
	abs := func(x int) int {
		if x < 0 {
			return -x
		}
		return x
	}

	diff := LineDiff{
		StartToStart: to.GetPositions().Pos.Line - from.GetPositions().Pos.Line,
		StartToEnd:   to.GetPositions().EndPos.Line - from.GetPositions().Pos.Line,
		EndToStart:   to.GetPositions().Pos.Line - from.GetPositions().EndPos.Line,
		EndToEnd:     to.GetPositions().EndPos.Line - from.GetPositions().EndPos.Line,
	}

	diff.AbsStartToStart = abs(diff.StartToStart)
	diff.AbsStartToEnd = abs(diff.StartToEnd)
	diff.AbsEndToStart = abs(diff.EndToStart)
	diff.AbsEndToEnd = abs(diff.EndToEnd)

	return diff
}

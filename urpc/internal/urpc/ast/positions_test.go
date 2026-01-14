package ast_test

import (
	"testing"

	"github.com/alecthomas/participle/v2/lexer"
	"github.com/stretchr/testify/assert"
	"github.com/uforg/uforpc/urpc/internal/urpc/ast"
)

func TestPositions_GetPositions(t *testing.T) {
	pos := lexer.Position{Line: 1, Column: 1, Offset: 0}
	endPos := lexer.Position{Line: 1, Column: 10, Offset: 10}
	p := ast.Positions{
		Pos:    pos,
		EndPos: endPos,
	}

	assert.Equal(t, p, p.GetPositions())
}

func TestGetLineDiff(t *testing.T) {
	tests := []struct {
		name     string
		from     ast.Positions
		to       ast.Positions
		expected ast.LineDiff
	}{
		{
			name: "same lines",
			from: ast.Positions{
				Pos:    lexer.Position{Line: 10},
				EndPos: lexer.Position{Line: 10},
			},
			to: ast.Positions{
				Pos:    lexer.Position{Line: 10},
				EndPos: lexer.Position{Line: 10},
			},
			expected: ast.LineDiff{
				StartToStart:    0,
				StartToEnd:      0,
				EndToStart:      0,
				EndToEnd:        0,
				AbsStartToStart: 0,
				AbsStartToEnd:   0,
				AbsEndToStart:   0,
				AbsEndToEnd:     0,
			},
		},
		{
			name: "different lines",
			from: ast.Positions{
				Pos:    lexer.Position{Line: 10},
				EndPos: lexer.Position{Line: 15},
			},
			to: ast.Positions{
				Pos:    lexer.Position{Line: 20},
				EndPos: lexer.Position{Line: 25},
			},
			expected: ast.LineDiff{
				StartToStart:    10, // 20 - 10
				StartToEnd:      15, // 25 - 10
				EndToStart:      5,  // 20 - 15
				EndToEnd:        10, // 25 - 15
				AbsStartToStart: 10,
				AbsStartToEnd:   15,
				AbsEndToStart:   5,
				AbsEndToEnd:     10,
			},
		},
		{
			name: "different lines (negative diffs)",
			from: ast.Positions{
				Pos:    lexer.Position{Line: 20},
				EndPos: lexer.Position{Line: 25},
			},
			to: ast.Positions{
				Pos:    lexer.Position{Line: 10},
				EndPos: lexer.Position{Line: 15},
			},
			expected: ast.LineDiff{
				StartToStart:    -10, // 10 - 20
				StartToEnd:      -5,  // 15 - 20
				EndToStart:      -15, // 10 - 25
				EndToEnd:        -10, // 15 - 25
				AbsStartToStart: 10,
				AbsStartToEnd:   5,
				AbsEndToStart:   15,
				AbsEndToEnd:     10,
			},
		},
		{
			name: "mixed offsets",
			from: ast.Positions{
				Pos:    lexer.Position{Line: 100},
				EndPos: lexer.Position{Line: 102},
			},
			to: ast.Positions{
				Pos:    lexer.Position{Line: 90},
				EndPos: lexer.Position{Line: 95},
			},
			expected: ast.LineDiff{
				StartToStart:    -10, // 90 - 100
				StartToEnd:      -5,  // 95 - 100
				EndToStart:      -12, // 90 - 102
				EndToEnd:        -7,  // 95 - 102
				AbsStartToStart: 10,
				AbsStartToEnd:   5,
				AbsEndToStart:   12,
				AbsEndToEnd:     7,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diff := ast.GetLineDiff(tt.from, tt.to)
			assert.Equal(t, tt.expected, diff)
		})
	}
}

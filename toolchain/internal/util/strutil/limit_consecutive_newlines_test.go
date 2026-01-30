package strutil

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLimitConsecutiveNewlines_DefaultOne(t *testing.T) {
	input := "Hello\n\nWorld\n\n\n!"
	expected := "Hello\nWorld\n!"
	result := LimitConsecutiveNewlines(input, 1)
	require.Equal(t, expected, result)
}

func TestLimitConsecutiveNewlines_MaxTwo(t *testing.T) {
	input := "Line1\n\n\nLine2\nLine3\n\n\n\nLine4"
	expected := "Line1\n\nLine2\nLine3\n\nLine4"
	result := LimitConsecutiveNewlines(input, 2)
	require.Equal(t, expected, result)
}

func TestLimitConsecutiveNewlines_NoChange(t *testing.T) {
	input := "No\nextra\nnewlines"
	result := LimitConsecutiveNewlines(input, 2)
	require.Equal(t, input, result)
}

func TestLimitConsecutiveNewlines_ZeroOrNegativeMax(t *testing.T) {
	input := "A\n\nB"
	// max <= 0 should behave like max == 1
	expected := "A\nB"
	result := LimitConsecutiveNewlines(input, 0)
	require.Equal(t, expected, result)

	result = LimitConsecutiveNewlines(input, -5)
	require.Equal(t, expected, result)
}

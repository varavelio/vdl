package strutil

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCenterText(t *testing.T) {
	// Define test cases in a table-driven format
	testCases := []struct {
		name         string // A description of the test case
		text         string // The input text to center
		desiredWidth int    // The desired total width
		expected     string // The expected output string
	}{
		{
			name:         "Single Line, Even Padding",
			text:         "hello",
			desiredWidth: 11,
			expected:     "   hello   ",
		},
		{
			name:         "Single Line, Odd Padding",
			text:         "go",
			desiredWidth: 11,
			expected:     "    go     ", // Extra space goes to the right
		},
		{
			name:         "Single Line, No Padding Needed",
			text:         "exact fit",
			desiredWidth: 9,
			expected:     "exact fit",
		},
		{
			name:         "Single Line, Text Wider Than Width",
			text:         "this text is definitely too long",
			desiredWidth: 10,
			expected:     "this text is definitely too long",
		},
		{
			name:         "Empty String",
			text:         "",
			desiredWidth: 8,
			expected:     "        ",
		},
		{
			name:         "Multi-line, Even Line Lengths",
			text:         "line one\nline two",
			desiredWidth: 20,
			expected:     "      line one      \n      line two      ",
		},
		{
			name:         "Multi-line, Uneven Line Lengths",
			text:         "short\nthis is a longer line",
			desiredWidth: 30,
			expected:     "    short                     \n    this is a longer line     ",
		},
		{
			name:         "Multi-line, Block Wider Than Width",
			text:         "this is a very long line\nand this one is shorter",
			desiredWidth: 20,
			expected:     "this is a very long line\nand this one is shorter",
		},
		{
			name:         "Multi-line with Empty Line",
			text:         "A\n\nB",
			desiredWidth: 5,
			expected:     "  A  \n     \n  B  ",
		},
		{
			name:         "Multi-line with Trailing Newline",
			text:         "line 1\nline 2\n",
			desiredWidth: 10,
			expected:     "  line 1  \n  line 2  \n          ",
		},
	}

	// Run the test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a new require object for each test case
			r := require.New(t)

			// Call the function we are testing
			actual := CenterText(tc.text, tc.desiredWidth)

			// Assert that the actual output matches the expected output
			r.Equal(tc.expected, actual)
		})
	}
}

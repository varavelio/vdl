package strutil

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNormalizeIndent(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "Standard code block with tabs",
			input: `
		// This is a comment.
		func myFunc() {
			// This has extra indentation.
			return true
		}
	`,
			expected: `
// This is a comment.
func myFunc() {
	// This has extra indentation.
	return true
}
	`,
		},
		{
			name: "Markdown code block with spaces",
			input: `
    Here is some text.

    ` + "```" + `typescript
      // Some indented code inside a markdown block
      const x = 1;
    ` + "```" + `
    `,
			expected: `
Here is some text.

` + "```" + `typescript
  // Some indented code inside a markdown block
  const x = 1;
` + "```" + `
`,
		},
		{
			name: "Text with no leading indentation",
			input: `Hello world.
This is a test.
  - With a list item.`,
			expected: `Hello world.
This is a test.
  - With a list item.`,
		},
		{
			name:     "Empty string input",
			input:    "",
			expected: "",
		},
		{
			name:     "String with only whitespace",
			input:    "\n    \n\t\n",
			expected: "\n    \n\t\n",
		},
		{
			name: "Text starting with empty lines",
			input: `

		// First line of actual content
		const value = 42;
`,
			expected: `

// First line of actual content
const value = 42;
`,
		},
		{
			name: "Mixed tabs and spaces for indentation",
			input: `
	  first line
	  	second line with more
	  third line
`,
			expected: `
first line
	second line with more
third line
`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := require.New(t)
			actual := NormalizeIndent(tc.input)
			r.Equal(tc.expected, actual)
		})
	}
}

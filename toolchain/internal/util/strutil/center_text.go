package strutil

import (
	"strings"
	"unicode/utf8"
)

// CenterText centers text within a given width.
// It handles both single and multi-line strings, treating multi-line
// strings as a single block. It prevents panics by not adding
// padding if the text exceeds the desired width.
func CenterText(text string, desiredWidth int) string {
	lines := strings.Split(text, "\n")

	// Find the widest line to determine the block's width
	var longestLineWidth int
	for _, line := range lines {
		lineWidth := utf8.RuneCountInString(line)
		if lineWidth > longestLineWidth {
			longestLineWidth = lineWidth
		}
	}

	// Calculate the left padding for the entire block
	blockLeftPaddingCount := 0
	if longestLineWidth < desiredWidth {
		blockLeftPaddingCount = (desiredWidth - longestLineWidth) / 2
	}
	blockLeftPadding := strings.Repeat(" ", blockLeftPaddingCount)

	// Build the result line by line
	var resultBuilder strings.Builder
	for i, line := range lines {
		// Add block padding and the line itself
		resultBuilder.WriteString(blockLeftPadding)
		resultBuilder.WriteString(line)

		// Calculate and add right padding to fill the remaining space
		currentWidth := blockLeftPaddingCount + utf8.RuneCountInString(line)
		rightPaddingCount := desiredWidth - currentWidth

		// Only add padding if the count is positive
		if rightPaddingCount > 0 {
			resultBuilder.WriteString(strings.Repeat(" ", rightPaddingCount))
		}

		// Add a newline if it's not the last line
		if i < len(lines)-1 {
			resultBuilder.WriteString("\n")
		}
	}

	return resultBuilder.String()
}

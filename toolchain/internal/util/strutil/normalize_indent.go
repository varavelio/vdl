package strutil

import (
	"strings"
	"unicode"
)

// NormalizeIndent removes the leading whitespace from each line of a string
// based on the indentation of the first non-empty line.
func NormalizeIndent(text string) string {
	// 1. Split the text into a slice of lines.
	lines := strings.Split(text, "\n")
	indentation := ""

	// 2. Find the first non-empty line to determine the base indentation.
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			// 3. Find the index of the first character that is not a space.
			// This efficiently captures the leading whitespace (tabs or spaces).
			endOfIndent := strings.IndexFunc(line, func(r rune) bool {
				return !unicode.IsSpace(r)
			})

			// If the line is all whitespace, the whole line is the indentation.
			// Otherwise, slice the string to get the indentation part.
			if endOfIndent == -1 {
				indentation = line
			} else {
				indentation = line[:endOfIndent]
			}
			break
		}
	}

	// If no indentation was found, return the original text.
	if indentation == "" {
		return text
	}

	// 4. Create a new slice to hold the dedented lines.
	dedentedLines := make([]string, len(lines))

	// 5. Remove the base indentation from every line.
	for i, line := range lines {
		dedentedLines[i] = strings.TrimPrefix(line, indentation)
	}

	// 6. Join the lines back into a single string.
	return strings.Join(dedentedLines, "\n")
}

package strutil

import (
	"strings"
	"unicode"
)

// ToPascalCase converts a string to PascalCase.
// It handles delimiters, camelCase transitions, and acronyms (e.g., "JSONBody" -> "JsonBody").
func ToPascalCase(s string) string {
	if s == "" {
		return ""
	}

	var sb strings.Builder
	sb.Grow(len(s))

	runes := []rune(s)
	length := len(runes)
	startOfWord := true

	for i := range length {
		r := runes[i]

		// Treat any non-alphanumeric char as a delimiter
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			startOfWord = true
			continue
		}

		if i > 0 && !startOfWord {
			if isWordBoundary(runes, i) {
				startOfWord = true
			}
		}

		if startOfWord {
			sb.WriteRune(unicode.ToUpper(r))
			startOfWord = false
		} else {
			sb.WriteRune(unicode.ToLower(r))
		}
	}

	return sb.String()
}

// isWordBoundary checks if the character at current index `i` marks the start of a new word
// based on camelCase or acronym rules. Assumes i > 0.
func isWordBoundary(runes []rune, i int) bool {
	prev := runes[i-1]
	curr := runes[i]

	// 1. Digit -> Letter transition (e.g., "123Test")
	if unicode.IsDigit(prev) && unicode.IsLetter(curr) {
		return true
	}

	// 2. Lower -> Upper transition (camelCase: "fooBar")
	if unicode.IsLower(prev) && unicode.IsUpper(curr) {
		return true
	}

	// 3. Acronym Boundary (Upper -> Upper -> Lower)
	// e.g., "JSONBody": 'B' is Upper, prev 'N' is Upper, next 'o' is Lower.
	// This marks 'B' as the start of "Body".
	if unicode.IsUpper(prev) && unicode.IsUpper(curr) {
		if i+1 < len(runes) && unicode.IsLower(runes[i+1]) {
			return true
		}
	}

	return false
}

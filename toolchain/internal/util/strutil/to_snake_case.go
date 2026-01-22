package strutil

import (
	"strings"
	"unicode"
)

// ToSnakeCase converts a string to snake_case.
func ToSnakeCase(str string) string {
	if str == "" {
		return ""
	}

	var sb strings.Builder
	sb.Grow(len(str) + 4) // Some pre-allocated extra space for underscores

	runes := []rune(str)
	length := len(runes)
	lastWasUnderscore := false

	for i := range length {
		r := runes[i]

		// Handle separators by normalizing them to underscores
		if isSeparator(r) {
			if !lastWasUnderscore && sb.Len() > 0 {
				sb.WriteRune('_')
				lastWasUnderscore = true
			}
			continue
		}

		// Insert underscore before current char if it marks a new word boundary
		if i > 0 && !lastWasUnderscore {
			if shouldInsertUnderscore(runes, i) {
				sb.WriteRune('_')
			}
		}

		sb.WriteRune(unicode.ToLower(r))
		lastWasUnderscore = false
	}

	return strings.TrimSuffix(sb.String(), "_")
}

// isSeparator checks if a rune should be treated as a word separator.
func isSeparator(r rune) bool {
	return unicode.IsSpace(r) || r == '-' || r == '_' || r == ':' || r == '.'
}

// shouldInsertUnderscore determines if an underscore is needed before the character at index i.
// It handles camelCase (fooBar -> foo_bar) and acronym boundaries (HTMLParser -> html_parser).
func shouldInsertUnderscore(runes []rune, i int) bool {
	curr := runes[i]
	prev := runes[i-1]

	// Only insert if current is Upper
	if !unicode.IsUpper(curr) {
		return false
	}

	// Case 1: camelCase (lower/digit -> Upper)
	// e.g. "fooBar" -> "foo_bar", "123Start" -> "123_start"
	if unicode.IsLower(prev) || unicode.IsDigit(prev) {
		return true
	}

	// Case 2: Acronym boundary (Upper -> Upper -> Lower)
	// e.g. "HTMLParser": At 'P', prev='L' (Upper), next='a' (Lower).
	// We want "html_parser", so 'P' starts a new word.
	if unicode.IsUpper(prev) {
		if i+1 < len(runes) && unicode.IsLower(runes[i+1]) {
			return true
		}
	}

	return false
}

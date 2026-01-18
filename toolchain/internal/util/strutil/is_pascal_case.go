package strutil

import "unicode"

// IsPascalCase checks if a string follows PascalCase naming convention.
// The first character must be uppercase and the string should contain only alphanumeric characters.
func IsPascalCase(s string) bool {
	if len(s) == 0 {
		return false
	}

	if !unicode.IsUpper(rune(s[0])) {
		return false
	}

	for _, r := range s {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			return false
		}
	}

	return true
}

package strutil

import "unicode"

// IsCamelCase checks if a string follows camelCase naming convention.
// The first character must be lowercase and the string should contain only alphanumeric characters.
func IsCamelCase(s string) bool {
	if len(s) == 0 {
		return false
	}

	if !unicode.IsLower(rune(s[0])) {
		return false
	}

	for _, r := range s {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			return false
		}
	}

	return true
}

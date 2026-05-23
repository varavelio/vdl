package strutil

import (
	"strings"
)

// GetStrAfter returns the substring that follows the specified delimiter in the input string.
// It discards the delimiter and any text before it.
//
// If the delimiter is not found, it returns an empty string.
func GetStrAfter(input, delimiter string) string {
	if delimiter == "" {
		return input
	}

	_, after, found := strings.Cut(input, delimiter)
	if !found {
		return ""
	}

	return after
}

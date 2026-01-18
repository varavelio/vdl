package strutil

import (
	"strings"
)

// GetStrBefore returns the substring that precedes the specified delimiter in the input string.
// It discards the delimiter and any text after it.
//
// If the delimiter is not found, it returns an empty string.
func GetStrBefore(input, delimiter string) string {
	if delimiter == "" {
		return input
	}

	index := strings.Index(input, delimiter)
	if index == -1 {
		return ""
	}

	return input[:index]
}

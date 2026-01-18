package strutil

import "strings"

// EscapeQuotes escapes quotes in a string, replacing them with a backslash.
// If a backslash is already present, it is doubled.
func EscapeQuotes(str string) string {
	str = strings.ReplaceAll(str, `\`, `\\`)
	str = strings.ReplaceAll(str, `"`, `\"`)
	return str
}

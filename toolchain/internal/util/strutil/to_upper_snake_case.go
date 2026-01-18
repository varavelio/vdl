package strutil

import "strings"

// ToUpperSnakeCase converts a string to UPPER_SNAKE_CASE.
func ToUpperSnakeCase(s string) string {
	return strings.ToUpper(ToSnakeCase(s))
}

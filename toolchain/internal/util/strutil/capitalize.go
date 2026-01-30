package strutil

import "strings"

// Capitalize returns the string with the first letter capitalized.
func Capitalize(str string) string {
	if len(str) == 0 {
		return str
	}
	if len(str) == 1 {
		return strings.ToUpper(str)
	}
	return strings.ToUpper(str[:1]) + str[1:]
}

// CapitalizeStrict returns the string with the first letter capitalized
// and the rest of the string always lowercase.
func CapitalizeStrict(str string) string {
	if len(str) == 0 {
		return str
	}
	return Capitalize(strings.ToLower(str))
}

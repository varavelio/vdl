package strutil

import "unicode"

// ToCamelCase converts a string to camelCase, it will interpret all
// space like characters, underscores and dashes as word boundaries.
func ToCamelCase(str string) string {
	pascalCase := ToPascalCase(str)
	if len(pascalCase) == 0 {
		return pascalCase
	}
	runes := []rune(pascalCase)
	runes[0] = unicode.ToLower(runes[0])
	return string(runes)
}

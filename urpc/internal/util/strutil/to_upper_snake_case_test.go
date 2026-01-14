package strutil

import "testing"

func TestToUpperSnakeCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", ""},
		{"simple", "SIMPLE"},
		{"camelCase", "CAMEL_CASE"},
		{"PascalCase", "PASCAL_CASE"},
		{"snake_case", "SNAKE_CASE"},
		{"UPPER_SNAKE_CASE", "UPPER_SNAKE_CASE"},
		{"HTMLParser", "HTML_PARSER"},
		{"JSONProvider", "JSON_PROVIDER"},
	}

	for _, test := range tests {
		result := ToUpperSnakeCase(test.input)
		if result != test.expected {
			t.Errorf("ToUpperSnakeCase(%q) = %q; want %q", test.input, result, test.expected)
		}
	}
}

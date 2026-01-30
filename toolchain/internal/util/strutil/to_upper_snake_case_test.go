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
		{"simpleXMLParser", "SIMPLE_XML_PARSER"},
		{"PDFLoader", "PDF_LOADER"},
		{"startMiddleEnd", "START_MIDDLE_END"},
		{"withNumber1", "WITH_NUMBER1"},
		{"123Start", "123_START"},
		{"Start123", "START123"},
		{"foo_Bar", "FOO_BAR"},
		{"foo__Bar", "FOO_BAR"},
		{"_foo", "FOO"},
		{"foo_", "FOO"},
		{"__foo__", "FOO"},
		{"  foo  ", "FOO"},
		{"foo.bar", "FOO_BAR"},
		{"foo:bar", "FOO_BAR"},
		{"foo-bar", "FOO_BAR"},
	}

	for _, test := range tests {
		result := ToUpperSnakeCase(test.input)
		if result != test.expected {
			t.Errorf("ToUpperSnakeCase(%q) = %q; want %q", test.input, result, test.expected)
		}
	}
}

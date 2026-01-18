package strutil

import "testing"

func TestToSnakeCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", ""},
		{"simple", "simple"},
		{"camelCase", "camel_case"},
		{"PascalCase", "pascal_case"},
		{"snake_case", "snake_case"},
		{"UPPER_SNAKE_CASE", "upper_snake_case"},
		{"HTMLParser", "html_parser"},
		{"JSONProvider", "json_provider"},
		{"simpleXMLParser", "simple_xml_parser"},
		{"PDFLoader", "pdf_loader"},
		{"startMiddleEnd", "start_middle_end"},
		{"withNumber1", "with_number1"},
		{"123Start", "123_start"},
		{"Start123", "start123"},
		{"foo_Bar", "foo_bar"},
		{"foo__Bar", "foo_bar"},
		{"_foo", "foo"},
		{"foo_", "foo"},
		{"__foo__", "foo"},
		{"  foo  ", "foo"},
		{"foo.bar", "foo_bar"},
		{"foo:bar", "foo_bar"},
		{"foo-bar", "foo_bar"},
	}

	for _, test := range tests {
		result := ToSnakeCase(test.input)
		if result != test.expected {
			t.Errorf("ToSnakeCase(%q) = %q; want %q", test.input, result, test.expected)
		}
	}
}

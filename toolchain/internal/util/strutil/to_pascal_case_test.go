package strutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToPascalCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello world", "HelloWorld"},
		{"hello_world", "HelloWorld"},
		{"hello-world", "HelloWorld"},
		{"hello   world", "HelloWorld"},
		{"HELLO WORLD", "HelloWorld"},
		{"HelloWorld", "HelloWorld"},
		{"hello-WORLD", "HelloWorld"},
		{"hello_world_test", "HelloWorldTest"},
		{"hello-world-test", "HelloWorldTest"},
		{"hello world test", "HelloWorldTest"},
		{"hello   world   test", "HelloWorldTest"},
		{"hello-world_test", "HelloWorldTest"},
		{"hello_world-test", "HelloWorldTest"},
		{"hello-world_test-case", "HelloWorldTestCase"},
		{"hello_world-test_case", "HelloWorldTestCase"},
		{"hello world-test case", "HelloWorldTestCase"},
		{"hello world test case", "HelloWorldTestCase"},
		{"123hello world", "123HelloWorld"},
		{"hello123 world", "Hello123World"},
		{"hello world123", "HelloWorld123"},
		{"", ""},
		{"singleword", "Singleword"},
		{"singleWord", "SingleWord"},
		{"SingleWord", "SingleWord"},
		{"singleWORD", "SingleWord"},
		{"single-WORD", "SingleWord"},
		{"single_WORD", "SingleWord"},
		{"single-word", "SingleWord"},
		{"single_word", "SingleWord"},
		{"single word", "SingleWord"},
		{"single   word", "SingleWord"},
		{"multiple   spaces", "MultipleSpaces"},
		{"multiple-spaces", "MultipleSpaces"},
		{"multiple_spaces", "MultipleSpaces"},
		{"multiple-_spaces", "MultipleSpaces"},
		{"multiple_-spaces", "MultipleSpaces"},
		{"multiple-_-spaces", "MultipleSpaces"},
		{"multiple--spaces", "MultipleSpaces"},
		{"multiple__spaces", "MultipleSpaces"},
		{"multiple__-spaces", "MultipleSpaces"},
		{"multiple-__spaces", "MultipleSpaces"},
		// Acronym handling
		{"HTMLParser", "HtmlParser"},
		{"JSONBody", "JsonBody"},
		{"HTTPRequest", "HttpRequest"},
		{"SimpleXMLParser", "SimpleXmlParser"},
		{"param_id", "ParamId"},
		{"param_ID", "ParamId"},
		{"ServeHTTP", "ServeHttp"},
	}

	for _, test := range tests {
		result := ToPascalCase(test.input)
		assert.Equal(t, test.expected, result, "Input: %s", test.input)
	}
}

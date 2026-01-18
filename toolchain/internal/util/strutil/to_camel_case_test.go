package strutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToCamelCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello world", "helloWorld"},
		{"HelloWorld", "helloWorld"},
		{"hello_world", "helloWorld"},
		{"hello-world", "helloWorld"},
		{"hello   world", "helloWorld"},
		{"HELLO WORLD", "helloWorld"},
		{"hello-WORLD", "helloWorld"},
		{"hello_world_test", "helloWorldTest"},
		{"hello-world-test", "helloWorldTest"},
		{"hello world test", "helloWorldTest"},
		{"hello   world   test", "helloWorldTest"},
		{"hello-world_test", "helloWorldTest"},
		{"hello_world-test", "helloWorldTest"},
		{"hello-world_test-case", "helloWorldTestCase"},
		{"hello_world-test_case", "helloWorldTestCase"},
		{"hello world-test case", "helloWorldTestCase"},
		{"hello world test case", "helloWorldTestCase"},
		{"123hello world", "123HelloWorld"},
		{"hello123 world", "hello123World"},
		{"hello world123", "helloWorld123"},
		{"", ""},
		{"singleword", "singleword"},
		{"singleWord", "singleWord"},
		{"SingleWord", "singleWord"},
		{"singleWORD", "singleWord"},
		{"single-WORD", "singleWord"},
		{"single_WORD", "singleWord"},
		{"single-word", "singleWord"},
		{"single_word", "singleWord"},
		{"single word", "singleWord"},
		{"single   word", "singleWord"},
		{"multiple   spaces", "multipleSpaces"},
		{"multiple-spaces", "multipleSpaces"},
		{"multiple_spaces", "multipleSpaces"},
		{"multiple-_spaces", "multipleSpaces"},
		{"multiple_-spaces", "multipleSpaces"},
		{"multiple-_-spaces", "multipleSpaces"},
		{"multiple--spaces", "multipleSpaces"},
		{"multiple__spaces", "multipleSpaces"},
		{"multiple__-spaces", "multipleSpaces"},
		{"multiple-__spaces", "multipleSpaces"},
		// Acronym handling
		{"HTMLParser", "htmlParser"},
		{"JSONBody", "jsonBody"},
		{"HTTPRequest", "httpRequest"},
		{"SimpleXMLParser", "simpleXmlParser"},
		{"param_id", "paramId"},
		{"param_ID", "paramId"},
		{"ServeHTTP", "serveHttp"},
	}

	for _, test := range tests {
		result := ToCamelCase(test.input)
		assert.Equal(t, test.expected, result, "Input: %s", test.input)
	}
}

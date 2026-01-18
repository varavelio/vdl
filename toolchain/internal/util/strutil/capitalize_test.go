package strutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCapitalize(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello", "Hello"},
		{"world", "World"},
		{"HeLLO", "HeLLO"},
		{"HELLO", "HELLO"},
		{"hello world", "Hello world"},
		{"", ""},
		{"123", "123"},
		{"123abc", "123abc"},
		{"123abc123", "123abc123"},
		{"!hello", "!hello"},
		{"@world", "@world"},
		{"#123", "#123"},
		{"$hello world", "$hello world"},
		{"%123abc", "%123abc"},
		{"^123abc123", "^123abc123"},
		{"&", "&"},
		{"*hello", "*hello"},
		{"(world)", "(world)"},
		{")hello world", ")hello world"},
		{"_123", "_123"},
		{"+123abc", "+123abc"},
		{"=123abc123", "=123abc123"},
		{"-hello", "-hello"},
		{"{world}", "{world}"},
		{"}hello world", "}hello world"},
		{"[123", "[123"},
		{"]123abc", "]123abc"},
		{"|123abc123", "|123abc123"},
		{":hello", ":hello"},
		{";world", ";world"},
		{"'123", "'123"},
		{"\"hello world\"", "\"hello world\""},
		{"<123abc", "<123abc"},
		{">123abc123", ">123abc123"},
		{",hello", ",hello"},
		{".world", ".world"},
		{"?123", "?123"},
		{"/hello world", "/hello world"},
		{"\\123abc", "\\123abc"},
		{"`123abc123", "`123abc123"},
		{"~hello", "~hello"},
		{"hello\tworld", "Hello\tworld"},
		{"hello\nworld", "Hello\nworld"},
		{"hello\rworld", "Hello\rworld"},
		{"hello\fworld", "Hello\fworld"},
		{"hello\vworld", "Hello\vworld"},
		{"hello\bworld", "Hello\bworld"},
		{"hello\aworld", "Hello\aworld"},
	}

	for _, test := range tests {
		result := Capitalize(test.input)
		assert.Equal(t, test.expected, result)
	}
}

func TestCapitalizeStrict(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello", "Hello"},
		{"world", "World"},
		{"HeLLO", "Hello"},
		{"HELLO", "Hello"},
		{"hello world", "Hello world"},
		{"Hello World", "Hello world"},
		{"", ""},
		{"123", "123"},
		{"123abc", "123abc"},
		{"123abc123", "123abc123"},
		{"!hello", "!hello"},
		{"@world", "@world"},
		{"#123", "#123"},
		{"$hello world", "$hello world"},
		{"%123abc", "%123abc"},
		{"^123abc123", "^123abc123"},
		{"&", "&"},
		{"*hello", "*hello"},
		{"(world)", "(world)"},
		{")hello world", ")hello world"},
		{"_123", "_123"},
		{"+123abc", "+123abc"},
		{"=123abc123", "=123abc123"},
		{"-hello", "-hello"},
		{"{world}", "{world}"},
		{")hello world", ")hello world"},
		{"[123", "[123"},
		{"]123abc", "]123abc"},
		{"|123abc123", "|123abc123"},
		{":hello", ":hello"},
		{";world", ";world"},
		{"'123", "'123"},
		{"\"hello world\"", "\"hello world\""},
		{"<123abc", "<123abc"},
		{">123abc123", ">123abc123"},
		{",hello", ",hello"},
		{".world", ".world"},
		{"?123", "?123"},
		{"/hello world", "/hello world"},
		{"\\123abc", "\\123abc"},
		{"`123abc123", "`123abc123"},
		{"~hello", "~hello"},
		{"hello\tworld", "Hello\tworld"},
		{"hello\nworld", "Hello\nworld"},
		{"hello\rworld", "Hello\rworld"},
		{"hello\fworld", "Hello\fworld"},
		{"hello\vworld", "Hello\vworld"},
		{"hello\bworld", "Hello\bworld"},
		{"hello\aworld", "Hello\aworld"},
	}

	for _, test := range tests {
		result := CapitalizeStrict(test.input)
		assert.Equal(t, test.expected, result)
	}
}

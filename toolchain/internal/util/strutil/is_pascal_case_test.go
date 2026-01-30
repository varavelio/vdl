package strutil

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsPascalCase(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"empty string", "", false},
		{"single lowercase letter", "a", false},
		{"single uppercase letter", "A", true},
		{"mixed case", "Abc", true},
		{"multiple words", "PascalCase", true},
		{"multiple words with numbers", "PascalCase123", true},
		{"leading underscore", "_PascalCase", false},
		{"trailing underscore", "PascalCase_", false},
		{"underscore in middle", "Pascal_Case", false},
		{"camel case", "camelCase", false},
		{"snake case", "snake_case", false},
		{"kebab case", "kebab-case", false},
		{"with spaces", "With Some Spaces", false},
		{"with special characters", "With$Special$Chars", false},
		{"with numbers", "With123Numbers", true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := IsPascalCase(test.input)
			require.Equal(t, test.expected, result)
		})
	}
}

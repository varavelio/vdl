package strutil

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsCamelCase(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"empty string", "", false},
		{"single lowercase letter", "a", true},
		{"single uppercase letter", "A", false},
		{"mixed case", "camelCase", true},
		{"multiple words", "camelCaseWord", true},
		{"multiple words with numbers", "camelCase123", true},
		{"leading underscore", "_camelCase", false},
		{"trailing underscore", "camelCase_", false},
		{"underscore in middle", "camel_Case", false},
		{"kebab case", "kebab-case", false},
		{"with spaces", "with some spaces", false},
		{"with special characters", "with$special$chars", false},
		{"with numbers", "with123numbers", true},
		{"PascalCase", "PascalCase", false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := IsCamelCase(test.input)
			require.Equal(t, test.expected, result)
		})
	}
}

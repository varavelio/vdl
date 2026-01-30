package strutil

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetStrBefore(t *testing.T) {
	tests := []struct {
		input       string
		delimiter   string
		expected    string
		description string
	}{
		{"hello-world", "-", "hello", "Delimiter found in the middle"},
		{"hello-world", "!", "", "Delimiter not found"},
		{"-world", "-", "", "Delimiter at the beginning"},
		{"hello-", "-", "hello", "Delimiter at the end"},
		{"hello--world", "-", "hello", "Multiple delimiters"},
		{"", "-", "", "Empty input string"},
		{"hello-world", "", "hello-world", "Empty delimiter"},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			result := GetStrBefore(test.input, test.delimiter)
			require.Equal(t, test.expected, result, fmt.Sprintf("failed for %s", test.description))
		})
	}
}

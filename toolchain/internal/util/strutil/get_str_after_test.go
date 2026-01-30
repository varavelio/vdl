package strutil

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetStrAfter(t *testing.T) {
	tests := []struct {
		input       string
		delimiter   string
		expected    string
		description string
	}{
		{"hello-world", "-", "world", "Delimiter found in the middle"},
		{"hello-world", "!", "", "Delimiter not found"},
		{"-world", "-", "world", "Delimiter at the beginning"},
		{"hello-", "-", "", "Delimiter at the end"},
		{"hello--world", "-", "-world", "Multiple delimiters"},
		{"", "-", "", "Empty input string"},
		{"hello-world", "", "hello-world", "Empty delimiter"},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			result := GetStrAfter(test.input, test.delimiter)
			assert.Equal(t, test.expected, result, fmt.Sprintf("failed for %s", test.description))
		})
	}
}

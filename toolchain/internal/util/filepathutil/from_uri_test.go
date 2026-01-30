package filepathutil

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFromURI(t *testing.T) {
	t.Run("Regular file paths", func(t *testing.T) {
		testCases := []struct {
			name     string
			input    string
			expected string
		}{
			{"Absolute path", "/path/to/file.txt", "/path/to/file.txt"},
			{"Path with redundant slashes", "/path//to/file.txt", "/path/to/file.txt"},
			{"Path with dot segments", "/path/./to/file.txt", "/path/to/file.txt"},
			{"Path with parent segments", "/path/to/../to/file.txt", "/path/to/file.txt"},
			{"Relative path", "path/to/file.txt", "path/to/file.txt"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				result := FromURI(tc.input)
				require.Equal(t, tc.expected, result)
			})
		}
	})

	t.Run("File URIs", func(t *testing.T) {
		testCases := []struct {
			name     string
			input    string
			expected string
		}{
			{"Simple file URI", "file:///path/to/file.txt", "/path/to/file.txt"},
			{"File URI with redundant slashes", "file:///path//to/file.txt", "/path/to/file.txt"},
			{"File URI with dot segments", "file:///path/./to/file.txt", "/path/to/file.txt"},
			{"File URI with parent segments", "file:///path/to/../to/file.txt", "/path/to/file.txt"},
			{"File URI with query parameters", "file:///path/to/file.txt?query=value", "/path/to/file.txt"},
			{"File URI with fragment", "file:///path/to/file.txt#fragment", "/path/to/file.txt"},
			{"File URI with encoded characters", "file:///path/to/file%20with spaces.txt", "/path/to/file with spaces.txt"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				result := FromURI(tc.input)
				require.Equal(t, tc.expected, result)
			})
		}
	})

	t.Run("Windows paths", func(t *testing.T) {
		// These tests will run on all platforms but simulate Windows paths
		testCases := []struct {
			name     string
			input    string
			expected string
		}{
			{"Windows drive path", "C:/path/to/file.txt", "C:/path/to/file.txt"},
			{"Windows drive path with backslashes", "C:\\path\\to\\file.txt", "C:\\path\\to\\file.txt"},
			{"Windows UNC path", "\\\\server\\share\\path\\file.txt", "\\\\server\\share\\path\\file.txt"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				result := FromURI(tc.input)
				require.Equal(t, tc.expected, result)
				require.NotEmpty(t, result)
			})
		}
	})

	t.Run("Windows file URIs", func(t *testing.T) {
		// These tests will run on all platforms but simulate Windows file URIs
		testCases := []struct {
			name     string
			input    string
			expected string
		}{
			{"Windows drive file URI", "file:///C:/path/to/file.txt", "C:/path/to/file.txt"},
			{"Windows drive file URI with spaces", "file:///C:/path/to/file%20with spaces.txt", "C:/path/to/file with spaces.txt"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				result := FromURI(tc.input)
				require.Equal(t, tc.expected, result)
				require.NotEmpty(t, result)
			})
		}
	})

	t.Run("Invalid URIs", func(t *testing.T) {
		testCases := []struct {
			name  string
			input string
		}{
			{"Invalid URI scheme", "http://example.com/file.txt"},
			{"Malformed URI", "file:/path/to/file.txt"},
			{"Empty URI", "file://"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// These should not panic, but return something
				result := FromURI(tc.input)
				require.NotPanics(t, func() { FromURI(tc.input) })
				require.NotEmpty(t, result)
			})
		}
	})
}

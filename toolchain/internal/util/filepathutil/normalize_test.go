package filepathutil

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNormalize(t *testing.T) {
	t.Run("Basic path normalization", func(t *testing.T) {
		testCases := []struct {
			name               string
			relativeToFilePath string
			filePath           string
			expectedPath       string
		}{
			{"Absolute path", "", "/path/to/file.txt", "/path/to/file.txt"},
			{"Path with redundant slashes", "", "/path//to/file.txt", "/path/to/file.txt"},
			{"Path with dot segments", "", "/path/./to/file.txt", "/path/to/file.txt"},
			{"Path with parent segments", "", "/path/to/../to/file.txt", "/path/to/file.txt"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				result, err := Normalize(tc.relativeToFilePath, tc.filePath)
				require.NoError(t, err)
				require.Equal(t, tc.expectedPath, result)
			})
		}
	})

	t.Run("Absolute paths with base path", func(t *testing.T) {
		testCases := []struct {
			name               string
			relativeToFilePath string
			filePath           string
			expectedPath       string
		}{
			{"Absolute path with empty base", "", "/absolute/path.txt", "/absolute/path.txt"},
			{"Absolute path with base", "/base/dir/file.txt", "/absolute/path.txt", "/absolute/path.txt"},
			{"Absolute path with base and redundant slashes", "/base/dir/file.txt", "/absolute//path.txt", "/absolute/path.txt"},
			{"Absolute path with base and dot segments", "/base/dir/file.txt", "/absolute/./path.txt", "/absolute/path.txt"},
			{"Absolute path with base and parent segments", "/base/dir/file.txt", "/absolute/../absolute/path.txt", "/absolute/path.txt"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				result, err := Normalize(tc.relativeToFilePath, tc.filePath)
				require.NoError(t, err)
				require.Equal(t, tc.expectedPath, result)
			})
		}
	})

	t.Run("Relative paths with base", func(t *testing.T) {
		testCases := []struct {
			name               string
			relativeToFilePath string
			filePath           string
			expectedPath       string
		}{
			{"Simple relative path", "/base/dir/file.txt", "other.txt", "/base/dir/other.txt"},
			{"Parent directory", "/base/dir/file.txt", "../other.txt", "/base/other.txt"},
			{"Multiple parent directories", "/base/dir/subdir/file.txt", "../../other.txt", "/base/other.txt"},
			{"Relative path with base", "/base/dir/file.txt", "relative/path.txt", "/base/dir/relative/path.txt"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				result, err := Normalize(tc.relativeToFilePath, tc.filePath)
				require.NoError(t, err)
				require.Equal(t, tc.expectedPath, result)
			})
		}
	})

	t.Run("URI handling", func(t *testing.T) {
		testCases := []struct {
			name               string
			relativeToFilePath string
			filePath           string
			expectedPath       string
		}{
			{"File URI", "", "file:///path/to/file.txt", "/path/to/file.txt"},
			{"File URI with base", "/base/dir/file.txt", "file:///path/to/file.txt", "/path/to/file.txt"},
			{"Base as file URI", "file:///base/dir/file.txt", "other.txt", "/base/dir/other.txt"},
			{"Both as file URIs", "file:///base/dir/file.txt", "file:///path/to/file.txt", "/path/to/file.txt"},
			{"Relative path with base as URI", "file:///base/dir/file.txt", "../other.txt", "/base/other.txt"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				result, err := Normalize(tc.relativeToFilePath, tc.filePath)
				require.NoError(t, err)
				require.Equal(t, tc.expectedPath, result)
			})
		}
	})

	t.Run("Error cases", func(t *testing.T) {
		testCases := []struct {
			name               string
			relativeToFilePath string
			filePath           string
			errorPattern       string
		}{
			{"Non-absolute relativeToFilePath", "relative/path", "file.txt", "relativeToFilePath must be an absolute path"},
			{"Non-absolute result path", "", "relative/path", "file path must be an absolute path"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				_, err := Normalize(tc.relativeToFilePath, tc.filePath)
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errorPattern)
			})
		}
	})
}

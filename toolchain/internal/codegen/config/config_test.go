package config

import (
	"embed"
	"io/fs"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

//go:embed testdata/*.yaml
var testdataFS embed.FS

func TestConfigSchema_AllTestFiles(t *testing.T) {
	entries, err := fs.ReadDir(testdataFS, "testdata")
	require.NoError(t, err, "failed to read testdata directory")

	var validCount, invalidCount int

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
			continue
		}

		shouldBeValid := strings.HasPrefix(entry.Name(), "valid_")
		shouldBeInvalid := strings.HasPrefix(entry.Name(), "invalid_")

		if !shouldBeValid && !shouldBeInvalid {
			t.Logf("Skipping file with unexpected prefix: %s", entry.Name())
			continue
		}

		if shouldBeValid {
			validCount++
		} else {
			invalidCount++
		}

		testName := friendlyTestName(entry.Name())
		t.Run(testName, func(t *testing.T) {
			data, err := testdataFS.ReadFile("testdata/" + entry.Name())
			require.NoError(t, err, "failed to read test file")

			_, err = ParseConfig(data)

			if shouldBeValid {
				require.NoError(t, err, "expected valid config but got error")
			} else {
				require.Error(t, err, "expected invalid config to fail validation")
			}
		})
	}

	// Ensure we found test files
	require.Greater(t, validCount, 0, "no valid_*.yaml test files found")
	require.Greater(t, invalidCount, 0, "no invalid_*.yaml test files found")
}

func TestTargetConstants(t *testing.T) {
	t.Run("constants have expected values", func(t *testing.T) {
		require.Equal(t, "go", TargetGo)
		require.Equal(t, "typescript", TargetTypeScript)
		require.Equal(t, "dart", TargetDart)
		require.Equal(t, "openapi", TargetOpenAPI)
		require.Equal(t, "playground", TargetPlayground)
	})
}

func friendlyTestName(filename string) string {
	// Remove extension
	name := strings.TrimSuffix(filename, ".yaml")

	// Remove prefix
	name = strings.TrimPrefix(name, "valid_")
	name = strings.TrimPrefix(name, "invalid_")

	// Replace underscores with spaces
	name = strings.ReplaceAll(name, "_", " ")

	return name
}

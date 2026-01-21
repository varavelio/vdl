package config

import (
	"embed"
	"encoding/json"
	"io/fs"
	"strings"
	"testing"

	"github.com/kaptinlin/jsonschema"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

//go:embed config.schema.json
var schemaJSON []byte

//go:embed testdata/*.yaml
var testdataFS embed.FS

func TestConfigSchema_AllTestFiles(t *testing.T) {
	compiler := jsonschema.NewCompiler()
	schema, err := compiler.Compile(schemaJSON)
	require.NoError(t, err, "failed to compile schema")

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
			data := loadTestdataAsJSON(t, entry.Name())
			result := schema.Validate(data)

			if shouldBeValid {
				require.True(t, result.IsValid(), "expected valid but got errors: %v", formatErrors(result))
			} else {
				require.False(t, result.IsValid(), "expected invalid config to fail validation")
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

func loadTestdataAsJSON(t *testing.T, filename string) any {
	t.Helper()

	yamlData, err := testdataFS.ReadFile("testdata/" + filename)
	require.NoError(t, err, "failed to read testdata file: %s", filename)

	var data any
	err = yaml.Unmarshal(yamlData, &data)
	require.NoError(t, err, "failed to parse YAML: %s", filename)

	// Convert to JSON and back to ensure proper type handling for the validator.
	jsonData, err := json.Marshal(data)
	require.NoError(t, err, "failed to convert to JSON: %s", filename)

	var result any
	err = json.Unmarshal(jsonData, &result)
	require.NoError(t, err, "failed to parse JSON: %s", filename)

	return result
}

func formatErrors(result *jsonschema.EvaluationResult) string {
	if result.IsValid() {
		return ""
	}

	var parts []string
	for path, err := range result.Errors {
		parts = append(parts, path+": "+err.Message)
	}
	return strings.Join(parts, "; ")
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

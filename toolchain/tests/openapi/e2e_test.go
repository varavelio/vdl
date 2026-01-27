package openapi_test

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

var (
	vdlBinaryPath string
	toolchainRoot string
	update        = flag.Bool("update", false, "update golden files")
)

func TestMain(m *testing.M) {
	flag.Parse()

	// Determine Toolchain Root
	_, filename, _, _ := runtime.Caller(0)
	toolchainRoot = filepath.Join(filepath.Dir(filename), "..", "..")
	toolchainRoot, _ = filepath.Abs(toolchainRoot)

	// Build VDL Binary
	if err := buildVDLBinary(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to build VDL binary: %v\n", err)
		os.Exit(1)
	}

	// Run Tests
	exitCode := m.Run()

	// Cleanup
	os.Remove(vdlBinaryPath)
	os.Exit(exitCode)
}

func buildVDLBinary() error {
	vdlBinaryPath = filepath.Join(os.TempDir(), "vdl-openapi-e2e-test")
	if runtime.GOOS == "windows" {
		vdlBinaryPath += ".exe"
	}

	cmd := exec.Command("go", "build", "-o", vdlBinaryPath, "./cmd/vdl")
	cmd.Dir = toolchainRoot
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func TestOpenAPI(t *testing.T) {
	wd, err := os.Getwd()
	require.NoError(t, err)
	testDataDir := filepath.Join(wd, "testdata")

	entries, err := os.ReadDir(testDataDir)
	require.NoError(t, err)

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		caseName := entry.Name()
		t.Run(caseName, func(t *testing.T) {
			t.Parallel()
			runTestCase(t, filepath.Join(testDataDir, caseName))
		})
	}
}

func runTestCase(t *testing.T, caseDir string) {
	// Clean output directory to ensure fresh generation
	os.RemoveAll(filepath.Join(caseDir, "gen"))

	// Run VDL Generate
	cmdGen := exec.Command(vdlBinaryPath, "generate")
	cmdGen.Dir = caseDir
	outGen, err := cmdGen.CombinedOutput()
	require.NoError(t, err, "vdl generate failed:\n%s", string(outGen))

	// Find generated OpenAPI file (yaml or json)
	generatedPath := findGeneratedFile(t, caseDir)
	require.NotEmpty(t, generatedPath, "no generated OpenAPI file found")

	// Find expected file
	expectedPath := findExpectedFile(t, caseDir)
	require.NotEmpty(t, expectedPath, "no expected file found")

	// Read generated content
	generatedBytes, err := os.ReadFile(generatedPath)
	require.NoError(t, err, "failed to read generated file")

	// Update mode: overwrite expected with generated
	if *update {
		err := os.WriteFile(expectedPath, generatedBytes, 0644)
		require.NoError(t, err, "failed to update expected file")
		t.Logf("updated expected file: %s", expectedPath)
		return
	}

	// Read expected content
	expectedBytes, err := os.ReadFile(expectedPath)
	require.NoError(t, err, "failed to read expected file: %s (run with -update to create)", expectedPath)

	// Parse and compare structurally
	generatedObj := parseContent(t, generatedPath, generatedBytes)
	expectedObj := parseContent(t, expectedPath, expectedBytes)

	assert.Equal(t, expectedObj, generatedObj, "OpenAPI spec mismatch")
}

// findGeneratedFile locates the generated OpenAPI file in the gen directory.
func findGeneratedFile(t *testing.T, caseDir string) string {
	t.Helper()

	genDir := filepath.Join(caseDir, "gen")
	patterns := []string{"*.yaml", "*.yml", "*.json"}

	for _, pattern := range patterns {
		matches, err := filepath.Glob(filepath.Join(genDir, pattern))
		if err != nil {
			continue
		}
		if len(matches) > 0 {
			return matches[0]
		}
	}

	return ""
}

// findExpectedFile locates the expected golden file in the test case directory.
func findExpectedFile(t *testing.T, caseDir string) string {
	t.Helper()

	candidates := []string{"expected.yaml", "expected.yml", "expected.json"}
	for _, name := range candidates {
		path := filepath.Join(caseDir, name)
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	// Fallback: create expected.yaml path for update mode
	return filepath.Join(caseDir, "expected.yaml")
}

// parseContent parses YAML or JSON content based on file extension.
func parseContent(t *testing.T, path string, content []byte) any {
	t.Helper()

	var result any
	var err error

	if strings.HasSuffix(path, ".json") {
		err = json.Unmarshal(content, &result)
	} else {
		err = yaml.Unmarshal(content, &result)
	}

	require.NoError(t, err, "failed to parse %s", path)
	return result
}

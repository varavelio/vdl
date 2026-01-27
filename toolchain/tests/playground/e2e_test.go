// Package playground_test provides end-to-end tests for the VDL playground generator.
//
// The playground generator outputs static files for a web-based VDL playground,
// including the formatted schema and optional configuration.
//
// Test cases verify:
// - Static playground files are generated
// - schema.vdl contains the formatted schema
// - config.json is generated when baseUrl or headers are configured
// - Multi-file schemas with includes work correctly
package playground_test

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	vdlBinaryPath string
	toolchainRoot string
	testdataDir   string
)

func TestMain(m *testing.M) {
	// Determine paths
	_, filename, _, _ := runtime.Caller(0)
	testdataDir = filepath.Join(filepath.Dir(filename), "testdata")
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
	vdlBinaryPath = filepath.Join(os.TempDir(), "vdl-playground-e2e-test")
	if runtime.GOOS == "windows" {
		vdlBinaryPath += ".exe"
	}

	cmd := exec.Command("go", "build", "-o", vdlBinaryPath, "./cmd/vdl")
	cmd.Dir = toolchainRoot
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// TestPlayground runs all playground test cases from testdata/.
func TestPlayground(t *testing.T) {
	entries, err := os.ReadDir(testdataDir)
	require.NoError(t, err)

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		caseName := entry.Name()
		t.Run(caseName, func(t *testing.T) {
			t.Parallel()
			runTestCase(t, filepath.Join(testdataDir, caseName))
		})
	}
}

func runTestCase(t *testing.T, caseDir string) {
	t.Helper()

	// Clean output directory
	genDir := filepath.Join(caseDir, "gen")
	os.RemoveAll(genDir)

	// Run VDL Generate
	cmd := exec.Command(vdlBinaryPath, "generate")
	cmd.Dir = caseDir
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "vdl generate failed:\n%s", string(out))

	// Verify output directory exists
	require.DirExists(t, genDir, "gen directory should exist after generation")

	// Load test expectations from expect.json
	expectPath := filepath.Join(caseDir, "expect.json")
	expectBytes, err := os.ReadFile(expectPath)
	require.NoError(t, err, "expect.json is required for each test case")

	var expect struct {
		// Files that must exist (relative to gen/)
		RequiredFiles []string `json:"requiredFiles"`
		// Files that must NOT exist
		ForbiddenFiles []string `json:"forbiddenFiles"`
		// Content checks: file -> list of substrings that must be present
		ContentContains map[string][]string `json:"contentContains"`
		// Exact content match for specific files
		ContentEquals map[string]string `json:"contentEquals"`
	}
	require.NoError(t, json.Unmarshal(expectBytes, &expect), "failed to parse expect.json")

	// Check required files exist
	for _, file := range expect.RequiredFiles {
		filePath := filepath.Join(genDir, file)
		assert.FileExists(t, filePath, "required file should exist: %s", file)
	}

	// Check forbidden files don't exist
	for _, file := range expect.ForbiddenFiles {
		filePath := filepath.Join(genDir, file)
		assert.NoFileExists(t, filePath, "forbidden file should not exist: %s", file)
	}

	// Check content contains
	for file, substrings := range expect.ContentContains {
		filePath := filepath.Join(genDir, file)
		content, err := os.ReadFile(filePath)
		require.NoError(t, err, "failed to read file for content check: %s", file)

		for _, substr := range substrings {
			assert.Contains(t, string(content), substr,
				"file %s should contain: %s", file, substr)
		}
	}

	// Check exact content matches
	for file, expectedContent := range expect.ContentEquals {
		filePath := filepath.Join(genDir, file)
		content, err := os.ReadFile(filePath)
		require.NoError(t, err, "failed to read file for exact match: %s", file)

		// Normalize line endings for comparison
		actual := strings.ReplaceAll(string(content), "\r\n", "\n")
		expected := strings.ReplaceAll(expectedContent, "\r\n", "\n")
		assert.Equal(t, expected, actual, "file %s content mismatch", file)
	}
}

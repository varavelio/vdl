package jsonschema_test

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	vdlBinaryPath = filepath.Join(os.TempDir(), "vdl-jsonschema-e2e-test")
	if runtime.GOOS == "windows" {
		vdlBinaryPath += ".exe"
	}

	cmd := exec.Command("go", "build", "-o", vdlBinaryPath, "./cmd/vdl")
	cmd.Dir = toolchainRoot
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func TestJSONSchema(t *testing.T) {
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

	// Find generated JSON Schema file
	generatedPath := findGeneratedFile(t, caseDir)
	require.NotEmpty(t, generatedPath, "no generated JSON Schema file found")

	// Expected file path
	expectedPath := filepath.Join(caseDir, "expected.json")

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
	var generatedObj, expectedObj any
	require.NoError(t, json.Unmarshal(generatedBytes, &generatedObj), "failed to parse generated JSON")
	require.NoError(t, json.Unmarshal(expectedBytes, &expectedObj), "failed to parse expected JSON")

	assert.Equal(t, expectedObj, generatedObj, "JSON Schema mismatch")
}

// findGeneratedFile locates the generated JSON Schema file in the gen directory.
func findGeneratedFile(t *testing.T, caseDir string) string {
	t.Helper()

	genDir := filepath.Join(caseDir, "gen")
	matches, err := filepath.Glob(filepath.Join(genDir, "*.json"))
	if err != nil || len(matches) == 0 {
		return ""
	}

	return matches[0]
}

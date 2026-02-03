package vdlschema_test

import (
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
	vdlBinaryPath = filepath.Join(os.TempDir(), "vdl-vdlschema-e2e-test")
	if runtime.GOOS == "windows" {
		vdlBinaryPath += ".exe"
	}

	cmd := exec.Command("go", "build", "-o", vdlBinaryPath, "./cmd/vdl")
	cmd.Dir = toolchainRoot
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func TestVdlSchema(t *testing.T) {
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

	// Find generated VDL schema file
	generatedPath := findGeneratedFile(t, caseDir)
	require.NotEmpty(t, generatedPath, "no generated VDL schema file found")

	// Expected file path
	expectedPath := filepath.Join(caseDir, "expected.vdl")

	// Read generated content
	generatedBytes, err := os.ReadFile(generatedPath)
	require.NoError(t, err, "failed to read generated file")

	// Normalize line endings for comparison
	generatedContent := normalizeLineEndings(string(generatedBytes))

	// Update mode: overwrite expected with generated
	if *update {
		err := os.WriteFile(expectedPath, []byte(generatedContent), 0644)
		require.NoError(t, err, "failed to update expected file")
		t.Logf("updated expected file: %s", expectedPath)
		return
	}

	// Read expected content
	expectedBytes, err := os.ReadFile(expectedPath)
	require.NoError(t, err, "failed to read expected file: %s (run with -update to create)", expectedPath)

	expectedContent := normalizeLineEndings(string(expectedBytes))

	assert.Equal(t, expectedContent, generatedContent, "VDL Schema mismatch")
}

// findGeneratedFile locates the generated VDL schema file in the gen directory.
func findGeneratedFile(t *testing.T, caseDir string) string {
	t.Helper()

	genDir := filepath.Join(caseDir, "gen")
	matches, err := filepath.Glob(filepath.Join(genDir, "*.vdl"))
	if err != nil || len(matches) == 0 {
		return ""
	}

	return matches[0]
}

// normalizeLineEndings converts all line endings to \n
func normalizeLineEndings(s string) string {
	return strings.ReplaceAll(strings.ReplaceAll(s, "\r\n", "\n"), "\r", "\n")
}

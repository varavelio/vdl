// Package plugin_test provides end-to-end tests for the VDL plugin system.
//
// Test Discovery:
// Tests are automatically discovered from the testdata/ directory based on folder prefixes:
//   - success_*  : Plugin should succeed, generated files are verified against expected/
//   - error_*    : Plugin should fail, error message is verified against expected_error.txt
//   - verify_*   : Plugin should succeed, then verify.py is executed for custom verification
//
// Test Structure:
// Each test case folder should contain:
//   - schema.vdl    : The VDL schema
//   - vdl.yaml      : The VDL configuration
//   - plugin.py     : The plugin script
//   - expected/     : (optional) Expected output files for comparison
//   - expected_error.txt : (error_ only) Expected error message substring
//   - verify.py     : (verify_ only) Custom verification script (exit 0 = pass)
package plugin_test

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
)

const (
	prefixSuccess = "success_"
	prefixError   = "error_"
	prefixVerify  = "verify_"
)

var (
	vdlBinaryPath string
	toolchainRoot string
	testdataDir   string
	update        = flag.Bool("update", false, "update golden files")
)

func TestMain(m *testing.M) {
	flag.Parse()

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
	vdlBinaryPath = filepath.Join(os.TempDir(), "vdl-plugin-e2e-test")
	if runtime.GOOS == "windows" {
		vdlBinaryPath += ".exe"
	}

	cmd := exec.Command("go", "build", "-o", vdlBinaryPath, "./cmd/vdl")
	cmd.Dir = toolchainRoot
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// discoverTestCases finds all test case directories with the given prefix.
func discoverTestCases(prefix string) ([]string, error) {
	entries, err := os.ReadDir(testdataDir)
	if err != nil {
		return nil, err
	}

	var cases []string
	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), prefix) {
			cases = append(cases, entry.Name())
		}
	}
	return cases, nil
}

// TestPluginSuccess tests all success_* test cases.
// Plugin should succeed and generated files are verified against expected/.
func TestPluginSuccess(t *testing.T) {
	cases, err := discoverTestCases(prefixSuccess)
	require.NoError(t, err, "failed to discover success test cases")
	require.NotEmpty(t, cases, "no success_* test cases found")

	for _, tc := range cases {
		t.Run(strings.TrimPrefix(tc, prefixSuccess), func(t *testing.T) {
			t.Parallel()
			runSuccessTestCase(t, tc)
		})
	}
}

// TestPluginErrors tests all error_* test cases.
// Plugin should fail and error message should contain expected_error.txt content.
func TestPluginErrors(t *testing.T) {
	cases, err := discoverTestCases(prefixError)
	require.NoError(t, err, "failed to discover error test cases")
	require.NotEmpty(t, cases, "no error_* test cases found")

	for _, tc := range cases {
		t.Run(strings.TrimPrefix(tc, prefixError), func(t *testing.T) {
			t.Parallel()
			runErrorTestCase(t, tc)
		})
	}
}

// TestPluginVerify tests all verify_* test cases.
// Plugin should succeed, then verify.py is executed for custom verification.
func TestPluginVerify(t *testing.T) {
	cases, err := discoverTestCases(prefixVerify)
	require.NoError(t, err, "failed to discover verify test cases")
	require.NotEmpty(t, cases, "no verify_* test cases found")

	for _, tc := range cases {
		t.Run(strings.TrimPrefix(tc, prefixVerify), func(t *testing.T) {
			t.Parallel()
			runVerifyTestCase(t, tc)
		})
	}
}

// runSuccessTestCase runs a success_* test case.
func runSuccessTestCase(t *testing.T, name string) {
	t.Helper()
	caseDir := filepath.Join(testdataDir, name)

	// Clean and run
	cleanGenDir(caseDir)
	makeScriptsExecutable(t, caseDir)

	cmd := exec.Command(vdlBinaryPath, "generate")
	cmd.Dir = caseDir
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "vdl generate failed:\n%s", string(out))

	// Verify generated files
	verifyGeneratedFiles(t, caseDir)
}

// runErrorTestCase runs an error_* test case.
func runErrorTestCase(t *testing.T, name string) {
	t.Helper()
	caseDir := filepath.Join(testdataDir, name)

	// Read expected error
	expectedErrorFile := filepath.Join(caseDir, "expected_error.txt")
	expectedErrorBytes, err := os.ReadFile(expectedErrorFile)
	require.NoError(t, err, "missing expected_error.txt in %s", name)
	expectedError := strings.TrimSpace(string(expectedErrorBytes))

	// Clean and run
	cleanGenDir(caseDir)
	makeScriptsExecutable(t, caseDir)

	cmd := exec.Command(vdlBinaryPath, "generate")
	cmd.Dir = caseDir
	out, err := cmd.CombinedOutput()
	require.Error(t, err, "expected vdl generate to fail")
	assert.Contains(t, string(out), expectedError, "expected error not found in output:\n%s", string(out))
}

// runVerifyTestCase runs a verify_* test case.
func runVerifyTestCase(t *testing.T, name string) {
	t.Helper()
	caseDir := filepath.Join(testdataDir, name)

	// Clean and run VDL generate
	cleanGenDir(caseDir)
	makeScriptsExecutable(t, caseDir)

	cmd := exec.Command(vdlBinaryPath, "generate")
	cmd.Dir = caseDir
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "vdl generate failed:\n%s", string(out))

	// Run verify.py script
	verifyScript := filepath.Join(caseDir, "verify.py")
	require.FileExists(t, verifyScript, "verify_* test case must have verify.py")

	verifyCmd := exec.Command("python3", verifyScript)
	verifyCmd.Dir = caseDir
	verifyOut, verifyErr := verifyCmd.CombinedOutput()
	require.NoError(t, verifyErr, "verification failed:\n%s", string(verifyOut))
	t.Logf("Verification output:\n%s", string(verifyOut))
}

// cleanGenDir removes the gen/ directory in a test case.
func cleanGenDir(caseDir string) {
	os.RemoveAll(filepath.Join(caseDir, "gen"))
}

// makeScriptsExecutable makes plugin and verify scripts executable.
func makeScriptsExecutable(t *testing.T, caseDir string) {
	t.Helper()
	scripts := []string{"plugin.py", "plugin.sh", "plugin", "verify.py"}
	for _, script := range scripts {
		path := filepath.Join(caseDir, script)
		if _, err := os.Stat(path); err == nil {
			if err := os.Chmod(path, 0o755); err != nil {
				t.Logf("warning: could not make %s executable: %v", script, err)
			}
		}
	}
}

// verifyGeneratedFiles compares generated files against expected files.
func verifyGeneratedFiles(t *testing.T, caseDir string) {
	t.Helper()

	expectedDir := filepath.Join(caseDir, "expected")
	genDir := filepath.Join(caseDir, "gen")

	// If expected directory exists, compare file-by-file
	if info, err := os.Stat(expectedDir); err == nil && info.IsDir() {
		compareDirectories(t, expectedDir, genDir)
		return
	}

	// If expected.json exists, compare the single output file
	expectedFile := filepath.Join(caseDir, "expected.json")
	if _, err := os.Stat(expectedFile); err == nil {
		genFiles, err := os.ReadDir(genDir)
		require.NoError(t, err, "failed to read gen directory")
		require.Len(t, genFiles, 1, "expected exactly one generated file")

		genPath := filepath.Join(genDir, genFiles[0].Name())
		compareJSONFiles(t, expectedFile, genPath)
		return
	}

	// If no expected files, just verify gen directory has content
	genFiles, err := os.ReadDir(genDir)
	require.NoError(t, err, "failed to read gen directory")
	require.NotEmpty(t, genFiles, "expected at least one generated file")
}

// compareDirectories compares files in expectedDir against genDir.
func compareDirectories(t *testing.T, expectedDir, genDir string) {
	t.Helper()

	err := filepath.Walk(expectedDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(expectedDir, path)
		if err != nil {
			return err
		}

		genPath := filepath.Join(genDir, relPath)

		// Update mode: copy generated to expected
		if *update {
			genContent, err := os.ReadFile(genPath)
			if err != nil {
				return fmt.Errorf("failed to read generated file %s: %w", genPath, err)
			}
			if err := os.WriteFile(path, genContent, 0o644); err != nil {
				return fmt.Errorf("failed to update expected file %s: %w", path, err)
			}
			t.Logf("updated expected file: %s", path)
			return nil
		}

		expectedContent, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read expected file: %w", err)
		}

		genContent, err := os.ReadFile(genPath)
		if err != nil {
			return fmt.Errorf("failed to read generated file %s: %w (run with -update to create)", genPath, err)
		}

		// For JSON files, compare structurally
		if strings.HasSuffix(path, ".json") {
			var expectedObj, genObj any
			require.NoError(t, json.Unmarshal(expectedContent, &expectedObj), "failed to parse expected JSON")
			require.NoError(t, json.Unmarshal(genContent, &genObj), "failed to parse generated JSON")
			assert.Equal(t, expectedObj, genObj, "mismatch in %s", relPath)
		} else {
			assert.Equal(t, string(expectedContent), string(genContent), "mismatch in %s", relPath)
		}

		return nil
	})
	require.NoError(t, err)
}

// compareJSONFiles compares two JSON files.
func compareJSONFiles(t *testing.T, expectedPath, genPath string) {
	t.Helper()

	if *update {
		genContent, err := os.ReadFile(genPath)
		require.NoError(t, err, "failed to read generated file")
		err = os.WriteFile(expectedPath, genContent, 0o644)
		require.NoError(t, err, "failed to update expected file")
		t.Logf("updated expected file: %s", expectedPath)
		return
	}

	expectedContent, err := os.ReadFile(expectedPath)
	require.NoError(t, err, "failed to read expected file")

	genContent, err := os.ReadFile(genPath)
	require.NoError(t, err, "failed to read generated file")

	var expectedObj, genObj any
	require.NoError(t, json.Unmarshal(expectedContent, &expectedObj), "failed to parse expected JSON")
	require.NoError(t, json.Unmarshal(genContent, &genObj), "failed to parse generated JSON")
	assert.Equal(t, expectedObj, genObj, "JSON content mismatch")
}

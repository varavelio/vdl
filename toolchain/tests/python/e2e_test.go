package python_test

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var (
	vdlBinaryPath   string
	toolchainRoot   string
	pythonTestsRoot string
)

func TestMain(m *testing.M) {
	// Determine paths
	_, filename, _, _ := runtime.Caller(0)
	pythonTestsRoot = filepath.Dir(filename)
	toolchainRoot = filepath.Join(pythonTestsRoot, "..", "..")
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
	vdlBinaryPath = filepath.Join(os.TempDir(), "vdl-python-e2e")
	if runtime.GOOS == "windows" {
		vdlBinaryPath += ".exe"
	}

	cmd := exec.Command("go", "build", "-o", vdlBinaryPath, "./cmd/vdl")
	cmd.Dir = toolchainRoot
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func TestPython(t *testing.T) {
	testDataDir := filepath.Join(pythonTestsRoot, "testdata")

	// Check if testdata exists
	if _, err := os.Stat(testDataDir); os.IsNotExist(err) {
		t.Skip("testdata directory does not exist")
	}

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
	// Run VDL Generate
	cmdGen := exec.Command(vdlBinaryPath, "generate")
	cmdGen.Dir = caseDir
	outGen, err := cmdGen.CombinedOutput()
	require.NoError(t, err, "vdl generate failed:\n%s", string(outGen))

	// Check if main.py exists
	mainPy := filepath.Join(caseDir, "main.py")
	if _, err := os.Stat(mainPy); os.IsNotExist(err) {
		// If no main.py, just verify generation succeeded
		return
	}

	// Run Python verification
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmdRun := exec.CommandContext(ctx, "python3", "main.py")
	cmdRun.Dir = caseDir
	outRun, err := cmdRun.CombinedOutput()
	if err != nil {
		t.Fatalf("python3 main.py failed:\nOutput:\n%s\nError: %v", string(outRun), err)
	}
}

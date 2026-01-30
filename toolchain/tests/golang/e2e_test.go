package golang_test

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
	vdlBinaryPath string
	toolchainRoot string
)

func TestMain(m *testing.M) {
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
	vdlBinaryPath = filepath.Join(os.TempDir(), "vdl-e2e-test")
	if runtime.GOOS == "windows" {
		vdlBinaryPath += ".exe"
	}

	cmd := exec.Command("go", "build", "-o", vdlBinaryPath, "./cmd/vdl")
	cmd.Dir = toolchainRoot
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func TestGolang(t *testing.T) {
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
	// Clean gen directory to ensure fresh generation
	os.RemoveAll(filepath.Join(caseDir, "gen"))

	// Run VDL Generate
	cmdGen := exec.Command(vdlBinaryPath, "generate")
	cmdGen.Dir = caseDir
	outGen, err := cmdGen.CombinedOutput()
	require.NoError(t, err, "vdl generate failed:\n%s", string(outGen))

	// Safety timeout to prevent deadlocks from hanging the test suite indefinitely.
	// 20s should be enough for compilation and execution.
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	cmdRun := exec.CommandContext(ctx, "go", "run", ".")
	cmdRun.Dir = caseDir
	outRun, err := cmdRun.CombinedOutput()
	if err != nil {
		t.Fatalf("go run failed:\nOutput:\n%s\nError: %v", string(outRun), err)
	}
}

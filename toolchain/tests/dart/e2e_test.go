package dart_test

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
	dartTestsRoot string
)

func TestMain(m *testing.M) {
	// Determine paths
	_, filename, _, _ := runtime.Caller(0)
	dartTestsRoot = filepath.Dir(filename)
	toolchainRoot = filepath.Join(dartTestsRoot, "..", "..")
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
	vdlBinaryPath = filepath.Join(os.TempDir(), "vdl-dart-e2e")
	if runtime.GOOS == "windows" {
		vdlBinaryPath += ".exe"
	}

	cmd := exec.Command("go", "build", "-o", vdlBinaryPath, "./cmd/vdl")
	cmd.Dir = toolchainRoot
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func TestDart(t *testing.T) {
	testDataDir := filepath.Join(dartTestsRoot, "testdata")

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
	// Clean gen directory
	os.RemoveAll(filepath.Join(caseDir, "gen"))

	// Run VDL Generate
	cmdGen := exec.Command(vdlBinaryPath, "generate")
	cmdGen.Dir = caseDir
	outGen, err := cmdGen.CombinedOutput()
	require.NoError(t, err, "vdl generate failed:\n%s", string(outGen))

	// Check if main.dart exists
	mainDart := filepath.Join(caseDir, "main.dart")
	if _, err := os.Stat(mainDart); os.IsNotExist(err) {
		// If no main.dart, just verify generation succeeded
		return
	}

	// Run Dart verification
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmdRun := exec.CommandContext(ctx, "dart", "run", "--enable-asserts", "main.dart")
	cmdRun.Dir = caseDir
	outRun, err := cmdRun.CombinedOutput()
	if err != nil {
		t.Fatalf("dart run main.dart failed:\nOutput:\n%s\nError: %v", string(outRun), err)
	}
}

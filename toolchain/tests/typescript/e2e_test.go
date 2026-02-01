package typescript_test

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
	tsTestsRoot   string
)

func TestMain(m *testing.M) {
	// Determine paths
	_, filename, _, _ := runtime.Caller(0)
	tsTestsRoot = filepath.Dir(filename)
	toolchainRoot = filepath.Join(tsTestsRoot, "..", "..")
	toolchainRoot, _ = filepath.Abs(toolchainRoot)

	// Build VDL Binary
	if err := buildVDLBinary(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to build VDL binary: %v\n", err)
		os.Exit(1)
	}

	// Install NPM deps for tests
	if err := installNPMDeps(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to install NPM deps: %v\n", err)
		os.Exit(1)
	}

	// Run Tests
	exitCode := m.Run()

	// Cleanup
	os.Remove(vdlBinaryPath)
	os.Exit(exitCode)
}

func buildVDLBinary() error {
	vdlBinaryPath = filepath.Join(os.TempDir(), "vdl-ts-e2e")
	if runtime.GOOS == "windows" {
		vdlBinaryPath += ".exe"
	}

	cmd := exec.Command("go", "build", "-o", vdlBinaryPath, "./cmd/vdl")
	cmd.Dir = toolchainRoot
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func installNPMDeps() error {
	nodeModulesPath := filepath.Join(tsTestsRoot, "node_modules")
	if _, err := os.Stat(nodeModulesPath); err == nil {
		return nil
	}

	cmd := exec.Command("npm", "install")
	cmd.Dir = tsTestsRoot
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func TestTypeScript(t *testing.T) {
	testDataDir := filepath.Join(tsTestsRoot, "testdata")

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

	// Type-check with tsc before running (use sh -c to expand globs)
	ctxTsc, cancelTsc := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelTsc()

	// Use the absolute path to tsc to avoid npx overhead
	tscCmd := "npx --no tsc ./*.ts ./gen/*.ts --noEmit --moduleResolution node --target es2022 --skipLibCheck --allowImportingTsExtensions"
	cmdTsc := exec.CommandContext(ctxTsc, "sh", "-c", tscCmd)
	cmdTsc.Dir = caseDir
	outTsc, err := cmdTsc.CombinedOutput()
	if err != nil {
		t.Fatalf("tsc type-check failed:\nOutput:\n%s\nError: %v", string(outTsc), err)
	}

	// Run TypeScript verification
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	tsxCmd := "npx --no tsx main.ts"
	cmdRun := exec.CommandContext(ctx, "sh", "-c", tsxCmd)
	cmdRun.Dir = caseDir
	outRun, err := cmdRun.CombinedOutput()
	if err != nil {
		t.Fatalf("tsx main.ts failed:\nOutput:\n%s\nError: %v", string(outRun), err)
	}
}

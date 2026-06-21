package e2e_test

import (
	"bytes"
	"encoding/json"
	"flag"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

const (
	casesDir       = "cases"
	configFileName = "vdl.config.vdl"
	pluginFileName = "plugin.js"
	inputFileName  = "input.vdl"
	outputFileName = "output.json"
)

var update = flag.Bool("update", false, "update golden output.json files")

// TestEndToEndIRGolden validates the plugin-facing IR produced by the real CLI.
func TestEndToEndIRGolden(t *testing.T) {
	caseNames := discoverCases(t)
	if len(caseNames) == 0 {
		t.Fatal("no e2e cases found")
	}

	vdlPath := buildVDLBinary(t)

	for _, caseName := range caseNames {
		t.Run(caseName, func(t *testing.T) {
			caseDir := filepath.Join(casesDir, caseName)
			projectDir := t.TempDir()
			copyCaseFiles(t, caseDir, projectDir)
			copyProjectFixtures(t, projectDir)
			runVDLGenerate(t, vdlPath, projectDir)

			actualPath := filepath.Join(projectDir, "gen", outputFileName)
			actual := normalizeJSONFile(t, actualPath)
			goldenPath := filepath.Join(caseDir, outputFileName)

			if *update {
				writeFile(t, goldenPath, actual)
				return
			}

			expected := normalizeJSONFile(t, goldenPath)
			if !bytes.Equal(expected, actual) {
				t.Fatalf(
					"IR golden mismatch for %s\nexpected: %s\nactual:   %s",
					caseName,
					goldenPath,
					actualPath,
				)
			}
		})
	}
}

// discoverCases returns all case directory names in deterministic order.
func discoverCases(t *testing.T) []string {
	t.Helper()

	entries, err := os.ReadDir(casesDir)
	if err != nil {
		t.Fatalf("read cases dir: %v", err)
	}

	var names []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		caseName := entry.Name()
		inputPath := filepath.Join(casesDir, caseName, inputFileName)
		if _, err := os.Stat(inputPath); err != nil {
			t.Fatalf("case %s must contain %s: %v", caseName, inputFileName, err)
		}

		names = append(names, caseName)
	}

	slices.Sort(names)
	return names
}

// buildVDLBinary builds an isolated CLI binary used by every E2E case.
func buildVDLBinary(t *testing.T) string {
	t.Helper()

	binPath := filepath.Join(t.TempDir(), "vdl")
	cmd := exec.Command("go", "build", "-o", binPath, "./cmd/vdl/.")
	cmd.Dir = filepath.Join("..", "toolchain")
	runCommand(t, cmd)
	return binPath
}

// copyCaseFiles copies a golden case into a temporary project, excluding output.json.
func copyCaseFiles(t *testing.T, srcDir, dstDir string) {
	t.Helper()

	err := filepath.WalkDir(srcDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}
		if relPath == "." || relPath == outputFileName {
			return nil
		}

		dstPath := filepath.Join(dstDir, relPath)
		if d.IsDir() {
			return os.MkdirAll(dstPath, 0o755)
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(dstPath), 0o755); err != nil {
			return err
		}
		return os.WriteFile(dstPath, data, 0o644)
	})
	if err != nil {
		t.Fatalf("copy case files from %s: %v", srcDir, err)
	}
}

// copyProjectFixtures installs the shared generation config and IR-dump plugin.
func copyProjectFixtures(t *testing.T, projectDir string) {
	t.Helper()

	copyFile(t, pluginFileName, filepath.Join(projectDir, pluginFileName))
	copyFile(t, configFileName, filepath.Join(projectDir, configFileName))
}

// runVDLGenerate executes the same command a user would run for generation.
func runVDLGenerate(t *testing.T, vdlPath, projectDir string) {
	t.Helper()

	cmd := exec.Command(vdlPath, "generate", projectDir)
	runCommand(t, cmd)
}

// runCommand runs a command and includes combined output in failures.
func runCommand(t *testing.T, cmd *exec.Cmd) {
	t.Helper()

	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &output
	if err := cmd.Run(); err != nil {
		t.Fatalf("%s failed: %v\n%s", strings.Join(cmd.Args, " "), err, output.String())
	}
}

// normalizeJSONFile parses and re-encodes JSON so comparisons ignore formatting.
func normalizeJSONFile(t *testing.T, path string) []byte {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read JSON file %s: %v", path, err)
	}

	var value any
	if err := json.Unmarshal(data, &value); err != nil {
		t.Fatalf("parse JSON file %s: %v", path, err)
	}

	normalized, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		t.Fatalf("normalize JSON file %s: %v", path, err)
	}

	return append(normalized, '\n')
}

// copyFile copies a fixture file while preserving its contents exactly.
func copyFile(t *testing.T, srcPath, dstPath string) {
	t.Helper()

	data, err := os.ReadFile(srcPath)
	if err != nil {
		t.Fatalf("read file %s: %v", srcPath, err)
	}
	writeFile(t, dstPath, data)
}

// writeFile writes a file after creating its parent directory.
func writeFile(t *testing.T, path string, content []byte) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("create parent dir for %s: %v", path, err)
	}
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatalf("write file %s: %v", path, err)
	}
}

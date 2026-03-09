package codegen

import (
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/varavelio/vdl/toolchain/internal/codegen/plugintypes"
)

const runPluginGoldenRoot = "testdata/run_plugin"

type runPluginGoldenCase struct {
	Name         string
	PluginPath   string
	ExpectedPath string
}

type runPluginExpected struct {
	Input         *plugintypes.PluginInput  `json:"input,omitempty"`
	Output        *plugintypes.PluginOutput `json:"output,omitempty"`
	ErrorContains string                    `json:"errorContains,omitempty"`
}

func TestRunPluginGolden(t *testing.T) {
	testCases := discoverRunPluginGoldenCases(t)
	require.NotEmpty(t, testCases, "no runPlugin golden cases found")

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			script := readTextFixture(t, tc.PluginPath)
			expected := readExpectedFixture(t, tc.ExpectedPath)
			input := resolveInput(t, expected)
			output, err := runPlugin("", script, input)

			if expected.ErrorContains != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), expected.ErrorContains)
				require.Equal(t, plugintypes.PluginOutput{}, output)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, expected.Output)
			require.Equal(t, *expected.Output, output)
		})
	}
}

func discoverRunPluginGoldenCases(t *testing.T) []runPluginGoldenCase {
	t.Helper()

	var testCases []runPluginGoldenCase
	err := filepath.WalkDir(runPluginGoldenRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			return nil
		}

		pluginPath := filepath.Join(path, "index.js")
		expectedPath := filepath.Join(path, "expected.json")
		if !isFile(pluginPath) || !isFile(expectedPath) {
			return nil
		}

		relPath, err := filepath.Rel(runPluginGoldenRoot, path)
		if err != nil {
			return err
		}

		testCases = append(testCases, runPluginGoldenCase{
			Name:         filepath.ToSlash(relPath),
			PluginPath:   pluginPath,
			ExpectedPath: expectedPath,
		})

		return nil
	})
	require.NoError(t, err)

	sort.Slice(testCases, func(i, j int) bool {
		return testCases[i].Name < testCases[j].Name
	})

	return testCases
}

func readExpectedFixture(t *testing.T, path string) runPluginExpected {
	t.Helper()

	raw := readTextFixture(t, path)
	var expected runPluginExpected
	require.NoError(t, json.Unmarshal([]byte(raw), &expected))

	if expected.Output == nil && expected.ErrorContains == "" {
		t.Fatalf("fixture %s must define output or errorContains", path)
	}
	if expected.Output != nil && expected.ErrorContains != "" {
		t.Fatalf("fixture %s cannot define both output and errorContains", path)
	}

	return expected
}

func resolveInput(t *testing.T, expected runPluginExpected) plugintypes.PluginInput {
	t.Helper()

	input := defaultRunPluginInput()
	if expected.Input != nil {
		input = *expected.Input
	}

	return input
}

func readTextFixture(t *testing.T, path string) string {
	t.Helper()

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	return string(data)
}

func isFile(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

func defaultRunPluginInput() plugintypes.PluginInput {
	return plugintypes.PluginInput{
		Version: "1.2.3",
		Ir: plugintypes.IrSchema{
			EntryPoint: "/schema/main.vdl",
			Constants:  []plugintypes.ConstantDef{},
			Enums:      []plugintypes.EnumDef{},
			Types:      []plugintypes.TypeDef{},
			Docs:       []plugintypes.TopLevelDoc{},
		},
		Options: map[string]string{
			"target": "typescript",
			"module": "esm",
		},
	}
}

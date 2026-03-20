package codegen

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRunWithConfigHooksUseGlobalSandwichOrder(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, filepath.Join(dir, "schema.vdl"), "type User {\n  name string\n}\n")
	writeTestFile(t, filepath.Join(dir, "plugin/index.js"), `exports.generate = () => ({ files: [{ path: "nested/generated.txt", content: "hello" }] })`)
	writeHookConfigFile(t, dir, []string{"pre-1", "pre-2"}, []string{"check-generated", "post-2"})

	var calls []string
	warningBuffer := &bytes.Buffer{}
	setHostHookTestDoubles(t, func(workDir, command string) error {
		calls = append(calls, command)
		switch command {
		case "pre-2":
			_, err := os.Stat(filepath.Join(workDir, "gen", "nested", "generated.txt"))
			if !errors.Is(err, os.ErrNotExist) {
				return fmt.Errorf("pre hook ran after output generation")
			}
		case "check-generated":
			if _, err := os.Stat(filepath.Join(workDir, "gen", "nested", "generated.txt")); err != nil {
				return fmt.Errorf("generated file not ready for post hook: %w", err)
			}
		}

		return nil
	}, warningBuffer)

	fileCount, err := Run(dir)
	require.NoError(t, err)
	require.Equal(t, 1, fileCount)
	require.Equal(t, []string{"pre-1", "pre-2", "check-generated", "post-2"}, calls)
	require.Empty(t, warningBuffer.String())
	require.FileExists(t, filepath.Join(dir, "gen", "nested", "generated.txt"))
}

func TestRunWithConfigPreGenerateHookFailureAbortsPipeline(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, filepath.Join(dir, "schema.vdl"), "type User {\n  name string\n}\n")
	writeTestFile(t, filepath.Join(dir, "plugin/index.js"), `exports.generate = () => ({ files: [{ path: "generated.txt", content: "hello" }] })`)
	writeHookConfigFile(t, dir, []string{"fail-pre"}, []string{"post-never"})

	var calls []string
	warningBuffer := &bytes.Buffer{}
	setHostHookTestDoubles(t, func(_ string, command string) error {
		calls = append(calls, command)
		if command == "fail-pre" {
			return errors.New("pre failed")
		}
		return nil
	}, warningBuffer)

	_, err := Run(dir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "preGenerate hook command 1 failed")
	require.Equal(t, []string{"fail-pre"}, calls)
	require.Empty(t, warningBuffer.String())
	require.NoFileExists(t, filepath.Join(dir, "gen", "generated.txt"))
}

func TestRunWithConfigPostGenerateHookFailureWarnsAndContinues(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, filepath.Join(dir, "schema.vdl"), "type User {\n  name string\n}\n")
	writeTestFile(t, filepath.Join(dir, "plugin/index.js"), `exports.generate = () => ({ files: [{ path: "generated.txt", content: "hello" }] })`)
	writeHookConfigFile(t, dir, nil, []string{"fail-post", "post-ok"})

	var calls []string
	warningBuffer := &bytes.Buffer{}
	setHostHookTestDoubles(t, func(_ string, command string) error {
		calls = append(calls, command)
		if command == "fail-post" {
			return errors.New("post failed")
		}
		return nil
	}, warningBuffer)

	fileCount, err := Run(dir)
	require.NoError(t, err)
	require.Equal(t, 1, fileCount)
	require.Equal(t, []string{"fail-post", "post-ok"}, calls)
	require.Contains(t, warningBuffer.String(), "VDL warning: postGenerate hook command 1 failed")
	require.FileExists(t, filepath.Join(dir, "gen", "generated.txt"))
}

func TestRunWithConfigSkipsHooksWhenDisabled(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, filepath.Join(dir, "schema.vdl"), "type User {\n  name string\n}\n")
	writeTestFile(t, filepath.Join(dir, "plugin/index.js"), `exports.generate = () => ({ files: [{ path: "generated.txt", content: "hello" }] })`)
	writeHookConfigFile(t, dir, []string{"pre-1"}, []string{"post-1"})

	var calls []string
	warningBuffer := &bytes.Buffer{}
	setHostHookTestDoubles(t, func(_ string, command string) error {
		calls = append(calls, command)
		return nil
	}, warningBuffer)
	t.Setenv("VDL_SKIP_HOST_HOOKS", "true")

	fileCount, err := Run(dir)
	require.NoError(t, err)
	require.Equal(t, 1, fileCount)
	require.Empty(t, calls)
	require.Empty(t, warningBuffer.String())
	require.FileExists(t, filepath.Join(dir, "gen", "generated.txt"))
}

func setHostHookTestDoubles(t *testing.T, runner func(workDir, command string) error, warningWriter *bytes.Buffer) {
	t.Helper()

	originalRunner := hostHookCommandRunner
	originalWarningWriter := hostHookWarningWriter
	hostHookCommandRunner = runner
	hostHookWarningWriter = warningWriter

	t.Cleanup(func() {
		hostHookCommandRunner = originalRunner
		hostHookWarningWriter = originalWarningWriter
	})
}

func writeHookConfigFile(t *testing.T, dir string, preGenerate, postGenerate []string) {
	t.Helper()

	writeTestFile(t, filepath.Join(dir, defaultConfigFileName), fmt.Sprintf(`
		const config = {
			version 1
			hooks {
				preGenerate [%s]
				postGenerate [%s]
			}
			plugins [
				{
					src "./plugin/index.js"
					schema "./schema.vdl"
					outDir "./gen"
				}
			]
		}
	`, hookCommandListLiteral(preGenerate), hookCommandListLiteral(postGenerate)))
}

func hookCommandListLiteral(commands []string) string {
	if len(commands) == 0 {
		return ""
	}

	quoted := make([]string, len(commands))
	for i, command := range commands {
		quoted[i] = strconv.Quote(command)
	}

	return strings.Join(quoted, " ")
}

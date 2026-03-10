package codegen

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/varavelio/vdl/toolchain/internal/codegen/plugintypes"
	"github.com/varavelio/vdl/toolchain/internal/version"
)

func TestApplyGeneratedFileHeaders(t *testing.T) {
	t.Run("prepends header to supported files", func(t *testing.T) {
		files := []plugintypes.PluginOutputFile{{
			Path:    "main.go",
			Content: "package main\n",
		}}
		results := []executedPlugin{{
			Plugin: runtimePlugin{Source: pluginSource{Original: "owner/vdl-plugin-demo@v1.0.0", ContentHash: "sha256-1234567890abcdef"}},
			Output: plugintypes.PluginOutput{Files: &files},
		}}

		applyGeneratedFileHeaders(results)

		expectedHeader := buildGeneratedFileHeader("owner/vdl-plugin-demo@v1.0.0 (hash 12345678)", lineSlashComment)
		require.Equal(t, expectedHeader+"package main\n", results[0].Output.GetFiles()[0].GetContent())
		require.Contains(t, results[0].Output.GetFiles()[0].GetContent(), "VDL v"+version.Version)
		require.Contains(t, results[0].Output.GetFiles()[0].GetContent(), "commit "+version.Commit)
		require.Contains(t, results[0].Output.GetFiles()[0].GetContent(), generatedFileHeaderRepoURL)
	})

	t.Run("uses matching comment syntax for sql files", func(t *testing.T) {
		files := []plugintypes.PluginOutputFile{{
			Path:    "query.sql",
			Content: "select 1;\n",
		}}
		results := []executedPlugin{{
			Plugin: runtimePlugin{Source: pluginSource{Original: "https://example.com/plugin.js"}},
			Output: plugintypes.PluginOutput{Files: &files},
		}}

		applyGeneratedFileHeaders(results)

		require.Equal(
			t,
			buildGeneratedFileHeader("https://example.com/plugin.js", lineDashComment)+"select 1;\n",
			results[0].Output.GetFiles()[0].GetContent(),
		)
	})

	t.Run("uses block comments for html-like files", func(t *testing.T) {
		files := []plugintypes.PluginOutputFile{{
			Path:    "index.html",
			Content: "<main>Hello</main>\n",
		}}
		results := []executedPlugin{{
			Plugin: runtimePlugin{Source: pluginSource{Original: "./plugin.js"}},
			Output: plugintypes.PluginOutput{Files: &files},
		}}

		applyGeneratedFileHeaders(results)

		require.Equal(
			t,
			buildGeneratedFileHeader("./plugin.js", htmlBlockComment)+"<main>Hello</main>\n",
			results[0].Output.GetFiles()[0].GetContent(),
		)
	})

	t.Run("supports basename-only formats", func(t *testing.T) {
		files := []plugintypes.PluginOutputFile{{
			Path:    "Dockerfile",
			Content: "FROM alpine:3.21\n",
		}}
		results := []executedPlugin{{
			Plugin: runtimePlugin{Source: pluginSource{Original: "./plugin.js"}},
			Output: plugintypes.PluginOutput{Files: &files},
		}}

		applyGeneratedFileHeaders(results)

		require.Equal(
			t,
			buildGeneratedFileHeader("./plugin.js", lineHashComment)+"FROM alpine:3.21\n",
			results[0].Output.GetFiles()[0].GetContent(),
		)
	})

	t.Run("leaves unsupported files unchanged", func(t *testing.T) {
		files := []plugintypes.PluginOutputFile{{
			Path:    "notes.txt",
			Content: "hello\n",
		}}
		results := []executedPlugin{{
			Plugin: runtimePlugin{Source: pluginSource{Original: "./plugin.js"}},
			Output: plugintypes.PluginOutput{Files: &files},
		}}

		applyGeneratedFileHeaders(results)

		require.Equal(t, "hello\n", results[0].Output.GetFiles()[0].GetContent())
	})

	t.Run("shortens plugin hashes for header labels", func(t *testing.T) {
		plugin := runtimePlugin{Source: pluginSource{Original: "./plugin.js", ContentHash: "sha256-abcdef1234567890"}}
		require.Equal(t, "./plugin.js (hash abcdef12)", generatedHeaderPluginLabel(plugin))
	})
}

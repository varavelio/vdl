package codegen

import (
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	"github.com/varavelio/vdl/toolchain/internal/version"
)

const generatedFileHeaderRepoURL = "https://github.com/varavelio/vdl"

type generatedFileCommentStyle struct {
	LinePrefix string
	BlockStart string
	BlockLine  string
	BlockEnd   string
}

type generatedFileHeaderLanguage struct {
	Name       string
	Extensions []string
	Basenames  []string
	Comment    generatedFileCommentStyle
}

var (
	lineSlashComment   = generatedFileCommentStyle{LinePrefix: "//"}
	lineHashComment    = generatedFileCommentStyle{LinePrefix: "#"}
	lineDashComment    = generatedFileCommentStyle{LinePrefix: "--"}
	lineSemiComment    = generatedFileCommentStyle{LinePrefix: ";"}
	lineBangComment    = generatedFileCommentStyle{LinePrefix: "!"}
	linePercentComment = generatedFileCommentStyle{LinePrefix: "%"}
	cBlockComment      = generatedFileCommentStyle{BlockStart: "/*", BlockLine: " *", BlockEnd: " */"}
	htmlBlockComment   = generatedFileCommentStyle{BlockStart: "<!--", BlockLine: "  ", BlockEnd: "-->"}
	ocamlBlockComment  = generatedFileCommentStyle{BlockStart: "(*", BlockLine: " *", BlockEnd: " *)"}
)

var generatedFileHeaderLanguages = []generatedFileHeaderLanguage{
	{Name: "VDL", Extensions: []string{".vdl"}, Comment: lineSlashComment},
	{Name: "JavaScript", Extensions: []string{".js", ".mjs", ".cjs"}, Comment: lineSlashComment},
	{Name: "TypeScript", Extensions: []string{".ts", ".mts", ".cts"}, Comment: lineSlashComment},
	{Name: "JSX", Extensions: []string{".jsx"}, Comment: lineSlashComment},
	{Name: "TSX", Extensions: []string{".tsx"}, Comment: lineSlashComment},
	{Name: "Go", Extensions: []string{".go"}, Comment: lineSlashComment},
	{Name: "Python", Extensions: []string{".py", ".pyi"}, Comment: lineHashComment},
	{Name: "Ruby", Extensions: []string{".rb", ".rake", ".gemspec"}, Comment: lineHashComment},
	{Name: "Perl", Extensions: []string{".pl", ".pm", ".t"}, Comment: lineHashComment},
	{Name: "PHP", Extensions: []string{".php", ".phtml"}, Comment: lineSlashComment},
	{Name: "Java", Extensions: []string{".java"}, Comment: lineSlashComment},
	{Name: "Kotlin", Extensions: []string{".kt", ".kts"}, Comment: lineSlashComment},
	{Name: "Scala", Extensions: []string{".scala", ".sc"}, Comment: lineSlashComment},
	{Name: "Groovy", Extensions: []string{".groovy", ".gradle"}, Comment: lineSlashComment},
	{Name: "Swift", Extensions: []string{".swift"}, Comment: lineSlashComment},
	{Name: "Rust", Extensions: []string{".rs"}, Comment: lineSlashComment},
	{Name: "C", Extensions: []string{".c", ".h"}, Comment: lineSlashComment},
	{Name: "C++", Extensions: []string{".cc", ".cpp", ".cxx", ".hpp", ".hh", ".hxx"}, Comment: lineSlashComment},
	{Name: "C#", Extensions: []string{".cs"}, Comment: lineSlashComment},
	{Name: "F#", Extensions: []string{".fs", ".fsi", ".fsx"}, Comment: lineSlashComment},
	{Name: "Dart", Extensions: []string{".dart"}, Comment: lineSlashComment},
	{Name: "Lua", Extensions: []string{".lua"}, Comment: lineDashComment},
	{Name: "Shell", Extensions: []string{".sh", ".bash", ".zsh", ".fish"}, Basenames: []string{".envrc"}, Comment: lineHashComment},
	{Name: "PowerShell", Extensions: []string{".ps1", ".psm1", ".psd1"}, Comment: lineHashComment},
	{Name: "SQL", Extensions: []string{".sql"}, Comment: lineDashComment},
	{Name: "R", Extensions: []string{".r"}, Comment: lineHashComment},
	{Name: "MATLAB", Extensions: []string{".m"}, Comment: linePercentComment},
	{Name: "Julia", Extensions: []string{".jl"}, Comment: lineHashComment},
	{Name: "Elixir", Extensions: []string{".ex", ".exs"}, Comment: lineHashComment},
	{Name: "Erlang", Extensions: []string{".erl", ".hrl"}, Comment: linePercentComment},
	{Name: "Haskell", Extensions: []string{".hs"}, Comment: lineDashComment},
	{Name: "Elm", Extensions: []string{".elm"}, Comment: lineDashComment},
	{Name: "Nim", Extensions: []string{".nim"}, Comment: lineHashComment},
	{Name: "Zig", Extensions: []string{".zig", ".zon"}, Comment: lineSlashComment},
	{Name: "Solidity", Extensions: []string{".sol"}, Comment: lineSlashComment},
	{Name: "GraphQL", Extensions: []string{".graphql", ".gql", ".graphqls"}, Comment: lineHashComment},
	{Name: "YAML", Extensions: []string{".yaml", ".yml"}, Comment: lineHashComment},
	{Name: "TOML", Extensions: []string{".toml"}, Comment: lineHashComment},
	{Name: "Nix", Extensions: []string{".nix"}, Comment: lineHashComment},
	{Name: "Tcl", Extensions: []string{".tcl"}, Comment: lineHashComment},
	{Name: "Ada", Extensions: []string{".adb", ".ads"}, Comment: lineDashComment},
	{Name: "Fortran", Extensions: []string{".f90", ".f95", ".f03", ".f08"}, Comment: lineBangComment},
	{Name: "Assembly", Extensions: []string{".asm", ".s", ".S"}, Comment: lineSemiComment},
	{Name: "Lisp", Extensions: []string{".lisp", ".lsp", ".el"}, Comment: lineSemiComment},
	{Name: "Scheme", Extensions: []string{".scm", ".ss"}, Comment: lineSemiComment},
	{Name: "Clojure", Extensions: []string{".clj", ".cljs", ".cljc", ".edn"}, Comment: generatedFileCommentStyle{LinePrefix: ";;"}},
	{Name: "Verilog", Extensions: []string{".v", ".sv", ".svh"}, Comment: lineSlashComment},
	{Name: "HCL", Extensions: []string{".hcl", ".tf", ".tfvars"}, Comment: lineHashComment},
	{Name: "Protocol Buffers", Extensions: []string{".proto"}, Comment: lineSlashComment},
	{Name: "Thrift", Extensions: []string{".thrift"}, Comment: lineSlashComment},
	{Name: "Avro IDL", Extensions: []string{".avdl"}, Comment: lineSlashComment},
	{Name: "Cap'n Proto", Extensions: []string{".capnp"}, Comment: lineHashComment},
	{Name: "FlatBuffers", Extensions: []string{".fbs"}, Comment: lineSlashComment},
	{Name: "Smithy", Extensions: []string{".smithy"}, Comment: lineSlashComment},
	{Name: "CUE", Extensions: []string{".cue"}, Comment: lineSlashComment},
	{Name: "Rego", Extensions: []string{".rego"}, Comment: lineHashComment},
	{Name: "Dhall", Extensions: []string{".dhall"}, Comment: lineDashComment},
	{Name: "Puppet", Extensions: []string{".pp"}, Comment: lineHashComment},
	{Name: "Bicep", Extensions: []string{".bicep"}, Comment: lineSlashComment},
	{Name: "Prisma", Extensions: []string{".prisma"}, Comment: lineSlashComment},
	{Name: "OCaml", Extensions: []string{".ml", ".mli", ".mll", ".mly"}, Comment: ocamlBlockComment},
	{Name: "Reason", Extensions: []string{".re", ".rei"}, Comment: ocamlBlockComment},
	{Name: "HTML", Extensions: []string{".html", ".htm"}, Comment: htmlBlockComment},
	{Name: "XML", Extensions: []string{".xml", ".xsd", ".xsl", ".xslt", ".svg", ".plist"}, Comment: htmlBlockComment},
	{Name: "CSS", Extensions: []string{".css", ".pcss", ".scss", ".less"}, Comment: cBlockComment},
	{Name: "Sass", Extensions: []string{".sass"}, Comment: lineSlashComment},
	{Name: "Vue", Extensions: []string{".vue"}, Comment: htmlBlockComment},
	{Name: "Svelte", Extensions: []string{".svelte"}, Comment: htmlBlockComment},
	{Name: "Astro", Extensions: []string{".astro"}, Comment: htmlBlockComment},
	{Name: "Markdown", Extensions: []string{".md", ".mdx", ".markdown"}, Comment: htmlBlockComment},
	{Name: "Docker", Basenames: []string{"Dockerfile"}, Extensions: []string{".dockerfile"}, Comment: lineHashComment},
	{Name: "Make", Basenames: []string{"Makefile", "GNUmakefile"}, Extensions: []string{".mk"}, Comment: lineHashComment},
	{Name: "CMake", Basenames: []string{"CMakeLists.txt"}, Extensions: []string{".cmake"}, Comment: lineHashComment},
	{Name: "Bazel", Basenames: []string{"BUILD", "WORKSPACE", "MODULE.bazel"}, Extensions: []string{".bzl", ".bazel"}, Comment: lineHashComment},
	{Name: "Properties", Extensions: []string{".properties", ".ini", ".cfg", ".conf"}, Comment: lineHashComment},
	{Name: "LaTeX", Extensions: []string{".tex", ".sty", ".bib"}, Comment: linePercentComment},
	{Name: "Crystal", Extensions: []string{".cr"}, Comment: lineHashComment},
	{Name: "Apex", Extensions: []string{".cls", ".trigger"}, Comment: lineSlashComment},
	{Name: "Pkl", Extensions: []string{".pkl"}, Comment: lineSlashComment},
}

// applyGeneratedFileHeaders prepends the standard VDL generated-file header to
// every plugin output whose path matches a supported language or IDL, unless
// the plugin explicitly disables generated headers.
func applyGeneratedFileHeaders(results []executedPlugin) {
	for i := range results {
		if !results[i].Plugin.GenerateHeader {
			continue
		}

		files := results[i].Output.GetFiles()
		if len(files) == 0 {
			continue
		}

		pluginName := generatedHeaderPluginLabel(results[i].Plugin)
		for fileIndex := range files {
			commentStyle, ok := generatedFileCommentStyleForPath(files[fileIndex].GetPath())
			if !ok {
				continue
			}
			files[fileIndex].Content = buildGeneratedFileHeader(pluginName, commentStyle) + files[fileIndex].GetContent()
		}

		results[i].Output.Files = &files
	}
}

// generatedFileCommentStyleForPath returns the comment style for path when its
// extension or basename matches one of the supported generated header formats.
func generatedFileCommentStyleForPath(path string) (generatedFileCommentStyle, bool) {
	trimmedPath := strings.TrimSpace(path)
	if trimmedPath == "" {
		return generatedFileCommentStyle{}, false
	}

	baseName := filepath.Base(trimmedPath)
	extension := strings.ToLower(filepath.Ext(trimmedPath))

	for _, language := range generatedFileHeaderLanguages {
		if slices.Contains(language.Basenames, baseName) {
			return language.Comment, true
		}
		if extension != "" && slices.Contains(language.Extensions, extension) {
			return language.Comment, true
		}
	}

	return generatedFileCommentStyle{}, false
}

// buildGeneratedFileHeader renders the generated-file banner using the comment
// style required by the target language.
func buildGeneratedFileHeader(pluginName string, commentStyle generatedFileCommentStyle) string {
	lines := generatedFileHeaderLines(pluginName)
	if commentStyle.LinePrefix != "" {
		return buildLineCommentHeader(lines, commentStyle.LinePrefix)
	}
	return buildBlockCommentHeader(lines, commentStyle)
}

// generatedFileHeaderLines returns the human-facing lines that make up the VDL
// generated-file banner.
func generatedFileHeaderLines(pluginName string) []string {
	return []string{
		fmt.Sprintf("Code generated by VDL v%s (commit %s) using %s", version.Version, version.Commit, pluginName),
		"Any changes will be overwritten the next time VDL is run. DO NOT EDIT.",
		fmt.Sprintf("Learn more: %s", generatedFileHeaderRepoURL),
	}
}

// buildLineCommentHeader renders header lines using a repeated single-line
// comment prefix.
func buildLineCommentHeader(lines []string, prefix string) string {
	var builder strings.Builder
	for _, line := range lines {
		builder.WriteString(prefix)
		builder.WriteByte(' ')
		builder.WriteString(line)
		builder.WriteByte('\n')
	}
	builder.WriteByte('\n')
	return builder.String()
}

// buildBlockCommentHeader renders header lines using a block comment style.
func buildBlockCommentHeader(lines []string, style generatedFileCommentStyle) string {
	var builder strings.Builder
	builder.WriteString(style.BlockStart)
	builder.WriteByte('\n')
	for _, line := range lines {
		builder.WriteString(style.BlockLine)
		if style.BlockLine != "" && !strings.HasSuffix(style.BlockLine, " ") {
			builder.WriteByte(' ')
		}
		builder.WriteString(line)
		builder.WriteByte('\n')
	}
	builder.WriteString(style.BlockEnd)
	builder.WriteString("\n\n")
	return builder.String()
}

// generatedHeaderPluginName returns the user-facing plugin identifier to embed
// in generated file headers.
func generatedHeaderPluginName(plugin runtimePlugin) string {
	pluginName := strings.TrimSpace(plugin.Source.Original)
	if pluginName != "" {
		return pluginName
	}
	return plugin.Source.DisplayName
}

// generatedHeaderPluginLabel returns the plugin label shown in generated file
// headers, including a short content hash when available.
func generatedHeaderPluginLabel(plugin runtimePlugin) string {
	pluginName := generatedHeaderPluginName(plugin)
	shortHash := shortenGeneratedHeaderHash(plugin.Source.ContentHash)
	if shortHash == "" {
		return pluginName
	}
	return fmt.Sprintf("%s (hash %s)", pluginName, shortHash)
}

// shortenGeneratedHeaderHash trims the digest prefix and returns at most the
// first eight characters for compact display in generated file headers.
func shortenGeneratedHeaderHash(hash string) string {
	hash = strings.TrimSpace(hash)
	hash = strings.TrimPrefix(hash, "sha256-")
	if len(hash) > 8 {
		return hash[:8]
	}
	return hash
}

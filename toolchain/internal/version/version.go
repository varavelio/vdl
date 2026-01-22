package version

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/varavelio/vdl/toolchain/internal/util/strutil"
)

func init() {
	// Clean up the Version
	Version = strings.TrimPrefix(strings.TrimSpace(Version), "v")

	// Set VersionMajor based on Version
	parts := strings.Split(Version, ".")
	if len(parts) > 0 {
		VersionMajor = parts[0]
	}

	// Initialize Schema IDs after VersionMajor is set
	SchemaIRID = fmt.Sprintf("https://vdl.varavel.com/schemas/v%s/ir.schema.json", VersionMajor)
	SchemaConfigID = fmt.Sprintf("https://vdl.varavel.com/schemas/v%s/config.schema.json", VersionMajor)
	SchemaPluginInputID = fmt.Sprintf("https://vdl.varavel.com/schemas/v%s/plugin_input.schema.json", VersionMajor)
	SchemaPluginOutputID = fmt.Sprintf("https://vdl.varavel.com/schemas/v%s/plugin_output.schema.json", VersionMajor)
}

var (
	// Version is replaced during the release process by the latest Git tag
	// and should not be manually edited.
	Version = "0.0.0-dev"

	// VersionMajor is the major version extracted from Version.
	VersionMajor = "0"
)

// Schema IDs
var (
	// SchemaIRID is the canonical URL for the IR JSON Schema.
	SchemaIRID string

	// SchemaConfigID is the canonical URL for the VDL Config JSON Schema.
	SchemaConfigID string

	// SchemaPluginInputID is the canonical URL for the Plugin Input JSON Schema.
	SchemaPluginInputID string

	// SchemaPluginOutputID is the canonical URL for the Plugin Output JSON Schema.
	SchemaPluginOutputID string
)

// asciiArtRaw is used to generate AsciiArt
var asciiArtRaw = strings.TrimSpace(`
██╗   ██╗██████╗ ██╗
██║   ██║██╔══██╗██║
██║   ██║██║  ██║██║
╚██╗ ██╔╝██║  ██║██║
 ╚████╔╝ ██████╔╝███████╗
  ╚═══╝  ╚═════╝ ╚══════╝
`)

// basicInfoRaw is used to generate AsciiArt
var basicInfoRaw = strings.Join([]string{
	"Star the repo: https://github.com/varavelio/vdl",
	"Show usage:    vdl --help",
	"Show version:  vdl --version",
}, "\n")

// AsciiArt is the ASCII art for the VDL logo.
// It is generated dynamically to ensure that the logo is always
// centered and the lines are always the same width.
var AsciiArt = func() string {
	maxWidth := 0
	for line := range strings.SplitSeq(basicInfoRaw, "\n") {
		if utf8.RuneCountInString(line) > maxWidth {
			maxWidth = utf8.RuneCountInString(line)
		}
	}
	horizontal := strings.Repeat("─", maxWidth)

	combined := strings.Join([]string{
		strutil.CenterText(asciiArtRaw, maxWidth),
		strutil.CenterText("v"+Version, maxWidth),
		"",
		basicInfoRaw,
	}, "\n")

	var combinedWithLines strings.Builder
	for line := range strings.SplitSeq(combined, "\n") {
		spaces := maxWidth - utf8.RuneCountInString(line)
		fmt.Fprintf(&combinedWithLines, "│ %s%s │\n", line, strings.Repeat(" ", spaces))
	}

	lines := []string{
		"┌─" + horizontal + "─┐",
		strings.TrimSpace(combinedWithLines.String()),
		"└─" + horizontal + "─┘",
	}

	return strings.Join(lines, "\n")
}()

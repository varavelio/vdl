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
}

var (
	// Version is replaced during the release process by the latest Git tag
	// and should not be manually edited.
	Version = "0.0.0-dev"

	// VersionMajor is the major version extracted from Version.
	VersionMajor = "0"
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

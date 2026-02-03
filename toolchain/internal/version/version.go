package version

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/varavelio/vdl/toolchain/internal/util/cliutil"
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
	//
	// This string does not contain the "v" prefix.
	Version = "0.0.0-dev"

	// VersionMajor is the major version extracted from Version.
	//
	// This string does not contain the "v" prefix.
	VersionMajor = "0"
)

// asciiArtRaw is used to generate AsciiArt
var asciiArtRaw = strings.TrimSpace(`
██╗  ██╗█████╗ ██╗
██║  ██║██╔═██╗██║
╚██╗██╔╝██║ ██║██║
 ╚███╔╝ █████╔╝█████╗
  ╚══╝  ╚════╝ ╚════╝
`)

// basicInfoRaw is used to generate AsciiArt
var basicInfoRaw = strings.Join([]string{
	"Star the repo: https://github.com/varavelio/vdl",
	"Show usage:    vdl --help",
	"Show version:  vdl --version",
}, "\n")

// AsciiArt is the ASCII art for the VDL logo, dynamically generated
// to ensure proper centering and consistent line widths.
var AsciiArt = func() string {
	maxWidth := 0
	for line := range strings.SplitSeq(basicInfoRaw, "\n") {
		if utf8.RuneCountInString(line) > maxWidth {
			maxWidth = utf8.RuneCountInString(line)
		}
	}

	logoSection := strutil.CenterText(asciiArtRaw, maxWidth)
	versionSection := strutil.CenterText("v"+Version, maxWidth)
	combined := strings.Join([]string{logoSection, versionSection, "", basicInfoRaw}, "\n")

	// Main box (bold)
	horizontal := cliutil.ColorizeBlueBold(strings.Repeat("━", maxWidth))
	borderLeft := cliutil.ColorizeBlueBold("┃")
	borderRight := cliutil.ColorizeBlueBold("┃")
	cornerTL := cliutil.ColorizeBlueBold("┏━")
	cornerTR := cliutil.ColorizeBlueBold("━┓")
	cornerBL := cliutil.ColorizeBlueBold("┗━")
	cornerBR := cliutil.ColorizeBlueBold("━┛")

	// Shadow box (non-bold for depth effect)
	doubleTopRight := cliutil.ColorizeBlue("╗")
	doubleRight := cliutil.ColorizeBlue("║")
	doubleBottom := cliutil.ColorizeBlue("╚" + strings.Repeat("═", maxWidth+2) + "╝")

	githubURL := "https://github.com/varavelio/vdl"

	var contentLines strings.Builder
	lineIndex := 0
	logoLineCount := strings.Count(logoSection, "\n") + 1
	versionLineIndex := logoLineCount

	for line := range strings.SplitSeq(combined, "\n") {
		spaces := maxWidth - utf8.RuneCountInString(line)
		padding := strings.Repeat(" ", spaces)

		var coloredLine string
		if lineIndex < logoLineCount {
			coloredLine = cliutil.ColorizeBlueBold(line)
		} else if lineIndex == versionLineIndex {
			coloredLine = cliutil.ColorizeGreenBold(line)
		} else if strings.Contains(line, githubURL) {
			coloredLine = strings.Replace(line, "Star the repo:", cliutil.ColorizeWhiteBold("Star the repo:"), 1)
			coloredLine = strings.Replace(coloredLine, githubURL, cliutil.ColorizeCyanBoldUnderline(githubURL), 1)
		} else if strings.HasPrefix(line, "Show usage:") {
			coloredLine = strings.Replace(line, "Show usage:", cliutil.ColorizeWhiteBold("Show usage:"), 1)
		} else if strings.HasPrefix(line, "Show version:") {
			coloredLine = strings.Replace(line, "Show version:", cliutil.ColorizeWhiteBold("Show version:"), 1)
		} else {
			coloredLine = line
		}

		if lineIndex == 0 {
			fmt.Fprintf(&contentLines, "%s %s%s %s%s\n", borderLeft, coloredLine, padding, borderRight, doubleTopRight)
		} else {
			fmt.Fprintf(&contentLines, "%s %s%s %s%s\n", borderLeft, coloredLine, padding, borderRight, doubleRight)
		}
		lineIndex++
	}

	return strings.Join([]string{
		cornerTL + horizontal + cornerTR,
		strings.TrimSuffix(contentLines.String(), "\n"),
		cornerBL + horizontal + cornerBR + doubleRight,
		" " + doubleBottom,
	}, "\n")
}()

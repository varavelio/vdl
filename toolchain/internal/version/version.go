package version

import (
	"strings"
	"unicode/utf8"

	"github.com/varavelio/tinta"
	"github.com/varavelio/vdl/toolchain/internal/util/strutil"
)

// These variables are set at build time using ldflags.
var (
	// Version is the vdl version, set at build time using ldflags. This string does not contain the "v" prefix.
	Version = "0.0.0-dev"
	// Commit is the git commit hash, set at build time using ldflags.
	Commit = "unknown"
	// Date is the build date, set at build time using ldflags.
	Date = "unknown"
)

// asciiArtRaw is used to generate AsciiArt.
var asciiArtRaw = strings.TrimSpace(`
██╗  ██╗█████╗ ██╗
██║  ██║██╔═██╗██║
╚██╗██╔╝██║ ██║██║
 ╚███╔╝ █████╔╝█████╗
  ╚══╝  ╚════╝ ╚════╝
`)

// basicInfoRaw is used to generate AsciiArt.
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
		if w := utf8.RuneCountInString(line); w > maxWidth {
			maxWidth = w
		}
	}

	logoSection := strutil.CenterText(asciiArtRaw, maxWidth)
	versionSection := strutil.CenterText("v"+Version, maxWidth)
	combined := strings.Join([]string{logoSection, versionSection, "", basicInfoRaw}, "\n")

	githubURL := "https://github.com/varavelio/vdl"
	logoLineCount := strings.Count(logoSection, "\n") + 1
	versionLineIndex := logoLineCount

	var contentLines strings.Builder
	for lineIndex, line := range strings.Split(combined, "\n") {
		var coloredLine string
		switch {
		case lineIndex < logoLineCount:
			coloredLine = tinta.Text().Blue().Bold().String(line)
		case lineIndex == versionLineIndex:
			coloredLine = tinta.Text().Green().Bold().String(line)
		case strings.Contains(line, githubURL):
			coloredLine = strings.Replace(line, "Star the repo:", tinta.Text().White().Bold().String("Star the repo:"), 1)
			coloredLine = strings.Replace(coloredLine, githubURL, tinta.Text().Cyan().Bold().Underline().String(githubURL), 1)
		case strings.HasPrefix(line, "Show usage:"):
			coloredLine = strings.Replace(line, "Show usage:", tinta.Text().White().Bold().String("Show usage:"), 1)
		case strings.HasPrefix(line, "Show version:"):
			coloredLine = strings.Replace(line, "Show version:", tinta.Text().White().Bold().String("Show version:"), 1)
		default:
			coloredLine = line
		}

		if contentLines.Len() > 0 {
			contentLines.WriteByte('\n')
		}
		contentLines.WriteString(coloredLine)
	}

	return tinta.Box().
		BorderHeavy().
		Blue().Bold().
		Shadow(tinta.ShadowBottomRight, tinta.ShadowStyle{
			TopRight:    tinta.BorderDouble.TopRight,
			TopLeft:     tinta.BorderDouble.TopLeft,
			BottomRight: tinta.BorderDouble.BottomRight,
			BottomLeft:  tinta.BorderDouble.BottomLeft,
			Vertical:    tinta.BorderDouble.Vertical,
			Horizontal:  tinta.BorderDouble.Horizontal,
		}).
		PaddingX(1).
		String(contentLines.String())
}()

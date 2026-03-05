package version

import (
	"strings"

	"github.com/varavelio/tinta"
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

// asciiLogo is used to generate AsciiArt.
var asciiLogo = []string{
	"██╗  ██╗█████╗ ██╗   ",
	"██║  ██║██╔═██╗██║   ",
	"╚██╗██╔╝██║ ██║██║   ",
	" ╚███╔╝ █████╔╝█████╗",
	"  ╚══╝  ╚════╝ ╚════╝",
}

// AsciiArt is the ASCII art for the VDL logo, dynamically generated
// to ensure proper centering and consistent line widths.
var AsciiArt = func() string {
	textBold := tinta.Text().White().Bold()
	textLink := tinta.Text().Cyan().Bold().Underline()
	textBlue := tinta.Text().Blue().Bold()
	textGreen := tinta.Text().Green().Bold()

	content := strings.Builder{}

	// Write ascii logo
	for _, line := range asciiLogo {
		content.WriteString(textBlue.String(line))
		content.WriteString("\n")
	}

	// Write version
	content.WriteString(textGreen.String("v" + Version))
	content.WriteString("\n\n")

	// Write basic info
	content.WriteString(textBold.String("Star the repo: ") + textLink.String("https://github.com/varavelio/vdl"))
	content.WriteString("\n")
	content.WriteString(textBold.String("Show usage:    ") + "vdl --help")
	content.WriteString("\n")
	content.WriteString(textBold.String("Show version:  ") + "vdl --version")

	mainBox := tinta.Box().
		Border(tinta.BorderHeavy).
		Blue().
		Bold().
		PaddingX(1).
		CenterLine(0).
		CenterLine(1).
		CenterLine(2).
		CenterLine(3).
		CenterLine(4).
		CenterLine(5).
		String(content.String())

	shadowBox := tinta.Box().
		Border(tinta.BorderDouble).
		Blue().
		PaddingX(1).
		String(content.String())

	return tinta.Canvas().Add(shadowBox, 1, 1).Add(mainBox, 0, 0).String()
}()

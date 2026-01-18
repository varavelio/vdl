package version

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/uforg/uforpc/urpc/internal/util/strutil"
)

// Version is replaced during the release process by the latest Git tag
// and should not be manually edited.
const Version = "0.0.0-dev"
const VersionWithPrefix = "v" + Version

// asciiArtRaw is used to generate AsciiArt
var asciiArtRaw = strings.Join([]string{
	"╦ ╦╔═╗╔═╗  ╦═╗╔═╗╔═╗",
	"║ ║╠╣ ║ ║  ╠╦╝╠═╝║  ",
	"╚═╝╚  ╚═╝  ╩╚═╩  ╚═╝",
}, "\n")

// basicInfoRaw is used to generate AsciiArt
var basicInfoRaw = strings.Join([]string{
	"Star the repo: https://github.com/uforg/uforpc",
	"Show usage:    urpc --help",
	"Show version:  urpc --version",
}, "\n")

// AsciiArt is the ASCII art for the UFO RPC logo.
// It is generated dynamically to ensure that the logo is always
// centered and the lines are always the same width.
var AsciiArt = func() string {
	maxWidth := 0
	for line := range strings.SplitSeq(basicInfoRaw, "\n") {
		if utf8.RuneCountInString(line) > maxWidth {
			maxWidth = utf8.RuneCountInString(line)
		}
	}
	dashes := strings.Repeat("-", maxWidth)

	combined := strings.Join([]string{
		strutil.CenterText(asciiArtRaw, maxWidth),
		strutil.CenterText(VersionWithPrefix, maxWidth),
		"",
		basicInfoRaw,
	}, "\n")

	combinedWithLines := ""
	for line := range strings.SplitSeq(combined, "\n") {
		spaces := maxWidth - utf8.RuneCountInString(line)
		combinedWithLines += fmt.Sprintf("| %s%s |\n", line, strings.Repeat(" ", spaces))
	}

	lines := []string{
		"+-" + dashes + "-+",
		strings.TrimSpace(combinedWithLines),
		"+-" + dashes + "-+",
	}

	return strings.Join(lines, "\n")
}()

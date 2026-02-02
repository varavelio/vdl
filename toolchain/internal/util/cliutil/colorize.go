package cliutil

import "fmt"

// ANSI color codes
const (
	reset   = "\033[0m"
	bold    = "\033[1m"
	black   = "\033[30m"
	red     = "\033[31m"
	green   = "\033[32m"
	yellow  = "\033[33m"
	blue    = "\033[34m"
	magenta = "\033[35m"
	cyan    = "\033[36m"
	white   = "\033[37m"
)

// ColorizeBlack returns the string wrapped in black color.
func ColorizeBlack(s string) string {
	return fmt.Sprintf("%s%s%s", black, s, reset)
}

// ColorizeRed returns the string wrapped in red color.
func ColorizeRed(s string) string {
	return fmt.Sprintf("%s%s%s", red, s, reset)
}

// ColorizeGreen returns the string wrapped in green color.
func ColorizeGreen(s string) string {
	return fmt.Sprintf("%s%s%s", green, s, reset)
}

// ColorizeYellow returns the string wrapped in yellow color.
func ColorizeYellow(s string) string {
	return fmt.Sprintf("%s%s%s", yellow, s, reset)
}

// ColorizeBlue returns the string wrapped in blue color.
func ColorizeBlue(s string) string {
	return fmt.Sprintf("%s%s%s", blue, s, reset)
}

// ColorizeMagenta returns the string wrapped in magenta color.
func ColorizeMagenta(s string) string {
	return fmt.Sprintf("%s%s%s", magenta, s, reset)
}

// ColorizeCyan returns the string wrapped in cyan color.
func ColorizeCyan(s string) string {
	return fmt.Sprintf("%s%s%s", cyan, s, reset)
}

// ColorizeWhite returns the string wrapped in white color.
func ColorizeWhite(s string) string {
	return fmt.Sprintf("%s%s%s", white, s, reset)
}

// ColorizeBold returns the string wrapped in bold.
func ColorizeBold(s string) string {
	return fmt.Sprintf("%s%s%s%s", bold, white, s, reset)
}

// ColorizeBlackBold returns the string wrapped in bold black color.
func ColorizeBlackBold(s string) string {
	return fmt.Sprintf("%s%s%s%s", bold, black, s, reset)
}

// ColorizeRedBold returns the string wrapped in bold red color.
func ColorizeRedBold(s string) string {
	return fmt.Sprintf("%s%s%s%s", bold, red, s, reset)
}

// ColorizeGreenBold returns the string wrapped in bold green color.
func ColorizeGreenBold(s string) string {
	return fmt.Sprintf("%s%s%s%s", bold, green, s, reset)
}

// ColorizeYellowBold returns the string wrapped in bold yellow color.
func ColorizeYellowBold(s string) string {
	return fmt.Sprintf("%s%s%s%s", bold, yellow, s, reset)
}

// ColorizeBlueBold returns the string wrapped in bold blue color.
func ColorizeBlueBold(s string) string {
	return fmt.Sprintf("%s%s%s%s", bold, blue, s, reset)
}

// ColorizeMagentaBold returns the string wrapped in bold magenta color.
func ColorizeMagentaBold(s string) string {
	return fmt.Sprintf("%s%s%s%s", bold, magenta, s, reset)
}

// ColorizeCyanBold returns the string wrapped in bold cyan color.
func ColorizeCyanBold(s string) string {
	return fmt.Sprintf("%s%s%s%s", bold, cyan, s, reset)
}

// ColorizeWhiteBold returns the string wrapped in bold white color.
func ColorizeWhiteBold(s string) string {
	return fmt.Sprintf("%s%s%s%s", bold, white, s, reset)
}

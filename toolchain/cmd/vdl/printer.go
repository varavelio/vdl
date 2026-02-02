package main

import (
	"fmt"
	"os"

	"github.com/varavelio/vdl/toolchain/internal/util/cliutil"
)

// printSuccess prints a success message in green bold to stdout.
func printSuccess(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	fmt.Println(cliutil.ColorizeGreenBold(msg))
}

// printError prints an error message in red bold to stderr.
func printError(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintln(os.Stderr, cliutil.ColorizeRedBold(msg))
}

// printFatal prints an error message in red bold to stderr and exits with code 1.
func printFatal(format string, args ...any) {
	printError(format, args...)
	os.Exit(1)
}

// printWarn prints a warning message in yellow bold to stderr.
func printWarn(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintln(os.Stderr, cliutil.ColorizeYellowBold(msg))
}

package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/varavelio/tinta"
)

var (
	styleSuccess = tinta.Text().Green().Bold()
	styleError   = tinta.Text().Red().Bold()
	styleWarn    = tinta.Text().Yellow().Bold()

	errorCodePattern  = regexp.MustCompile(`error\[[^\]]+\]`)
	didYouMeanPattern = regexp.MustCompile(`did you mean[^?]+\?`)
)

// printSuccess prints a success message in green bold to stdout.
func printSuccess(format string, args ...any) {
	styleSuccess.Fprintf(os.Stdout, format+"\n", args...)
}

// printError prints an error message in red bold to stderr.
func printError(format string, args ...any) {
	styleError.Fprintf(os.Stderr, format+"\n", args...)
}

// printFatal prints an error message in red bold to stderr and exits with code 1.
func printFatal(format string, args ...any) {
	printError(format, args...)
	os.Exit(1)
}

// printWarn prints a warning message in yellow bold to stderr.
func printWarn(format string, args ...any) {
	styleWarn.Fprintf(os.Stderr, format+"\n", args...)
}

// printVDLError formats and prints a VDL error string to stderr with colors:
// the first line is red bold, error[XXXX] codes are red, and "did you mean...?"
// hints are cyan. Subsequent lines are indented with two spaces.
func printVDLError(errMsg string) {
	errStr := "VDL error: " + errMsg

	if idx := strings.Index(errStr, "\n"); idx != -1 {
		errStr = tinta.Text().Red().Bold().String(errStr[:idx]) + errStr[idx:]
	} else {
		errStr = tinta.Text().Red().Bold().String(errStr)
	}

	errStr = strings.ReplaceAll(errStr, "\n", "\n  ")

	errStr = errorCodePattern.ReplaceAllStringFunc(errStr, func(s string) string {
		return tinta.Text().Red().String(s)
	})

	errStr = didYouMeanPattern.ReplaceAllStringFunc(errStr, func(s string) string {
		return tinta.Text().Cyan().String(s)
	})

	fmt.Fprintf(os.Stderr, "%s\n", errStr)
}

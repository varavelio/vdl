package main

import (
	"os"

	"github.com/varavelio/tinta"
)

var (
	styleSuccess = tinta.Text().Green().Bold()
	styleError   = tinta.Text().Red().Bold()
	styleWarn    = tinta.Text().Yellow().Bold()
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

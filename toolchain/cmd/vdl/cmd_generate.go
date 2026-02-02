package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/varavelio/vdl/toolchain/internal/codegen"
	"github.com/varavelio/vdl/toolchain/internal/util/cliutil"
)

type cmdGenerateArgs struct {
	ConfigPath string `arg:"positional" help:"The config file path (default: vdl.yaml, vdl.yml, .vdl.yaml, .vdl.yml)"`
}

func cmdGenerate(args *cmdGenerateArgs) {
	startTime := time.Now()
	candidates := []string{"vdl.yaml", "vdl.yml", ".vdl.yaml", ".vdl.yml"}

	if args.ConfigPath == "" {
		for _, c := range candidates {
			if _, err := os.Stat(c); err == nil {
				args.ConfigPath = c
				break
			}
		}
	}

	if args.ConfigPath == "" {
		fmt.Fprintf(os.Stderr, "VDL could not find the configuration file (searched: %s)\n", strings.Join(candidates, ", "))
		os.Exit(1)
	}

	fileCount, err := codegen.Run(args.ConfigPath)
	if err != nil {
		errStr := "VDL error: " + err.Error()

		// Make the first line red bold
		if idx := strings.Index(errStr, "\n"); idx != -1 {
			errStr = cliutil.ColorizeRedBold(errStr[:idx]) + errStr[idx:]
		} else {
			errStr = cliutil.ColorizeRedBold(errStr)
		}

		// Add 2 spaces after each newline for better indentation
		errStr = strings.ReplaceAll(errStr, "\n", "\n  ")

		// Paint error[XXXX] patterns in red
		errorCodePattern := regexp.MustCompile(`error\[[^\]]+\]`)
		errStr = errorCodePattern.ReplaceAllStringFunc(errStr, cliutil.ColorizeRed)

		// Make "did you mean ... ?" patterns bold
		didYouMeanPattern := regexp.MustCompile(`did you mean[^?]+\?`)
		errStr = didYouMeanPattern.ReplaceAllStringFunc(errStr, cliutil.ColorizeCyan)

		fmt.Fprintf(os.Stderr, "%s\n", errStr)
		os.Exit(1)
	}

	filesText := "files"
	if fileCount == 1 {
		filesText = "file"
	}

	fmt.Printf("VDL generated %d %s in %s\n", fileCount, filesText, time.Since(startTime))
}

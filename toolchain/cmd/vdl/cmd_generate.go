package main

import (
	"os"
	"strings"
	"time"

	"github.com/varavelio/vdl/toolchain/internal/codegen"
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
		printFatal("VDL could not find the configuration file (searched: %s)", strings.Join(candidates, ", "))
	}

	fileCount, err := codegen.Run(args.ConfigPath)
	if err != nil {
		printVDLError(err.Error())
		os.Exit(1)
	}

	filesText := "files"
	if fileCount == 1 {
		filesText = "file"
	}

	printSuccess("VDL generated %d %s in %s", fileCount, filesText, time.Since(startTime))
}

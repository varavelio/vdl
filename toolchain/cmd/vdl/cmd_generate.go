package main

import (
	"os"
	"time"

	"github.com/varavelio/vdl/toolchain/internal/codegen"
)

type cmdGenerateArgs struct {
	Path string `arg:"positional" help:"Directory or config file path (default: current directory, searching for vdl.config.vdl)"`
}

func cmdGenerate(args *cmdGenerateArgs) {
	startTime := time.Now()
	fileCount, err := codegen.Run(args.Path)
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

package main

import (
	"os"
	"time"

	"github.com/varavelio/vdl/toolchain/internal/codegen"
)

type cmdGenerateArgs struct {
	Path   string `arg:"positional" help:"Directory or config file path (default: current directory, searching for vdl.config.vdl)"`
	DryRun bool   `arg:"--dry-run"  help:"Validate pipeline without writing output files (useful for lint/CI)"`
}

func cmdGenerate(args *cmdGenerateArgs) {
	startTime := time.Now()
	fileCount, err := codegen.Run(args.Path, args.DryRun)
	if err != nil {
		printVDLError(err.Error())
		os.Exit(1)
	}

	filesText := "files"
	if fileCount == 1 {
		filesText = "file"
	}

	if args.DryRun {
		printSuccess(
			"VDL would generate %d %s in %s (dry-run)",
			fileCount,
			filesText,
			time.Since(startTime),
		)
		return
	}

	printSuccess(
		"VDL generated %d %s in %s",
		fileCount,
		filesText,
		time.Since(startTime),
	)
}

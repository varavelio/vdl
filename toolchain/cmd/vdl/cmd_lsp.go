package main

import (
	"fmt"
	"os"
	"slices"

	"github.com/varavelio/vdl/toolchain/internal/lsp"
)

type cmdLSPArgs struct {
	LogPath bool `arg:"--log-path" help:"Print the path to the LSP log file"`
}

func cmdLSP(_ *cmdLSPArgs) {
	// Manual parsing for flags that need to bypass the strict lsp mode
	if slices.Contains(os.Args, "--log-path") {
		fmt.Println(lsp.GetLogFilePath())
		return
	}

	lspInstance := lsp.New(os.Stdin, os.Stdout)
	if err := lspInstance.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "VDL lsp error: %v\n", err)
		os.Exit(1)
	}
}

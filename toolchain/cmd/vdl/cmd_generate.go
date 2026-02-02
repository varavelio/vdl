package main

import (
	"fmt"
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
		fmt.Fprintf(os.Stderr, "VDL could not find the configuration file (searched: %s)\n", strings.Join(candidates, ", "))
		os.Exit(1)
	}

	if err := codegen.Run(args.ConfigPath); err != nil {
		fmt.Fprintf(os.Stderr, "VDL error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("VDL code generation finished in %s\n", time.Since(startTime))
}

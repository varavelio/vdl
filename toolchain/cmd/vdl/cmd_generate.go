package main

import (
	"log"
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
		log.Fatal("VDL: no configuration file found. Searched for: " + strings.Join(candidates, ", "))
	}

	if err := codegen.Run(args.ConfigPath); err != nil {
		log.Fatalf("VDL: failed to run code generator: %s", err)
	}

	log.Printf("VDL: code generation finished in %s", time.Since(startTime))
}

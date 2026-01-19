package main

import (
	"log"
	"time"

	"github.com/varavelio/vdl/toolchain/internal/codegen"
)

type cmdGenerateArgs struct {
	ConfigPath string `arg:"positional" help:"The config file path (default: ./vdl.toml)"`
}

func cmdGenerate(args *cmdGenerateArgs) {
	startTime := time.Now()

	if args.ConfigPath == "" {
		args.ConfigPath = "./vdl.toml"
	}

	if err := codegen.Run(args.ConfigPath); err != nil {
		log.Fatalf("VDL: failed to run code generator: %s", err)
	}

	log.Printf("VDL: code generation finished in %s", time.Since(startTime))
}

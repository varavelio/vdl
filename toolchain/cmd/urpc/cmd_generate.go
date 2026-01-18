package main

import (
	"log"
	"time"

	"github.com/uforg/uforpc/urpc/internal/codegen"
)

type cmdGenerateArgs struct {
	ConfigPath string `arg:"positional" help:"The config file path (default: ./uforpc.toml)"`
}

func cmdGenerate(args *cmdGenerateArgs) {
	startTime := time.Now()

	if args.ConfigPath == "" {
		args.ConfigPath = "./uforpc.toml"
	}

	if err := codegen.Run(args.ConfigPath); err != nil {
		log.Fatalf("UFO RPC: failed to run code generator: %s", err)
	}

	log.Printf("UFO RPC: code generation finished in %s", time.Since(startTime))
}

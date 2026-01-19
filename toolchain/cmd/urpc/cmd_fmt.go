package main

import (
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/varavelio/vdl/toolchain/internal/formatter"
)

type cmdFmtArgs struct {
	Pattern string `arg:"positional" help:"The file pattern to format (support globs e.g. './rpc/**/*.vdl')"`
	Verbose bool   `arg:"-v,--verbose" help:"Verbose output prints all formatted files"`
}

func cmdFmt(args *cmdFmtArgs) {
	var matches []string
	var err error
	startTime := time.Now()

	matches, err = filepath.Glob(args.Pattern)
	if err != nil {
		log.Fatalf("VDL: failed to glob pattern: %s", err)
	}

	for _, match := range matches {
		fileBytes, err := os.ReadFile(match)
		if err != nil {
			log.Fatalf("VDL: failed to read file: %s", err)
		}

		formatted, err := formatter.Format(match, string(fileBytes))
		if err != nil {
			log.Fatalf("VDL: failed to format file: %s", err)
		}

		if err := os.WriteFile(match, []byte(formatted), 0644); err != nil {
			log.Fatalf("VDL: failed to write file: %s", err)
		}

		if args.Verbose {
			log.Println("VDL: formatted", match)
		}
	}

	log.Printf("VDL: formatted %d files in %s", len(matches), time.Since(startTime))
}

package main

import (
	"log"
	"os"
	"strings"
	"time"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/varavelio/vdl/toolchain/internal/formatter"
)

type cmdFormatArgs struct {
	Patterns []string `arg:"positional" help:"The file patterns to format (supports recursive globs e.g. './**/*.vdl')"`
	Verbose  bool     `arg:"-v,--verbose" help:"Verbose output prints all formatted files"`
}

func cmdFmt(args *cmdFormatArgs) {
	var allMatches []string
	startTime := time.Now()

	// Default pattern if none provided
	if len(args.Patterns) == 0 {
		args.Patterns = []string{"./**/*.vdl"}
	}

	for _, pattern := range args.Patterns {
		// Check if the pattern is actually a directory
		info, err := os.Stat(pattern)
		if err == nil && info.IsDir() {
			// If it's a directory, look for .vdl files recursively inside it
			// We ensure the pattern ends with separator before appending **/*.vdl
			dirPattern := pattern
			if !strings.HasSuffix(dirPattern, string(os.PathSeparator)) {
				dirPattern += string(os.PathSeparator)
			}
			dirPattern += "**/*.vdl"

			matches, err := doublestar.FilepathGlob(dirPattern)
			if err != nil {
				log.Printf("VDL: warning: failed to glob directory '%s': %s", pattern, err)
				continue
			}

			allMatches = append(allMatches, matches...)
			continue
		}

		// Normal glob processing (files or glob patterns)
		matches, err := doublestar.FilepathGlob(pattern)
		if err != nil {
			log.Fatalf("VDL: failed to glob pattern '%s': %s", pattern, err)
		}
		allMatches = append(allMatches, matches...)
	}

	// Deduplicate matches to avoid formatting the same file twice
	uniqueMatches := make(map[string]bool)
	var dedupedMatches []string
	for _, m := range allMatches {
		if _, exists := uniqueMatches[m]; !exists {
			uniqueMatches[m] = true
			dedupedMatches = append(dedupedMatches, m)
		}
	}

	formattedCount := 0
	for _, match := range dedupedMatches {
		if !strings.HasSuffix(match, ".vdl") {
			continue
		}

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

		formattedCount++
	}

	log.Printf("VDL: formatted %d files in %s", formattedCount, time.Since(startTime))
}

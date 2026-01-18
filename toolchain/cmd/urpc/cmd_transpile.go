package main

import (
	"encoding/json"
	"log"
	"os"
	"strings"

	"github.com/uforg/uforpc/urpc/internal/schema"
	"github.com/uforg/uforpc/urpc/internal/transpile"
	"github.com/uforg/uforpc/urpc/internal/urpc/formatter"
	"github.com/uforg/uforpc/urpc/internal/urpc/parser"
)

type cmdTranspileArgs struct {
	Path string `arg:"positional" help:"The file to be transpiled, if it ends with '.urpc' it will be transpiled to JSON and if it ends with '.json' it will be transpiled to URPC"`
}

func cmdTranspile(args *cmdTranspileArgs) {
	isJSON := strings.HasSuffix(args.Path, ".json")
	isURPC := strings.HasSuffix(args.Path, ".urpc")

	if !isJSON && !isURPC {
		log.Fatalf("UFO RPC: file must end with '.urpc' or '.json'")
	}

	fileBytes, err := os.ReadFile(args.Path)
	if err != nil {
		log.Fatalf("UFO RPC: failed to read file: %s", err)
	}

	if isJSON {
		parsed, err := schema.ParseSchema(string(fileBytes))
		if err != nil {
			log.Fatalf("UFO RPC: failed to parse JSON schema: %s", err)
		}

		urpc, err := transpile.ToURPC(parsed)
		if err != nil {
			log.Fatalf("UFO RPC: failed to transpile JSON to URPC: %s", err)
		}

		formatted := formatter.FormatSchema(&urpc)
		os.Stdout.WriteString(formatted)
	}

	if isURPC {
		parsed, err := parser.ParserInstance.ParseBytes(args.Path, fileBytes)
		if err != nil {
			log.Fatalf("UFO RPC: failed to parse URPC schema: %s", err)
		}

		jsonSch, err := transpile.ToJSON(*parsed)
		if err != nil {
			log.Fatalf("UFO RPC: failed to transpile URPC to JSON: %s", err)
		}

		jsonBytes, err := json.MarshalIndent(jsonSch, "", "  ")
		if err != nil {
			log.Fatalf("UFO RPC: failed to marshal JSON schema: %s", err)
		}

		os.Stdout.WriteString(string(jsonBytes))
	}
}

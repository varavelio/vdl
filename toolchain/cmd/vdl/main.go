package main

import (
	"fmt"
	"log"
	"os"

	"github.com/alexflint/go-arg"
	"github.com/varavelio/vdl/toolchain/internal/version"
)

type allArgs struct {
	Init     *cmdInitArgs     `arg:"subcommand:init" help:"Initialize a new VDL schema in the specified path"`
	Format   *cmdFormatArgs   `arg:"subcommand:format" help:"Format the VDL schema in the specified path"`
	Generate *cmdGenerateArgs `arg:"subcommand:generate" help:"Generate code from the VDL schema"`
	LSP      *cmdLSPArgs      `arg:"subcommand:lsp" help:"Start the VDL Language Server"`
	Version  *struct{}        `arg:"subcommand:version" help:"Show vdl version information"`
}

func printVersion() {
	fmt.Printf("%s\n\n", version.AsciiArt)
}

func main() {
	// If the LSP is called, then omit the arg parser to avoid taking
	// control of the stdin/stdout because the LSP will need it.
	//
	// If the command is not lsp, then continue with the rest of the logic.
	if len(os.Args) > 1 && os.Args[1] == "lsp" {
		cmdLSP(nil)
		return
	}

	// Check for version flags before argument parsing
	if len(os.Args) > 1 {
		arg := os.Args[1]
		if arg == "--version" || arg == "-v" {
			printVersion()
			return
		}
	}

	var args allArgs
	p, err := arg.NewParser(arg.Config{}, &args)
	if err != nil {
		log.Fatalf("failed to create arg parser: %s", err)
	}

	err = p.Parse(os.Args[1:])
	switch {
	case err == arg.ErrHelp: // indicates that user wrote "--help" on command line
		printVersion()
		p.WriteHelp(os.Stdout)
		os.Exit(0)
	case err != nil:
		fmt.Printf("error: %v\n", err)
		p.WriteUsage(os.Stdout)
		os.Exit(1)
	}

	if args.Init != nil {
		cmdInit(args.Init)
		return
	}

	if args.Format != nil {
		cmdFmt(args.Format)
		return
	}

	if args.Generate != nil {
		cmdGenerate(args.Generate)
		return
	}

	// If no subcommand was specified, show version by default
	printVersion()
}

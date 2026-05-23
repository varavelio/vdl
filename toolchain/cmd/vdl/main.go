package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/alexflint/go-arg"
	"github.com/varavelio/vdl/toolchain/internal/version"
)

type allArgs struct {
	Init     *cmdInitArgs     `arg:"subcommand:init"     help:"Initialize a new VDL project in the specified directory"`
	Format   *cmdFormatArgs   `arg:"subcommand:format"   help:"Format VDL files matching the given glob patterns"`
	Generate *cmdGenerateArgs `arg:"subcommand:generate" help:"Run code generation from a vdl.config.vdl project"`
	Compile  *cmdCompileArgs  `arg:"subcommand:compile"  help:"Compile a VDL file and emit its IR as JSON"`
	LSP      *cmdLSPArgs      `arg:"subcommand:lsp"      help:"Start the VDL Language Server"`
	Version  *struct{}        `arg:"subcommand:version"  help:"Show VDL version information"`
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
		printFatal("VDL failed to create arg parser: %v", err)
	}

	err = p.Parse(os.Args[1:])
	switch {
	case errors.Is(err, arg.ErrHelp): // indicates that user wrote "--help" on command line
		printVersion()
		p.WriteHelp(os.Stdout)
		os.Exit(0)
	case err != nil:
		printError("VDL error: %v", err)
		p.WriteUsage(os.Stderr)
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

	if args.Compile != nil {
		cmdCompile(args.Compile)
		return
	}

	// If no subcommand was specified, show version by default
	printVersion()
}

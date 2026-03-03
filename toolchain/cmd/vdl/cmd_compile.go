package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/varavelio/vdl/toolchain/internal/core/analysis"
	"github.com/varavelio/vdl/toolchain/internal/core/ir"
	"github.com/varavelio/vdl/toolchain/internal/core/vfs"
)

type cmdCompileArgs struct {
	File string `arg:"positional,required" help:"Path to the .vdl file to compile"`
}

func cmdCompile(args *cmdCompileArgs) {
	fs := vfs.New()
	program, diagnostics := analysis.Analyze(fs, args.File)

	if len(diagnostics) > 0 {
		for _, d := range diagnostics {
			printVDLError(d.Error())
		}
		os.Exit(1)
	}

	schema := ir.FromProgram(program)

	data, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		printFatal("VDL error: failed to marshal IR to JSON: %v", err)
	}

	fmt.Println(string(data))
}

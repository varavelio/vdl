package main

import (
	"fmt"
	"os"

	"test/gen"
)

func main() {
	// Verify VDLPaths structure (only procs and streams, no service root path)
	if gen.VDLPaths.MyService.MyProc != "/MyService/MyProc" {
		fail("VDLPaths.MyService.MyProc", "/MyService/MyProc", gen.VDLPaths.MyService.MyProc)
	}
	if gen.VDLPaths.MyService.MyStream != "/MyService/MyStream" {
		fail("VDLPaths.MyService.MyStream", "/MyService/MyStream", gen.VDLPaths.MyService.MyStream)
	}

	// Verify paths using OperationDefinition.Path()
	var foundProc, foundStream bool
	for _, op := range gen.VDLProcedures {
		if op.RPCName == "MyService" && op.Name == "MyProc" {
			foundProc = true
			if op.Path() != "/MyService/MyProc" {
				fail("op.Path() for MyProc", "/MyService/MyProc", op.Path())
			}
		}
	}
	if !foundProc {
		fmt.Fprintf(os.Stderr, "MyProc operation not found in VDLProcedures\n")
		os.Exit(1)
	}

	for _, op := range gen.VDLStreams {
		if op.RPCName == "MyService" && op.Name == "MyStream" {
			foundStream = true
			if op.Path() != "/MyService/MyStream" {
				fail("op.Path() for MyStream", "/MyService/MyStream", op.Path())
			}
		}
	}
	if !foundStream {
		fmt.Fprintf(os.Stderr, "MyStream operation not found in VDLStreams\n")
		os.Exit(1)
	}

	fmt.Println("Paths verification successful")
}

func fail(name, expected, actual string) {
	fmt.Fprintf(os.Stderr, "%s mismatch: expected %q, got %q\n", name, expected, actual)
	os.Exit(1)
}

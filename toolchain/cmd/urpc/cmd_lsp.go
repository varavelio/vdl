package main

import (
	"log"
	"os"

	"github.com/uforg/uforpc/urpc/internal/urpc/lsp"
)

type cmdLSPArgs struct{}

func cmdLSP(_ *cmdLSPArgs) {
	lspInstance := lsp.New(os.Stdin, os.Stdout)
	if err := lspInstance.Run(); err != nil {
		log.Fatalf("failed to run lsp: %s", err)
	}
}

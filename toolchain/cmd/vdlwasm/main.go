//go:build js && wasm

package main

import (
	"log"
	"syscall/js"
)

var wrappers map[string]js.Func = map[string]js.Func{
	// "cmdTranspile":     cmdTranspileWrapper(),
	"cmdFmt":           cmdFmtWrapper(),
	"cmdCodegen":       cmdCodegenWrapper(),
	"cmdExpandTypes":   cmdExpandTypesWrapper(),
	"cmdExtractType":   cmdExtractTypeWrapper(),
	"cmdExtractProc":   cmdExtractProcWrapper(),
	"cmdExtractStream": cmdExtractStreamWrapper(),
}

func main() {
	log.Println("VDL WASM: Initializing...")

	for name, wrapper := range wrappers {
		js.Global().Set(name, wrapper)
	}

	log.Println("VDL WASM: Initialized")
	<-make(chan any)
}

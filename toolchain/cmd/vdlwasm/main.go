//go:build js && wasm

package main

import (
	"log"
	"syscall/js"

	"github.com/varavelio/vdl/toolchain/internal/wasm"
)

const JAVASCRIPT_FUNCTION_NAME = "wasmExecuteFunction"

func main() {
	log.Println("VDL WASM: Initializing...")
	js.Global().Set(JAVASCRIPT_FUNCTION_NAME, jsWrapper())
	log.Println("VDL WASM: Initialized")
	<-make(chan any)
}

func jsWrapper() js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) any {
		return js.Global().Get("Promise").New(js.FuncOf(func(_ js.Value, promArgs []js.Value) any {
			input := args[0].String()
			resolve := promArgs[0]
			reject := promArgs[1]

			go func() {
				if len(args) != 1 {
					reject.Invoke("missing input")
					return
				}
				if output, err := wasm.RunString(input); err != nil {
					reject.Invoke("failed to run wasm function: " + err.Error())
				} else {
					resolve.Invoke(output)
				}
			}()

			return nil
		}))
	})
}

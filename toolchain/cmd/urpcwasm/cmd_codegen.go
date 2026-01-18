//go:build js && wasm

package main

import (
	"encoding/json"
	"syscall/js"

	"github.com/uforg/uforpc/urpc/internal/codegen"
)

func cmdCodegenWrapper() js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) any {
		return js.Global().Get("Promise").New(js.FuncOf(func(_ js.Value, promArgs []js.Value) any {
			resolve := promArgs[0]
			reject := promArgs[1]

			go func() {
				if len(args) != 1 {
					reject.Invoke("missing options argument")
					return
				}
				var opts codegen.RunWasmOptions
				if err := json.Unmarshal([]byte(args[0].String()), &opts); err != nil {
					reject.Invoke("failed to parse options JSON: " + err.Error())
					return
				}
				if code, err := codegen.RunWasmString(opts); err != nil {
					reject.Invoke("failed to generate: " + err.Error())
				} else {
					resolve.Invoke(code)
				}
			}()

			return nil
		}))
	})
}

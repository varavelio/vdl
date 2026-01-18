//go:build js && wasm

package main

import (
	"syscall/js"

	"github.com/varavelio/vdl/toolchain/internal/core/transform"
)

/*
	Expands all custom type references to inline objects in the URPC schema.

	Available command:
	cmdExpandTypes(input: string): Promise<string>

	Example:
	const expanded = await cmdExpandTypes(schemaContent);
*/

func cmdExpandTypesWrapper() js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) any {
		return js.Global().Get("Promise").New(js.FuncOf(func(_ js.Value, promArgs []js.Value) any {
			resolve := promArgs[0]
			reject := promArgs[1]

			go func() {
				if len(args) != 1 {
					reject.Invoke("missing input")
					return
				}
				if expanded, err := cmdExpandTypes(args[0].String()); err != nil {
					reject.Invoke("failed to expand types: " + err.Error())
				} else {
					resolve.Invoke(expanded)
				}
			}()

			return nil
		}))
	})
}

func cmdExpandTypes(input string) (string, error) {
	return transform.ExpandTypesStr("schema.urpc", input)
}

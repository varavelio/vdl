//go:build js && wasm

package main

import (
	"fmt"
	"syscall/js"

	"github.com/uforg/uforpc/urpc/internal/urpc/formatter"
)

func cmdFmtWrapper() js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) any {
		return js.Global().Get("Promise").New(js.FuncOf(func(_ js.Value, promArgs []js.Value) any {
			resolve := promArgs[0]
			reject := promArgs[1]

			go func() {
				if len(args) != 1 {
					reject.Invoke("missing input")
					return
				}
				if formatted, err := cmdFmt(args[0].String()); err != nil {
					reject.Invoke("failed to format: " + err.Error())
				} else {
					resolve.Invoke(formatted)
				}
			}()

			return nil
		}))
	})
}

func cmdFmt(input string) (string, error) {
	formatted, err := formatter.Format("schema.urpc", input)
	if err != nil {
		return "", fmt.Errorf("failed to format file: %w", err)
	}
	return formatted, nil
}

//go:build js && wasm

package main

import (
	"syscall/js"

	"github.com/varavelio/vdl/toolchain/internal/transform"
)

/*
	Extract commands for VDL schema manipulation.

	Available commands:

	cmdExtractType(input: string, typeName: string): Promise<string>
		Extracts a specific type declaration from the schema by name.
		Example:
			const userType = await cmdExtractType(schemaContent, "User");

	cmdExtractProc(input: string, procName: string): Promise<string>
		Extracts a specific proc declaration from the schema by name.
		Example:
			const getUser = await cmdExtractProc(schemaContent, "GetUser");

	cmdExtractStream(input: string, streamName: string): Promise<string>
		Extracts a specific stream declaration from the schema by name.
		Example:
			const chatStream = await cmdExtractStream(schemaContent, "ChatStream");
*/

func cmdExtractTypeWrapper() js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) any {
		return js.Global().Get("Promise").New(js.FuncOf(func(_ js.Value, promArgs []js.Value) any {
			resolve := promArgs[0]
			reject := promArgs[1]

			go func() {
				if len(args) != 2 {
					reject.Invoke("missing input and/or type name")
					return
				}
				if extracted, err := cmdExtractType(args[0].String(), args[1].String()); err != nil {
					reject.Invoke("failed to extract type: " + err.Error())
				} else {
					resolve.Invoke(extracted)
				}
			}()

			return nil
		}))
	})
}

func cmdExtractType(input, typeName string) (string, error) {
	return transform.ExtractTypeStr("schema.vdl", input, typeName)
}

func cmdExtractProcWrapper() js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) any {
		return js.Global().Get("Promise").New(js.FuncOf(func(_ js.Value, promArgs []js.Value) any {
			resolve := promArgs[0]
			reject := promArgs[1]

			go func() {
				if len(args) != 2 {
					reject.Invoke("missing input and/or proc name")
					return
				}
				if extracted, err := cmdExtractProc(args[0].String(), args[1].String(), args[2].String()); err != nil {
					reject.Invoke("failed to extract proc: " + err.Error())
				} else {
					resolve.Invoke(extracted)
				}
			}()

			return nil
		}))
	})
}

func cmdExtractProc(input, rpcName string, procName string) (string, error) {
	return transform.ExtractProcStr("schema.vdl", input, rpcName, procName)
}

func cmdExtractStreamWrapper() js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) any {
		return js.Global().Get("Promise").New(js.FuncOf(func(_ js.Value, promArgs []js.Value) any {
			resolve := promArgs[0]
			reject := promArgs[1]

			go func() {
				if len(args) != 2 {
					reject.Invoke("missing input and/or stream name")
					return
				}
				if extracted, err := cmdExtractStream(args[0].String(), args[1].String(), args[2].String()); err != nil {
					reject.Invoke("failed to extract stream: " + err.Error())
				} else {
					resolve.Invoke(extracted)
				}
			}()

			return nil
		}))
	})
}

func cmdExtractStream(input, rpcName string, streamName string) (string, error) {
	return transform.ExtractStreamStr("schema.vdl", input, rpcName, streamName)
}

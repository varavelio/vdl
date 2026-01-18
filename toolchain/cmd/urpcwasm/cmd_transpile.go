//go:build js && wasm

package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"syscall/js"

	"github.com/uforg/uforpc/urpc/internal/schema"
	"github.com/uforg/uforpc/urpc/internal/transpile"
	"github.com/uforg/uforpc/urpc/internal/urpc/formatter"
	"github.com/uforg/uforpc/urpc/internal/urpc/parser"
)

func cmdTranspileWrapper() js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) any {
		return js.Global().Get("Promise").New(js.FuncOf(func(_ js.Value, promArgs []js.Value) any {
			resolve := promArgs[0]
			reject := promArgs[1]

			go func() {
				if len(args) != 2 {
					reject.Invoke("missing original extension and/or input")
					return
				}
				if transpiled, err := cmdTranspile(args[0].String(), args[1].String()); err != nil {
					reject.Invoke("failed to transpile: " + err.Error())
				} else {
					resolve.Invoke(transpiled)
				}
			}()

			return nil
		}))
	})
}

func cmdTranspile(originalExtension string, input string) (string, error) {
	isJSON := strings.HasSuffix(originalExtension, "json")
	isURPC := strings.HasSuffix(originalExtension, "urpc")

	if !isJSON && !isURPC {
		return "", fmt.Errorf("original extension must be '.urpc' or '.json'")
	}

	if isJSON {
		parsed, err := schema.ParseSchema(input)
		if err != nil {
			return "", fmt.Errorf("failed to parse JSON schema: %s", err)
		}

		urpc, err := transpile.ToURPC(parsed)
		if err != nil {
			return "", fmt.Errorf("failed to transpile JSON to URPC: %s", err)
		}

		formatted := formatter.FormatSchema(&urpc)
		return formatted, nil
	}

	if isURPC {
		parsed, err := parser.ParserInstance.ParseString("schema.urpc", input)
		if err != nil {
			return "", fmt.Errorf("failed to parse URPC schema: %s", err)
		}

		jsonSch, err := transpile.ToJSON(*parsed)
		if err != nil {
			return "", fmt.Errorf("failed to transpile URPC to JSON: %s", err)
		}

		jsonBytes, err := json.MarshalIndent(jsonSch, "", "  ")
		if err != nil {
			return "", fmt.Errorf("failed to marshal JSON schema: %s", err)
		}

		return string(jsonBytes), nil
	}

	return "", nil
}

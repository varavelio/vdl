package wasm

import (
	"encoding/json"
	"fmt"

	"github.com/varavelio/vdl/toolchain/internal/wasm/wasmtypes"
)

// RunString takes an input as a JSON string, runs the WASM logic and returns the
// output as a JSON string
func RunString(input string) (string, error) {
	typedInput := wasmtypes.WasmInput{}
	if err := json.Unmarshal([]byte(input), &typedInput); err != nil {
		return "", fmt.Errorf("error unmarshaling JSON input: %w", err)
	}

	typedOutput, err := Run(typedInput)
	if err != nil {
		return "", fmt.Errorf("error running the function: %w", err)
	}

	output, err := json.Marshal(typedOutput)
	if err != nil {
		return "", fmt.Errorf("error marshaling JSON output: %w", err)
	}

	return string(output), nil
}

// Run runs the WASM logic
func Run(input wasmtypes.WasmInput) (wasmtypes.WasmOutput, error) {
	return wasmtypes.WasmOutput{}, nil
}

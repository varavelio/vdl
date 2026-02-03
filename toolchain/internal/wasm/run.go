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
func Run(input wasmtypes.WasmInput) (*wasmtypes.WasmOutput, error) {
	switch input.FunctionName {
	case wasmtypes.WasmFunctionNameExpandTypes:
		out, err := runExpandTypes(input.GetExpandTypes())
		if err != nil {
			return nil, fmt.Errorf("error while running expand types function: %w", err)
		}
		return &wasmtypes.WasmOutput{ExpandTypes: out}, nil

	case wasmtypes.WasmFunctionNameExtractType:
		out, err := runExtractType(input.GetExtractType())
		if err != nil {
			return nil, fmt.Errorf("error while running extract type function: %w", err)
		}
		return &wasmtypes.WasmOutput{ExtractType: out}, nil

	case wasmtypes.WasmFunctionNameExtractProc:
		out, err := runExtractProc(input.GetExtractProc())
		if err != nil {
			return nil, fmt.Errorf("error while running extract procedure function: %w", err)
		}
		return &wasmtypes.WasmOutput{ExtractProc: out}, nil

	case wasmtypes.WasmFunctionNameExtractStream:
		out, err := runExtractStream(input.GetExtractStream())
		if err != nil {
			return nil, fmt.Errorf("error while running extract stream function: %w", err)
		}
		return &wasmtypes.WasmOutput{ExtractStream: out}, nil

	case wasmtypes.WasmFunctionNameIrgen:
		out, err := runIrgen(input.GetIrgen())
		if err != nil {
			return nil, fmt.Errorf("error while running irgen function: %w", err)
		}
		return &wasmtypes.WasmOutput{Irgen: out}, nil

	case wasmtypes.WasmFunctionNameCodegen:
		out, err := runCodegen(input.GetCodegen())
		if err != nil {
			return nil, fmt.Errorf("error while running codegen function: %w", err)
		}
		return &wasmtypes.WasmOutput{Codegen: out}, nil
	}

	return nil, fmt.Errorf("%s function is not supported", input.FunctionName)
}

package wasm

import (
	"fmt"

	"github.com/varavelio/vdl/toolchain/internal/transform"
	"github.com/varavelio/vdl/toolchain/internal/wasm/wasmtypes"
)

func runExtractStream(input wasmtypes.ExtractStreamInput) (*wasmtypes.ExtractStreamOutput, error) {
	extracted, err := transform.ExtractStreamStr("schema.vdl", input.VdlSchema, input.RpcName, input.StreamName)
	if err != nil {
		return nil, fmt.Errorf("error extracting procedure: %w", err)
	}

	return &wasmtypes.ExtractStreamOutput{StreamSchema: extracted}, nil
}

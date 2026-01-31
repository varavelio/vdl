package wasm

import (
	"fmt"

	"github.com/varavelio/vdl/toolchain/internal/transform"
	"github.com/varavelio/vdl/toolchain/internal/wasm/wasmtypes"
)

func runExtractProc(input wasmtypes.ExtractProcInput) (*wasmtypes.ExtractProcOutput, error) {
	extracted, err := transform.ExtractProcStr("schema.vdl", input.VdlSchema, input.RpcName, input.ProcName)
	if err != nil {
		return nil, fmt.Errorf("error extracting procedure: %w", err)
	}

	return &wasmtypes.ExtractProcOutput{ProcSchema: extracted}, nil
}

package wasm

import (
	"fmt"

	"github.com/varavelio/vdl/toolchain/internal/transform"
	"github.com/varavelio/vdl/toolchain/internal/wasm/wasmtypes"
)

func runExtractType(input wasmtypes.ExtractTypeInput) (*wasmtypes.ExtractTypeOutput, error) {
	extracted, err := transform.ExtractTypeStr("schema.vdl", input.VdlSchema, input.TypeName)
	if err != nil {
		return nil, fmt.Errorf("error extracting type: %w", err)
	}

	return &wasmtypes.ExtractTypeOutput{TypeSchema: extracted}, nil
}

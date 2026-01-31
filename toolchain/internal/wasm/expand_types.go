package wasm

import (
	"fmt"

	"github.com/varavelio/vdl/toolchain/internal/transform"
	"github.com/varavelio/vdl/toolchain/internal/wasm/wasmtypes"
)

func runExpandTypes(input wasmtypes.ExpandTypesInput) (*wasmtypes.ExpandTypesOutput, error) {
	expanded, err := transform.ExpandTypesStr("schema.vdl", input.VdlSchema)
	if err != nil {
		return nil, fmt.Errorf("error expanding types: %w", err)
	}

	return &wasmtypes.ExpandTypesOutput{ExpandedSchema: expanded}, nil
}

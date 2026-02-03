package python

import (
	_ "embed"

	"github.com/varavelio/vdl/toolchain/internal/codegen/config/configtypes"
	"github.com/varavelio/vdl/toolchain/internal/core/ir/irtypes"
)

//go:embed pieces/core.py
var coreTypesContent string

func generateCore(_ *irtypes.IrSchema, _ *configtypes.PythonTargetConfig) (string, error) {
	return coreTypesContent, nil
}

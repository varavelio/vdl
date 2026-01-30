package python

import (
	_ "embed"

	"github.com/varavelio/vdl/toolchain/internal/codegen/config"
	"github.com/varavelio/vdl/toolchain/internal/core/ir"
)

//go:embed pieces/core_types.py
var coreTypesContent string

func generateCore(schema *ir.Schema, cfg *config.PythonConfig) (string, error) {
	return coreTypesContent, nil
}

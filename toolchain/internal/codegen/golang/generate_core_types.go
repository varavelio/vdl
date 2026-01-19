package golang

import (
	_ "embed"
	"fmt"

	"github.com/varavelio/vdl/toolchain/internal/core/ir"
	"github.com/varavelio/vdl/toolchain/internal/util/strutil"
)

//go:embed pieces/core_types.go
var coreTypesRawPiece string

func generateCoreTypes(_ *ir.Schema, _ *flatSchema, _ Config) (string, error) {
	piece := strutil.GetStrAfter(coreTypesRawPiece, "/** START FROM HERE **/")
	if piece == "" {
		return "", fmt.Errorf("core_types.go: could not find start delimiter")
	}
	return piece, nil
}

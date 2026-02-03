package golang

import (
	_ "embed"
	"fmt"

	"github.com/varavelio/vdl/toolchain/internal/codegen/config"
	"github.com/varavelio/vdl/toolchain/internal/core/ir/irtypes"
	"github.com/varavelio/vdl/toolchain/internal/util/strutil"
)

//go:embed pieces/core.go
var coreTypesRawPiece string

func generateCoreTypes(_ *irtypes.IrSchema, _ *config.GoConfig) (string, error) {
	piece := strutil.GetStrAfter(coreTypesRawPiece, "/** START FROM HERE **/")
	if piece == "" {
		return "", fmt.Errorf("core.go: could not find start delimiter")
	}
	return piece, nil
}

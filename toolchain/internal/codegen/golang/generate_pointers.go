package golang

import (
	_ "embed"
	"fmt"

	"github.com/varavelio/vdl/toolchain/internal/codegen/config"
	"github.com/varavelio/vdl/toolchain/internal/core/ir/irtypes"
	"github.com/varavelio/vdl/toolchain/internal/util/strutil"
)

//go:embed pieces/pointers.go
var pointersRawPiece string

func generatePointers(_ *irtypes.IrSchema, _ *config.GoConfig) (string, error) {
	piece := strutil.GetStrAfter(pointersRawPiece, "/** START FROM HERE **/")
	if piece == "" {
		return "", fmt.Errorf("pointers.go: could not find start delimiter")
	}
	return piece, nil
}

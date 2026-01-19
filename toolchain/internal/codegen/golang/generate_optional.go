package golang

import (
	_ "embed"
	"fmt"

	"github.com/varavelio/vdl/toolchain/internal/core/ir"
	"github.com/varavelio/vdl/toolchain/internal/util/strutil"
)

//go:embed pieces/optional.go
var optionalRawPiece string

func generateOptional(_ *ir.Schema, _ *flatSchema, _ Config) (string, error) {
	piece := strutil.GetStrAfter(optionalRawPiece, "/** START FROM HERE **/")
	if piece == "" {
		return "", fmt.Errorf("optional.go: could not find start delimiter")
	}
	return piece, nil
}

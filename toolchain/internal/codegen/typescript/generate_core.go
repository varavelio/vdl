package typescript

import (
	_ "embed"
	"fmt"

	"github.com/varavelio/vdl/toolchain/internal/codegen/config"
	"github.com/varavelio/vdl/toolchain/internal/core/ir"
	"github.com/varavelio/vdl/toolchain/internal/util/strutil"
)

//go:embed pieces/core.ts
var coreTypesRawPiece string

func generateCoreTypes(_ *ir.Schema, _ *config.TypeScriptConfig) (string, error) {
	piece := strutil.GetStrAfter(coreTypesRawPiece, "/** START FROM HERE **/")
	if piece == "" {
		return "", fmt.Errorf("core.ts: could not find start delimiter")
	}
	return piece, nil
}

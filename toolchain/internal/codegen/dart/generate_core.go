package dart

import (
	_ "embed"

	"github.com/varavelio/vdl/toolchain/internal/codegen/config"
	"github.com/varavelio/vdl/toolchain/internal/core/ir/irtypes"
)

//go:embed pieces/core.dart
var coreTypesRawPiece string

// generateCore returns the core types content (Response, VdlError).
// The header is added by the caller.
func generateCore(_ *irtypes.IrSchema, _ *config.DartConfig) (string, error) {
	return coreTypesRawPiece, nil
}

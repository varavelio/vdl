package golang

import (
	_ "embed"
	"fmt"

	"github.com/uforg/uforpc/urpc/internal/schema"
	"github.com/uforg/uforpc/urpc/internal/util/strutil"
)

//go:embed pieces/core_types.go
var coreTypesRawPiece string

func generateCoreTypes(_ schema.Schema, _ Config) (string, error) {
	piece := strutil.GetStrAfter(coreTypesRawPiece, "/** START FROM HERE **/")
	if piece == "" {
		return "", fmt.Errorf("core_types.go: could not find start delimiter")
	}
	return piece, nil
}

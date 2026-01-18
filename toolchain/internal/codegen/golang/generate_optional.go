package golang

import (
	_ "embed"
	"fmt"

	"github.com/uforg/uforpc/urpc/internal/schema"
	"github.com/uforg/uforpc/urpc/internal/util/strutil"
)

//go:embed pieces/optional.go
var optionalRawPiece string

func generateOptional(_ schema.Schema, _ Config) (string, error) {
	piece := strutil.GetStrAfter(optionalRawPiece, "/** START FROM HERE **/")
	if piece == "" {
		return "", fmt.Errorf("optional.go: could not find start delimiter")
	}
	return piece, nil
}

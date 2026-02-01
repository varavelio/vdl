package typescript

import (
	_ "embed"
	"fmt"

	"github.com/varavelio/gen"
	"github.com/varavelio/vdl/toolchain/internal/codegen/config"
	"github.com/varavelio/vdl/toolchain/internal/util/strutil"
)

//go:embed pieces/adapters/fetch.ts
var fetchAdapterRawPiece string

//go:embed pieces/adapters/node.ts
var nodeAdapterRawPiece string

// generateFetchAdapter generates the Fetch API compatible adapter.
func generateFetchAdapter(cfg *config.TypeScriptConfig) (string, error) {
	piece := strutil.GetStrAfter(fetchAdapterRawPiece, "/** START FROM HERE **/")
	if piece == "" {
		return "", fmt.Errorf("adapters/fetch.ts: could not find start delimiter")
	}

	g := gen.New().WithSpaces(2)

	// Imports
	generateImport(g, []string{"HTTPAdapter"}, "../server", true, cfg)
	generateImport(g, []string{"Server"}, "../server", false, cfg)
	g.Break()

	// Core adapter piece
	g.Raw(piece)

	return g.String(), nil
}

// generateNodeAdapter generates the Node.js HTTP compatible adapter.
func generateNodeAdapter(cfg *config.TypeScriptConfig) (string, error) {
	piece := strutil.GetStrAfter(nodeAdapterRawPiece, "/** START FROM HERE **/")
	if piece == "" {
		return "", fmt.Errorf("adapters/node.ts: could not find start delimiter")
	}

	g := gen.New().WithSpaces(2)

	// Imports
	generateImport(g, []string{"HTTPAdapter"}, "../server", true, cfg)
	generateImport(g, []string{"Server"}, "../server", false, cfg)
	g.Line("import type { IncomingMessage, ServerResponse } from \"node:http\";")
	g.Break()

	// Core adapter piece
	g.Raw(piece)

	return g.String(), nil
}

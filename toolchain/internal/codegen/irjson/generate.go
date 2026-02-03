package irjson

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/varavelio/vdl/toolchain/internal/codegen/config/configtypes"
	"github.com/varavelio/vdl/toolchain/internal/core/ir/irtypes"
)

// File represents a generated file.
type File struct {
	RelativePath string
	Content      []byte
}

// Generator implements the IR JSON generator.
type Generator struct {
	config *configtypes.IrConfig
}

// New creates a new IR JSON generator with the given config.
func New(config *configtypes.IrConfig) *Generator {
	return &Generator{config: config}
}

// Name returns the generator name.
func (g *Generator) Name() string {
	return "ir"
}

// Generate produces an IR JSON file from the IR schema.
func (g *Generator) Generate(ctx context.Context, schema *irtypes.IrSchema) ([]File, error) {
	cfg := g.config

	// Encode schema to JSON
	var content []byte
	var err error

	if cfg.Minify != nil && *cfg.Minify {
		// Minified output (no indentation)
		content, err = json.Marshal(schema)
	} else {
		// Pretty-printed output with 2-space indentation (default)
		content, err = json.MarshalIndent(schema, "", "  ")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to encode IR schema to JSON: %w", err)
	}

	// Determine output filename
	filename := "ir.json"
	if cfg.Filename != nil && *cfg.Filename != "" {
		filename = *cfg.Filename
	}

	return []File{
		{
			RelativePath: filename,
			Content:      content,
		},
	}, nil
}

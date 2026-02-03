package vdl

import (
	"context"

	"github.com/varavelio/vdl/toolchain/internal/codegen/config/configtypes"
	"github.com/varavelio/vdl/toolchain/internal/core/ir/irtypes"
)

// File represents a generated file.
type File struct {
	RelativePath string
	Content      []byte
}

// Generator implements the unified VDL schema generator.
type Generator struct {
	config          *configtypes.VdlTargetConfig
	formattedSchema string
}

// New creates a new VDL schema generator with the given config and pre-formatted schema.
func New(config *configtypes.VdlTargetConfig, formattedSchema string) *Generator {
	return &Generator{
		config:          config,
		formattedSchema: formattedSchema,
	}
}

// Name returns the generator name.
func (g *Generator) Name() string {
	return "vdl"
}

// Generate produces a unified VDL schema file.
// The schema parameter is included for interface consistency but the actual output
// comes from the pre-merged and formatted schema string.
func (g *Generator) Generate(ctx context.Context, schema *irtypes.IrSchema) ([]File, error) {
	cfg := g.config

	// Determine output filename
	filename := "schema.vdl"
	if cfg.Filename != nil && *cfg.Filename != "" {
		filename = *cfg.Filename
	}

	return []File{
		{
			RelativePath: filename,
			Content:      []byte(g.formattedSchema),
		},
	}, nil
}

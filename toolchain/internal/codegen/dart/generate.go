package dart

import (
	"context"
	_ "embed"
	"strings"

	"github.com/varavelio/gen"
	"github.com/varavelio/vdl/toolchain/internal/codegen/config"
	"github.com/varavelio/vdl/toolchain/internal/core/ir"
	"github.com/varavelio/vdl/toolchain/internal/util/strutil"
)

// File represents a generated file. This mirrors codegen.File to avoid import cycles.
type File struct {
	RelativePath string
	Content      []byte
}

// Generator implements the Dart code generator.
type Generator struct {
	config *config.DartConfig
}

// New creates a new Dart generator with the given config.
func New(config *config.DartConfig) *Generator {
	return &Generator{config: config}
}

// Name returns the generator name.
func (g *Generator) Name() string {
	return "dart"
}

// Generate produces Dart source files from the IR schema.
func (g *Generator) Generate(ctx context.Context, schema *ir.Schema) ([]File, error) {
	subGenerators := []func(*ir.Schema, *config.DartConfig) (string, error){
		generateCore,
		generateEnums,
		generateConstants,
		generatePatterns,
		generateDomainTypes,
		generateProcedureTypes,
		generateStreamTypes,
	}

	// 1) Generate lib/client.dart
	builder := gen.New().WithSpaces(2)
	for _, generator := range subGenerators {
		codeChunk, err := generator(schema, g.config)
		if err != nil {
			return nil, err
		}

		codeChunk = strings.TrimSpace(codeChunk)
		if codeChunk != "" {
			builder.Raw(codeChunk)
			builder.Break()
			builder.Break()
		}
	}
	libClientContent := builder.String()
	libClientContent = strutil.LimitConsecutiveNewlines(libClientContent, 2)

	dartClient := File{
		RelativePath: "client.dart",
		Content:      []byte(libClientContent),
	}

	return []File{
		dartClient,
	}, nil
}

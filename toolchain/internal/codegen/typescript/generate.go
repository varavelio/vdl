package typescript

import (
	"context"
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

// Generator implements the TypeScript code generator.
type Generator struct {
	config *config.TypeScriptConfig
}

// New creates a new TypeScript generator with the given config.
func New(config *config.TypeScriptConfig) *Generator {
	return &Generator{config: config}
}

// Name returns the generator name.
func (g *Generator) Name() string {
	return "typescript"
}

// Generate produces TypeScript source files from the IR schema.
func (g *Generator) Generate(ctx context.Context, schema *ir.Schema) ([]File, error) {
	subGenerators := []func(*ir.Schema, *config.TypeScriptConfig) (string, error){
		generateCoreTypes,
		generateEnums,
		generateConstants,
		generatePatterns,
		generateDomainTypes,
		generateProcedureTypes,
		generateStreamTypes,
		generateClient,
	}

	builder := gen.New().WithSpaces(2)
	for _, generator := range subGenerators {
		codeChunk, err := generator(schema, g.config)
		if err != nil {
			return nil, err
		}

		codeChunk = strings.TrimSpace(codeChunk)
		if codeChunk == "" {
			continue
		}
		builder.Raw(codeChunk)
		builder.Break()
		builder.Break()
	}

	generatedCode := builder.String()
	generatedCode = strutil.LimitConsecutiveNewlines(generatedCode, 2)

	outputFile := g.config.Output
	if outputFile == "" {
		outputFile = "vdl.ts"
	}

	return []File{
		{
			RelativePath: outputFile,
			Content:      []byte(generatedCode),
		},
	}, nil
}

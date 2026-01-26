package golang

import (
	"context"
	"strings"

	"github.com/varavelio/gen"
	"github.com/varavelio/vdl/toolchain/internal/codegen/config"
	"github.com/varavelio/vdl/toolchain/internal/core/ir"
	"golang.org/x/tools/imports"
)

// File represents a generated file. This mirrors codegen.File to avoid import cycles.
type File struct {
	RelativePath string
	Content      []byte
}

// Generator implements the Go code generator.
type Generator struct {
	config *config.GoConfig
}

// New creates a new Go generator with the given config.
func New(config *config.GoConfig) *Generator {
	return &Generator{config: config}
}

// Name returns the generator name.
func (g *Generator) Name() string {
	return "golang"
}

// Generate produces Go source files from the IR schema.
func (g *Generator) Generate(ctx context.Context, schema *ir.Schema) ([]File, error) {
	// Map of filename -> generator
	builders := make(map[string]*gen.Generator)

	// Helper to get or create a builder for a file
	getBuilder := func(filename string) *gen.Generator {
		if b, ok := builders[filename]; ok {
			return b
		}
		b := gen.New().WithTabs()

		// Always start with package declaration
		pkgCode := generatePackage(schema, g.config)
		b.Raw(pkgCode)
		b.Break()

		builders[filename] = b
		return b
	}

	// Core Types (core_types.go)
	code, err := generateCoreTypes(schema, g.config)
	if err != nil {
		return nil, err
	}
	if code = strings.TrimSpace(code); code != "" {
		b := getBuilder("core_types.go")
		b.Raw(code)
		b.Break()
	}

	// Optional utility type (optional.go)
	code, err = generateOptional(schema, g.config)
	if err != nil {
		return nil, err
	}
	if code = strings.TrimSpace(code); code != "" {
		b := getBuilder("optional.go")
		b.Raw(code)
		b.Break()
	}

	// Types (types.go) - Domain types
	// Contains: Enums, Domain Types, Procedure Types, Stream Types
	typeGenerators := []func(*ir.Schema, *config.GoConfig) (string, error){
		generateEnums,
		generateDomainTypes,
		generateProcedureTypes,
		generateStreamTypes,
	}
	for _, generator := range typeGenerators {
		code, err := generator(schema, g.config)
		if err != nil {
			return nil, err
		}
		if code = strings.TrimSpace(code); code != "" {
			b := getBuilder("types.go")
			b.Raw(code)
			b.Break()
		}
	}

	// Constants (consts.go)
	code, err = generateConstants(schema, g.config)
	if err != nil {
		return nil, err
	}
	if code = strings.TrimSpace(code); code != "" {
		b := getBuilder("consts.go")
		b.Raw(code)
		b.Break()
	}

	// Patterns (patterns.go)
	code, err = generatePatterns(schema, g.config)
	if err != nil {
		return nil, err
	}
	if code = strings.TrimSpace(code); code != "" {
		b := getBuilder("patterns.go")
		b.Raw(code)
		b.Break()
	}

	// RPC Server (rpc_server.go) - Core + All RPCs
	code, err = generateServerCore(schema, g.config)
	if err != nil {
		return nil, err
	}
	if code = strings.TrimSpace(code); code != "" {
		b := getBuilder("rpc_server.go")
		b.Raw(code)
		b.Break()
	}
	for _, rpc := range schema.RPCs {
		code, err := generateServerRPC(rpc, g.config)
		if err != nil {
			return nil, err
		}
		if code = strings.TrimSpace(code); code != "" {
			b := getBuilder("rpc_server.go")
			b.Raw(code)
			b.Break()
		}
	}

	// RPC Client (rpc_client.go) - Core + All RPCs
	code, err = generateClientCore(schema, g.config)
	if err != nil {
		return nil, err
	}
	if code = strings.TrimSpace(code); code != "" {
		b := getBuilder("rpc_client.go")
		b.Raw(code)
		b.Break()
	}
	for _, rpc := range schema.RPCs {
		code, err := generateClientRPC(rpc, g.config)
		if err != nil {
			return nil, err
		}
		if code = strings.TrimSpace(code); code != "" {
			b := getBuilder("rpc_client.go")
			b.Raw(code)
			b.Break()
		}
	}

	// Convert builders to files
	var files []File
	for filename, builder := range builders {
		content := builder.String()

		// Format code
		formattedCode, err := imports.Process("", []byte(content), nil)
		if err == nil {
			content = string(formattedCode)
		}

		files = append(files, File{
			RelativePath: filename,
			Content:      []byte(content),
		})
	}

	return files, nil
}

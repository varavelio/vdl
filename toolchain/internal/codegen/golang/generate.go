package golang

import (
	"context"
	"strings"

	"github.com/varavelio/gen"
	"github.com/varavelio/vdl/toolchain/internal/codegen/config"
	"github.com/varavelio/vdl/toolchain/internal/core/ir/irtypes"
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
func (g *Generator) Generate(ctx context.Context, schema *irtypes.IrSchema) ([]File, error) {
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

	// Core Types (core.go)
	code, err := generateCoreTypes(schema, g.config)
	if err != nil {
		return nil, err
	}
	if code = strings.TrimSpace(code); code != "" {
		b := getBuilder("core.go")
		b.Raw(code)
		b.Break()
	}

	// Pointer utility functions (pointers.go)
	code, err = generatePointers(schema, g.config)
	if err != nil {
		return nil, err
	}
	if code = strings.TrimSpace(code); code != "" {
		b := getBuilder("pointers.go")
		b.Raw(code)
		b.Break()
	}

	// Types (types.go) - Domain types
	// Contains: Enums, Domain Types, Procedure Types, Stream Types
	typeGenerators := []func(*irtypes.IrSchema, *config.GoConfig) (string, error){
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

	// Constants (constants.go)
	code, err = generateConstants(schema, g.config)
	if err != nil {
		return nil, err
	}
	if code = strings.TrimSpace(code); code != "" {
		b := getBuilder("constants.go")
		b.Raw(code)
		b.Break()
	}

	// RPC Catalog (catalog.go)
	code, err = generateRPCCatalog(schema, g.config)
	if err != nil {
		return nil, err
	}
	if code = strings.TrimSpace(code); code != "" {
		b := getBuilder("catalog.go")
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

	// RPC Server (server.go) - Core + All RPCs
	code, err = generateServerCore(schema, g.config)
	if err != nil {
		return nil, err
	}
	if code = strings.TrimSpace(code); code != "" {
		b := getBuilder("server.go")
		b.Raw(code)
		b.Break()
	}

	// Build maps of procedures and streams by RPC name for efficient lookup
	rpcProcs := make(map[string][]irtypes.ProcedureDef)
	rpcStreams := make(map[string][]irtypes.StreamDef)
	for _, proc := range schema.Procedures {
		rpcProcs[proc.RpcName] = append(rpcProcs[proc.RpcName], proc)
	}
	for _, stream := range schema.Streams {
		rpcStreams[stream.RpcName] = append(rpcStreams[stream.RpcName], stream)
	}

	for _, rpc := range schema.Rpcs {
		code, err := generateServerRPC(rpc.Name, rpcProcs[rpc.Name], rpcStreams[rpc.Name], g.config)
		if err != nil {
			return nil, err
		}
		if code = strings.TrimSpace(code); code != "" {
			b := getBuilder("server.go")
			b.Raw(code)
			b.Break()
		}
	}

	// RPC Client (client.go) - Core + All RPCs
	code, err = generateClientCore(schema, g.config)
	if err != nil {
		return nil, err
	}
	if code = strings.TrimSpace(code); code != "" {
		b := getBuilder("client.go")
		b.Raw(code)
		b.Break()
	}
	for _, rpc := range schema.Rpcs {
		code, err := generateClientRPC(rpc.Name, rpcProcs[rpc.Name], rpcStreams[rpc.Name], g.config)
		if err != nil {
			return nil, err
		}
		if code = strings.TrimSpace(code); code != "" {
			b := getBuilder("client.go")
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

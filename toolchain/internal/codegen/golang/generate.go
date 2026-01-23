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
	// Flatten the schema for easier iteration
	flat := flattenSchema(schema)

	// Map of filename -> generator
	builders := make(map[string]*gen.Generator)

	// Helper to get or create a builder for a file
	getBuilder := func(filename string) *gen.Generator {
		if b, ok := builders[filename]; ok {
			return b
		}
		b := gen.New().WithTabs()

		// Always start with package declaration
		pkgCode := generatePackage(schema, flat, g.config)
		b.Raw(pkgCode)
		b.Break()

		builders[filename] = b
		return b
	}

	// Core Types (types_core.go)
	code, err := generateCoreTypes(schema, flat, g.config)
	if err != nil {
		return nil, err
	}
	if code = strings.TrimSpace(code); code != "" {
		b := getBuilder("types_core.go")
		b.Raw(code)
		b.Break()
	}

	// Optional utility type (optional.go)
	code, err = generateOptional(schema, flat, g.config)
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
	typeGenerators := []func(*ir.Schema, *flatSchema, *config.GoConfig) (string, error){
		generateEnums,
		generateDomainTypes,
		generateProcedureTypes,
		generateStreamTypes,
	}

	for _, generator := range typeGenerators {
		code, err := generator(schema, flat, g.config)
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
	code, err = generateConstants(schema, flat, g.config)
	if err != nil {
		return nil, err
	}
	if code = strings.TrimSpace(code); code != "" {
		b := getBuilder("consts.go")
		b.Raw(code)
		b.Break()
	}

	// Patterns (patterns.go)
	code, err = generatePatterns(schema, flat, g.config)
	if err != nil {
		return nil, err
	}
	if code = strings.TrimSpace(code); code != "" {
		b := getBuilder("patterns.go")
		b.Raw(code)
		b.Break()
	}

	// RPC Server (rpc_server.go) - Core + All RPCs
	code, err = generateServerCore(schema, flat, g.config)
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
	code, err = generateClientCore(schema, flat, g.config)
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

// flatSchema provides pre-computed flattened views of the schema for easier iteration.
// This avoids nested loops throughout the generators.
type flatSchema struct {
	// Procedures contains all procedures from all RPCs with their parent RPC name.
	Procedures []flatProcedure
	// Streams contains all streams from all RPCs with their parent RPC name.
	Streams []flatStream
}

// flatProcedure represents a procedure with its parent RPC context.
type flatProcedure struct {
	RPCName   string
	Procedure ir.Procedure
}

// flatStream represents a stream with its parent RPC context.
type flatStream struct {
	RPCName string
	Stream  ir.Stream
}

// flattenSchema creates flattened views of procedures and streams for easier iteration.
func flattenSchema(schema *ir.Schema) *flatSchema {
	flat := &flatSchema{
		Procedures: []flatProcedure{},
		Streams:    []flatStream{},
	}

	for _, rpc := range schema.RPCs {
		for _, proc := range rpc.Procs {
			flat.Procedures = append(flat.Procedures, flatProcedure{
				RPCName:   rpc.Name,
				Procedure: proc,
			})
		}
		for _, stream := range rpc.Streams {
			flat.Streams = append(flat.Streams, flatStream{
				RPCName: rpc.Name,
				Stream:  stream,
			})
		}
	}

	return flat
}

// fullProcName returns the fully qualified procedure name: {RPC}{Proc}
func fullProcName(rpcName, procName string) string {
	return rpcName + procName
}

// fullStreamName returns the fully qualified stream name: {RPC}{Stream}
func fullStreamName(rpcName, streamName string) string {
	return rpcName + streamName
}

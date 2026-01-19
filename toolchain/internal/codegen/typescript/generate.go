package typescript

import (
	"context"
	"strings"

	"github.com/varavelio/gen"
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
	config Config
}

// New creates a new TypeScript generator with the given config.
func New(config Config) *Generator {
	return &Generator{config: config}
}

// Name returns the generator name.
func (g *Generator) Name() string {
	return "typescript"
}

// Generate produces TypeScript source files from the IR schema.
func (g *Generator) Generate(ctx context.Context, schema *ir.Schema) ([]File, error) {
	// Flatten the schema for easier iteration
	flat := flattenSchema(schema)

	subGenerators := []func(*ir.Schema, *flatSchema, Config) (string, error){
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
		codeChunk, err := generator(schema, flat, g.config)
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

	outputFile := g.config.OutputFile
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

// rpcProcPath returns the RPC path for a procedure: {rpcName}/{procName}
func rpcProcPath(rpcName, procName string) string {
	return strutil.ToCamelCase(rpcName) + "/" + strutil.ToCamelCase(procName)
}

// rpcStreamPath returns the RPC path for a stream: {rpcName}/{streamName}
func rpcStreamPath(rpcName, streamName string) string {
	return strutil.ToCamelCase(rpcName) + "/" + strutil.ToCamelCase(streamName)
}

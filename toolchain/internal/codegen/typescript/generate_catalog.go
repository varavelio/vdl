package typescript

import (
	"github.com/varavelio/gen"
	"github.com/varavelio/vdl/toolchain/internal/codegen/config"
	"github.com/varavelio/vdl/toolchain/internal/core/ir"
)

// generateRPCCatalog generates introspection data: VDLProcedures, VDLStreams, and VDLPaths.
func generateRPCCatalog(schema *ir.Schema, config *config.TypeScriptConfig) (string, error) {
	if len(schema.RPCs) == 0 {
		return "", nil
	}

	g := gen.New().WithSpaces(2)

	generateImport(g, []string{"OperationDefinition"}, "./core", true, config)
	g.Break()

	g.Line("// -----------------------------------------------------------------------------")
	g.Line("// VDL RPC Catalog")
	g.Line("// -----------------------------------------------------------------------------")
	g.Break()

	// VDLProcedures
	g.Line("/**")
	g.Line(" * VDLProcedures is a list of all procedure definitions.")
	g.Line(" *")
	g.Line(" * It allows introspection of RPC procedures at runtime.")
	g.Line(" */")
	g.Line("export const VDLProcedures: OperationDefinition[] = [")
	g.Block(func() {
		for _, rpc := range schema.RPCs {
			for _, proc := range rpc.Procs {
				g.Linef("{ rpcName: \"%s\", name: \"%s\", type: \"proc\" },", rpc.Name, proc.Name)
			}
		}
	})
	g.Line("];")
	g.Break()

	// VDLStreams
	g.Line("/**")
	g.Line(" * VDLStreams is a list of all stream definitions.")
	g.Line(" *")
	g.Line(" * It allows introspection of RPC streams at runtime.")
	g.Line(" */")
	g.Line("export const VDLStreams: OperationDefinition[] = [")
	g.Block(func() {
		for _, rpc := range schema.RPCs {
			for _, stream := range rpc.Streams {
				g.Linef("{ rpcName: \"%s\", name: \"%s\", type: \"stream\" },", rpc.Name, stream.Name)
			}
		}
	})
	g.Line("];")
	g.Break()

	// VDLPaths
	g.Line("/**")
	g.Line(" * VDLPaths holds the URL paths for all RPCs and their operations.")
	g.Line(" *")
	g.Line(" * It provides type-safe access to the URL paths.")
	g.Line(" */")
	g.Line("export const VDLPaths = {")
	g.Block(func() {
		for _, rpc := range schema.RPCs {
			g.Linef("%s: {", rpc.Name)
			g.Block(func() {
				for _, proc := range rpc.Procs {
					g.Linef("%s: \"/%s/%s\",", proc.Name, rpc.Name, proc.Name)
				}
				for _, stream := range rpc.Streams {
					g.Linef("%s: \"/%s/%s\",", stream.Name, rpc.Name, stream.Name)
				}
			})
			g.Line("},")
		}
	})
	g.Line("} as const;")
	g.Break()

	return g.String(), nil
}

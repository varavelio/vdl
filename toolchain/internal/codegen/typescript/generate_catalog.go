package typescript

import (
	"github.com/varavelio/gen"
	"github.com/varavelio/vdl/toolchain/internal/codegen/config/configtypes"
	"github.com/varavelio/vdl/toolchain/internal/core/ir/irtypes"
)

// generateCatalog generates introspection data: VDLProcedures, VDLStreams, and VDLPaths.
func generateCatalog(schema *irtypes.IrSchema, config *configtypes.TypeScriptConfig) (string, error) {
	if len(schema.Rpcs) == 0 {
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
		for _, proc := range schema.Procedures {
			g.Linef("{ rpcName: \"%s\", name: \"%s\", type: \"proc\" },", proc.RpcName, proc.Name)
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
		for _, stream := range schema.Streams {
			g.Linef("{ rpcName: \"%s\", name: \"%s\", type: \"stream\" },", stream.RpcName, stream.Name)
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
		for _, rpc := range schema.Rpcs {
			g.Linef("%s: {", rpc.Name)
			g.Block(func() {
				for _, proc := range schema.Procedures {
					if proc.RpcName == rpc.Name {
						g.Linef("%s: \"/%s/%s\",", proc.Name, rpc.Name, proc.Name)
					}
				}
				for _, stream := range schema.Streams {
					if stream.RpcName == rpc.Name {
						g.Linef("%s: \"/%s/%s\",", stream.Name, rpc.Name, stream.Name)
					}
				}
			})
			g.Line("},")
		}
	})
	g.Line("} as const;")
	g.Break()

	return g.String(), nil
}

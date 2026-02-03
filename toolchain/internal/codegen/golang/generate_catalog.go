package golang

import (
	"github.com/varavelio/gen"
	"github.com/varavelio/vdl/toolchain/internal/codegen/config"
	"github.com/varavelio/vdl/toolchain/internal/core/ir/irtypes"
)

// generateRPCCatalog generates introspection data: VDLProcedures, VDLStreams, and VDLPaths.
func generateRPCCatalog(schema *irtypes.IrSchema, _ *config.GoConfig) (string, error) {
	if len(schema.Rpcs) == 0 {
		return "", nil
	}

	// Build a map of RPC name to procedures and streams for efficient lookup
	rpcProcs := make(map[string][]irtypes.ProcedureDef)
	rpcStreams := make(map[string][]irtypes.StreamDef)
	for _, proc := range schema.Procedures {
		rpcProcs[proc.RpcName] = append(rpcProcs[proc.RpcName], proc)
	}
	for _, stream := range schema.Streams {
		rpcStreams[stream.RpcName] = append(rpcStreams[stream.RpcName], stream)
	}

	g := gen.New().WithTabs()

	g.Line("// -----------------------------------------------------------------------------")
	g.Line("// VDL RPC Catalog")
	g.Line("// -----------------------------------------------------------------------------")
	g.Break()

	// VDLProcedures
	g.Line("// VDLProcedures is a list of all procedure definitions.")
	g.Line("//")
	g.Line("// It allows introspection of RPC procedures at runtime.")
	g.Line("var VDLProcedures = []OperationDefinition{")
	g.Block(func() {
		for _, proc := range schema.Procedures {
			g.Linef("{RPCName: %q, Name: %q, Type: OperationTypeProc},", proc.RpcName, proc.Name)
		}
	})
	g.Line("}")
	g.Break()

	// VDLStreams
	g.Line("// VDLStreams is a list of all stream definitions.")
	g.Line("//")
	g.Line("// It allows introspection of RPC streams at runtime.")
	g.Line("var VDLStreams = []OperationDefinition{")
	g.Block(func() {
		for _, stream := range schema.Streams {
			g.Linef("{RPCName: %q, Name: %q, Type: OperationTypeStream},", stream.RpcName, stream.Name)
		}
	})
	g.Line("}")
	g.Break()

	// VDLPaths
	g.Line("// VDLPaths holds the URL paths for all RPCs and their operations.")
	g.Line("//")
	g.Line("// It provides type-safe access to the URL paths.")
	g.Line("var VDLPaths = struct {")
	g.Block(func() {
		for _, rpc := range schema.Rpcs {
			g.Linef("%s struct {", rpc.Name)
			g.Block(func() {
				for _, proc := range rpcProcs[rpc.Name] {
					g.Linef("%s string", proc.Name)
				}
				for _, stream := range rpcStreams[rpc.Name] {
					g.Linef("%s string", stream.Name)
				}
			})
			g.Line("}")
		}
	})
	g.Line("}{")
	g.Block(func() {
		for _, rpc := range schema.Rpcs {
			g.Linef("%s: struct {", rpc.Name)
			g.Block(func() {
				for _, proc := range rpcProcs[rpc.Name] {
					g.Linef("%s string", proc.Name)
				}
				for _, stream := range rpcStreams[rpc.Name] {
					g.Linef("%s string", stream.Name)
				}
			})
			g.Line("}{")
			g.Block(func() {
				for _, proc := range rpcProcs[rpc.Name] {
					g.Linef("%s: \"/%s/%s\",", proc.Name, rpc.Name, proc.Name)
				}
				for _, stream := range rpcStreams[rpc.Name] {
					g.Linef("%s: \"/%s/%s\",", stream.Name, rpc.Name, stream.Name)
				}
			})
			g.Line("},")
		}
	})
	g.Line("}")
	g.Break()

	return g.String(), nil
}

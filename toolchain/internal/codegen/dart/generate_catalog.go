package dart

import (
	"github.com/varavelio/gen"
	"github.com/varavelio/vdl/toolchain/internal/codegen/config/configtypes"
	"github.com/varavelio/vdl/toolchain/internal/core/ir/irtypes"
)

// generateRPCCatalog generates introspection data: VDLProcedures, VDLStreams, and VDLPaths.
func generateRPCCatalog(schema *irtypes.IrSchema, _ *configtypes.DartConfig) (string, error) {
	if len(schema.Rpcs) == 0 {
		return "", nil
	}

	g := gen.New().WithSpaces(2)

	g.Line("// -----------------------------------------------------------------------------")
	g.Line("// VDL RPC Catalog")
	g.Line("// -----------------------------------------------------------------------------")
	g.Break()

	// OperationType enum
	g.Line("/// OperationType defines the type of RPC operation.")
	g.Line("enum OperationType {")
	g.Block(func() {
		g.Line("proc,")
		g.Line("stream;")
	})
	g.Line("}")
	g.Break()

	// OperationDefinition class
	g.Line("/// OperationDefinition describes an RPC operation.")
	g.Line("class OperationDefinition {")
	g.Block(func() {
		g.Line("/// The name of the RPC service.")
		g.Line("final String rpcName;")
		g.Line("/// The name of the operation.")
		g.Line("final String name;")
		g.Line("/// The type of operation (proc or stream).")
		g.Line("final OperationType type;")
		g.Break()
		g.Line("const OperationDefinition({")
		g.Block(func() {
			g.Line("required this.rpcName,")
			g.Line("required this.name,")
			g.Line("required this.type,")
		})
		g.Line("});")
		g.Break()
		g.Line("/// Returns the full path for this operation.")
		g.Line("String get path => '/$rpcName/$name';")
	})
	g.Line("}")
	g.Break()

	// VDLProcedures - now using flattened schema.Procedures
	g.Line("/// VDLProcedures is a list of all procedure definitions.")
	g.Line("///")
	g.Line("/// It allows introspection of RPC procedures at runtime.")
	g.Line("const List<OperationDefinition> vdlProcedures = [")
	g.Block(func() {
		for _, proc := range schema.Procedures {
			g.Linef("OperationDefinition(rpcName: '%s', name: '%s', type: OperationType.proc),", proc.RpcName, proc.Name)
		}
	})
	g.Line("];")
	g.Break()

	// VDLStreams - now using flattened schema.Streams
	g.Line("/// VDLStreams is a list of all stream definitions.")
	g.Line("///")
	g.Line("/// It allows introspection of RPC streams at runtime.")
	g.Line("const List<OperationDefinition> vdlStreams = [")
	g.Block(func() {
		for _, stream := range schema.Streams {
			g.Linef("OperationDefinition(rpcName: '%s', name: '%s', type: OperationType.stream),", stream.RpcName, stream.Name)
		}
	})
	g.Line("];")
	g.Break()

	// VDLPaths class
	g.Line("/// VDLPaths provides type-safe access to all RPC operation paths.")
	g.Line("abstract class VDLPaths {")
	g.Block(func() {
		for _, rpc := range schema.Rpcs {
			g.Linef("/// Paths for the %s RPC.", rpc.Name)
			g.Linef("static const %s = _%sPaths._();", lowercaseFirst(rpc.Name), rpc.Name)
		}
	})
	g.Line("}")
	g.Break()

	// Generate path classes for each RPC
	for _, rpc := range schema.Rpcs {
		g.Linef("class _%sPaths {", rpc.Name)
		g.Block(func() {
			g.Linef("const _%sPaths._();", rpc.Name)
			g.Break()
			// Find procedures for this RPC
			for _, proc := range schema.Procedures {
				if proc.RpcName == rpc.Name {
					g.Linef("/// Path for the %s procedure.", proc.Name)
					g.Linef("String get %s => '/%s/%s';", lowercaseFirst(proc.Name), rpc.Name, proc.Name)
				}
			}
			// Find streams for this RPC
			for _, stream := range schema.Streams {
				if stream.RpcName == rpc.Name {
					g.Linef("/// Path for the %s stream.", stream.Name)
					g.Linef("String get %s => '/%s/%s';", lowercaseFirst(stream.Name), rpc.Name, stream.Name)
				}
			}
		})
		g.Line("}")
		g.Break()
	}

	return g.String(), nil
}

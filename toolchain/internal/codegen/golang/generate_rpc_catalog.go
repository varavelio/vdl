package golang

import (
	"github.com/varavelio/gen"
	"github.com/varavelio/vdl/toolchain/internal/codegen/config"
	"github.com/varavelio/vdl/toolchain/internal/core/ir"
)

// generateRPCCatalog generates introspection data: VDLProcedures, VDLStreams, and VDLPaths.
func generateRPCCatalog(schema *ir.Schema, _ *config.GoConfig) (string, error) {
	if len(schema.RPCs) == 0 {
		return "", nil
	}

	g := gen.New().WithTabs()

	g.Line("// -----------------------------------------------------------------------------")
	g.Line("// VDL RPC Catalog")
	g.Line("// -----------------------------------------------------------------------------")
	g.Break()

	// VDLProcedures
	g.Line("// VDLProcedures is a list of all procedure definitions.")
	g.Line("var VDLProcedures = []OperationDefinition{")
	g.Block(func() {
		for _, rpc := range schema.RPCs {
			for _, proc := range rpc.Procs {
				g.Linef("{RPCName: %q, Name: %q, Type: OperationTypeProc},", rpc.Name, proc.Name)
			}
		}
	})
	g.Line("}")
	g.Break()

	// VDLStreams
	g.Line("// VDLStreams is a list of all stream definitions.")
	g.Line("var VDLStreams = []OperationDefinition{")
	g.Block(func() {
		for _, rpc := range schema.RPCs {
			for _, stream := range rpc.Streams {
				g.Linef("{RPCName: %q, Name: %q, Type: OperationTypeStream},", rpc.Name, stream.Name)
			}
		}
	})
	g.Line("}")
	g.Break()

	// VDLPaths
	g.Line("// VDLPaths holds the URL paths for all RPCs and their operations.")
	g.Line("var VDLPaths = struct {")
	g.Block(func() {
		for _, rpc := range schema.RPCs {
			g.Linef("%s struct {", rpc.Name)
			g.Block(func() {
				for _, proc := range rpc.Procs {
					g.Linef("%s string", proc.Name)
				}
				for _, stream := range rpc.Streams {
					g.Linef("%s string", stream.Name)
				}
			})
			g.Line("}")
		}
	})
	g.Line("}{")
	g.Block(func() {
		for _, rpc := range schema.RPCs {
			g.Linef("%s: struct {", rpc.Name)
			g.Block(func() {
				for _, proc := range rpc.Procs {
					g.Linef("%s string", proc.Name)
				}
				for _, stream := range rpc.Streams {
					g.Linef("%s string", stream.Name)
				}
			})
			g.Line("}{")
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
	g.Line("}")
	g.Break()

	return g.String(), nil
}

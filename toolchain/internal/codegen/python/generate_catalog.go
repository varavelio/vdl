package python

import (
	"github.com/varavelio/gen"
	"github.com/varavelio/vdl/toolchain/internal/codegen/config/configtypes"
	"github.com/varavelio/vdl/toolchain/internal/core/ir/irtypes"
	"github.com/varavelio/vdl/toolchain/internal/util/strutil"
)

func generateRPCCatalog(schema *irtypes.IrSchema, _ *configtypes.PythonTargetConfig) (string, error) {
	if len(schema.Rpcs) == 0 {
		return "", nil
	}

	g := gen.New()
	g.Line("from dataclasses import dataclass")
	g.Line("from enum import Enum")
	g.Line("from typing import Any, List, Type")
	g.Break()

	// Import all types
	g.Line("from .types import *")
	g.Break()

	g.Line("class OperationType(Enum):")
	g.Line("    PROC = 'proc'")
	g.Line("    STREAM = 'stream'")
	g.Break()

	g.Line("@dataclass")
	g.Line("class OperationDefinition:")
	g.Line("    rpc_name: str")
	g.Line("    name: str")
	g.Line("    type: OperationType")
	g.Break()
	g.Line("    @property")
	g.Line("    def path(self) -> str:")
	g.Line("        return f'/{self.rpc_name}/{self.name}'")
	g.Break()

	// VDLPaths
	g.Line("class VDLPaths:")
	if len(schema.Rpcs) == 0 {
		g.Line("    pass")
	}

	// Group procedures and streams by RPC name
	procsByRpc := make(map[string][]irtypes.ProcedureDef)
	for _, proc := range schema.Procedures {
		procsByRpc[proc.RpcName] = append(procsByRpc[proc.RpcName], proc)
	}
	streamsByRpc := make(map[string][]irtypes.StreamDef)
	for _, stream := range schema.Streams {
		streamsByRpc[stream.RpcName] = append(streamsByRpc[stream.RpcName], stream)
	}

	for _, rpc := range schema.Rpcs {
		g.Linef("    class %s:", strutil.ToSnakeCase(rpc.Name))
		hasOperations := false
		for _, proc := range procsByRpc[rpc.Name] {
			path := "/" + rpc.Name + "/" + proc.Name
			g.Linef("        %s = %q", strutil.ToSnakeCase(proc.Name), path)
			hasOperations = true
		}
		for _, stream := range streamsByRpc[rpc.Name] {
			path := "/" + rpc.Name + "/" + stream.Name
			g.Linef("        %s = %q", strutil.ToSnakeCase(stream.Name), path)
			hasOperations = true
		}
		if !hasOperations {
			g.Line("        pass")
		}
		g.Break()
	}
	if len(schema.Rpcs) > 0 {
		for _, rpc := range schema.Rpcs {
			rpcName := strutil.ToSnakeCase(rpc.Name)
			g.Linef("    %s = %s()", rpcName, rpcName)
		}
	}
	g.Break()

	// VDL_PROCEDURES
	g.Line("VDL_PROCEDURES: List[OperationDefinition] = [")
	for _, proc := range schema.Procedures {
		g.Line("    OperationDefinition(")
		g.Linef("        rpc_name=%q,", proc.RpcName)
		g.Linef("        name=%q,", proc.Name)
		g.Line("        type=OperationType.PROC,")
		g.Line("    ),")
	}
	g.Line("]")
	g.Break()

	// VDL_STREAMS
	g.Line("VDL_STREAMS: List[OperationDefinition] = [")
	for _, stream := range schema.Streams {
		g.Line("    OperationDefinition(")
		g.Linef("        rpc_name=%q,", stream.RpcName)
		g.Linef("        name=%q,", stream.Name)
		g.Line("        type=OperationType.STREAM,")
		g.Line("    ),")
	}
	g.Line("]")

	return g.String(), nil
}

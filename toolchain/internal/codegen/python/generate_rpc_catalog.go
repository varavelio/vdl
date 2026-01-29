package python

import (
	"github.com/varavelio/gen"
	"github.com/varavelio/vdl/toolchain/internal/codegen/config"
	"github.com/varavelio/vdl/toolchain/internal/core/ir"
	"github.com/varavelio/vdl/toolchain/internal/util/strutil"
)

func generateRPCCatalog(schema *ir.Schema, cfg *config.PythonConfig) (string, error) {
	g := gen.New()
	g.Line("from dataclasses import dataclass")
	g.Line("from typing import Any, List, Type")
	g.Break()

	// Import all types
	g.Line("from .types import *")
	g.Break()

	g.Line("@dataclass")
	g.Line("class OperationDefinition:")
	g.Line("    name: str")
	g.Line("    path: str")
	g.Line("    input_type: Type[Any]")
	g.Line("    output_type: Type[Any]")
	g.Line("    doc: str")
	g.Line("    is_stream: bool")
	g.Break()

	// VDLPaths
	g.Line("class VDLPaths:")
	if len(schema.Procedures) == 0 && len(schema.Streams) == 0 {
		g.Line("    pass")
	}
	for _, proc := range schema.Procedures {
		pathName := strutil.ToPascalCase(proc.RPCName) + "_" + strutil.ToPascalCase(proc.Name)
		g.Linef("    %s = %q", pathName, proc.Path())
	}
	for _, stream := range schema.Streams {
		pathName := strutil.ToPascalCase(stream.RPCName) + "_" + strutil.ToPascalCase(stream.Name)
		g.Linef("    %s = %q", pathName, stream.Path())
	}
	g.Break()

	// VDL_PROCEDURES
	g.Line("VDL_PROCEDURES: List[OperationDefinition] = [")
	for _, proc := range schema.Procedures {
		fullName := proc.FullName()
		g.Line("    OperationDefinition(")
		g.Linef("        name=%q,", fullName)
		g.Linef("        path=VDLPaths.%s_%s,", strutil.ToPascalCase(proc.RPCName), strutil.ToPascalCase(proc.Name))
		g.Linef("        input_type=%sInput,", fullName)
		g.Linef("        output_type=%sOutput,", fullName)
		g.Linef("        doc=%q,", proc.Doc)
		g.Line("        is_stream=False,")
		g.Line("    ),")
	}
	g.Line("]")
	g.Break()

	// VDL_STREAMS
	g.Line("VDL_STREAMS: List[OperationDefinition] = [")
	for _, stream := range schema.Streams {
		fullName := stream.FullName()
		g.Line("    OperationDefinition(")
		g.Linef("        name=%q,", fullName)
		g.Linef("        path=VDLPaths.%s_%s,", strutil.ToPascalCase(stream.RPCName), strutil.ToPascalCase(stream.Name))
		g.Linef("        input_type=%sInput,", fullName)
		g.Linef("        output_type=%sOutput,", fullName)
		g.Linef("        doc=%q,", stream.Doc)
		g.Line("        is_stream=True,")
		g.Line("    ),")
	}
	g.Line("]")

	return g.String(), nil
}

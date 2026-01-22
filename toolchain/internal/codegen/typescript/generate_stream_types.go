package typescript

import (
	"fmt"

	"github.com/varavelio/gen"
	"github.com/varavelio/vdl/toolchain/internal/codegen/config"
	"github.com/varavelio/vdl/toolchain/internal/core/ir"
)

func generateStreamTypes(_ *ir.Schema, flat *flatSchema, _ *config.TypeScriptConfig) (string, error) {
	if len(flat.Streams) == 0 {
		return "", nil
	}

	g := gen.New().WithSpaces(2)

	g.Line("// -----------------------------------------------------------------------------")
	g.Line("// Stream Types")
	g.Line("// -----------------------------------------------------------------------------")
	g.Break()

	for _, fs := range flat.Streams {
		stream := fs.Stream
		fullName := fullStreamName(fs.RPCName, stream.Name)

		inputName := fmt.Sprintf("%sInput", fullName)
		outputName := fmt.Sprintf("%sOutput", fullName)
		responseName := fmt.Sprintf("%sResponse", fullName)

		inputDesc := fmt.Sprintf("%s represents the input parameters for the %s stream.", inputName, fullName)
		outputDesc := fmt.Sprintf("%s represents the output parameters for the %s stream.", outputName, fullName)
		responseDesc := fmt.Sprintf("%s represents the response for the %s stream.", responseName, fullName)

		g.Line(renderType("", inputName, inputDesc, stream.Input))
		g.Break()

		g.Line(renderType("", outputName, outputDesc, stream.Output))
		g.Break()

		g.Line(renderHydrateType("", outputName, stream.Output))
		g.Break()

		g.Linef("// %s", responseDesc)
		g.Linef("export type %s = Response<%s>", responseName, outputName)
		g.Break()
	}

	// Generate stream names list
	g.Line("// vdlStreamNames is a list of all stream names.")
	g.Line("const vdlStreamNames: string[] = [")
	g.Block(func() {
		for _, fs := range flat.Streams {
			path := rpcStreamPath(fs.RPCName, fs.Stream.Name)
			g.Linef("\"%s\",", path)
		}
	})
	g.Line("]")
	g.Break()

	return g.String(), nil
}

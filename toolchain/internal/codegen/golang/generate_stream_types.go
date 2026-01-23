package golang

import (
	"fmt"

	"github.com/varavelio/gen"
	"github.com/varavelio/vdl/toolchain/internal/codegen/config"
	"github.com/varavelio/vdl/toolchain/internal/core/ir"
)

func generateStreamTypes(_ *ir.Schema, flat *flatSchema, _ *config.GoConfig) (string, error) {
	if len(flat.Streams) == 0 {
		return "", nil
	}

	g := gen.New().WithTabs()

	g.Line("// -----------------------------------------------------------------------------")
	g.Line("// Stream Types")
	g.Line("// -----------------------------------------------------------------------------")
	g.Break()

	for _, fs := range flat.Streams {
		streamName := fullStreamName(fs.RPCName, fs.Stream.Name)
		inputName := fmt.Sprintf("%sInput", streamName)
		outputName := fmt.Sprintf("%sOutput", streamName)
		responseName := fmt.Sprintf("%sResponse", streamName)

		inputDesc := fmt.Sprintf("%s represents the input parameters for the %s.%s stream.", inputName, fs.RPCName, fs.Stream.Name)
		outputDesc := fmt.Sprintf("%s represents the output parameters for the %s.%s stream.", outputName, fs.RPCName, fs.Stream.Name)
		responseDesc := fmt.Sprintf("%s represents the response for the %s.%s stream.", responseName, fs.RPCName, fs.Stream.Name)

		g.Line(renderType("", inputName, inputDesc, fs.Stream.Input))
		g.Break()

		g.Line(renderPreType("", inputName, fs.Stream.Input))
		g.Break()

		g.Line(renderType("", outputName, outputDesc, fs.Stream.Output))
		g.Break()

		g.Linef("// %s", responseDesc)
		g.Linef("type %s = Response[%s]", responseName, outputName)
		g.Break()
	}

	// Generate list of all stream names
	g.Line("// VDLStreamNames is a list of all stream names.")
	g.Line("var VDLStreamNames = []string{")
	g.Block(func() {
		for _, fs := range flat.Streams {
			streamName := fullStreamName(fs.RPCName, fs.Stream.Name)
			g.Linef("%q,", streamName)
		}
	})
	g.Line("}")
	g.Break()

	return g.String(), nil
}

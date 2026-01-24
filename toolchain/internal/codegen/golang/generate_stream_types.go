package golang

import (
	"fmt"

	"github.com/varavelio/gen"
	"github.com/varavelio/vdl/toolchain/internal/codegen/config"
	"github.com/varavelio/vdl/toolchain/internal/core/ir"
)

func generateStreamTypes(schema *ir.Schema, _ *config.GoConfig) (string, error) {
	if len(schema.Streams) == 0 {
		return "", nil
	}

	g := gen.New().WithTabs()

	g.Line("// -----------------------------------------------------------------------------")
	g.Line("// Stream Types")
	g.Line("// -----------------------------------------------------------------------------")
	g.Break()

	for _, stream := range schema.Streams {
		streamName := stream.FullName()
		inputName := fmt.Sprintf("%sInput", streamName)
		outputName := fmt.Sprintf("%sOutput", streamName)
		responseName := fmt.Sprintf("%sResponse", streamName)

		inputDesc := fmt.Sprintf("%s represents the input parameters for the %s.%s stream.", inputName, stream.RPCName, stream.Name)
		outputDesc := fmt.Sprintf("%s represents the output parameters for the %s.%s stream.", outputName, stream.RPCName, stream.Name)
		responseDesc := fmt.Sprintf("%s represents the response for the %s.%s stream.", responseName, stream.RPCName, stream.Name)

		g.Line(renderType("", inputName, inputDesc, stream.Input))
		g.Break()

		g.Line(renderPreType("", inputName, stream.Input))
		g.Break()

		g.Line(renderType("", outputName, outputDesc, stream.Output))
		g.Break()

		g.Linef("// %s", responseDesc)
		g.Linef("type %s = Response[%s]", responseName, outputName)
		g.Break()
	}

	// Generate list of all stream names
	g.Line("// VDLStreamNames is a list of all stream definitions.")
	g.Line("var VDLStreamNames = []OperationDefinition{")
	g.Block(func() {
		for _, stream := range schema.Streams {
			g.Linef("{RPCName: %q, Name: %q, Type: OperationTypeStream},", stream.RPCName, stream.Name)
		}
	})
	g.Line("}")
	g.Break()

	return g.String(), nil
}

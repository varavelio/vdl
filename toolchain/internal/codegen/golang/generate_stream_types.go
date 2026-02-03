package golang

import (
	"fmt"

	"github.com/varavelio/gen"
	"github.com/varavelio/vdl/toolchain/internal/codegen/config/configtypes"
	"github.com/varavelio/vdl/toolchain/internal/core/ir/irtypes"
)

func generateStreamTypes(schema *irtypes.IrSchema, _ *configtypes.GoTargetConfig) (string, error) {
	if len(schema.Streams) == 0 {
		return "", nil
	}

	g := gen.New().WithTabs()

	g.Line("// -----------------------------------------------------------------------------")
	g.Line("// Stream Types")
	g.Line("// -----------------------------------------------------------------------------")
	g.Break()

	for _, stream := range schema.Streams {
		streamName := stream.RpcName + stream.Name
		inputName := fmt.Sprintf("%sInput", streamName)
		outputName := fmt.Sprintf("%sOutput", streamName)
		responseName := fmt.Sprintf("%sResponse", streamName)

		inputDesc := fmt.Sprintf("%s represents the input parameters for the %s.%s stream.", inputName, stream.RpcName, stream.Name)
		outputDesc := fmt.Sprintf("%s represents the output parameters for the %s.%s stream.", outputName, stream.RpcName, stream.Name)
		responseDesc := fmt.Sprintf("%s represents the response for the %s.%s stream.", responseName, stream.RpcName, stream.Name)

		if stream.GetDoc() != "" {
			inputDesc += "\n\n" + stream.GetDoc()
			outputDesc += "\n\n" + stream.GetDoc()
		}

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

	return g.String(), nil
}

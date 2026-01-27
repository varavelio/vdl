package typescript

import (
	"fmt"

	"github.com/varavelio/gen"
	"github.com/varavelio/vdl/toolchain/internal/codegen/config"
	"github.com/varavelio/vdl/toolchain/internal/core/ir"
)

func generateStreamTypes(schema *ir.Schema, _ *config.TypeScriptConfig) (string, error) {
	g := gen.New().WithSpaces(2)

	if len(schema.Streams) > 0 {
		g.Line("// -----------------------------------------------------------------------------")
		g.Line("// Stream Types")
		g.Line("// -----------------------------------------------------------------------------")
		g.Break()

		for _, stream := range schema.Streams {
			fullName := stream.FullName()

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
	}

	return g.String(), nil
}

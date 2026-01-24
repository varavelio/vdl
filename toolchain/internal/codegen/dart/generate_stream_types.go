package dart

import (
	"fmt"

	"github.com/varavelio/gen"
	"github.com/varavelio/vdl/toolchain/internal/codegen/config"
	"github.com/varavelio/vdl/toolchain/internal/core/ir"
)

func generateStreamTypes(schema *ir.Schema, _ *config.DartConfig) (string, error) {
	if len(schema.Streams) == 0 {
		return "", nil
	}

	g := gen.New().WithSpaces(2)

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
		responseDesc := fmt.Sprintf("%s is the typed event wrapper yielded by the %s stream.", responseName, fullName)

		g.Line(renderDartType("", inputName, inputDesc, stream.Input))
		g.Break()

		g.Line(renderDartType("", outputName, outputDesc, stream.Output))
		g.Break()

		g.Linef("/// %s", responseDesc)
		g.Linef("typedef %s = Response<%s>;", responseName, outputName)
		g.Break()
	}

	g.Line("/// __vdlStreamNames lists all stream identifiers available in this client.")
	g.Line("const List<String> __vdlStreamNames = [")
	g.Block(func() {
		for _, stream := range schema.Streams {
			path := stream.Path()
			g.Linef("'%s',", path)
		}
	})
	g.Line("];")
	g.Break()

	return g.String(), nil
}

package dart

import (
	"fmt"

	"github.com/varavelio/gen"
	"github.com/varavelio/vdl/toolchain/internal/codegen/config"
	"github.com/varavelio/vdl/toolchain/internal/core/ir"
)

func generateProcedureTypes(schema *ir.Schema, _ *config.DartConfig) (string, error) {
	if len(schema.Procedures) == 0 {
		return "", nil
	}

	g := gen.New().WithSpaces(2)

	g.Line("// -----------------------------------------------------------------------------")
	g.Line("// Procedure Types")
	g.Line("// -----------------------------------------------------------------------------")
	g.Break()

	for _, proc := range schema.Procedures {
		fullName := proc.FullName()

		inputName := fmt.Sprintf("%sInput", fullName)
		outputName := fmt.Sprintf("%sOutput", fullName)
		responseName := fmt.Sprintf("%sResponse", fullName)

		inputDesc := fmt.Sprintf("%s represents the input parameters for the %s procedure.", inputName, fullName)
		outputDesc := fmt.Sprintf("%s represents the output parameters for the %s procedure.", outputName, fullName)
		responseDesc := fmt.Sprintf("%s is the typed result wrapper returned by %s calls.", responseName, fullName)

		g.Line(renderDartType("", inputName, inputDesc, proc.Input))
		g.Break()

		g.Line(renderDartType("", outputName, outputDesc, proc.Output))
		g.Break()

		g.Linef("/// %s", responseDesc)
		g.Linef("typedef %s = Response<%s>;", responseName, outputName)
		g.Break()
	}

	g.Line("/// __vdlProcedureNames lists all procedure identifiers available in this client.")
	g.Line("const List<String> __vdlProcedureNames = [")
	g.Block(func() {
		for _, proc := range schema.Procedures {
			path := proc.Path()
			g.Linef("'%s',", path)
		}
	})
	g.Line("];")
	g.Break()

	return g.String(), nil
}

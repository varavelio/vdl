package dart

import (
	"fmt"

	"github.com/varavelio/gen"
	"github.com/varavelio/vdl/toolchain/internal/codegen/config"
	"github.com/varavelio/vdl/toolchain/internal/core/ir/irtypes"
)

func generateProcedureTypes(schema *irtypes.IrSchema, _ *config.DartConfig) (string, error) {
	if len(schema.Procedures) == 0 {
		return "", nil
	}

	g := gen.New().WithSpaces(2)

	g.Line("// -----------------------------------------------------------------------------")
	g.Line("// Procedure Types")
	g.Line("// -----------------------------------------------------------------------------")
	g.Break()

	for _, proc := range schema.Procedures {
		fullName := proc.RpcName + proc.Name

		inputName := fmt.Sprintf("%sInput", fullName)
		outputName := fmt.Sprintf("%sOutput", fullName)
		responseName := fmt.Sprintf("%sResponse", fullName)

		inputDesc := fmt.Sprintf("%s represents the input parameters for the %s procedure.", inputName, fullName)
		outputDesc := fmt.Sprintf("%s represents the output parameters for the %s procedure.", outputName, fullName)
		responseDesc := fmt.Sprintf("%s is the typed result wrapper returned by %s calls.", responseName, fullName)

		g.Line(renderDartType("", inputName, inputDesc, proc.InputFields))
		g.Break()

		g.Line(renderDartType("", outputName, outputDesc, proc.OutputFields))
		g.Break()

		g.Linef("/// %s", responseDesc)
		g.Linef("typedef %s = Response<%s>;", responseName, outputName)
		g.Break()
	}

	return g.String(), nil
}

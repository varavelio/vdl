package python

import (
	"fmt"

	"github.com/varavelio/gen"
	"github.com/varavelio/vdl/toolchain/internal/codegen/config"
	"github.com/varavelio/vdl/toolchain/internal/core/ir"
)

func generateProcedureTypes(schema *ir.Schema, cfg *config.PythonConfig) (string, error) {
	if len(schema.Procedures) == 0 {
		return "", nil
	}

	g := gen.New()

	for _, proc := range schema.Procedures {
		fullName := proc.FullName()

		inputName := fmt.Sprintf("%sInput", fullName)
		outputName := fmt.Sprintf("%sOutput", fullName)

		inputDesc := fmt.Sprintf("%s represents the input parameters for the %s procedure.", inputName, fullName)
		outputDesc := fmt.Sprintf("%s represents the output parameters for the %s procedure.", outputName, fullName)

		g.Raw(GenerateDataclass(inputName, inputDesc, proc.Input))
		g.Break()

		g.Raw(GenerateDataclass(outputName, outputDesc, proc.Output))
		g.Break()

		// Response Type Alias
		responseName := fmt.Sprintf("%sResponse", fullName)
		g.Linef("%s = Response[%s]", responseName, outputName)
		g.Break()
	}

	return g.String(), nil
}

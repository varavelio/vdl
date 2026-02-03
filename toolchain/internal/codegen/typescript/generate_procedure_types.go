package typescript

import (
	"fmt"

	"github.com/varavelio/gen"
	"github.com/varavelio/vdl/toolchain/internal/codegen/config/configtypes"
	"github.com/varavelio/vdl/toolchain/internal/core/ir/irtypes"
	"github.com/varavelio/vdl/toolchain/internal/util/strutil"
)

func generateProcedureTypes(schema *irtypes.IrSchema, _ *configtypes.TypeScriptConfig) (string, error) {
	g := gen.New().WithSpaces(2)

	if len(schema.Procedures) > 0 {
		g.Line("// -----------------------------------------------------------------------------")
		g.Line("// Procedure Types")
		g.Line("// -----------------------------------------------------------------------------")
		g.Break()

		for _, proc := range schema.Procedures {
			fullName := strutil.ToPascalCase(proc.RpcName) + strutil.ToPascalCase(proc.Name)

			inputName := fmt.Sprintf("%sInput", fullName)
			outputName := fmt.Sprintf("%sOutput", fullName)
			responseName := fmt.Sprintf("%sResponse", fullName)

			inputDesc := fmt.Sprintf("%s represents the input parameters for the %s procedure.", inputName, fullName)
			outputDesc := fmt.Sprintf("%s represents the output parameters for the %s procedure.", outputName, fullName)
			responseDesc := fmt.Sprintf("%s represents the response for the %s procedure.", responseName, fullName)

			g.Line(renderType("", inputName, inputDesc, proc.Input))
			g.Break()

			g.Line(renderType("", outputName, outputDesc, proc.Output))
			g.Break()

			g.Line(renderHydrateType("", outputName, proc.Output))
			g.Break()

			g.Line(renderValidateType("", inputName, proc.Input))
			g.Break()

			g.Linef("// %s", responseDesc)
			g.Linef("export type %s = Response<%s>", responseName, outputName)
			g.Break()
		}
	}

	return g.String(), nil
}

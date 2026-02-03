package golang

import (
	"fmt"

	"github.com/varavelio/gen"
	"github.com/varavelio/vdl/toolchain/internal/codegen/config/configtypes"
	"github.com/varavelio/vdl/toolchain/internal/core/ir/irtypes"
)

func generateProcedureTypes(schema *irtypes.IrSchema, _ *configtypes.GoTargetConfig) (string, error) {
	if len(schema.Procedures) == 0 {
		return "", nil
	}

	g := gen.New().WithTabs()

	g.Line("// -----------------------------------------------------------------------------")
	g.Line("// Procedure Types")
	g.Line("// -----------------------------------------------------------------------------")
	g.Break()

	for _, proc := range schema.Procedures {
		procName := proc.RpcName + proc.Name
		inputName := fmt.Sprintf("%sInput", procName)
		outputName := fmt.Sprintf("%sOutput", procName)
		responseName := fmt.Sprintf("%sResponse", procName)

		inputDesc := fmt.Sprintf("%s represents the input parameters for the %s.%s procedure.", inputName, proc.RpcName, proc.Name)
		outputDesc := fmt.Sprintf("%s represents the output parameters for the %s.%s procedure.", outputName, proc.RpcName, proc.Name)
		responseDesc := fmt.Sprintf("%s represents the response for the %s.%s procedure.", responseName, proc.RpcName, proc.Name)

		if proc.GetDoc() != "" {
			inputDesc += "\n\n" + proc.GetDoc()
			outputDesc += "\n\n" + proc.GetDoc()
		}

		g.Line(renderType("", inputName, inputDesc, proc.Input))
		g.Break()

		g.Line(renderPreType("", inputName, proc.Input))
		g.Break()

		g.Line(renderType("", outputName, outputDesc, proc.Output))
		g.Break()

		g.Linef("// %s", responseDesc)
		g.Linef("type %s = Response[%s]", responseName, outputName)
		g.Break()
	}

	return g.String(), nil
}

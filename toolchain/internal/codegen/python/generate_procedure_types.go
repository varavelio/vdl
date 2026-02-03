package python

import (
	"fmt"

	"github.com/varavelio/gen"
	"github.com/varavelio/vdl/toolchain/internal/codegen/config"
	"github.com/varavelio/vdl/toolchain/internal/core/ir/irtypes"
	"github.com/varavelio/vdl/toolchain/internal/util/strutil"
)

func generateProcedureTypes(schema *irtypes.IrSchema, _ *config.PythonConfig) (string, error) {
	if len(schema.Procedures) == 0 {
		return "", nil
	}

	g := gen.New()

	for _, proc := range schema.Procedures {
		fullName := strutil.ToPascalCase(proc.RpcName) + strutil.ToPascalCase(proc.Name)

		inputName := fmt.Sprintf("%sInput", fullName)
		outputName := fmt.Sprintf("%sOutput", fullName)

		inputDesc := fmt.Sprintf("%s represents the input parameters for the %s procedure.", inputName, fullName)
		outputDesc := fmt.Sprintf("%s represents the output parameters for the %s procedure.", outputName, fullName)

		g.Raw(GenerateDataclass(inputName, inputDesc, proc.Input))
		g.Break()
		g.Raw(renderInlineTypes(inputName, proc.Input))
		g.Break()

		g.Raw(GenerateDataclass(outputName, outputDesc, proc.Output))
		g.Break()
		g.Raw(renderInlineTypes(outputName, proc.Output))
		g.Break()

		// Response Type Alias
		responseName := fmt.Sprintf("%sResponse", fullName)
		g.Linef("%s = Response[%s]", responseName, outputName)
		g.Break()
	}

	return g.String(), nil
}

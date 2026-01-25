package golang

import (
	"fmt"

	"github.com/varavelio/gen"
	"github.com/varavelio/vdl/toolchain/internal/codegen/config"
	"github.com/varavelio/vdl/toolchain/internal/core/ir"
)

func generateProcedureTypes(schema *ir.Schema, _ *config.GoConfig) (string, error) {
	g := gen.New().WithTabs()

	if len(schema.Procedures) > 0 {
		g.Line("// -----------------------------------------------------------------------------")
		g.Line("// Procedure Types")
		g.Line("// -----------------------------------------------------------------------------")
		g.Break()

		for _, proc := range schema.Procedures {
			procName := proc.FullName()
			inputName := fmt.Sprintf("%sInput", procName)
			outputName := fmt.Sprintf("%sOutput", procName)
			responseName := fmt.Sprintf("%sResponse", procName)

			inputDesc := fmt.Sprintf("%s represents the input parameters for the %s.%s procedure.", inputName, proc.RPCName, proc.Name)
			outputDesc := fmt.Sprintf("%s represents the output parameters for the %s.%s procedure.", outputName, proc.RPCName, proc.Name)
			responseDesc := fmt.Sprintf("%s represents the response for the %s.%s procedure.", responseName, proc.RPCName, proc.Name)

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
	}

	// Generate list of all procedure names
	g.Line("// VDLProcedureNames is a list of all procedure definitions.")
	g.Line("var VDLProcedureNames = []OperationDefinition{")
	g.Block(func() {
		for _, proc := range schema.Procedures {
			g.Linef("{RPCName: %q, Name: %q, Type: OperationTypeProc},", proc.RPCName, proc.Name)
		}
	})
	g.Line("}")
	g.Break()

	return g.String(), nil
}

package golang

import (
	"fmt"

	"github.com/varavelio/gen"
	"github.com/varavelio/vdl/toolchain/internal/core/ir"
)

func generateProcedureTypes(_ *ir.Schema, flat *flatSchema, _ Config) (string, error) {
	if len(flat.Procedures) == 0 {
		return "", nil
	}

	g := gen.New().WithTabs()

	g.Line("// -----------------------------------------------------------------------------")
	g.Line("// Procedure Types")
	g.Line("// -----------------------------------------------------------------------------")
	g.Break()

	for _, fp := range flat.Procedures {
		procName := fullProcName(fp.RPCName, fp.Procedure.Name)
		inputName := fmt.Sprintf("%sInput", procName)
		outputName := fmt.Sprintf("%sOutput", procName)
		responseName := fmt.Sprintf("%sResponse", procName)

		inputDesc := fmt.Sprintf("%s represents the input parameters for the %s.%s procedure.", inputName, fp.RPCName, fp.Procedure.Name)
		outputDesc := fmt.Sprintf("%s represents the output parameters for the %s.%s procedure.", outputName, fp.RPCName, fp.Procedure.Name)
		responseDesc := fmt.Sprintf("%s represents the response for the %s.%s procedure.", responseName, fp.RPCName, fp.Procedure.Name)

		g.Line(renderType("", inputName, inputDesc, fp.Procedure.Input))
		g.Break()

		g.Line(renderPreType("", inputName, fp.Procedure.Input))
		g.Break()

		g.Line(renderType("", outputName, outputDesc, fp.Procedure.Output))
		g.Break()

		g.Linef("// %s", responseDesc)
		g.Linef("type %s = Response[%s]", responseName, outputName)
		g.Break()
	}

	// Generate list of all procedure names
	g.Line("// vdlProcedureNames is a list of all procedure names.")
	g.Line("var vdlProcedureNames = []string{")
	g.Block(func() {
		for _, fp := range flat.Procedures {
			procName := fullProcName(fp.RPCName, fp.Procedure.Name)
			g.Linef("%q,", procName)
		}
	})
	g.Line("}")
	g.Break()

	return g.String(), nil
}

package typescript

import (
	"fmt"

	"github.com/varavelio/gen"
	"github.com/varavelio/vdl/toolchain/internal/codegen/config"
	"github.com/varavelio/vdl/toolchain/internal/core/ir"
)

func generateProcedureTypes(schema *ir.Schema, _ *config.TypeScriptConfig) (string, error) {
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
		responseDesc := fmt.Sprintf("%s represents the response for the %s procedure.", responseName, fullName)

		g.Line(renderType("", inputName, inputDesc, proc.Input))
		g.Break()

		g.Line(renderType("", outputName, outputDesc, proc.Output))
		g.Break()

		g.Line(renderHydrateType("", outputName, proc.Output))
		g.Break()

		g.Linef("// %s", responseDesc)
		g.Linef("export type %s = Response<%s>", responseName, outputName)
		g.Break()
	}

	// Generate procedure names list
	g.Line("// vdlProcedureNames is a list of all procedure names.")
	g.Line("const vdlProcedureNames: string[] = [")
	g.Block(func() {
		for _, proc := range schema.Procedures {
			path := proc.Path()
			g.Linef("\"%s\",", path)
		}
	})
	g.Line("]")
	g.Break()

	return g.String(), nil
}

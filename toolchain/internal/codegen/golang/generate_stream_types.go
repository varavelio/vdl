package golang

import (
	"fmt"

	"github.com/uforg/ufogenkit"
	"github.com/uforg/uforpc/urpc/internal/schema"
	"github.com/uforg/uforpc/urpc/internal/util/strutil"
)

func generateStreamTypes(sch schema.Schema, _ Config) (string, error) {
	g := ufogenkit.NewGenKit().WithTabs()

	g.Line("// -----------------------------------------------------------------------------")
	g.Line("// Stream Types")
	g.Line("// -----------------------------------------------------------------------------")
	g.Break()

	for _, streamNode := range sch.GetStreamNodes() {
		namePascal := strutil.ToPascalCase(streamNode.Name)
		inputName := fmt.Sprintf("%sInput", namePascal)
		outputName := fmt.Sprintf("%sOutput", namePascal)
		responseName := fmt.Sprintf("%sResponse", namePascal)

		inputDesc := fmt.Sprintf("%s represents the input parameters for the %s stream.", inputName, namePascal)
		outputDesc := fmt.Sprintf("%s represents the output parameters for the %s stream.", outputName, namePascal)
		responseDesc := fmt.Sprintf("%s represents the response for the %s stream.", responseName, namePascal)

		g.Line(renderType("", inputName, inputDesc, streamNode.Input))
		g.Break()

		g.Line(renderPreType("", inputName, streamNode.Input))
		g.Break()

		g.Line(renderType("", outputName, outputDesc, streamNode.Output))
		g.Break()

		g.Linef("// %s", responseDesc)
		g.Linef("type %s = Response[%s]", responseName, outputName)
		g.Break()
	}

	g.Line("// ufoStreamNames is a list of all stream names.")
	g.Line("var ufoStreamNames = []string{")
	g.Block(func() {
		for _, streamNode := range sch.GetStreamNodes() {
			g.Linef("\"%s\",", streamNode.Name)
		}
	})
	g.Line("}")
	g.Break()

	return g.String(), nil
}

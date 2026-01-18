package dart

import (
	"fmt"

	"github.com/uforg/ufogenkit"
	"github.com/uforg/uforpc/urpc/internal/schema"
	"github.com/uforg/uforpc/urpc/internal/util/strutil"
)

func generateStreamTypes(sch schema.Schema, _ Config) (string, error) {
	g := ufogenkit.NewGenKit().WithSpaces(2)

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
		responseDesc := fmt.Sprintf("%s is the typed event wrapper yielded by the %s stream.", responseName, namePascal)

		g.Line(renderDartType("", inputName, inputDesc, streamNode.Input))
		g.Break()

		g.Line(renderDartType("", outputName, outputDesc, streamNode.Output))
		g.Break()

		g.Linef("/// %s", responseDesc)
		g.Linef("typedef %s = Response<%s>;", responseName, outputName)
		g.Break()
	}

	g.Line("/// __ufoStreamNames lists all stream identifiers available in this client.")
	g.Line("const List<String> __ufoStreamNames = [")
	g.Block(func() {
		for _, streamNode := range sch.GetStreamNodes() {
			g.Linef("'%s',", streamNode.Name)
		}
	})
	g.Line("];")
	g.Break()

	return g.String(), nil
}

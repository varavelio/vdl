package python

import (
	"fmt"

	"github.com/varavelio/gen"
	"github.com/varavelio/vdl/toolchain/internal/codegen/config"
	"github.com/varavelio/vdl/toolchain/internal/core/ir"
)

func generateStreamTypes(schema *ir.Schema, cfg *config.PythonConfig) (string, error) {
	if len(schema.Streams) == 0 {
		return "", nil
	}

	g := gen.New()

	for _, stream := range schema.Streams {
		fullName := stream.FullName()

		inputName := fmt.Sprintf("%sInput", fullName)
		outputName := fmt.Sprintf("%sOutput", fullName)

		inputDesc := fmt.Sprintf("%s represents the input parameters for the %s stream.", inputName, fullName)
		outputDesc := fmt.Sprintf("%s represents the output event data for the %s stream.", outputName, fullName)

		g.Raw(GenerateDataclass(inputName, inputDesc, stream.Input))
		g.Break()

		g.Raw(GenerateDataclass(outputName, outputDesc, stream.Output))
		g.Break()
	}

	return g.String(), nil
}

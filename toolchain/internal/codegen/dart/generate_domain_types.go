package dart

import (
	"strings"

	"github.com/varavelio/gen"
	"github.com/varavelio/vdl/toolchain/internal/codegen/config"
	"github.com/varavelio/vdl/toolchain/internal/core/ir/irtypes"
)

func generateDomainTypes(schema *irtypes.IrSchema, _ *config.DartConfig) (string, error) {
	if len(schema.Types) == 0 {
		return "", nil
	}

	g := gen.New().WithSpaces(2)

	g.Line("// -----------------------------------------------------------------------------")
	g.Line("// Domain Types")
	g.Line("// -----------------------------------------------------------------------------")
	g.Break()

	for _, typeNode := range schema.Types {
		desc := "is a domain type defined in VDL with no documentation."
		if typeNode.GetDoc() != "" {
			desc = strings.TrimSpace(typeNode.GetDoc())
		}
		if typeNode.Deprecation != nil {
			desc += "\n\n@deprecated "
			if *typeNode.Deprecation == "" {
				desc += "This type is deprecated and should not be used in new code."
			} else {
				desc += *typeNode.Deprecation
			}
		}

		g.Line(renderDartType("", typeNode.Name, desc, typeNode.Fields))
		g.Break()
	}

	return g.String(), nil
}

package golang

import (
	"github.com/varavelio/gen"
	"github.com/varavelio/vdl/toolchain/internal/codegen/config/configtypes"
	"github.com/varavelio/vdl/toolchain/internal/core/ir/irtypes"
)

func generateDomainTypes(schema *irtypes.IrSchema, _ *configtypes.GoTargetConfig) (string, error) {
	if len(schema.Types) == 0 {
		return "", nil
	}

	g := gen.New().WithTabs()

	g.Line("// -----------------------------------------------------------------------------")
	g.Line("// Domain Types")
	g.Line("// -----------------------------------------------------------------------------")
	g.Break()

	for _, typeNode := range schema.Types {
		desc := typeNode.Name + " is a domain type defined in VDL."
		if typeNode.GetDoc() != "" {
			desc = typeNode.GetDoc()
		}

		if typeNode.Deprecated != nil {
			desc += "\n\nDeprecated: "
			if *typeNode.Deprecated == "" {
				desc += "This type is deprecated and should not be used in new code."
			} else {
				desc += *typeNode.Deprecated
			}
		}

		g.Line(renderType("", typeNode.Name, desc, typeNode.Fields))
		g.Break()

		g.Line(renderPreType("", typeNode.Name, typeNode.Fields))
		g.Break()
	}

	return g.String(), nil
}

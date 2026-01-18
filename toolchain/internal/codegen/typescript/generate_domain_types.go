package typescript

import (
	"strings"

	"github.com/uforg/ufogenkit"
	"github.com/uforg/uforpc/urpc/internal/schema"
)

func generateDomainTypes(sch schema.Schema, config Config) (string, error) {
	g := ufogenkit.NewGenKit().WithSpaces(2)

	g.Line("// -----------------------------------------------------------------------------")
	g.Line("// Domain Types")
	g.Line("// -----------------------------------------------------------------------------")
	g.Break()

	// Generate typescript types
	for _, typeNode := range sch.GetTypeNodes() {
		desc := "is a domain type defined in UFO RPC with no documentation."
		if typeNode.Doc != nil {
			desc = strings.TrimSpace(*typeNode.Doc)
		}

		if typeNode.Deprecated != nil {
			desc += "\n\n@deprecated "
			if *typeNode.Deprecated == "" {
				desc += "This type is deprecated and should not be used in new code."
			} else {
				desc += *typeNode.Deprecated
			}
		}

		g.Line(renderType("", typeNode.Name, desc, typeNode.Fields))
		g.Break()

		g.Line(renderHydrateType("", typeNode.Name, typeNode.Fields))
		g.Break()
	}

	return g.String(), nil
}

package golang

import (
	"strings"

	"github.com/uforg/ufogenkit"
	"github.com/uforg/uforpc/urpc/internal/schema"
	"github.com/uforg/uforpc/urpc/internal/util/strutil"
)

func generateDomainTypes(sch schema.Schema, config Config) (string, error) {
	g := ufogenkit.NewGenKit().WithTabs()

	g.Line("// -----------------------------------------------------------------------------")
	g.Line("// Domain Types")
	g.Line("// -----------------------------------------------------------------------------")
	g.Break()

	for _, typeNode := range sch.GetTypeNodes() {
		desc := "is a domain type defined in UFO RPC with no documentation."
		if typeNode.Doc != nil {
			desc = strings.TrimSpace(strutil.NormalizeIndent(*typeNode.Doc))
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

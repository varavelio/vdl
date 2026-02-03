package python

import (
	"fmt"
	"strings"

	"github.com/varavelio/gen"
	"github.com/varavelio/vdl/toolchain/internal/codegen/config/configtypes"
	"github.com/varavelio/vdl/toolchain/internal/core/ir/irtypes"
	"github.com/varavelio/vdl/toolchain/internal/util/strutil"
)

func generateConstants(schema *irtypes.IrSchema, _ *configtypes.PythonConfig) (string, error) {
	g := gen.New()

	for _, c := range schema.Constants {
		if c.GetDoc() != "" {
			g.Linef("# %s", strings.ReplaceAll(c.GetDoc(), "\n", "\n# "))
		}
		name := strutil.ToUpperSnakeCase(c.Name)
		val := c.Value

		switch c.ConstType {
		case irtypes.ConstTypeString:
			val = fmt.Sprintf("%q", c.Value)
		case irtypes.ConstTypeBool:
			if c.Value == "true" {
				val = "True"
			} else {
				val = "False"
			}
		}
		// int and float are just string representations of numbers in IR, so they work as is.

		g.Linef("%s = %s", name, val)
		g.Break()
	}

	return g.String(), nil
}

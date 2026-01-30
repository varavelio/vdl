package python

import (
	"fmt"
	"strings"

	"github.com/varavelio/gen"
	"github.com/varavelio/vdl/toolchain/internal/codegen/config"
	"github.com/varavelio/vdl/toolchain/internal/core/ir"
	"github.com/varavelio/vdl/toolchain/internal/util/strutil"
)

func generateConstants(schema *ir.Schema, cfg *config.PythonConfig) (string, error) {
	g := gen.New()

	for _, c := range schema.Constants {
		if c.Doc != "" {
			g.Linef("# %s", strings.ReplaceAll(c.Doc, "\n", "\n# "))
		}
		name := strutil.ToUpperSnakeCase(c.Name)
		val := c.Value

		switch c.ValueType {
		case ir.ConstValueTypeString:
			val = fmt.Sprintf("%q", c.Value)
		case ir.ConstValueTypeBool:
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

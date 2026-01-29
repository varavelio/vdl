package python

import (
	"fmt"

	"github.com/varavelio/gen"
	"github.com/varavelio/vdl/toolchain/internal/codegen/config"
	"github.com/varavelio/vdl/toolchain/internal/core/ir"
	"github.com/varavelio/vdl/toolchain/internal/util/strutil"
)

func generateEnums(schema *ir.Schema, cfg *config.PythonConfig) (string, error) {
	g := gen.New()

	for _, e := range schema.Enums {
		className := strutil.ToPascalCase(e.Name)
		baseClass := "str, Enum"
		if e.ValueType == ir.EnumValueTypeInt {
			baseClass = "IntEnum"
		}

		g.Linef("class %s(%s):", className, baseClass)
		if e.Doc != "" {
			g.Linef("    \"\"\"%s\"\"\"", e.Doc)
		}

		for _, m := range e.Members {
			memberName := strutil.ToUpperSnakeCase(m.Name)
			val := m.Value
			if e.ValueType == ir.EnumValueTypeString {
				val = fmt.Sprintf("%q", m.Value)
			}
			g.Linef("    %s = %s", memberName, val)
		}
		g.Break()
		if e.ValueType == ir.EnumValueTypeString {
			g.Line("    @classmethod")
			g.Linef("    def from_value(cls, value: str) -> %s | None:", className)
			g.Line("        try:")
			g.Line("            return cls(value)")
			g.Line("        except Exception:")
			g.Line("            return None")
		} else {
			g.Line("    @classmethod")
			g.Linef("    def from_value(cls, value: int) -> %s | None:", className)
			g.Line("        try:")
			g.Line("            return cls(value)")
			g.Line("        except Exception:")
			g.Line("            return None")
		}
		g.Break()
	}

	return g.String(), nil
}

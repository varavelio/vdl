package python

import (
	"fmt"

	"github.com/varavelio/gen"
	"github.com/varavelio/vdl/toolchain/internal/codegen/config"
	"github.com/varavelio/vdl/toolchain/internal/core/ir/irtypes"
	"github.com/varavelio/vdl/toolchain/internal/util/strutil"
)

func generateEnums(schema *irtypes.IrSchema, _ *config.PythonConfig) (string, error) {
	g := gen.New()

	for _, e := range schema.Enums {
		className := strutil.ToPascalCase(e.Name)
		baseClass := "str, Enum"
		if e.EnumType == irtypes.EnumTypeInt {
			baseClass = "IntEnum"
		}

		g.Linef("class %s(%s):", className, baseClass)
		if e.GetDoc() != "" {
			g.Linef("    \"\"\"%s\"\"\"", e.GetDoc())
		}

		for _, m := range e.Members {
			memberName := strutil.ToUpperSnakeCase(m.Name)
			val := m.Value
			if e.EnumType == irtypes.EnumTypeString {
				val = fmt.Sprintf("%q", m.Value)
			}
			g.Linef("    %s = %s", memberName, val)
		}
		g.Break()
		if e.EnumType == irtypes.EnumTypeString {
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

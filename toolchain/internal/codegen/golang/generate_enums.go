package golang

import (
	"strconv"

	"github.com/varavelio/gen"
	"github.com/varavelio/vdl/toolchain/internal/codegen/config"
	"github.com/varavelio/vdl/toolchain/internal/core/ir"
)

func generateEnums(schema *ir.Schema, config *config.GoConfig) (string, error) {
	if len(schema.Enums) == 0 {
		return "", nil
	}

	g := gen.New().WithTabs()

	g.Line("// -----------------------------------------------------------------------------")
	g.Line("// Enumerations")
	g.Line("// -----------------------------------------------------------------------------")
	g.Break()

	for _, enum := range schema.Enums {
		generateEnum(g, enum)
	}

	return g.String(), nil
}

// generateEnum generates Go code for a single enum type.
func generateEnum(g *gen.Generator, enum ir.Enum) {
	// Documentation
	if enum.Doc != "" {
		doc := enum.Doc
		renderMultilineComment(g, doc)
	} else {
		g.Linef("// %s is an enumeration type.", enum.Name)
	}

	// Deprecation
	renderDeprecated(g, enum.Deprecated)

	// Type definition
	if enum.ValueType == ir.EnumValueTypeString {
		g.Linef("type %s string", enum.Name)
	} else {
		g.Linef("type %s int", enum.Name)
	}
	g.Break()

	// Constants block
	g.Linef("// %s enum values", enum.Name)
	g.Line("const (")
	g.Block(func() {
		for _, member := range enum.Members {
			constName := enum.Name + member.Name
			if enum.ValueType == ir.EnumValueTypeString {
				g.Linef("%s %s = %q", constName, enum.Name, member.Value)
			} else {
				// Integer value
				intVal, _ := strconv.Atoi(member.Value)
				g.Linef("%s %s = %d", constName, enum.Name, intVal)
			}
		}
	})
	g.Line(")")
	g.Break()

	// String() method for the enum
	g.Linef("// String returns the string representation of %s.", enum.Name)
	g.Linef("func (e %s) String() string {", enum.Name)
	g.Block(func() {
		if enum.ValueType == ir.EnumValueTypeString {
			g.Line("return string(e)")
		} else {
			// For int enums, create a switch statement
			g.Line("switch e {")
			for _, member := range enum.Members {
				constName := enum.Name + member.Name
				g.Linef("case %s:", constName)
				g.Block(func() {
					g.Linef("return %q", member.Name)
				})
			}
			g.Line("default:")
			g.Block(func() {
				g.Linef("return fmt.Sprintf(\"%s(%%d)\", e)", enum.Name)
			})
			g.Line("}")
		}
	})
	g.Line("}")
	g.Break()

	// IsValid() method
	g.Linef("// IsValid returns true if the value is a valid %s.", enum.Name)
	g.Linef("func (e %s) IsValid() bool {", enum.Name)
	g.Block(func() {
		g.Line("switch e {")
		g.Line("case")
		g.Block(func() {
			for i, member := range enum.Members {
				constName := enum.Name + member.Name
				if i < len(enum.Members)-1 {
					g.Linef("%s,", constName)
				} else {
					g.Linef("%s:", constName)
				}
			}
		})
		g.Block(func() {
			g.Line("return true")
		})
		g.Line("}")
		g.Line("return false")
	})
	g.Line("}")
	g.Break()
}

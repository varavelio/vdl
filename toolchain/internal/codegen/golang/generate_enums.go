package golang

import (
	"strconv"
	"strings"

	"github.com/varavelio/gen"
	"github.com/varavelio/vdl/toolchain/internal/codegen/config/configtypes"
	"github.com/varavelio/vdl/toolchain/internal/core/ir/irtypes"
)

func generateEnums(schema *irtypes.IrSchema, config *configtypes.GoConfig) (string, error) {
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
func generateEnum(g *gen.Generator, enum irtypes.EnumDef) {
	// Documentation
	if enum.GetDoc() != "" {
		doc := enum.GetDoc()
		renderMultilineComment(g, doc)
	} else {
		g.Linef("// %s is an enumeration of values.", enum.Name)
	}

	// Deprecation
	renderDeprecated(g, enum.Deprecated)

	// Type definition
	if enum.EnumType == irtypes.EnumTypeString {
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
			if enum.EnumType == irtypes.EnumTypeString {
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

	// List variable with all enum values
	g.Linef("// %sList contains all valid %s values.", enum.Name, enum.Name)
	g.Linef("var %sList = []%s{", enum.Name, enum.Name)
	g.Block(func() {
		for _, member := range enum.Members {
			constName := enum.Name + member.Name
			g.Linef("%s,", constName)
		}
	})
	g.Line("}")
	g.Break()

	// String() method for the enum
	g.Linef("// String returns the string representation of %s.", enum.Name)
	g.Linef("func (e %s) String() string {", enum.Name)
	g.Block(func() {
		if enum.EnumType == irtypes.EnumTypeString {
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
		g.Linef("case %s:", enumCaseList(enum))
		g.Block(func() {
			g.Line("return true")
		})
		g.Line("}")
		g.Line("return false")
	})
	g.Line("}")
	g.Break()

	// MarshalJSON method
	generateEnumMarshalJSON(g, enum)

	// UnmarshalJSON method
	generateEnumUnmarshalJSON(g, enum)
}

// enumCaseList returns a comma-separated list of all enum constant names.
func enumCaseList(enum irtypes.EnumDef) string {
	var result strings.Builder
	for i, member := range enum.Members {
		if i > 0 {
			result.WriteString(", ")
		}
		result.WriteString(enum.Name + member.Name)
	}
	return result.String()
}

// generateEnumMarshalJSON generates the json.Marshaler implementation.
func generateEnumMarshalJSON(g *gen.Generator, enum irtypes.EnumDef) {
	g.Linef("// MarshalJSON implements json.Marshaler.")
	g.Linef("// Returns an error if the value is not a valid %s member.", enum.Name)
	g.Linef("func (e %s) MarshalJSON() ([]byte, error) {", enum.Name)
	g.Block(func() {
		g.Line("if !e.IsValid() {")
		g.Block(func() {
			if enum.EnumType == irtypes.EnumTypeString {
				g.Linef("return nil, fmt.Errorf(\"cannot marshal invalid value '%%s' for enum %s\", string(e))", enum.Name)
			} else {
				g.Linef("return nil, fmt.Errorf(\"cannot marshal invalid value '%%d' for enum %s\", int(e))", enum.Name)
			}
		})
		g.Line("}")
		if enum.EnumType == irtypes.EnumTypeString {
			g.Line("return json.Marshal(string(e))")
		} else {
			g.Line("return json.Marshal(int(e))")
		}
	})
	g.Line("}")
	g.Break()
}

// generateEnumUnmarshalJSON generates the json.Unmarshaler implementation.
func generateEnumUnmarshalJSON(g *gen.Generator, enum irtypes.EnumDef) {
	g.Linef("// UnmarshalJSON implements json.Unmarshaler.")
	g.Linef("// Returns an error if the value is not a valid %s member.", enum.Name)
	g.Linef("func (e *%s) UnmarshalJSON(data []byte) error {", enum.Name)
	g.Block(func() {
		if enum.EnumType == irtypes.EnumTypeString {
			g.Line("var s string")
			g.Line("if err := json.Unmarshal(data, &s); err != nil {")
			g.Block(func() {
				g.Line("return err")
			})
			g.Line("}")
			g.Break()
			g.Linef("v := %s(s)", enum.Name)
		} else {
			g.Line("var n int")
			g.Line("if err := json.Unmarshal(data, &n); err != nil {")
			g.Block(func() {
				g.Line("return err")
			})
			g.Line("}")
			g.Break()
			g.Linef("v := %s(n)", enum.Name)
		}
		g.Line("if !v.IsValid() {")
		g.Block(func() {
			if enum.EnumType == irtypes.EnumTypeString {
				g.Linef("return fmt.Errorf(\"invalid value '%%s' for enum %s\", s)", enum.Name)
			} else {
				g.Linef("return fmt.Errorf(\"invalid value '%%d' for enum %s\", n)", enum.Name)
			}
		})
		g.Line("}")
		g.Break()
		g.Line("*e = v")
		g.Line("return nil")
	})
	g.Line("}")
	g.Break()
}

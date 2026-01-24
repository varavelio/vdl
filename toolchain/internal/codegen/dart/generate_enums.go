package dart

import (
	"strconv"
	"strings"

	"github.com/varavelio/gen"
	"github.com/varavelio/vdl/toolchain/internal/codegen/config"
	"github.com/varavelio/vdl/toolchain/internal/core/ir"
	"github.com/varavelio/vdl/toolchain/internal/util/strutil"
)

func generateEnums(schema *ir.Schema, _ *config.DartConfig) (string, error) {
	if len(schema.Enums) == 0 {
		return "", nil
	}

	g := gen.New().WithSpaces(2)

	g.Line("// -----------------------------------------------------------------------------")
	g.Line("// Enums")
	g.Line("// -----------------------------------------------------------------------------")
	g.Break()

	for _, enum := range schema.Enums {
		renderDartEnum(g, enum)
	}

	return g.String(), nil
}

// renderDartEnum renders a single Dart enum definition.
func renderDartEnum(g *gen.Generator, enum ir.Enum) {
	// Generate doc comment
	if enum.Doc != "" {
		doc := strings.TrimSpace(strutil.NormalizeIndent(enum.Doc))
		renderMultilineCommentDart(g, doc)
	} else {
		g.Linef("/// %s is an enumeration type.", enum.Name)
	}

	// Deprecation
	if enum.Deprecated != nil {
		renderDeprecatedDart(g, enum.Deprecated)
	}

	if enum.ValueType == ir.EnumValueTypeString {
		// String enums - use Dart enum with string values
		g.Linef("enum %s {", enum.Name)
		g.Block(func() {
			for i, member := range enum.Members {
				suffix := ","
				if i == len(enum.Members)-1 {
					suffix = ";"
				}
				g.Linef("%s('%s')%s", member.Name, member.Value, suffix)
			}
			g.Break()
			g.Line("final String value;")
			g.Linef("const %s(this.value);", enum.Name)
			g.Break()
			g.Linef("static %s? fromValue(String value) {", enum.Name)
			g.Block(func() {
				g.Linef("try { return %s.values.firstWhere((e) => e.value == value); } catch (_) { return null; }", enum.Name)
			})
			g.Line("}")
		})
		g.Linef("}")
	} else {
		// Int enums - use Dart enum with int values
		g.Linef("enum %s {", enum.Name)
		g.Block(func() {
			for i, member := range enum.Members {
				suffix := ","
				if i == len(enum.Members)-1 {
					suffix = ";"
				}
				intVal, _ := strconv.Atoi(member.Value)
				g.Linef("%s(%d)%s", member.Name, intVal, suffix)
			}
			g.Break()
			g.Line("final int value;")
			g.Linef("const %s(this.value);", enum.Name)
			g.Break()
			g.Linef("static %s? fromValue(int value) {", enum.Name)
			g.Block(func() {
				g.Linef("try { return %s.values.firstWhere((e) => e.value == value); } catch (_) { return null; }", enum.Name)
			})
			g.Line("}")
		})
		g.Linef("}")
	}
	g.Break()

	// Generate extension for JSON serialization
	g.Linef("extension %sJson on %s {", enum.Name, enum.Name)
	g.Block(func() {
		if enum.ValueType == ir.EnumValueTypeString {
			g.Line("String toJson() => value;")
			g.Break()
			g.Linef("static %s fromJson(String json) {", enum.Name)
			g.Block(func() {
				g.Linef("final result = %s.fromValue(json);", enum.Name)
				g.Linef("if (result == null) throw FormatException('Invalid %s value: $json');", enum.Name)
				g.Line("return result;")
			})
			g.Line("}")
		} else {
			g.Line("int toJson() => value;")
			g.Break()
			g.Linef("static %s fromJson(int json) {", enum.Name)
			g.Block(func() {
				g.Linef("final result = %s.fromValue(json);", enum.Name)
				g.Linef("if (result == null) throw FormatException('Invalid %s value: $json');", enum.Name)
				g.Line("return result;")
			})
			g.Line("}")
		}
	})
	g.Linef("}")
	g.Break()

	// Generate list of all enum values
	g.Linef("/// List of all %s values.", enum.Name)
	g.Linef("const List<%s> %sList = %s.values;", enum.Name, lowercaseFirst(enum.Name), enum.Name)
	g.Break()
}

// lowercaseFirst returns the string with the first character lowercased.
func lowercaseFirst(s string) string {
	if s == "" {
		return s
	}
	return strings.ToLower(s[:1]) + s[1:]
}

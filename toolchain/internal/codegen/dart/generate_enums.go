package dart

import (
	"fmt"
	"strings"

	"github.com/varavelio/gen"
	"github.com/varavelio/vdl/toolchain/internal/core/ir"
)

// generateEnums generates Dart enum definitions.
func generateEnums(schema *ir.Schema, _ *flatSchema, _ Config) (string, error) {
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
		g.Line("/// " + strings.ReplaceAll(strings.TrimSpace(enum.Doc), "\n", "\n/// "))
	}
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
				g.Linef("for (final v in %s.values) {", enum.Name)
				g.Block(func() {
					g.Line("if (v.value == value) return v;")
				})
				g.Line("}")
				g.Line("return null;")
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
				g.Linef("%s(%s)%s", member.Name, member.Value, suffix)
			}
			g.Break()
			g.Line("final int value;")
			g.Linef("const %s(this.value);", enum.Name)
			g.Break()
			g.Linef("static %s? fromValue(int value) {", enum.Name)
			g.Block(func() {
				g.Linef("for (final v in %s.values) {", enum.Name)
				g.Block(func() {
					g.Line("if (v.value == value) return v;")
				})
				g.Line("}")
				g.Line("return null;")
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

// generateConstants generates Dart constant definitions.
func generateConstants(schema *ir.Schema, _ *flatSchema, _ Config) (string, error) {
	if len(schema.Constants) == 0 {
		return "", nil
	}

	g := gen.New().WithSpaces(2)

	g.Line("// -----------------------------------------------------------------------------")
	g.Line("// Constants")
	g.Line("// -----------------------------------------------------------------------------")
	g.Break()

	for _, constant := range schema.Constants {
		renderDartConstant(g, constant)
	}

	return g.String(), nil
}

// renderDartConstant renders a single Dart constant definition.
func renderDartConstant(g *gen.Generator, constant ir.Constant) {
	// Generate doc comment
	if constant.Doc != "" {
		g.Line("/// " + strings.ReplaceAll(strings.TrimSpace(constant.Doc), "\n", "\n/// "))
	}
	if constant.Deprecated != nil {
		renderDeprecatedDart(g, constant.Deprecated)
	}

	// Determine the Dart type and value format
	var dartType, dartValue string
	switch constant.ValueType {
	case ir.ConstValueTypeString:
		dartType = "String"
		dartValue = fmt.Sprintf("'%s'", constant.Value)
	case ir.ConstValueTypeInt:
		dartType = "int"
		dartValue = constant.Value
	case ir.ConstValueTypeFloat:
		dartType = "double"
		dartValue = constant.Value
	case ir.ConstValueTypeBool:
		dartType = "bool"
		dartValue = constant.Value
	default:
		dartType = "dynamic"
		dartValue = constant.Value
	}

	g.Linef("const %s %s = %s;", dartType, constant.Name, dartValue)
	g.Break()
}

// generatePatterns generates Dart pattern template functions.
func generatePatterns(schema *ir.Schema, _ *flatSchema, _ Config) (string, error) {
	if len(schema.Patterns) == 0 {
		return "", nil
	}

	g := gen.New().WithSpaces(2)

	g.Line("// -----------------------------------------------------------------------------")
	g.Line("// Patterns")
	g.Line("// -----------------------------------------------------------------------------")
	g.Break()

	for _, pattern := range schema.Patterns {
		renderDartPattern(g, pattern)
	}

	return g.String(), nil
}

// renderDartPattern renders a single Dart pattern template function.
func renderDartPattern(g *gen.Generator, pattern ir.Pattern) {
	// Generate doc comment
	if pattern.Doc != "" {
		g.Line("/// " + strings.ReplaceAll(strings.TrimSpace(pattern.Doc), "\n", "\n/// "))
	}
	if pattern.Deprecated != nil {
		renderDeprecatedDart(g, pattern.Deprecated)
	}

	// Generate function signature with parameters
	params := make([]string, len(pattern.Placeholders))
	for i, placeholder := range pattern.Placeholders {
		params[i] = fmt.Sprintf("String %s", placeholder)
	}

	g.Linef("String %s(%s) {", pattern.Name, strings.Join(params, ", "))
	g.Block(func() {
		// Convert template to Dart string interpolation
		templateLiteral := convertPatternToDartInterpolation(pattern.Template, pattern.Placeholders)
		g.Linef("return %s;", templateLiteral)
	})
	g.Linef("}")
	g.Break()
}

// convertPatternToDartInterpolation converts a VDL pattern template to a Dart string interpolation.
// Pattern format: "Hello, {name}!" -> 'Hello, $name!'
func convertPatternToDartInterpolation(template string, placeholders []string) string {
	result := template

	// Replace each {placeholder} with $placeholder
	for _, placeholder := range placeholders {
		result = strings.ReplaceAll(result, "{"+placeholder+"}", "$"+placeholder)
	}

	// Wrap in single quotes for Dart string
	return "'" + result + "'"
}

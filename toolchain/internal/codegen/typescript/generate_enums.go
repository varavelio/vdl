package typescript

import (
	"fmt"
	"strings"

	"github.com/varavelio/gen"
	"github.com/varavelio/vdl/toolchain/internal/core/ir"
)

// generateEnums generates TypeScript enum type definitions.
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
		renderEnum(g, enum)
	}

	return g.String(), nil
}

// renderEnum renders a single enum definition.
func renderEnum(g *gen.Generator, enum ir.Enum) {
	// Generate doc comment
	if enum.Doc != "" {
		g.Linef("/**")
		renderPartialMultilineComment(g, strings.TrimSpace(enum.Doc))
		if enum.Deprecated != nil {
			renderDeprecated(g, enum.Deprecated)
		}
		g.Linef(" */")
	} else if enum.Deprecated != nil {
		g.Linef("/**")
		renderDeprecated(g, enum.Deprecated)
		g.Linef(" */")
	}

	// TypeScript enums as union types for better type safety
	if len(enum.Members) == 0 {
		g.Linef("export type %s = never;", enum.Name)
	} else {
		values := make([]string, len(enum.Members))
		for i, member := range enum.Members {
			if enum.ValueType == ir.EnumValueTypeString {
				values[i] = fmt.Sprintf("\"%s\"", member.Value)
			} else {
				values[i] = member.Value
			}
		}
		g.Linef("export type %s = %s;", enum.Name, strings.Join(values, " | "))
	}
	g.Break()

	// Generate const object with enum values for runtime access
	g.Linef("export const %sValues = {", enum.Name)
	g.Block(func() {
		for _, member := range enum.Members {
			if enum.ValueType == ir.EnumValueTypeString {
				g.Linef("%s: \"%s\",", member.Name, member.Value)
			} else {
				g.Linef("%s: %s,", member.Name, member.Value)
			}
		}
	})
	g.Linef("} as const;")
	g.Break()

	// Generate list of all enum values
	g.Linef("export const %sList: %s[] = [", enum.Name, enum.Name)
	g.Block(func() {
		for _, member := range enum.Members {
			if enum.ValueType == ir.EnumValueTypeString {
				g.Linef("\"%s\",", member.Value)
			} else {
				g.Linef("%s,", member.Value)
			}
		}
	})
	g.Linef("];")
	g.Break()

	// Generate type guard function
	g.Linef("export function is%s(value: unknown): value is %s {", enum.Name, enum.Name)
	g.Block(func() {
		g.Linef("return %sList.includes(value as %s);", enum.Name, enum.Name)
	})
	g.Linef("}")
	g.Break()
}

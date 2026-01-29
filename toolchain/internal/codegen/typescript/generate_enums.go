package typescript

import (
	"strconv"
	"strings"

	"github.com/varavelio/gen"
	"github.com/varavelio/vdl/toolchain/internal/codegen/config"
	"github.com/varavelio/vdl/toolchain/internal/core/ir"
)

func generateEnums(schema *ir.Schema, _ *config.TypeScriptConfig) (string, error) {
	if len(schema.Enums) == 0 {
		return "", nil
	}

	g := gen.New().WithSpaces(2)

	g.Line("// -----------------------------------------------------------------------------")
	g.Line("// Enumerations")
	g.Line("// -----------------------------------------------------------------------------")
	g.Break()

	for _, enum := range schema.Enums {
		generateEnum(g, enum)
	}

	return g.String(), nil
}

// generateEnum generates TypeScript code for a single enum type.
// It generates:
// 1. A type definition (union of literal types)
// 2. An array of all values
// 3. A type guard function
func generateEnum(g *gen.Generator, enum ir.Enum) {
	// Documentation
	if enum.Doc != "" {
		renderMultilineComment(g, enum.Doc)
	} else {
		g.Linef("/** %s is an enumeration type. */", enum.Name)
	}

	// Deprecation
	renderDeprecated(g, enum.Deprecated)

	// 1. Type definition
	// export type Status = "active" | "inactive";
	if len(enum.Members) == 0 {
		g.Linef("export type %s = never;", enum.Name)
	} else {
		var values []string
		for _, member := range enum.Members {
			if enum.ValueType == ir.EnumValueTypeString {
				values = append(values, strconv.Quote(member.Value))
			} else {
				// Integer value
				// Don't quote numbers in TS union types
				values = append(values, member.Value)
			}
		}
		g.Linef("export type %s = %s;", enum.Name, strings.Join(values, " | "))
	}
	g.Break()

	// 2. List of values
	// export const StatusList: Status[] = ["active", "inactive"];
	g.Linef("export const %sList: %s[] = [", enum.Name, enum.Name)
	g.Block(func() {
		for _, member := range enum.Members {
			if enum.ValueType == ir.EnumValueTypeString {
				g.Linef("%q,", member.Value)
			} else {
				// Integer value
				g.Linef("%s,", member.Value)
			}
		}
	})
	g.Line("];")
	g.Break()

	// 3. Type guard
	// export function isStatus(value: unknown): value is Status { ... }
	g.Linef("export function is%s(value: unknown): value is %s {", enum.Name, enum.Name)
	g.Block(func() {
		g.Linef("return %sList.includes(value as %s);", enum.Name, enum.Name)
	})
	g.Line("}")
	g.Break()
}

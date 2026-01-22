package typescript

import (
	"fmt"
	"strings"

	"github.com/varavelio/gen"
	"github.com/varavelio/vdl/toolchain/internal/core/ir"
	"github.com/varavelio/vdl/toolchain/internal/util/strutil"
)

// =============================================================================
// Type Conversion: IR TypeRef → TypeScript Type String
// =============================================================================

// typeRefToTS converts an IR TypeRef to its TypeScript type string representation.
// parentTypeName is used to generate names for inline object types.
func typeRefToTS(parentTypeName string, tr ir.TypeRef) string {
	switch tr.Kind {
	case ir.TypeKindPrimitive:
		return primitiveToTS(tr.Primitive)

	case ir.TypeKindType:
		return tr.Type

	case ir.TypeKindEnum:
		return tr.Enum

	case ir.TypeKindArray:
		// Build the element type first, then wrap with array suffix
		elementType := typeRefToTS(parentTypeName, *tr.ArrayItem)
		// Add array suffix for each dimension
		return elementType + strings.Repeat("[]", tr.ArrayDimensions)

	case ir.TypeKindMap:
		valueType := typeRefToTS(parentTypeName, *tr.MapValue)
		return fmt.Sprintf("Record<string, %s>", valueType)

	case ir.TypeKindObject:
		// Inline objects get a generated name based on parent
		return parentTypeName
	}

	return "any"
}

// primitiveToTS converts an IR primitive type to its TypeScript equivalent.
func primitiveToTS(p ir.PrimitiveType) string {
	switch p {
	case ir.PrimitiveString:
		return "string"
	case ir.PrimitiveInt:
		return "number"
	case ir.PrimitiveFloat:
		return "number"
	case ir.PrimitiveBool:
		return "boolean"
	case ir.PrimitiveDatetime:
		return "Date"
	}
	return "any"
}

// =============================================================================
// Field Rendering
// =============================================================================

// renderField generates the TypeScript code for a single field.
func renderField(parentTypeName string, field ir.Field) string {
	namePascal := strutil.ToPascalCase(field.Name)
	nameCamel := strutil.ToCamelCase(field.Name)

	// Calculate the inline type name for objects
	inlineTypeName := parentTypeName + namePascal
	typeLiteral := typeRefToTS(inlineTypeName, field.Type)

	finalName := nameCamel
	if field.Optional {
		finalName += "?"
	}

	return fmt.Sprintf("%s: %s", finalName, typeLiteral)
}

// =============================================================================
// Type Structure Rendering
// =============================================================================

// renderType renders a complete type definition with all its fields.
func renderType(parentName, name, desc string, fields []ir.Field) string {
	fullName := parentName + name

	g := gen.New().WithSpaces(2)
	if desc != "" {
		g.Linef("/**")
		renderPartialMultilineComment(g, fmt.Sprintf("%s %s", fullName, desc))
		g.Linef(" */")
	}
	g.Linef("export type %s = {", fullName)
	g.Block(func() {
		for _, field := range fields {
			g.Line(renderField(fullName, field))
		}
	})
	g.Line("}")
	g.Break()

	// Render children inline types
	for _, field := range fields {
		if field.Type.Kind == ir.TypeKindObject && field.Type.Object != nil {
			childName := fullName + strutil.ToPascalCase(field.Name)
			g.Line(renderType("", childName, "", field.Type.Object.Fields))
		}
	}

	return g.String()
}

// =============================================================================
// Hydration Functions
// =============================================================================

// renderHydrateField generates the code for a field in a hydrate function.
func renderHydrateField(parentTypeName string, field ir.Field) string {
	namePascal := strutil.ToPascalCase(field.Name)
	nameCamel := strutil.ToCamelCase(field.Name)
	hydratedName := "hydrated" + namePascal

	// Build a formatter for a single value hydration expression. Use "%s" placeholder for the value.
	valueFmt := buildHydrationExpr(parentTypeName, field)

	// Compose the final value literal, handling arrays vs single values.
	valueLiteral := fmt.Sprintf(valueFmt, "input."+nameCamel)

	if field.Type.Kind == ir.TypeKindArray && field.Type.ArrayDimensions > 0 {
		// Handle array types
		valueLiteral = buildArrayHydration(parentTypeName, field)
	} else if field.Type.Kind == ir.TypeKindMap {
		// Handle map types
		valueLiteral = buildMapHydration(parentTypeName, field)
	}

	if field.Optional {
		valueLiteral = fmt.Sprintf("input.%s ? %s : input.%s", nameCamel, valueLiteral, nameCamel)
	}

	return fmt.Sprintf("const %s = %s", hydratedName, valueLiteral)
}

// buildHydrationExpr returns a format string for hydrating a single value.
func buildHydrationExpr(parentTypeName string, field ir.Field) string {
	namePascal := strutil.ToPascalCase(field.Name)

	switch field.Type.Kind {
	case ir.TypeKindObject:
		return fmt.Sprintf("hydrate%s%s(%%s)", parentTypeName, namePascal)

	case ir.TypeKindType:
		typePascal := strutil.ToPascalCase(field.Type.Type)
		return fmt.Sprintf("hydrate%s(%%s)", typePascal)

	case ir.TypeKindPrimitive:
		if field.Type.Primitive == ir.PrimitiveDatetime {
			return "new Date(%s)"
		}
		return "%s"

	default:
		return "%s"
	}
}

// buildArrayHydration builds the hydration expression for array types.
func buildArrayHydration(parentTypeName string, field ir.Field) string {
	nameCamel := strutil.ToCamelCase(field.Name)
	itemType := *field.Type.ArrayItem

	// Get the hydration expression for the base item
	itemExpr := getItemHydrationExpr(parentTypeName, field.Name, itemType)

	if itemExpr == "el" {
		// No transformation needed
		return fmt.Sprintf("input.%s", nameCamel)
	}

	// Build nested map calls for multi-dimensional arrays
	result := fmt.Sprintf("input.%s", nameCamel)
	for i := 0; i < field.Type.ArrayDimensions; i++ {
		if i == field.Type.ArrayDimensions-1 {
			result = fmt.Sprintf("%s.map(el => %s)", result, itemExpr)
		} else {
			result = fmt.Sprintf("%s.map(arr%d => arr%d", result, i, i)
		}
	}

	// Close nested maps
	if field.Type.ArrayDimensions > 1 {
		for i := 0; i < field.Type.ArrayDimensions-1; i++ {
			result += ")"
		}
	}

	return result
}

// buildMapHydration builds the hydration expression for map types.
func buildMapHydration(parentTypeName string, field ir.Field) string {
	nameCamel := strutil.ToCamelCase(field.Name)
	valueType := *field.Type.MapValue

	valueExpr := getItemHydrationExpr(parentTypeName, field.Name, valueType)

	if valueExpr == "v" {
		// No transformation needed
		return fmt.Sprintf("input.%s", nameCamel)
	}

	return fmt.Sprintf("Object.fromEntries(Object.entries(input.%s).map(([k, v]) => [k, %s]))",
		nameCamel, valueExpr)
}

// getItemHydrationExpr returns the hydration expression for an array/map item.
func getItemHydrationExpr(parentTypeName, fieldName string, tr ir.TypeRef) string {
	switch tr.Kind {
	case ir.TypeKindObject:
		fieldPascal := strutil.ToPascalCase(fieldName)
		return fmt.Sprintf("hydrate%s%s(el)", parentTypeName, fieldPascal)

	case ir.TypeKindType:
		typePascal := strutil.ToPascalCase(tr.Type)
		return fmt.Sprintf("hydrate%s(el)", typePascal)

	case ir.TypeKindPrimitive:
		if tr.Primitive == ir.PrimitiveDatetime {
			return "new Date(el)"
		}
		return "el"

	case ir.TypeKindArray:
		// Nested arrays need recursive handling
		innerExpr := getItemHydrationExpr(parentTypeName, fieldName, *tr.ArrayItem)
		if innerExpr == "el" {
			return "el"
		}
		return fmt.Sprintf("el.map(inner => %s)", strings.ReplaceAll(innerExpr, "el", "inner"))

	case ir.TypeKindMap:
		innerExpr := getItemHydrationExpr(parentTypeName, fieldName, *tr.MapValue)
		if innerExpr == "el" {
			return "el"
		}
		return fmt.Sprintf("Object.fromEntries(Object.entries(el).map(([k, v]) => [k, %s]))",
			strings.ReplaceAll(innerExpr, "el", "v"))

	default:
		return "el"
	}
}

// renderHydrateType renders a function used to transform a type returned from JSON.parse
// to its final type.
func renderHydrateType(parentName string, name string, fields []ir.Field) string {
	fullName := parentName + name

	g := gen.New().WithSpaces(2)
	g.Linef("function hydrate%s(input: %s): %s {", fullName, fullName, fullName)
	g.Block(func() {
		for _, field := range fields {
			g.Line(renderHydrateField(fullName, field))
		}
		g.Linef("return {")
		g.Block(func() {
			for _, field := range fields {
				nameCamel := strutil.ToCamelCase(field.Name)
				hydratedName := "hydrated" + strutil.ToPascalCase(field.Name)
				g.Linef("%s: %s,", nameCamel, hydratedName)
			}
		})
		g.Linef("}")
	})
	g.Line("}")
	g.Break()

	// Render children inline types hydration functions
	for _, field := range fields {
		if field.Type.Kind == ir.TypeKindObject && field.Type.Object != nil {
			childName := fullName + strutil.ToPascalCase(field.Name)
			g.Line(renderHydrateType("", childName, field.Type.Object.Fields))
		}
	}

	return g.String()
}

// =============================================================================
// Documentation and Comments
// =============================================================================

// renderMultilineComment renders a complete multiline comment.
func renderMultilineComment(g *gen.Generator, text string) {
	g.Line("/**")
	renderPartialMultilineComment(g, text)
	g.Line(" */")
}

// renderPartialMultilineComment renders text as a partial multiline comment.
func renderPartialMultilineComment(g *gen.Generator, text string) {
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		g.Linef(" * %s", line)
	}
}

// renderDeprecated renders a deprecation comment for TypeScript.
func renderDeprecated(g *gen.Generator, deprecated *ir.Deprecation) {
	if deprecated == nil {
		return
	}

	desc := "@deprecated "
	if deprecated.Message == "" {
		desc += "This is deprecated and should not be used in new code."
	} else {
		desc += deprecated.Message
	}

	g.Line(" *")
	renderPartialMultilineComment(g, desc)
}

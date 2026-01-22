package dart

import (
	"fmt"
	"strings"

	"github.com/varavelio/gen"
	"github.com/varavelio/vdl/toolchain/internal/core/ir"
	"github.com/varavelio/vdl/toolchain/internal/util/strutil"
)

// =============================================================================
// Type Conversion: IR TypeRef → Dart Type String
// =============================================================================

// typeRefToDart converts an IR TypeRef to its Dart type string representation.
// parentTypeName is used to generate names for inline object types.
func typeRefToDart(parentTypeName string, tr ir.TypeRef) string {
	switch tr.Kind {
	case ir.TypeKindPrimitive:
		return primitiveToDart(tr.Primitive)

	case ir.TypeKindType:
		return tr.Type

	case ir.TypeKindEnum:
		return tr.Enum

	case ir.TypeKindArray:
		// Build nested List types for multi-dimensional arrays
		elementType := typeRefToDart(parentTypeName, *tr.ArrayItem)
		result := elementType
		for i := 0; i < tr.ArrayDimensions; i++ {
			result = fmt.Sprintf("List<%s>", result)
		}
		return result

	case ir.TypeKindMap:
		valueType := typeRefToDart(parentTypeName, *tr.MapValue)
		return fmt.Sprintf("Map<String, %s>", valueType)

	case ir.TypeKindObject:
		// Inline objects get a generated name based on parent
		return parentTypeName
	}

	return "dynamic"
}

// primitiveToDart converts an IR primitive type to its Dart equivalent.
func primitiveToDart(p ir.PrimitiveType) string {
	switch p {
	case ir.PrimitiveString:
		return "String"
	case ir.PrimitiveInt:
		return "int"
	case ir.PrimitiveFloat:
		return "double"
	case ir.PrimitiveBool:
		return "bool"
	case ir.PrimitiveDatetime:
		return "DateTime"
	}
	return "dynamic"
}

// =============================================================================
// JSON Parsing Expressions
// =============================================================================

// dartFromJsonExpr returns the Dart expression to parse a single field from JSON value.
func dartFromJsonExpr(parentTypeName string, field ir.Field, jsonAccessor string) string {
	return buildFromJsonExpr(parentTypeName, field.Name, field.Type, jsonAccessor)
}

// buildFromJsonExpr builds the fromJson expression for a TypeRef.
func buildFromJsonExpr(parentTypeName, fieldName string, tr ir.TypeRef, jsonAccessor string) string {
	switch tr.Kind {
	case ir.TypeKindPrimitive:
		return buildPrimitiveFromJson(tr.Primitive, jsonAccessor)

	case ir.TypeKindType:
		return fmt.Sprintf("%s.fromJson((%s as Map).cast<String, dynamic>())", tr.Type, jsonAccessor)

	case ir.TypeKindEnum:
		// Enums are just the raw value in Dart
		return jsonAccessor

	case ir.TypeKindArray:
		return buildArrayFromJson(parentTypeName, fieldName, tr, jsonAccessor)

	case ir.TypeKindMap:
		return buildMapFromJson(parentTypeName, fieldName, tr, jsonAccessor)

	case ir.TypeKindObject:
		inlineName := parentTypeName + strutil.ToPascalCase(fieldName)
		return fmt.Sprintf("%s.fromJson((%s as Map).cast<String, dynamic>())", inlineName, jsonAccessor)
	}

	return jsonAccessor
}

// buildPrimitiveFromJson builds the fromJson expression for primitive types.
func buildPrimitiveFromJson(p ir.PrimitiveType, jsonAccessor string) string {
	switch p {
	case ir.PrimitiveString:
		return fmt.Sprintf("%s as String", jsonAccessor)
	case ir.PrimitiveInt:
		return fmt.Sprintf("(%s as num).toInt()", jsonAccessor)
	case ir.PrimitiveFloat:
		return fmt.Sprintf("(%s as num).toDouble()", jsonAccessor)
	case ir.PrimitiveBool:
		return fmt.Sprintf("%s as bool", jsonAccessor)
	case ir.PrimitiveDatetime:
		return fmt.Sprintf("DateTime.parse(%s as String)", jsonAccessor)
	}
	return jsonAccessor
}

// buildArrayFromJson builds the fromJson expression for array types.
func buildArrayFromJson(parentTypeName, fieldName string, tr ir.TypeRef, jsonAccessor string) string {
	itemExpr := buildItemFromJsonExpr(parentTypeName, fieldName, *tr.ArrayItem, "e")

	// For multi-dimensional arrays, we need nested maps
	if tr.ArrayDimensions > 1 {
		// Build nested map expression
		result := fmt.Sprintf("((%s as List).map((e) => %s).toList())", jsonAccessor, buildNestedArrayFromJson(parentTypeName, fieldName, *tr.ArrayItem, tr.ArrayDimensions-1, "e"))
		return result
	}

	return fmt.Sprintf("((%s as List).map((e) => %s).toList())", jsonAccessor, itemExpr)
}

// buildNestedArrayFromJson builds nested array parsing for multi-dimensional arrays.
func buildNestedArrayFromJson(parentTypeName, fieldName string, itemType ir.TypeRef, remainingDims int, varName string) string {
	if remainingDims == 0 {
		return buildItemFromJsonExpr(parentTypeName, fieldName, itemType, varName)
	}

	innerExpr := buildNestedArrayFromJson(parentTypeName, fieldName, itemType, remainingDims-1, "inner")
	return fmt.Sprintf("((%s as List).map((inner) => %s).toList())", varName, innerExpr)
}

// buildItemFromJsonExpr builds the expression for parsing a single array/map item.
func buildItemFromJsonExpr(parentTypeName, fieldName string, tr ir.TypeRef, varName string) string {
	switch tr.Kind {
	case ir.TypeKindPrimitive:
		switch tr.Primitive {
		case ir.PrimitiveString:
			return fmt.Sprintf("%s as String", varName)
		case ir.PrimitiveInt:
			return fmt.Sprintf("(%s as num).toInt()", varName)
		case ir.PrimitiveFloat:
			return fmt.Sprintf("(%s as num).toDouble()", varName)
		case ir.PrimitiveBool:
			return fmt.Sprintf("%s as bool", varName)
		case ir.PrimitiveDatetime:
			return fmt.Sprintf("DateTime.parse(%s as String)", varName)
		}

	case ir.TypeKindType:
		return fmt.Sprintf("%s.fromJson((%s as Map).cast<String, dynamic>())", tr.Type, varName)

	case ir.TypeKindEnum:
		return varName

	case ir.TypeKindObject:
		inlineName := parentTypeName + strutil.ToPascalCase(fieldName)
		return fmt.Sprintf("%s.fromJson((%s as Map).cast<String, dynamic>())", inlineName, varName)

	case ir.TypeKindArray:
		innerExpr := buildItemFromJsonExpr(parentTypeName, fieldName, *tr.ArrayItem, "inner")
		return fmt.Sprintf("((%s as List).map((inner) => %s).toList())", varName, innerExpr)

	case ir.TypeKindMap:
		innerExpr := buildItemFromJsonExpr(parentTypeName, fieldName, *tr.MapValue, "v")
		return fmt.Sprintf("((%s as Map).cast<String, dynamic>().map((k, v) => MapEntry(k, %s)))", varName, innerExpr)
	}

	return varName
}

// buildMapFromJson builds the fromJson expression for map types.
func buildMapFromJson(parentTypeName, fieldName string, tr ir.TypeRef, jsonAccessor string) string {
	valueExpr := buildItemFromJsonExpr(parentTypeName, fieldName, *tr.MapValue, "v")
	return fmt.Sprintf("((%s as Map).cast<String, dynamic>().map((k, v) => MapEntry(k, %s)))", jsonAccessor, valueExpr)
}

// =============================================================================
// JSON Serialization Expressions
// =============================================================================

// dartToJsonExpr returns the Dart expression to serialise a field to JSON.
func dartToJsonExpr(field ir.Field, varName string) string {
	return buildToJsonExpr(field.Type, varName)
}

// buildToJsonExpr builds the toJson expression for a TypeRef.
func buildToJsonExpr(tr ir.TypeRef, varName string) string {
	switch tr.Kind {
	case ir.TypeKindPrimitive:
		if tr.Primitive == ir.PrimitiveDatetime {
			return fmt.Sprintf("%s.toUtc().toIso8601String()", varName)
		}
		return varName

	case ir.TypeKindType, ir.TypeKindObject:
		return fmt.Sprintf("%s.toJson()", varName)

	case ir.TypeKindEnum:
		return varName

	case ir.TypeKindArray:
		itemExpr := buildItemToJsonExpr(*tr.ArrayItem, "e")
		if itemExpr == "e" {
			return varName
		}
		return fmt.Sprintf("%s.map((e) => %s).toList()", varName, itemExpr)

	case ir.TypeKindMap:
		valueExpr := buildItemToJsonExpr(*tr.MapValue, "v")
		if valueExpr == "v" {
			return varName
		}
		return fmt.Sprintf("%s.map((k, v) => MapEntry(k, %s))", varName, valueExpr)
	}

	return varName
}

// buildItemToJsonExpr builds the toJson expression for a single array/map item.
func buildItemToJsonExpr(tr ir.TypeRef, varName string) string {
	switch tr.Kind {
	case ir.TypeKindPrimitive:
		if tr.Primitive == ir.PrimitiveDatetime {
			return fmt.Sprintf("%s.toUtc().toIso8601String()", varName)
		}
		return varName

	case ir.TypeKindType, ir.TypeKindObject:
		return fmt.Sprintf("%s.toJson()", varName)

	case ir.TypeKindEnum:
		return varName

	case ir.TypeKindArray:
		innerExpr := buildItemToJsonExpr(*tr.ArrayItem, "inner")
		if innerExpr == "inner" {
			return varName
		}
		return fmt.Sprintf("%s.map((inner) => %s).toList()", varName, innerExpr)

	case ir.TypeKindMap:
		innerExpr := buildItemToJsonExpr(*tr.MapValue, "v2")
		if innerExpr == "v2" {
			return varName
		}
		return fmt.Sprintf("%s.map((k, v2) => MapEntry(k, %s))", varName, innerExpr)
	}

	return varName
}

// =============================================================================
// Type Structure Rendering
// =============================================================================

// renderDartType renders a Dart class for given fields, including a short description,
// a factory constructor to hydrate from JSON and a toJson method for serialisation.
func renderDartType(parentName, name, desc string, fields []ir.Field) string {
	fullName := parentName + name

	g := gen.New().WithSpaces(2)
	if desc != "" {
		g.Line("/// " + strings.ReplaceAll(desc, "\n", "\n/// "))
	}
	g.Linef("class %s {", fullName)
	g.Block(func() {
		// Fields
		for _, field := range fields {
			fieldName := strutil.ToCamelCase(field.Name)
			inlineTypeName := fullName + strutil.ToPascalCase(field.Name)
			typeLit := typeRefToDart(inlineTypeName, field.Type)
			if field.Optional {
				typeLit = typeLit + "?"
			}
			// Field description if present
			if field.Doc != "" {
				g.Line("/// " + strings.ReplaceAll(strings.TrimSpace(field.Doc), "\n", "\n/// "))
			}
			g.Linef("final %s %s;", typeLit, fieldName)
		}
		g.Break()

		// Constructor
		g.Linef("/// Creates a new %s instance.", fullName)
		if len(fields) == 0 {
			g.Linef("const %s();", fullName)
		} else {
			g.Linef("const %s({", fullName)
			g.Block(func() {
				for _, field := range fields {
					fieldName := strutil.ToCamelCase(field.Name)
					isRequired := !field.Optional
					if isRequired {
						g.Linef("required this.%s,", fieldName)
					} else {
						g.Linef("this.%s,", fieldName)
					}
				}
			})
			g.Line("});")
		}

		g.Break()

		// fromJson factory
		g.Linef("/// Hydrates a %s from a JSON map.", fullName)
		g.Linef("factory %s.fromJson(Map<String, dynamic> json) {", fullName)
		g.Block(func() {
			for _, field := range fields {
				fieldName := strutil.ToCamelCase(field.Name)
				jsonKey := strutil.ToCamelCase(field.Name)
				jsonAccessor := fmt.Sprintf("json['%s']", jsonKey)
				parseExpr := dartFromJsonExpr(fullName, field, jsonAccessor)
				if field.Optional {
					g.Linef("final %s = json.containsKey('%s') && %s != null ? %s : null;", fieldName, jsonKey, jsonAccessor, parseExpr)
				} else {
					g.Linef("final %s = %s;", fieldName, parseExpr)
				}
			}
			// return
			g.Linef("return %s(", fullName)
			g.Block(func() {
				for _, field := range fields {
					fieldName := strutil.ToCamelCase(field.Name)
					g.Linef("%s: %s,", fieldName, fieldName)
				}
			})
			g.Line(");")
		})
		g.Line("}")
		g.Break()

		// toJson method
		g.Linef("/// Serialises this %s to a JSON map compatible with the server.", fullName)
		g.Line("Map<String, dynamic> toJson() {")
		g.Block(func() {
			g.Line("final _data = <String, dynamic>{};")
			for _, field := range fields {
				fieldName := strutil.ToCamelCase(field.Name)
				jsonKey := strutil.ToCamelCase(field.Name)
				if field.Optional {
					local := "__v_" + fieldName
					g.Linef("final %s = %s;", local, fieldName)
					ser := dartToJsonExpr(field, local)
					g.Linef("if (%s != null) _data['%s'] = %s;", local, jsonKey, ser)
				} else {
					ser := dartToJsonExpr(field, fieldName)
					g.Linef("_data['%s'] = %s;", jsonKey, ser)
				}
			}
			g.Line("return _data;")
		})
		g.Line("}")
	})
	g.Line("}")
	g.Break()

	// Children inline types
	for _, field := range fields {
		if field.Type.Kind == ir.TypeKindObject && field.Type.Object != nil {
			childName := fullName + strutil.ToPascalCase(field.Name)
			childDesc := field.Doc
			g.Line(renderDartType("", childName, childDesc, field.Type.Object.Fields))
		}
	}

	return g.String()
}

// =============================================================================
// Documentation and Comments
// =============================================================================

// renderDeprecatedDart writes a deprecated doc line if provided.
func renderDeprecatedDart(g *gen.Generator, deprecated *ir.Deprecation) {
	if deprecated == nil {
		return
	}
	desc := "@deprecated "
	if deprecated.Message == "" {
		desc += "This is deprecated and should not be used in new code."
	} else {
		desc += deprecated.Message
	}
	g.Line("///")
	for _, line := range strings.Split(desc, "\n") {
		g.Linef("/// %s", line)
	}
}

// renderMultilineCommentDart renders a complete multiline comment for Dart.
func renderMultilineCommentDart(g *gen.Generator, text string) {
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		g.Linef("/// %s", line)
	}
}

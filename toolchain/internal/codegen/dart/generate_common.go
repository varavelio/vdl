package dart

import (
	"fmt"
	"strings"

	"github.com/varavelio/gen"
	"github.com/varavelio/vdl/toolchain/internal/core/ir/irtypes"
	"github.com/varavelio/vdl/toolchain/internal/util/strutil"
)

// =============================================================================
// Type Conversion: IR TypeRef → Dart Type String
// =============================================================================

// typeRefToDart converts an IR TypeRef to its Dart type string representation.
// parentTypeName is used to generate names for inline object types.
func typeRefToDart(parentTypeName string, tr irtypes.TypeRef) string {
	switch tr.Kind {
	case irtypes.TypeKindPrimitive:
		return primitiveToDart(tr.GetPrimitiveName())

	case irtypes.TypeKindType:
		return tr.GetTypeName()

	case irtypes.TypeKindEnum:
		return tr.GetEnumName()

	case irtypes.TypeKindArray:
		// Build nested List types for multi-dimensional arrays
		elementType := typeRefToDart(parentTypeName, *tr.ArrayType)
		result := elementType
		for i := int64(0); i < tr.GetArrayDims(); i++ {
			result = fmt.Sprintf("List<%s>", result)
		}
		return result

	case irtypes.TypeKindMap:
		valueType := typeRefToDart(parentTypeName, *tr.MapType)
		return fmt.Sprintf("Map<String, %s>", valueType)

	case irtypes.TypeKindObject:
		// Inline objects get a generated name based on parent
		return parentTypeName
	}

	return "dynamic"
}

// primitiveToDart converts an IR primitive type to its Dart equivalent.
func primitiveToDart(p irtypes.PrimitiveType) string {
	switch p {
	case irtypes.PrimitiveTypeString:
		return "String"
	case irtypes.PrimitiveTypeInt:
		return "int"
	case irtypes.PrimitiveTypeFloat:
		return "double"
	case irtypes.PrimitiveTypeBool:
		return "bool"
	case irtypes.PrimitiveTypeDatetime:
		return "DateTime"
	}
	return "dynamic"
}

// =============================================================================
// JSON Parsing Expressions
// =============================================================================

// dartFromJsonExpr returns the Dart expression to parse a single field from JSON value.
func dartFromJsonExpr(parentTypeName string, field irtypes.Field, jsonAccessor string) string {
	return buildFromJsonExpr(parentTypeName, field.Name, field.TypeRef, jsonAccessor)
}

// buildFromJsonExpr builds the fromJson expression for a TypeRef.
func buildFromJsonExpr(parentTypeName, fieldName string, tr irtypes.TypeRef, jsonAccessor string) string {
	switch tr.Kind {
	case irtypes.TypeKindPrimitive:
		return buildPrimitiveFromJson(tr.GetPrimitiveName(), jsonAccessor)

	case irtypes.TypeKindType:
		return fmt.Sprintf("%s.fromJson((%s as Map).cast<String, dynamic>())", tr.GetTypeName(), jsonAccessor)

	case irtypes.TypeKindEnum:
		// Enums need to be converted from JSON value using the extension's fromJson method
		return fmt.Sprintf("%sJson.fromJson(%s)", tr.GetEnumName(), jsonAccessor)

	case irtypes.TypeKindArray:
		return buildArrayFromJson(parentTypeName, fieldName, tr, jsonAccessor)

	case irtypes.TypeKindMap:
		return buildMapFromJson(parentTypeName, fieldName, tr, jsonAccessor)

	case irtypes.TypeKindObject:
		inlineName := parentTypeName + strutil.ToPascalCase(fieldName)
		return fmt.Sprintf("%s.fromJson((%s as Map).cast<String, dynamic>())", inlineName, jsonAccessor)
	}

	return jsonAccessor
}

// buildPrimitiveFromJson builds the fromJson expression for primitive types.
func buildPrimitiveFromJson(p irtypes.PrimitiveType, jsonAccessor string) string {
	switch p {
	case irtypes.PrimitiveTypeString:
		return fmt.Sprintf("%s as String", jsonAccessor)
	case irtypes.PrimitiveTypeInt:
		return fmt.Sprintf("(%s as num).toInt()", jsonAccessor)
	case irtypes.PrimitiveTypeFloat:
		return fmt.Sprintf("(%s as num).toDouble()", jsonAccessor)
	case irtypes.PrimitiveTypeBool:
		return fmt.Sprintf("%s as bool", jsonAccessor)
	case irtypes.PrimitiveTypeDatetime:
		return fmt.Sprintf("DateTime.parse(%s as String)", jsonAccessor)
	}
	return jsonAccessor
}

// buildArrayFromJson builds the fromJson expression for array types.
func buildArrayFromJson(parentTypeName, fieldName string, tr irtypes.TypeRef, jsonAccessor string) string {
	itemExpr := buildItemFromJsonExpr(parentTypeName, fieldName, *tr.ArrayType, "e")

	// For multi-dimensional arrays, we need nested maps
	if tr.GetArrayDims() > 1 {
		// Build nested map expression
		result := fmt.Sprintf("((%s as List).map((e) => %s).toList())", jsonAccessor, buildNestedArrayFromJson(parentTypeName, fieldName, *tr.ArrayType, tr.GetArrayDims()-1, "e"))
		return result
	}

	return fmt.Sprintf("((%s as List).map((e) => %s).toList())", jsonAccessor, itemExpr)
}

// buildNestedArrayFromJson builds nested array parsing for multi-dimensional arrays.
func buildNestedArrayFromJson(parentTypeName, fieldName string, itemType irtypes.TypeRef, remainingDims int64, varName string) string {
	if remainingDims == 0 {
		return buildItemFromJsonExpr(parentTypeName, fieldName, itemType, varName)
	}

	innerExpr := buildNestedArrayFromJson(parentTypeName, fieldName, itemType, remainingDims-1, "inner")
	return fmt.Sprintf("((%s as List).map((inner) => %s).toList())", varName, innerExpr)
}

// buildItemFromJsonExpr builds the expression for parsing a single array/map item.
func buildItemFromJsonExpr(parentTypeName, fieldName string, tr irtypes.TypeRef, varName string) string {
	switch tr.Kind {
	case irtypes.TypeKindPrimitive:
		switch tr.GetPrimitiveName() {
		case irtypes.PrimitiveTypeString:
			return fmt.Sprintf("%s as String", varName)
		case irtypes.PrimitiveTypeInt:
			return fmt.Sprintf("(%s as num).toInt()", varName)
		case irtypes.PrimitiveTypeFloat:
			return fmt.Sprintf("(%s as num).toDouble()", varName)
		case irtypes.PrimitiveTypeBool:
			return fmt.Sprintf("%s as bool", varName)
		case irtypes.PrimitiveTypeDatetime:
			return fmt.Sprintf("DateTime.parse(%s as String)", varName)
		}

	case irtypes.TypeKindType:
		return fmt.Sprintf("%s.fromJson((%s as Map).cast<String, dynamic>())", tr.GetTypeName(), varName)

	case irtypes.TypeKindEnum:
		// Enums need to be converted using the extension's fromJson method
		return fmt.Sprintf("%sJson.fromJson(%s)", tr.GetEnumName(), varName)

	case irtypes.TypeKindObject:
		inlineName := parentTypeName + strutil.ToPascalCase(fieldName)
		return fmt.Sprintf("%s.fromJson((%s as Map).cast<String, dynamic>())", inlineName, varName)

	case irtypes.TypeKindArray:
		innerExpr := buildItemFromJsonExpr(parentTypeName, fieldName, *tr.ArrayType, "inner")
		return fmt.Sprintf("((%s as List).map((inner) => %s).toList())", varName, innerExpr)

	case irtypes.TypeKindMap:
		innerExpr := buildItemFromJsonExpr(parentTypeName, fieldName, *tr.MapType, "v")
		return fmt.Sprintf("((%s as Map).cast<String, dynamic>().map((k, v) => MapEntry(k, %s)))", varName, innerExpr)
	}

	return varName
}

// buildMapFromJson builds the fromJson expression for map types.
func buildMapFromJson(parentTypeName, fieldName string, tr irtypes.TypeRef, jsonAccessor string) string {
	valueExpr := buildItemFromJsonExpr(parentTypeName, fieldName, *tr.MapType, "v")
	return fmt.Sprintf("((%s as Map).cast<String, dynamic>().map((k, v) => MapEntry(k, %s)))", jsonAccessor, valueExpr)
}

// =============================================================================
// JSON Serialization Expressions
// =============================================================================

// buildNestedArrayToJson builds nested array serialization for multi-dimensional arrays.
func buildNestedArrayToJson(itemType irtypes.TypeRef, remainingDims int64, varName string) string {
	if remainingDims == 1 {
		// Base case: innermost dimension
		itemExpr := buildItemToJsonExpr(itemType, "e")
		if itemExpr == "e" {
			return varName
		}
		return fmt.Sprintf("%s.map((e) => %s).toList()", varName, itemExpr)
	}

	// Recursive case: more dimensions to process
	innerExpr := buildNestedArrayToJson(itemType, remainingDims-1, "inner")
	if innerExpr == "inner" {
		return varName
	}
	return fmt.Sprintf("%s.map((inner) => %s).toList()", varName, innerExpr)
}

// dartToJsonExpr returns the Dart expression to serialise a field to JSON.
func dartToJsonExpr(field irtypes.Field, varName string) string {
	return buildToJsonExpr(field.TypeRef, varName)
}

// buildToJsonExpr builds the toJson expression for a TypeRef.
func buildToJsonExpr(tr irtypes.TypeRef, varName string) string {
	switch tr.Kind {
	case irtypes.TypeKindPrimitive:
		if tr.GetPrimitiveName() == irtypes.PrimitiveTypeDatetime {
			return fmt.Sprintf("%s.toUtc().toIso8601String()", varName)
		}
		return varName

	case irtypes.TypeKindType, irtypes.TypeKindObject:
		return fmt.Sprintf("%s.toJson()", varName)

	case irtypes.TypeKindEnum:
		// Enums use the toJson method from the extension
		return fmt.Sprintf("%s.toJson()", varName)

	case irtypes.TypeKindArray:
		// For multi-dimensional arrays, we need nested maps
		if tr.GetArrayDims() > 1 {
			return buildNestedArrayToJson(*tr.ArrayType, tr.GetArrayDims(), varName)
		}
		itemExpr := buildItemToJsonExpr(*tr.ArrayType, "e")
		if itemExpr == "e" {
			return varName
		}
		return fmt.Sprintf("%s.map((e) => %s).toList()", varName, itemExpr)

	case irtypes.TypeKindMap:
		valueExpr := buildItemToJsonExpr(*tr.MapType, "v")
		if valueExpr == "v" {
			return varName
		}
		return fmt.Sprintf("%s.map((k, v) => MapEntry(k, %s))", varName, valueExpr)
	}

	return varName
}

// buildItemToJsonExpr builds the toJson expression for a single array/map item.
func buildItemToJsonExpr(tr irtypes.TypeRef, varName string) string {
	switch tr.Kind {
	case irtypes.TypeKindPrimitive:
		if tr.GetPrimitiveName() == irtypes.PrimitiveTypeDatetime {
			return fmt.Sprintf("%s.toUtc().toIso8601String()", varName)
		}
		return varName

	case irtypes.TypeKindType, irtypes.TypeKindObject:
		return fmt.Sprintf("%s.toJson()", varName)

	case irtypes.TypeKindEnum:
		// Enums use the toJson method from the extension
		return fmt.Sprintf("%s.toJson()", varName)

	case irtypes.TypeKindArray:
		innerExpr := buildItemToJsonExpr(*tr.ArrayType, "inner")
		if innerExpr == "inner" {
			return varName
		}
		return fmt.Sprintf("%s.map((inner) => %s).toList()", varName, innerExpr)

	case irtypes.TypeKindMap:
		innerExpr := buildItemToJsonExpr(*tr.MapType, "v2")
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
func renderDartType(parentName, name, desc string, fields []irtypes.Field) string {
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
			typeLit := typeRefToDart(inlineTypeName, field.TypeRef)
			if field.Optional {
				typeLit = typeLit + "?"
			}
			// Field description if present
			if field.GetDoc() != "" {
				g.Line("/// " + strings.ReplaceAll(strings.TrimSpace(field.GetDoc()), "\n", "\n/// "))
			}
			g.Linef("final %s %s;", typeLit, fieldName)
		}
		g.Break()

		// Constructor
		g.Linef("/// Creates a new [%s] instance.", fullName)
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
		g.Linef("/// Creates a [%s] from a JSON map.", fullName)
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
		g.Linef("/// Converts this [%s] to a JSON map.", fullName)
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
		g.Break()

		// copyWith method
		if len(fields) > 0 {
			g.Linef("/// Creates a copy of this [%s] with the given fields replaced.", fullName)
			g.Linef("%s copyWith({", fullName)
			g.Block(func() {
				for _, field := range fields {
					fieldName := strutil.ToCamelCase(field.Name)
					inlineTypeName := fullName + strutil.ToPascalCase(field.Name)
					typeLit := typeRefToDart(inlineTypeName, field.TypeRef)
					// All fields are optional in copyWith
					g.Linef("%s? %s,", typeLit, fieldName)
				}
			})
			g.Line("}) {")
			g.Block(func() {
				g.Linef("return %s(", fullName)
				g.Block(func() {
					for _, field := range fields {
						fieldName := strutil.ToCamelCase(field.Name)
						g.Linef("%s: %s ?? this.%s,", fieldName, fieldName, fieldName)
					}
				})
				g.Line(");")
			})
			g.Line("}")
			g.Break()
		}

		// == operator
		g.Line("@override")
		g.Line("bool operator ==(Object other) {")
		g.Block(func() {
			g.Line("if (identical(this, other)) return true;")
			g.Linef("return other is %s", fullName)
			if len(fields) > 0 {
				for i, field := range fields {
					fieldName := strutil.ToCamelCase(field.Name)
					if i == len(fields)-1 {
						g.Linef("    && %s == other.%s;", fieldName, fieldName)
					} else {
						g.Linef("    && %s == other.%s", fieldName, fieldName)
					}
				}
			} else {
				g.Line(";")
			}
		})
		g.Line("}")
		g.Break()

		// hashCode
		g.Line("@override")
		if len(fields) == 0 {
			g.Line("int get hashCode => 0;")
		} else if len(fields) == 1 {
			fieldName := strutil.ToCamelCase(fields[0].Name)
			g.Linef("int get hashCode => %s.hashCode;", fieldName)
		} else {
			g.Line("int get hashCode => Object.hash(")
			g.Block(func() {
				for i, field := range fields {
					fieldName := strutil.ToCamelCase(field.Name)
					if i == len(fields)-1 {
						g.Linef("%s,", fieldName)
					} else {
						g.Linef("%s,", fieldName)
					}
				}
			})
			g.Line(");")
		}
		g.Break()

		// toString
		g.Line("@override")
		if len(fields) == 0 {
			g.Linef("String toString() => '%s()';", fullName)
		} else {
			g.Linef("String toString() {")
			g.Block(func() {
				g.Linef("return '%s('", fullName)
				for i, field := range fields {
					fieldName := strutil.ToCamelCase(field.Name)
					if i == len(fields)-1 {
						g.Linef("    '%s: $%s'", fieldName, fieldName)
					} else {
						g.Linef("    '%s: $%s, '", fieldName, fieldName)
					}
				}
				g.Line("    ')';")
			})
			g.Line("}")
		}
	})
	g.Line("}")
	g.Break()

	// Children inline types - recursively extract from arrays, maps, and nested objects
	inlineTypes := extractAllInlineTypes(fullName, fields)
	for _, inlineType := range inlineTypes {
		g.Line(renderInlineType(inlineType.name, inlineType.doc, inlineType.fields))
	}

	return g.String()
}

// =============================================================================
// Inline Type Extraction
// =============================================================================

// inlineTypeInfo represents an inline type that needs to be generated.
type inlineTypeInfo struct {
	name   string
	doc    string
	fields []irtypes.Field
}

// extractInlineTypes recursively extracts all inline object types from a TypeRef.
// parentName is the full name prefix for the inline type.
func extractInlineTypes(parentName string, tr irtypes.TypeRef) []inlineTypeInfo {
	var result []inlineTypeInfo

	switch tr.Kind {
	case irtypes.TypeKindObject:
		if tr.ObjectFields != nil {
			result = append(result, inlineTypeInfo{
				name:   parentName,
				doc:    "",
				fields: *tr.ObjectFields,
			})
			// Recursively extract from child fields
			for _, f := range *tr.ObjectFields {
				childName := parentName + strutil.ToPascalCase(f.Name)
				result = append(result, extractInlineTypes(childName, f.TypeRef)...)
			}
		}

	case irtypes.TypeKindArray:
		if tr.ArrayType != nil {
			// For arrays, the inline type name is the same as parentName
			result = append(result, extractInlineTypes(parentName, *tr.ArrayType)...)
		}

	case irtypes.TypeKindMap:
		if tr.MapType != nil {
			// For maps, the inline type name is the same as parentName
			result = append(result, extractInlineTypes(parentName, *tr.MapType)...)
		}
	}

	return result
}

// extractAllInlineTypes extracts all inline types from a list of fields.
func extractAllInlineTypes(parentName string, fields []irtypes.Field) []inlineTypeInfo {
	var result []inlineTypeInfo
	for _, field := range fields {
		childName := parentName + strutil.ToPascalCase(field.Name)
		inlines := extractInlineTypes(childName, field.TypeRef)
		result = append(result, inlines...)
	}
	return result
}

// renderInlineType renders a single inline type class without recursively rendering children
// (since extractAllInlineTypes already flattens the hierarchy).
func renderInlineType(name, desc string, fields []irtypes.Field) string {
	g := gen.New().WithSpaces(2)
	if desc != "" {
		g.Line("/// " + strings.ReplaceAll(desc, "\n", "\n/// "))
	}
	g.Linef("class %s {", name)
	g.Block(func() {
		// Fields
		for _, field := range fields {
			fieldName := strutil.ToCamelCase(field.Name)
			inlineTypeName := name + strutil.ToPascalCase(field.Name)
			typeLit := typeRefToDart(inlineTypeName, field.TypeRef)
			if field.Optional {
				typeLit = typeLit + "?"
			}
			if field.GetDoc() != "" {
				g.Line("/// " + strings.ReplaceAll(strings.TrimSpace(field.GetDoc()), "\n", "\n/// "))
			}
			g.Linef("final %s %s;", typeLit, fieldName)
		}
		g.Break()

		// Constructor
		g.Linef("/// Creates a new [%s] instance.", name)
		if len(fields) == 0 {
			g.Linef("const %s();", name)
		} else {
			g.Linef("const %s({", name)
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
		g.Linef("/// Creates a [%s] from a JSON map.", name)
		g.Linef("factory %s.fromJson(Map<String, dynamic> json) {", name)
		g.Block(func() {
			for _, field := range fields {
				fieldName := strutil.ToCamelCase(field.Name)
				jsonKey := strutil.ToCamelCase(field.Name)
				jsonAccessor := fmt.Sprintf("json['%s']", jsonKey)
				parseExpr := dartFromJsonExpr(name, field, jsonAccessor)
				if field.Optional {
					g.Linef("final %s = json.containsKey('%s') && %s != null ? %s : null;", fieldName, jsonKey, jsonAccessor, parseExpr)
				} else {
					g.Linef("final %s = %s;", fieldName, parseExpr)
				}
			}
			g.Linef("return %s(", name)
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
		g.Linef("/// Converts this [%s] to a JSON map.", name)
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
		g.Break()

		// copyWith method
		if len(fields) > 0 {
			g.Linef("/// Creates a copy of this [%s] with the given fields replaced.", name)
			g.Linef("%s copyWith({", name)
			g.Block(func() {
				for _, field := range fields {
					fieldName := strutil.ToCamelCase(field.Name)
					inlineTypeName := name + strutil.ToPascalCase(field.Name)
					typeLit := typeRefToDart(inlineTypeName, field.TypeRef)
					g.Linef("%s? %s,", typeLit, fieldName)
				}
			})
			g.Line("}) {")
			g.Block(func() {
				g.Linef("return %s(", name)
				g.Block(func() {
					for _, field := range fields {
						fieldName := strutil.ToCamelCase(field.Name)
						g.Linef("%s: %s ?? this.%s,", fieldName, fieldName, fieldName)
					}
				})
				g.Line(");")
			})
			g.Line("}")
			g.Break()
		}

		// == operator
		g.Line("@override")
		g.Line("bool operator ==(Object other) {")
		g.Block(func() {
			g.Line("if (identical(this, other)) return true;")
			g.Linef("return other is %s", name)
			if len(fields) > 0 {
				for i, field := range fields {
					fieldName := strutil.ToCamelCase(field.Name)
					if i == len(fields)-1 {
						g.Linef("    && %s == other.%s;", fieldName, fieldName)
					} else {
						g.Linef("    && %s == other.%s", fieldName, fieldName)
					}
				}
			} else {
				g.Line(";")
			}
		})
		g.Line("}")
		g.Break()

		// hashCode
		g.Line("@override")
		if len(fields) == 0 {
			g.Line("int get hashCode => 0;")
		} else if len(fields) == 1 {
			fieldName := strutil.ToCamelCase(fields[0].Name)
			g.Linef("int get hashCode => %s.hashCode;", fieldName)
		} else {
			g.Line("int get hashCode => Object.hash(")
			g.Block(func() {
				for _, field := range fields {
					fieldName := strutil.ToCamelCase(field.Name)
					g.Linef("%s,", fieldName)
				}
			})
			g.Line(");")
		}
		g.Break()

		// toString
		g.Line("@override")
		if len(fields) == 0 {
			g.Linef("String toString() => '%s()';", name)
		} else {
			g.Linef("String toString() {")
			g.Block(func() {
				g.Linef("return '%s('", name)
				for i, field := range fields {
					fieldName := strutil.ToCamelCase(field.Name)
					if i == len(fields)-1 {
						g.Linef("    '%s: $%s'", fieldName, fieldName)
					} else {
						g.Linef("    '%s: $%s, '", fieldName, fieldName)
					}
				}
				g.Line("    ')';")
			})
			g.Line("}")
		}
	})
	g.Line("}")
	g.Break()

	return g.String()
}

// =============================================================================
// Documentation and Comments
// =============================================================================

// renderDeprecatedDart writes a deprecated doc line if provided.
// deprecated is a *string (nil if not deprecated, message if deprecated)
func renderDeprecatedDart(g *gen.Generator, deprecated *string) {
	if deprecated == nil {
		return
	}
	desc := "@deprecated "
	if *deprecated == "" {
		desc += "This is deprecated and should not be used in new code."
	} else {
		desc += *deprecated
	}
	g.Line("///")
	for _, line := range strings.Split(desc, "\n") {
		g.Linef("/// %s", line)
	}
}

// renderMultilineCommentDart renders a complete multiline comment for Dart.
func renderMultilineCommentDart(g *gen.Generator, text string) {
	lines := strings.SplitSeq(text, "\n")
	for line := range lines {
		g.Linef("/// %s", line)
	}
}

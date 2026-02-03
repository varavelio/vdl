package openapi

import (
	"github.com/varavelio/vdl/toolchain/internal/core/ir/irtypes"
)

// generateTypeRefSchema converts an IR TypeRef to a JSON Schema representation.
func generateTypeRefSchema(t irtypes.TypeRef) map[string]any {
	switch t.Kind {
	case irtypes.TypeKindPrimitive:
		return primitiveToJSONSchema(t.GetPrimitiveName())

	case irtypes.TypeKindType:
		return map[string]any{
			"$ref": "#/components/schemas/" + t.GetTypeName(),
		}

	case irtypes.TypeKindEnum:
		return map[string]any{
			"$ref": "#/components/schemas/" + t.GetEnumName(),
		}

	case irtypes.TypeKindArray:
		itemSchema := generateTypeRefSchema(t.GetArrayType())
		// For multi-dimensional arrays, we need to nest the array schema
		dims := t.GetArrayDims()
		for i := int64(1); i < dims; i++ {
			itemSchema = map[string]any{
				"type":  "array",
				"items": itemSchema,
			}
		}
		return map[string]any{
			"type":  "array",
			"items": itemSchema,
		}

	case irtypes.TypeKindMap:
		return map[string]any{
			"type":                 "object",
			"additionalProperties": generateTypeRefSchema(t.GetMapType()),
		}

	case irtypes.TypeKindObject:
		props, required := generatePropertiesFromFields(t.GetObjectFields())
		schema := map[string]any{
			"type":       "object",
			"properties": props,
		}
		if len(required) > 0 {
			schema["required"] = required
		}
		return schema
	}

	return map[string]any{}
}

// primitiveToJSONSchema converts an IR primitive type to JSON Schema.
func primitiveToJSONSchema(p irtypes.PrimitiveType) map[string]any {
	switch p {
	case irtypes.PrimitiveTypeString:
		return map[string]any{"type": "string"}
	case irtypes.PrimitiveTypeInt:
		return map[string]any{"type": "integer"}
	case irtypes.PrimitiveTypeFloat:
		return map[string]any{"type": "number"}
	case irtypes.PrimitiveTypeBool:
		return map[string]any{"type": "boolean"}
	case irtypes.PrimitiveTypeDatetime:
		return map[string]any{"type": "string", "format": "date-time"}
	}
	return map[string]any{"type": "string"}
}

// generatePropertiesFromFields generates JSON schema properties from IR fields.
// Returns the properties map and a list of required field names.
func generatePropertiesFromFields(fields []irtypes.Field) (map[string]any, []string) {
	properties := map[string]any{}
	required := []string{}

	for _, field := range fields {
		prop := generateTypeRefSchema(field.TypeRef)

		// Add description if present
		doc := field.GetDoc()
		if doc != "" {
			// If prop is a $ref, we need to wrap it in allOf to add description
			if _, hasRef := prop["$ref"]; hasRef {
				prop = map[string]any{
					"allOf": []map[string]any{
						prop,
						{"description": doc},
					},
				}
			} else {
				prop["description"] = doc
			}
		}

		properties[field.Name] = prop

		if !field.Optional {
			required = append(required, field.Name)
		}
	}

	return properties, required
}

// generateOutputProperties generates the output wrapper with ok/error structure.
// This follows the VDL response lifecycle spec.
func generateOutputProperties(fields []irtypes.Field) (map[string]any, []string) {
	outputProperties, outputRequiredFields := generatePropertiesFromFields(fields)
	output := map[string]any{
		"type":       "object",
		"properties": outputProperties,
	}
	if len(outputRequiredFields) > 0 {
		output["required"] = outputRequiredFields
	}

	properties := map[string]any{
		"ok":     map[string]any{"type": "boolean"},
		"output": output,
		"error": map[string]any{
			"type": "object",
			"properties": map[string]any{
				"message": map[string]any{
					"type": "string",
				},
				"category": map[string]any{
					"type": "string",
				},
				"code": map[string]any{
					"type": "string",
				},
				"details": map[string]any{
					"type":                 "object",
					"properties":           map[string]any{},
					"additionalProperties": true,
				},
			},
			"required": []string{"message"},
		},
	}

	return properties, []string{"ok"}
}

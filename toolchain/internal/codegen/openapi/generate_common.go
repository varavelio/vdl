package openapi

import (
	"github.com/varavelio/vdl/toolchain/internal/core/ir"
)

// generateTypeRefSchema converts an IR TypeRef to a JSON Schema representation.
func generateTypeRefSchema(t ir.TypeRef) map[string]any {
	switch t.Kind {
	case ir.TypeKindPrimitive:
		return primitiveToJSONSchema(t.Primitive)

	case ir.TypeKindType:
		return map[string]any{
			"$ref": "#/components/schemas/" + t.Type,
		}

	case ir.TypeKindEnum:
		return map[string]any{
			"$ref": "#/components/schemas/" + t.Enum,
		}

	case ir.TypeKindArray:
		itemSchema := generateTypeRefSchema(*t.ArrayItem)
		// For multi-dimensional arrays, we need to nest the array schema
		for i := 1; i < t.ArrayDimensions; i++ {
			itemSchema = map[string]any{
				"type":  "array",
				"items": itemSchema,
			}
		}
		return map[string]any{
			"type":  "array",
			"items": itemSchema,
		}

	case ir.TypeKindMap:
		return map[string]any{
			"type":                 "object",
			"additionalProperties": generateTypeRefSchema(*t.MapValue),
		}

	case ir.TypeKindObject:
		props, required := generatePropertiesFromFields(t.Object.Fields)
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
func primitiveToJSONSchema(p ir.PrimitiveType) map[string]any {
	switch p {
	case ir.PrimitiveString:
		return map[string]any{"type": "string"}
	case ir.PrimitiveInt:
		return map[string]any{"type": "integer"}
	case ir.PrimitiveFloat:
		return map[string]any{"type": "number"}
	case ir.PrimitiveBool:
		return map[string]any{"type": "boolean"}
	case ir.PrimitiveDatetime:
		return map[string]any{"type": "string", "format": "date-time"}
	}
	return map[string]any{"type": "string"}
}

// generatePropertiesFromFields generates JSON schema properties from IR fields.
// Returns the properties map and a list of required field names.
func generatePropertiesFromFields(fields []ir.Field) (map[string]any, []string) {
	properties := map[string]any{}
	required := []string{}

	for _, field := range fields {
		prop := generateTypeRefSchema(field.Type)

		// Add description if present
		if field.Doc != "" {
			// If prop is a $ref, we need to wrap it in allOf to add description
			if _, hasRef := prop["$ref"]; hasRef {
				prop = map[string]any{
					"allOf": []map[string]any{
						prop,
						{"description": field.Doc},
					},
				}
			} else {
				prop["description"] = field.Doc
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
func generateOutputProperties(fields []ir.Field) (map[string]any, []string) {
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

package openapi

import (
	"fmt"
	"strings"

	"github.com/uforg/uforpc/urpc/internal/schema"
	"github.com/uforg/uforpc/urpc/internal/urpc/ast"
	"github.com/uforg/uforpc/urpc/internal/util/strutil"
)

// generateProperties generates the JSON schema properties for a given list of fields.
//
// It returns a map of the JSON schema properties and a list of required fields.
func generateProperties(fields []schema.FieldDefinition) (map[string]any, []string) {
	properties := map[string]any{}
	requiredFields := []string{}

	for _, field := range fields {
		properties[field.Name] = map[string]any{}

		isInline := field.TypeInline != nil
		isNamed := field.TypeName != nil
		isArray := field.IsArray
		hasDoc := field.Doc != nil

		doc := ""
		if hasDoc {
			doc = strings.TrimSpace(strutil.NormalizeIndent(*field.Doc))
		}

		if isNamed && ast.IsPrimitiveType(*field.TypeName) {
			fieldType := *field.TypeName
			switch *field.TypeName {
			case ast.PrimitiveTypeString:
				fieldType = "string"
			case ast.PrimitiveTypeInt:
				fieldType = "integer"
			case ast.PrimitiveTypeFloat:
				fieldType = "number"
			case ast.PrimitiveTypeBool:
				fieldType = "boolean"
			case ast.PrimitiveTypeDatetime:
				fieldType = "string"
			}

			prop := map[string]any{
				"type": fieldType,
			}

			if *field.TypeName == ast.PrimitiveTypeDatetime {
				prop["format"] = "date-time"
			}

			if hasDoc {
				prop["description"] = doc
			}

			properties[field.Name] = prop
		}

		if isNamed && !ast.IsPrimitiveType(*field.TypeName) {
			allOf := []map[string]any{
				{
					"$ref": fmt.Sprintf("#/components/schemas/%s", *field.TypeName),
				},
			}

			if hasDoc {
				allOf = append(allOf, map[string]any{
					"description": doc,
				})
			}

			properties[field.Name] = map[string]any{
				"allOf": allOf,
			}
		}

		if isInline {
			childProps, childRequired := generateProperties(field.TypeInline.Fields)

			prop := map[string]any{
				"type":       "object",
				"properties": childProps,
			}

			if len(childRequired) > 0 {
				prop["required"] = childRequired
			}

			if hasDoc {
				prop["description"] = doc
			}

			properties[field.Name] = prop
		}

		if isArray {
			arrayProp := map[string]any{
				"type":  "array",
				"items": properties[field.Name],
			}

			if hasDoc {
				arrayProp["description"] = doc
			}

			properties[field.Name] = arrayProp
		}

		if !field.Optional {
			requiredFields = append(requiredFields, field.Name)
		}
	}

	return properties, requiredFields
}

// generateOutputProperties generates the output properties for a given list of fields.
//
// It includes the `ok` field and the `error` field to handle both success and failure cases.
//
// It returns a map of the JSON schema properties and a list of required fields.
func generateOutputProperties(fields []schema.FieldDefinition) (map[string]any, []string) {
	outputProperties, outputRequiredFields := generateProperties(fields)
	output := componentRequestBodySchema{
		Type:       "object",
		Properties: outputProperties,
		Required:   outputRequiredFields,
	}

	properties := map[string]any{
		"ok": map[string]any{
			"type": "boolean",
		},
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

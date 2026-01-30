package jsonschema

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"

	"github.com/varavelio/vdl/toolchain/internal/codegen/config"
	"github.com/varavelio/vdl/toolchain/internal/core/ir"
)

// File represents a generated file.
type File struct {
	RelativePath string
	Content      []byte
}

// Generator implements the JSON Schema generator.
type Generator struct {
	config *config.JSONSchemaConfig
}

// New creates a new JSON Schema generator with the given config.
func New(config *config.JSONSchemaConfig) *Generator {
	return &Generator{config: config}
}

// Name returns the generator name.
func (g *Generator) Name() string {
	return "jsonschema"
}

// Generate produces JSON Schema files from the IR schema.
func (g *Generator) Generate(ctx context.Context, schema *ir.Schema) ([]File, error) {
	cfg := g.config

	// Root schema structure
	root := map[string]any{
		"$schema": "https://json-schema.org/draft/2020-12/schema",
		"$defs":   map[string]any{},
	}

	if cfg.ID != "" {
		root["$id"] = cfg.ID
	}

	definitions := root["$defs"].(map[string]any)

	// Generate schemas for custom types
	for _, t := range schema.Types {
		definitions[t.Name] = generateTypeSchema(t)
	}

	// Generate schemas for enums
	for _, e := range schema.Enums {
		definitions[e.Name] = generateEnumSchema(e)
	}

	// Generate input/output types for RPCs
	for _, rpc := range schema.RPCs {
		// Procedures
		for _, proc := range rpc.Procs {
			inputName := rpc.Name + proc.Name + "Input"
			outputName := rpc.Name + proc.Name + "Output"

			definitions[inputName] = generateObjectSchema(proc.Input, fmt.Sprintf("Input for %s/%s procedure", rpc.Name, proc.Name))
			definitions[outputName] = generateObjectSchema(proc.Output, fmt.Sprintf("Output for %s/%s procedure", rpc.Name, proc.Name))
		}

		// Streams
		for _, stream := range rpc.Streams {
			inputName := rpc.Name + stream.Name + "Input"
			outputName := rpc.Name + stream.Name + "Output"

			definitions[inputName] = generateObjectSchema(stream.Input, fmt.Sprintf("Input for %s/%s stream", rpc.Name, stream.Name))
			definitions[outputName] = generateObjectSchema(stream.Output, fmt.Sprintf("Output for %s/%s stream", rpc.Name, stream.Name))
		}
	}

	// Sort definitions to ensure deterministic output (map iteration is random)
	// Actually, encoding/json sorts map keys automatically, so we are good.

	// Encode schema
	content, err := json.MarshalIndent(root, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to encode json schema: %w", err)
	}

	filename := cfg.Filename
	if filename == "" {
		filename = "schema.json"
	}

	return []File{
		{
			RelativePath: filename,
			Content:      content,
		},
	}, nil
}

// generateTypeSchema generates a JSON Schema for an IR type.
func generateTypeSchema(t ir.Type) map[string]any {
	properties, required := generatePropertiesFromFields(t.Fields)

	schema := map[string]any{
		"type":       "object",
		"properties": properties,
	}

	if t.Doc != "" {
		schema["description"] = t.Doc
	}

	if len(required) > 0 {
		schema["required"] = required
	}

	return schema
}

// generateEnumSchema generates a JSON Schema for an IR enum.
func generateEnumSchema(e ir.Enum) map[string]any {
	schema := map[string]any{}

	if e.ValueType == ir.EnumValueTypeString {
		values := []string{}
		for _, m := range e.Members {
			values = append(values, m.Value)
		}
		schema["type"] = "string"
		schema["enum"] = values
	} else {
		values := []int{}
		for _, m := range e.Members {
			v, _ := strconv.Atoi(m.Value)
			values = append(values, v)
		}
		schema["type"] = "integer"
		schema["enum"] = values
	}

	if e.Doc != "" {
		schema["description"] = e.Doc
	}

	return schema
}

// generateObjectSchema generates a generic object schema from fields.
func generateObjectSchema(fields []ir.Field, description string) map[string]any {
	properties, required := generatePropertiesFromFields(fields)

	schema := map[string]any{
		"type":        "object",
		"properties":  properties,
		"description": description,
	}

	if len(required) > 0 {
		schema["required"] = required
	}

	return schema
}

// generatePropertiesFromFields generates JSON schema properties from IR fields.
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

	// Sort required fields for deterministic output
	sort.Strings(required)

	return properties, required
}

// generateTypeRefSchema converts an IR TypeRef to a JSON Schema representation.
func generateTypeRefSchema(t ir.TypeRef) map[string]any {
	switch t.Kind {
	case ir.TypeKindPrimitive:
		return primitiveToJSONSchema(t.Primitive)

	case ir.TypeKindType:
		return map[string]any{
			"$ref": "#/$defs/" + t.Type,
		}

	case ir.TypeKindEnum:
		return map[string]any{
			"$ref": "#/$defs/" + t.Enum,
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

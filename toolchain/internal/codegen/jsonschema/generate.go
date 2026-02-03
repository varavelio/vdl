package jsonschema

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/varavelio/vdl/toolchain/internal/codegen/config/configtypes"
	"github.com/varavelio/vdl/toolchain/internal/core/ir/irtypes"
	"github.com/varavelio/vdl/toolchain/internal/util/strutil"
)

// File represents a generated file.
type File struct {
	RelativePath string
	Content      []byte
}

// Generator implements the JSON Schema generator.
type Generator struct {
	config *configtypes.JsonSchemaConfig
}

// New creates a new JSON Schema generator with the given config.
func New(config *configtypes.JsonSchemaConfig) *Generator {
	return &Generator{config: config}
}

// Name returns the generator name.
func (g *Generator) Name() string {
	return "jsonschema"
}

// Generate produces JSON Schema files from the IR schema.
func (g *Generator) Generate(ctx context.Context, schema *irtypes.IrSchema) ([]File, error) {
	cfg := g.config

	// Root schema structure
	root := map[string]any{
		"$schema": "https://json-schema.org/draft/2020-12/schema",
		"$defs":   map[string]any{},
	}

	if cfg.Id != nil && *cfg.Id != "" {
		root["$id"] = *cfg.Id
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

	// Generate input/output types for procedures (flattened list with RpcName)
	for _, proc := range schema.Procedures {
		inputName := proc.RpcName + proc.Name + "Input"
		outputName := proc.RpcName + proc.Name + "Output"

		definitions[inputName] = generateObjectSchema(proc.Input, fmt.Sprintf("Input for %s/%s procedure", proc.RpcName, proc.Name))
		definitions[outputName] = generateObjectSchema(proc.Output, fmt.Sprintf("Output for %s/%s procedure", proc.RpcName, proc.Name))
	}

	// Generate input/output types for streams (flattened list with RpcName)
	for _, stream := range schema.Streams {
		inputName := stream.RpcName + stream.Name + "Input"
		outputName := stream.RpcName + stream.Name + "Output"

		definitions[inputName] = generateObjectSchema(stream.Input, fmt.Sprintf("Input for %s/%s stream", stream.RpcName, stream.Name))
		definitions[outputName] = generateObjectSchema(stream.Output, fmt.Sprintf("Output for %s/%s stream", stream.RpcName, stream.Name))
	}

	// Sort definitions to ensure deterministic output (map iteration is random)
	// Actually, encoding/json sorts map keys automatically, so we are good.

	// Validate and add root $ref if specified
	if cfg.Root != nil && *cfg.Root != "" {
		rootName := *cfg.Root

		// Check if the root type exists in definitions
		if _, exists := definitions[rootName]; !exists {
			// Build list of available definition names for fuzzy search
			availableNames := make([]string, 0, len(definitions))
			for name := range definitions {
				availableNames = append(availableNames, name)
			}

			// Use fuzzy search to find similar names
			suggestions, _ := strutil.FuzzySearch(availableNames, rootName)

			errMsg := fmt.Sprintf("root type '%s' not found in schema definitions", rootName)
			if len(suggestions) > 0 {
				suggestionStr := ""
				switch len(suggestions) {
				case 1:
					suggestionStr = fmt.Sprintf("'%s'", suggestions[0])
				case 2:
					suggestionStr = fmt.Sprintf("'%s' or '%s'", suggestions[0], suggestions[1])
				default:
					suggestionStr = fmt.Sprintf("'%s' or '%s'", strings.Join(suggestions[:len(suggestions)-1], "', '"), suggestions[len(suggestions)-1])
				}
				errMsg += fmt.Sprintf(". Did you mean %s?", suggestionStr)
			}
			return nil, fmt.Errorf("%s", errMsg)
		}

		// Add $ref at the root level pointing to the specified definition
		root["$ref"] = "#/$defs/" + rootName
	}

	// Encode schema
	content, err := json.MarshalIndent(root, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to encode json schema: %w", err)
	}

	filename := cfg.Filename
	if filename == nil || *filename == "" {
		defaultFilename := "schema.json"
		filename = &defaultFilename
	}

	return []File{
		{
			RelativePath: *filename,
			Content:      content,
		},
	}, nil
}

// generateTypeSchema generates a JSON Schema for an IR type.
func generateTypeSchema(t irtypes.TypeDef) map[string]any {
	properties, required := generatePropertiesFromFields(t.Fields)

	schema := map[string]any{
		"type":       "object",
		"properties": properties,
	}

	if t.Doc != nil && *t.Doc != "" {
		schema["description"] = *t.Doc
	}

	if len(required) > 0 {
		schema["required"] = required
	}

	return schema
}

// generateEnumSchema generates a JSON Schema for an IR enum.
func generateEnumSchema(e irtypes.EnumDef) map[string]any {
	schema := map[string]any{}

	if e.EnumType == irtypes.EnumTypeString {
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

	if e.Doc != nil && *e.Doc != "" {
		schema["description"] = *e.Doc
	}

	return schema
}

// generateObjectSchema generates a generic object schema from fields.
func generateObjectSchema(fields []irtypes.Field, description string) map[string]any {
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
func generatePropertiesFromFields(fields []irtypes.Field) (map[string]any, []string) {
	properties := map[string]any{}
	required := []string{}

	for _, field := range fields {
		prop := generateTypeRefSchema(field.TypeRef)

		// Add description if present
		if field.Doc != nil && *field.Doc != "" {
			// If prop is a $ref, we need to wrap it in allOf to add description
			if _, hasRef := prop["$ref"]; hasRef {
				prop = map[string]any{
					"allOf": []map[string]any{
						prop,
						{"description": *field.Doc},
					},
				}
			} else {
				prop["description"] = *field.Doc
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
func generateTypeRefSchema(t irtypes.TypeRef) map[string]any {
	switch t.Kind {
	case irtypes.TypeKindPrimitive:
		if t.PrimitiveName != nil {
			return primitiveToJSONSchema(*t.PrimitiveName)
		}
		return map[string]any{"type": "string"}

	case irtypes.TypeKindType:
		if t.TypeName != nil {
			return map[string]any{
				"$ref": "#/$defs/" + *t.TypeName,
			}
		}
		return map[string]any{}

	case irtypes.TypeKindEnum:
		if t.EnumName != nil {
			return map[string]any{
				"$ref": "#/$defs/" + *t.EnumName,
			}
		}
		return map[string]any{}

	case irtypes.TypeKindArray:
		if t.ArrayType != nil {
			itemSchema := generateTypeRefSchema(*t.ArrayType)
			// For multi-dimensional arrays, we need to nest the array schema
			dims := int64(1)
			if t.ArrayDims != nil {
				dims = *t.ArrayDims
			}
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
		}
		return map[string]any{}

	case irtypes.TypeKindMap:
		if t.MapType != nil {
			return map[string]any{
				"type":                 "object",
				"additionalProperties": generateTypeRefSchema(*t.MapType),
			}
		}
		return map[string]any{}

	case irtypes.TypeKindObject:
		if t.ObjectFields != nil {
			props, required := generatePropertiesFromFields(*t.ObjectFields)
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

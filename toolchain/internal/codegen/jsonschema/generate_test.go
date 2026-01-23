package jsonschema

import (
	"context"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/varavelio/vdl/toolchain/internal/codegen/config"
	"github.com/varavelio/vdl/toolchain/internal/core/analysis"
	"github.com/varavelio/vdl/toolchain/internal/core/ir"
	"github.com/varavelio/vdl/toolchain/internal/core/vfs"
)

var update = flag.Bool("update", false, "update golden files")

// TestGenerate_Golden runs golden file tests for the JSON Schema generator.
func TestGenerate_Golden(t *testing.T) {
	inputs, err := filepath.Glob("testdata/*.vdl")
	require.NoError(t, err)
	require.NotEmpty(t, inputs, "no .vdl files found in testdata/")

	for _, input := range inputs {
		name := strings.TrimSuffix(filepath.Base(input), ".vdl")
		t.Run(name, func(t *testing.T) {
			// Get absolute path for analysis
			absInput, err := filepath.Abs(input)
			require.NoError(t, err)

			// Parse + Analyze
			fs := vfs.New()
			program, diags := analysis.Analyze(fs, absInput)
			require.Empty(t, diags, "analysis errors: %v", diags)

			// Build IR
			schema := ir.FromProgram(program)

			// Generate JSON Schema
			gen := New(&config.JSONSchemaConfig{
				Filename: "schema.json",
			})

			files, err := gen.Generate(context.Background(), schema)
			require.NoError(t, err)
			require.Len(t, files, 1)

			gotBytes := files[0].Content

			// Golden file path (same name, .json extension)
			goldenPath := strings.TrimSuffix(input, ".vdl") + ".json"

			// Update or compare
			if *update {
				err := os.WriteFile(goldenPath, gotBytes, 0644)
				require.NoError(t, err)
				t.Logf("updated golden file: %s", goldenPath)
				return
			}

			// Read and compare with golden
			wantBytes, err := os.ReadFile(goldenPath)
			if os.IsNotExist(err) {
				t.Fatalf("golden file not found: %s (run with -update to create)", goldenPath)
			}
			require.NoError(t, err)

			assert.JSONEq(t, string(wantBytes), string(gotBytes))
		})
	}
}

func TestGenerator_Name(t *testing.T) {
	gen := New(&config.JSONSchemaConfig{})
	assert.Equal(t, "jsonschema", gen.Name())
}

func TestGenerator_Generate(t *testing.T) {
	gen := New(&config.JSONSchemaConfig{
		ID:       "https://example.com/schema.json",
		Filename: "test.json",
	})

	schema := &ir.Schema{
		Types: []ir.Type{
			{
				Name: "User",
				Fields: []ir.Field{
					{
						Name: "id",
						Type: ir.TypeRef{Kind: ir.TypeKindPrimitive, Primitive: ir.PrimitiveString},
					},
				},
			},
		},
		RPCs: []ir.RPC{
			{
				Name: "UserService",
				Procs: []ir.Procedure{
					{
						Name: "Create",
						Input: []ir.Field{
							{
								Name: "name",
								Type: ir.TypeRef{Kind: ir.TypeKindPrimitive, Primitive: ir.PrimitiveString},
							},
						},
						Output: []ir.Field{},
					},
				},
			},
		},
	}

	files, err := gen.Generate(context.Background(), schema)
	require.NoError(t, err)
	require.Len(t, files, 1)
	assert.Equal(t, "test.json", files[0].RelativePath)

	content := string(files[0].Content)
	assert.Contains(t, content, `"$id": "https://example.com/schema.json"`)
	assert.Contains(t, content, `"$schema": "https://json-schema.org/draft/2020-12/schema"`)
	assert.Contains(t, content, `"User": {`)
	assert.Contains(t, content, `"UserService_CreateInput": {`)
	assert.Contains(t, content, `"UserService_CreateOutput": {`)
}

func TestGenerateTypeRefSchema(t *testing.T) {
	tests := []struct {
		name     string
		typeRef  ir.TypeRef
		expected map[string]any
	}{
		{
			name: "primitive string",
			typeRef: ir.TypeRef{
				Kind:      ir.TypeKindPrimitive,
				Primitive: ir.PrimitiveString,
			},
			expected: map[string]any{"type": "string"},
		},
		{
			name: "primitive datetime",
			typeRef: ir.TypeRef{
				Kind:      ir.TypeKindPrimitive,
				Primitive: ir.PrimitiveDatetime,
			},
			expected: map[string]any{"type": "string", "format": "date-time"},
		},
		{
			name: "custom type reference",
			typeRef: ir.TypeRef{
				Kind: ir.TypeKindType,
				Type: "User",
			},
			expected: map[string]any{"$ref": "#/$defs/User"},
		},
		{
			name: "enum reference",
			typeRef: ir.TypeRef{
				Kind: ir.TypeKindEnum,
				Enum: "Status",
			},
			expected: map[string]any{"$ref": "#/$defs/Status"},
		},
		{
			name: "simple array",
			typeRef: ir.TypeRef{
				Kind:            ir.TypeKindArray,
				ArrayDimensions: 1,
				ArrayItem: &ir.TypeRef{
					Kind:      ir.TypeKindPrimitive,
					Primitive: ir.PrimitiveString,
				},
			},
			expected: map[string]any{
				"type":  "array",
				"items": map[string]any{"type": "string"},
			},
		},
		{
			name: "nested array (matrix)",
			typeRef: ir.TypeRef{
				Kind:            ir.TypeKindArray,
				ArrayDimensions: 2,
				ArrayItem: &ir.TypeRef{
					Kind:      ir.TypeKindPrimitive,
					Primitive: ir.PrimitiveFloat,
				},
			},
			expected: map[string]any{
				"type": "array",
				"items": map[string]any{
					"type":  "array",
					"items": map[string]any{"type": "number"},
				},
			},
		},
		{
			name: "map of strings",
			typeRef: ir.TypeRef{
				Kind: ir.TypeKindMap,
				MapValue: &ir.TypeRef{
					Kind:      ir.TypeKindPrimitive,
					Primitive: ir.PrimitiveString,
				},
			},
			expected: map[string]any{
				"type":                 "object",
				"additionalProperties": map[string]any{"type": "string"},
			},
		},
		{
			name: "map of objects",
			typeRef: ir.TypeRef{
				Kind: ir.TypeKindMap,
				MapValue: &ir.TypeRef{
					Kind: ir.TypeKindType,
					Type: "User",
				},
			},
			expected: map[string]any{
				"type":                 "object",
				"additionalProperties": map[string]any{"$ref": "#/$defs/User"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateTypeRefSchema(tt.typeRef)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGeneratePropertiesFromFields(t *testing.T) {
	fields := []ir.Field{
		{
			Name: "id",
			Type: ir.TypeRef{Kind: ir.TypeKindPrimitive, Primitive: ir.PrimitiveString},
		},
		{
			Name:     "user",
			Doc:      "The user object",
			Type:     ir.TypeRef{Kind: ir.TypeKindType, Type: "User"},
			Optional: false,
		},
	}

	props, required := generatePropertiesFromFields(fields)

	// Check required fields
	assert.Equal(t, []string{"id", "user"}, required)

	// Check id property
	idProp := props["id"].(map[string]any)
	assert.Equal(t, "string", idProp["type"])

	// Check user property uses allOf for doc with $ref
	userProp := props["user"].(map[string]any)
	allOf := userProp["allOf"].([]map[string]any)
	assert.Len(t, allOf, 2)
	assert.Equal(t, "#/$defs/User", allOf[0]["$ref"])
	assert.Equal(t, "The user object", allOf[1]["description"])
}

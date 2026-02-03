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
	"github.com/varavelio/vdl/toolchain/internal/core/ir/irtypes"
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

	primString := irtypes.PrimitiveTypeString
	schema := &irtypes.IrSchema{
		Types: []irtypes.TypeDef{
			{
				Name: "User",
				Fields: []irtypes.Field{
					{
						Name:    "id",
						TypeRef: irtypes.TypeRef{Kind: irtypes.TypeKindPrimitive, PrimitiveName: &primString},
					},
				},
			},
		},
		Rpcs: []irtypes.RpcDef{
			{
				Name: "UserService",
			},
		},
		Procedures: []irtypes.ProcedureDef{
			{
				RpcName: "UserService",
				Name:    "Create",
				InputFields: []irtypes.Field{
					{
						Name:    "name",
						TypeRef: irtypes.TypeRef{Kind: irtypes.TypeKindPrimitive, PrimitiveName: &primString},
					},
				},
				OutputFields: []irtypes.Field{},
			},
		},
		Streams:   []irtypes.StreamDef{},
		Enums:     []irtypes.EnumDef{},
		Constants: []irtypes.ConstantDef{},
		Patterns:  []irtypes.PatternDef{},
		Docs:      []irtypes.DocDef{},
	}

	files, err := gen.Generate(context.Background(), schema)
	require.NoError(t, err)
	require.Len(t, files, 1)
	assert.Equal(t, "test.json", files[0].RelativePath)

	content := string(files[0].Content)
	assert.Contains(t, content, `"$id": "https://example.com/schema.json"`)
	assert.Contains(t, content, `"$schema": "https://json-schema.org/draft/2020-12/schema"`)
	assert.Contains(t, content, `"User": {`)
	assert.Contains(t, content, `"UserServiceCreateInput": {`)
	assert.Contains(t, content, `"UserServiceCreateOutput": {`)
}

func TestGenerateTypeRefSchema(t *testing.T) {
	primString := irtypes.PrimitiveTypeString
	primDatetime := irtypes.PrimitiveTypeDatetime
	primFloat := irtypes.PrimitiveTypeFloat
	typeName := "User"
	enumName := "Status"
	arrayDims1 := int64(1)
	arrayDims2 := int64(2)

	tests := []struct {
		name     string
		typeRef  irtypes.TypeRef
		expected map[string]any
	}{
		{
			name: "primitive string",
			typeRef: irtypes.TypeRef{
				Kind:          irtypes.TypeKindPrimitive,
				PrimitiveName: &primString,
			},
			expected: map[string]any{"type": "string"},
		},
		{
			name: "primitive datetime",
			typeRef: irtypes.TypeRef{
				Kind:          irtypes.TypeKindPrimitive,
				PrimitiveName: &primDatetime,
			},
			expected: map[string]any{"type": "string", "format": "date-time"},
		},
		{
			name: "custom type reference",
			typeRef: irtypes.TypeRef{
				Kind:     irtypes.TypeKindType,
				TypeName: &typeName,
			},
			expected: map[string]any{"$ref": "#/$defs/User"},
		},
		{
			name: "enum reference",
			typeRef: irtypes.TypeRef{
				Kind:     irtypes.TypeKindEnum,
				EnumName: &enumName,
			},
			expected: map[string]any{"$ref": "#/$defs/Status"},
		},
		{
			name: "simple array",
			typeRef: irtypes.TypeRef{
				Kind:      irtypes.TypeKindArray,
				ArrayDims: &arrayDims1,
				ArrayType: &irtypes.TypeRef{
					Kind:          irtypes.TypeKindPrimitive,
					PrimitiveName: &primString,
				},
			},
			expected: map[string]any{
				"type":  "array",
				"items": map[string]any{"type": "string"},
			},
		},
		{
			name: "nested array (matrix)",
			typeRef: irtypes.TypeRef{
				Kind:      irtypes.TypeKindArray,
				ArrayDims: &arrayDims2,
				ArrayType: &irtypes.TypeRef{
					Kind:          irtypes.TypeKindPrimitive,
					PrimitiveName: &primFloat,
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
			typeRef: irtypes.TypeRef{
				Kind: irtypes.TypeKindMap,
				MapType: &irtypes.TypeRef{
					Kind:          irtypes.TypeKindPrimitive,
					PrimitiveName: &primString,
				},
			},
			expected: map[string]any{
				"type":                 "object",
				"additionalProperties": map[string]any{"type": "string"},
			},
		},
		{
			name: "map of objects",
			typeRef: irtypes.TypeRef{
				Kind: irtypes.TypeKindMap,
				MapType: &irtypes.TypeRef{
					Kind:     irtypes.TypeKindType,
					TypeName: &typeName,
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
	primString := irtypes.PrimitiveTypeString
	typeName := "User"
	userDoc := "The user object"

	fields := []irtypes.Field{
		{
			Name:    "id",
			TypeRef: irtypes.TypeRef{Kind: irtypes.TypeKindPrimitive, PrimitiveName: &primString},
		},
		{
			Name:     "user",
			Doc:      &userDoc,
			TypeRef:  irtypes.TypeRef{Kind: irtypes.TypeKindType, TypeName: &typeName},
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

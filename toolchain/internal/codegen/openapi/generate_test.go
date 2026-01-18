package openapi

import (
	"context"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/varavelio/vdl/toolchain/internal/core/analysis"
	"github.com/varavelio/vdl/toolchain/internal/core/ir"
	"github.com/varavelio/vdl/toolchain/internal/core/vfs"
)

var update = flag.Bool("update", false, "update golden files")

// TestGenerate_Golden runs golden file tests for the OpenAPI generator.
// Each .vdl file in testdata is parsed, analyzed, converted to IR,
// and the OpenAPI output is compared against its corresponding .yaml file.
//
// Run with -update flag to regenerate golden files:
//
//	go test -run TestGenerate_Golden -update
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

			// 1. Parse + Analyze
			fs := vfs.New()
			program, diags := analysis.Analyze(fs, absInput)
			require.Empty(t, diags, "analysis errors: %v", diags)

			// 2. Build IR
			schema := ir.FromProgram(program)

			// 3. Generate OpenAPI
			gen := New(Config{
				OutputFile: "openapi.yaml",
				Title:      "Test API",
				Version:    "1.0.0",
			})

			files, err := gen.Generate(context.Background(), schema)
			require.NoError(t, err)
			require.Len(t, files, 1)

			got := files[0].Content

			// Golden file path (same name, .yaml extension)
			goldenPath := strings.TrimSuffix(input, ".vdl") + ".yaml"

			// 4. Update or compare
			if *update {
				err := os.WriteFile(goldenPath, got, 0644)
				require.NoError(t, err)
				t.Logf("updated golden file: %s", goldenPath)
				return
			}

			// 5. Read and compare with golden
			want, err := os.ReadFile(goldenPath)
			if os.IsNotExist(err) {
				t.Fatalf("golden file not found: %s (run with -update to create)", goldenPath)
			}
			require.NoError(t, err)

			assert.Equal(t, string(want), string(got))
		})
	}
}

// TestGenerator_Name tests that the generator returns the correct name.
func TestGenerator_Name(t *testing.T) {
	gen := New(Config{})
	assert.Equal(t, "openapi", gen.Name())
}

// TestGenerator_DefaultConfig tests that defaults are applied.
func TestGenerator_DefaultConfig(t *testing.T) {
	gen := New(Config{})

	schema := &ir.Schema{
		Types: []ir.Type{},
		Enums: []ir.Enum{},
		RPCs:  []ir.RPC{},
	}

	files, err := gen.Generate(context.Background(), schema)
	require.NoError(t, err)
	require.Len(t, files, 1)

	// Check that default output file is used
	assert.Equal(t, "openapi.yaml", files[0].RelativePath)

	// Check that default title is in the output
	assert.Contains(t, string(files[0].Content), "VDL RPC API")
}

// TestGenerator_JSONOutput tests JSON output format.
func TestGenerator_JSONOutput(t *testing.T) {
	gen := New(Config{
		OutputFile: "api.json",
		Title:      "JSON Test API",
	})

	schema := &ir.Schema{
		Types: []ir.Type{},
		Enums: []ir.Enum{},
		RPCs:  []ir.RPC{},
	}

	files, err := gen.Generate(context.Background(), schema)
	require.NoError(t, err)
	require.Len(t, files, 1)

	assert.Equal(t, "api.json", files[0].RelativePath)
	// JSON should start with {
	assert.True(t, strings.HasPrefix(string(files[0].Content), "{"))
}

// TestGenerateTags tests tag generation from RPCs.
func TestGenerateTags(t *testing.T) {
	schema := &ir.Schema{
		RPCs: []ir.RPC{
			{
				Name: "Users",
				Doc:  "User management",
				Procs: []ir.Procedure{
					{Name: "GetUser"},
				},
				Streams: []ir.Stream{
					{Name: "UserUpdates"},
				},
			},
			{
				Name: "Chat",
				Procs: []ir.Procedure{
					{Name: "Send"},
				},
			},
		},
	}

	tags := generateTags(schema)

	// Should have 3 tags: ChatProcedures, UsersProcedures, UsersStreams (sorted)
	require.Len(t, tags, 3)
	assert.Equal(t, "ChatProcedures", tags[0].Name)
	assert.Equal(t, "UsersProcedures", tags[1].Name)
	assert.Equal(t, "UsersStreams", tags[2].Name)

	// Check that RPC doc is used for description
	assert.Equal(t, "User management", tags[1].Description)
}

// TestGeneratePaths tests path generation with correct RPC structure.
func TestGeneratePaths(t *testing.T) {
	schema := &ir.Schema{
		RPCs: []ir.RPC{
			{
				Name: "Users",
				Procs: []ir.Procedure{
					{
						Name: "CreateUser",
						Doc:  "Creates a user",
					},
					{
						Name:       "DeleteUser",
						Deprecated: &ir.Deprecation{Message: "Use RemoveUser"},
					},
				},
				Streams: []ir.Stream{
					{
						Name: "UserEvents",
					},
				},
			},
		},
	}

	paths := generatePaths(schema)

	// Check procedure paths
	require.Contains(t, paths, "/Users/CreateUser")
	require.Contains(t, paths, "/Users/DeleteUser")
	require.Contains(t, paths, "/Users/UserEvents")

	// Check CreateUser operation
	createPath := paths["/Users/CreateUser"].(map[string]any)
	createOp := createPath["post"].(map[string]any)
	assert.Equal(t, []string{"UsersProcedures"}, createOp["tags"])
	assert.Equal(t, "Creates a user", createOp["description"])

	// Check deprecated operation
	deletePath := paths["/Users/DeleteUser"].(map[string]any)
	deleteOp := deletePath["post"].(map[string]any)
	assert.Equal(t, true, deleteOp["deprecated"])

	// Check stream uses Streams tag
	streamPath := paths["/Users/UserEvents"].(map[string]any)
	streamOp := streamPath["post"].(map[string]any)
	assert.Equal(t, []string{"UsersStreams"}, streamOp["tags"])
}

// TestGenerateEnumSchema tests enum schema generation.
func TestGenerateEnumSchema(t *testing.T) {
	t.Run("string enum", func(t *testing.T) {
		e := ir.Enum{
			Name:      "Status",
			Doc:       "Order status",
			ValueType: ir.EnumValueTypeString,
			Members: []ir.EnumMember{
				{Name: "Pending", Value: "Pending"},
				{Name: "Active", Value: "Active"},
			},
		}

		schema := generateEnumSchema(e)

		assert.Equal(t, "string", schema["type"])
		assert.Equal(t, []string{"Pending", "Active"}, schema["enum"])
		assert.Equal(t, "Order status", schema["description"])
	})

	t.Run("int enum", func(t *testing.T) {
		e := ir.Enum{
			Name:      "Priority",
			ValueType: ir.EnumValueTypeInt,
			Members: []ir.EnumMember{
				{Name: "Low", Value: "1"},
				{Name: "High", Value: "10"},
			},
		}

		schema := generateEnumSchema(e)

		assert.Equal(t, "integer", schema["type"])
		assert.Equal(t, []int{1, 10}, schema["enum"])
	})
}

// TestGenerateTypeRefSchema tests type reference conversion.
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
			name: "primitive int",
			typeRef: ir.TypeRef{
				Kind:      ir.TypeKindPrimitive,
				Primitive: ir.PrimitiveInt,
			},
			expected: map[string]any{"type": "integer"},
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
			expected: map[string]any{"$ref": "#/components/schemas/User"},
		},
		{
			name: "enum reference",
			typeRef: ir.TypeRef{
				Kind: ir.TypeKindEnum,
				Enum: "Status",
			},
			expected: map[string]any{"$ref": "#/components/schemas/Status"},
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
			name: "map type",
			typeRef: ir.TypeRef{
				Kind: ir.TypeKindMap,
				MapValue: &ir.TypeRef{
					Kind:      ir.TypeKindPrimitive,
					Primitive: ir.PrimitiveInt,
				},
			},
			expected: map[string]any{
				"type":                 "object",
				"additionalProperties": map[string]any{"type": "integer"},
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

// TestGeneratePropertiesFromFields tests field to property conversion.
func TestGeneratePropertiesFromFields(t *testing.T) {
	fields := []ir.Field{
		{
			Name: "id",
			Type: ir.TypeRef{Kind: ir.TypeKindPrimitive, Primitive: ir.PrimitiveString},
		},
		{
			Name:     "email",
			Optional: true,
			Doc:      "User email",
			Type:     ir.TypeRef{Kind: ir.TypeKindPrimitive, Primitive: ir.PrimitiveString},
		},
		{
			Name: "user",
			Doc:  "The user object",
			Type: ir.TypeRef{Kind: ir.TypeKindType, Type: "User"},
		},
	}

	props, required := generatePropertiesFromFields(fields)

	// Check required fields
	assert.Equal(t, []string{"id", "user"}, required)

	// Check id property
	idProp := props["id"].(map[string]any)
	assert.Equal(t, "string", idProp["type"])

	// Check email property has description
	emailProp := props["email"].(map[string]any)
	assert.Equal(t, "User email", emailProp["description"])

	// Check user property uses allOf for doc with $ref
	userProp := props["user"].(map[string]any)
	allOf := userProp["allOf"].([]map[string]any)
	assert.Len(t, allOf, 2)
	assert.Equal(t, "#/components/schemas/User", allOf[0]["$ref"])
	assert.Equal(t, "The user object", allOf[1]["description"])
}

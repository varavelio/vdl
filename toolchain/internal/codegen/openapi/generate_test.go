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
	"github.com/varavelio/vdl/toolchain/internal/codegen/config"
	"github.com/varavelio/vdl/toolchain/internal/core/analysis"
	"github.com/varavelio/vdl/toolchain/internal/core/ir"
	"github.com/varavelio/vdl/toolchain/internal/core/ir/irtypes"
	"github.com/varavelio/vdl/toolchain/internal/core/vfs"
	"gopkg.in/yaml.v3"
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
			gen := New(&config.OpenAPIConfig{
				Filename: "openapi.yaml",
				Title:    "Test API",
				Version:  "1.0.0",
			})

			files, err := gen.Generate(context.Background(), schema)
			require.NoError(t, err)
			require.Len(t, files, 1)

			gotBytes := files[0].Content

			// Golden file path (same name, .yaml extension)
			goldenPath := strings.TrimSuffix(input, ".vdl") + ".yaml"

			// 4. Update or compare
			if *update {
				err := os.WriteFile(goldenPath, gotBytes, 0644)
				require.NoError(t, err)
				t.Logf("updated golden file: %s", goldenPath)
				return
			}

			// 5. Read and compare with golden
			wantBytes, err := os.ReadFile(goldenPath)
			if os.IsNotExist(err) {
				t.Fatalf("golden file not found: %s (run with -update to create)", goldenPath)
			}
			require.NoError(t, err)

			// Perform structural comparison instead of string comparison
			// to be resilient to formatting changes (e.g. Prettier)
			var wantObj, gotObj any

			err = yaml.Unmarshal(wantBytes, &wantObj)
			require.NoError(t, err, "failed to unmarshal golden file")

			err = yaml.Unmarshal(gotBytes, &gotObj)
			require.NoError(t, err, "failed to unmarshal generated output")

			assert.Equal(t, wantObj, gotObj)
		})
	}
}

// TestGenerator_Name tests that the generator returns the correct name.
func TestGenerator_Name(t *testing.T) {
	gen := New(&config.OpenAPIConfig{})
	assert.Equal(t, "openapi", gen.Name())
}

// TestGenerator_DefaultConfig tests that defaults are applied.
func TestGenerator_DefaultConfig(t *testing.T) {
	gen := New(&config.OpenAPIConfig{})

	schema := &irtypes.IrSchema{
		Types:      []irtypes.TypeDef{},
		Enums:      []irtypes.EnumDef{},
		Rpcs:       []irtypes.RpcDef{},
		Procedures: []irtypes.ProcedureDef{},
		Streams:    []irtypes.StreamDef{},
		Constants:  []irtypes.ConstantDef{},
		Patterns:   []irtypes.PatternDef{},
		Docs:       []irtypes.DocDef{},
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
	gen := New(&config.OpenAPIConfig{
		Filename: "api.json",
		Title:    "JSON Test API",
	})

	schema := &irtypes.IrSchema{
		Types:      []irtypes.TypeDef{},
		Enums:      []irtypes.EnumDef{},
		Rpcs:       []irtypes.RpcDef{},
		Procedures: []irtypes.ProcedureDef{},
		Streams:    []irtypes.StreamDef{},
		Constants:  []irtypes.ConstantDef{},
		Patterns:   []irtypes.PatternDef{},
		Docs:       []irtypes.DocDef{},
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
	usersDoc := "User management"
	schema := &irtypes.IrSchema{
		Rpcs: []irtypes.RpcDef{
			{
				Name: "Users",
				Doc:  &usersDoc,
			},
			{
				Name: "Chat",
			},
		},
		Procedures: []irtypes.ProcedureDef{
			{RpcName: "Users", Name: "GetUser"},
			{RpcName: "Chat", Name: "Send"},
		},
		Streams: []irtypes.StreamDef{
			{RpcName: "Users", Name: "UserUpdates"},
		},
		Types:     []irtypes.TypeDef{},
		Enums:     []irtypes.EnumDef{},
		Constants: []irtypes.ConstantDef{},
		Patterns:  []irtypes.PatternDef{},
		Docs:      []irtypes.DocDef{},
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
	createUserDoc := "Creates a user"
	deleteUserDeprecation := "Use RemoveUser"
	schema := &irtypes.IrSchema{
		Rpcs: []irtypes.RpcDef{
			{Name: "Users"},
		},
		Procedures: []irtypes.ProcedureDef{
			{
				RpcName: "Users",
				Name:    "CreateUser",
				Doc:     &createUserDoc,
			},
			{
				RpcName:     "Users",
				Name:        "DeleteUser",
				Deprecation: &deleteUserDeprecation,
			},
		},
		Streams: []irtypes.StreamDef{
			{
				RpcName: "Users",
				Name:    "UserEvents",
			},
		},
		Types:     []irtypes.TypeDef{},
		Enums:     []irtypes.EnumDef{},
		Constants: []irtypes.ConstantDef{},
		Patterns:  []irtypes.PatternDef{},
		Docs:      []irtypes.DocDef{},
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
		doc := "Order status"
		e := irtypes.EnumDef{
			Name:     "Status",
			Doc:      &doc,
			EnumType: irtypes.EnumTypeString,
			Members: []irtypes.EnumDefMember{
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
		e := irtypes.EnumDef{
			Name:     "Priority",
			EnumType: irtypes.EnumTypeInt,
			Members: []irtypes.EnumDefMember{
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
		typeRef  irtypes.TypeRef
		expected map[string]any
	}{
		{
			name: "primitive string",
			typeRef: irtypes.TypeRef{
				Kind:          irtypes.TypeKindPrimitive,
				PrimitiveName: irtypes.Ptr(irtypes.PrimitiveTypeString),
			},
			expected: map[string]any{"type": "string"},
		},
		{
			name: "primitive int",
			typeRef: irtypes.TypeRef{
				Kind:          irtypes.TypeKindPrimitive,
				PrimitiveName: irtypes.Ptr(irtypes.PrimitiveTypeInt),
			},
			expected: map[string]any{"type": "integer"},
		},
		{
			name: "primitive datetime",
			typeRef: irtypes.TypeRef{
				Kind:          irtypes.TypeKindPrimitive,
				PrimitiveName: irtypes.Ptr(irtypes.PrimitiveTypeDatetime),
			},
			expected: map[string]any{"type": "string", "format": "date-time"},
		},
		{
			name: "custom type reference",
			typeRef: irtypes.TypeRef{
				Kind:     irtypes.TypeKindType,
				TypeName: ptrString("User"),
			},
			expected: map[string]any{"$ref": "#/components/schemas/User"},
		},
		{
			name: "enum reference",
			typeRef: irtypes.TypeRef{
				Kind:     irtypes.TypeKindEnum,
				EnumName: ptrString("Status"),
			},
			expected: map[string]any{"$ref": "#/components/schemas/Status"},
		},
		{
			name: "simple array",
			typeRef: irtypes.TypeRef{
				Kind:      irtypes.TypeKindArray,
				ArrayDims: ptrInt64(1),
				ArrayType: &irtypes.TypeRef{
					Kind:          irtypes.TypeKindPrimitive,
					PrimitiveName: irtypes.Ptr(irtypes.PrimitiveTypeString),
				},
			},
			expected: map[string]any{
				"type":  "array",
				"items": map[string]any{"type": "string"},
			},
		},
		{
			name: "map type",
			typeRef: irtypes.TypeRef{
				Kind: irtypes.TypeKindMap,
				MapType: &irtypes.TypeRef{
					Kind:          irtypes.TypeKindPrimitive,
					PrimitiveName: irtypes.Ptr(irtypes.PrimitiveTypeInt),
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

func ptrString(s string) *string {
	return &s
}

func ptrInt64(i int64) *int64 {
	return &i
}

// TestGeneratePropertiesFromFields tests field to property conversion.
func TestGeneratePropertiesFromFields(t *testing.T) {
	emailDoc := "User email"
	userDoc := "The user object"
	fields := []irtypes.Field{
		{
			Name:    "id",
			TypeRef: irtypes.TypeRef{Kind: irtypes.TypeKindPrimitive, PrimitiveName: irtypes.Ptr(irtypes.PrimitiveTypeString)},
		},
		{
			Name:     "email",
			Optional: true,
			Doc:      &emailDoc,
			TypeRef:  irtypes.TypeRef{Kind: irtypes.TypeKindPrimitive, PrimitiveName: irtypes.Ptr(irtypes.PrimitiveTypeString)},
		},
		{
			Name:    "user",
			Doc:     &userDoc,
			TypeRef: irtypes.TypeRef{Kind: irtypes.TypeKindType, TypeName: ptrString("User")},
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

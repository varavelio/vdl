package ir

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/varavelio/vdl/toolchain/internal/codegen/config/configtypes"
	"github.com/varavelio/vdl/toolchain/internal/core/ir/irtypes"
)

func TestGenerator_Name(t *testing.T) {
	gen := New(&configtypes.IrTargetConfig{})
	assert.Equal(t, "ir", gen.Name())
}

func TestGenerator_Generate_DefaultFilename(t *testing.T) {
	gen := New(&configtypes.IrTargetConfig{
		Output: "dist",
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

	// Default filename should be ir.json
	assert.Equal(t, "ir.json", files[0].RelativePath)
}

func TestGenerator_Generate_CustomFilename(t *testing.T) {
	filename := "custom-schema.json"
	gen := New(&configtypes.IrTargetConfig{
		Output:   "dist",
		Filename: &filename,
	})

	schema := &irtypes.IrSchema{}

	files, err := gen.Generate(context.Background(), schema)
	require.NoError(t, err)
	require.Len(t, files, 1)

	assert.Equal(t, "custom-schema.json", files[0].RelativePath)
}

func TestGenerator_Generate_EmptyFilenameUsesDefault(t *testing.T) {
	empty := ""
	gen := New(&configtypes.IrTargetConfig{
		Output:   "dist",
		Filename: &empty,
	})

	schema := &irtypes.IrSchema{}

	files, err := gen.Generate(context.Background(), schema)
	require.NoError(t, err)
	require.Len(t, files, 1)

	// Empty filename should fallback to default
	assert.Equal(t, "ir.json", files[0].RelativePath)
}

func TestGenerator_Generate_ValidJSON(t *testing.T) {
	gen := New(&configtypes.IrTargetConfig{
		Output: "dist",
	})

	primString := irtypes.PrimitiveTypeString
	primInt := irtypes.PrimitiveTypeInt

	schema := &irtypes.IrSchema{
		Types: []irtypes.TypeDef{
			{
				Name: "User",
				Fields: []irtypes.Field{
					{
						Name:    "id",
						TypeRef: irtypes.TypeRef{Kind: irtypes.TypeKindPrimitive, PrimitiveName: &primString},
					},
					{
						Name:    "age",
						TypeRef: irtypes.TypeRef{Kind: irtypes.TypeKindPrimitive, PrimitiveName: &primInt},
					},
				},
			},
		},
		Enums: []irtypes.EnumDef{
			{
				Name:     "Status",
				EnumType: irtypes.EnumTypeString,
				Members:  []irtypes.EnumDefMember{{Name: "Active"}, {Name: "Inactive"}},
			},
		},
		Rpcs:       []irtypes.RpcDef{{Name: "UserService"}},
		Procedures: []irtypes.ProcedureDef{},
		Streams:    []irtypes.StreamDef{},
		Constants:  []irtypes.ConstantDef{},
		Patterns:   []irtypes.PatternDef{},
		Docs:       []irtypes.DocDef{},
	}

	files, err := gen.Generate(context.Background(), schema)
	require.NoError(t, err)
	require.Len(t, files, 1)

	// Verify output is valid JSON
	var result map[string]any
	err = json.Unmarshal(files[0].Content, &result)
	require.NoError(t, err)

	// Check key structure
	assert.Contains(t, result, "types")
	assert.Contains(t, result, "enums")
	assert.Contains(t, result, "rpcs")

	types := result["types"].([]any)
	require.Len(t, types, 1)

	userType := types[0].(map[string]any)
	assert.Equal(t, "User", userType["name"])
}

func TestGenerator_Generate_JSONContainsAllFields(t *testing.T) {
	gen := New(&configtypes.IrTargetConfig{
		Output: "dist",
	})

	primString := irtypes.PrimitiveTypeString
	docText := "User documentation"
	constValue := "1.0.0"
	template := "/users/{id}"

	schema := &irtypes.IrSchema{
		Types: []irtypes.TypeDef{
			{
				Name: "User",
				Doc:  &docText,
				Fields: []irtypes.Field{
					{
						Name:    "id",
						TypeRef: irtypes.TypeRef{Kind: irtypes.TypeKindPrimitive, PrimitiveName: &primString},
					},
				},
			},
		},
		Enums: []irtypes.EnumDef{
			{
				Name:     "Status",
				EnumType: irtypes.EnumTypeString,
				Members:  []irtypes.EnumDefMember{{Name: "Active"}},
			},
		},
		Rpcs: []irtypes.RpcDef{
			{Name: "UserService"},
		},
		Procedures: []irtypes.ProcedureDef{
			{
				RpcName: "UserService",
				Name:    "GetUser",
				Input: []irtypes.Field{
					{Name: "id", TypeRef: irtypes.TypeRef{Kind: irtypes.TypeKindPrimitive, PrimitiveName: &primString}},
				},
				Output: []irtypes.Field{},
			},
		},
		Streams: []irtypes.StreamDef{
			{
				RpcName: "UserService",
				Name:    "WatchUsers",
				Input:   []irtypes.Field{},
				Output:  []irtypes.Field{},
			},
		},
		Constants: []irtypes.ConstantDef{
			{Name: "VERSION", Value: constValue, ConstType: irtypes.ConstTypeString},
		},
		Patterns: []irtypes.PatternDef{
			{Name: "UserPath", Template: template, Placeholders: []string{"id"}},
		},
		Docs: []irtypes.DocDef{
			{Content: "# Hello"},
		},
	}

	files, err := gen.Generate(context.Background(), schema)
	require.NoError(t, err)

	content := string(files[0].Content)

	// Check all sections are present
	assert.Contains(t, content, `"types"`)
	assert.Contains(t, content, `"enums"`)
	assert.Contains(t, content, `"rpcs"`)
	assert.Contains(t, content, `"procedures"`)
	assert.Contains(t, content, `"streams"`)
	assert.Contains(t, content, `"constants"`)
	assert.Contains(t, content, `"patterns"`)
	assert.Contains(t, content, `"docs"`)

	// Check specific values
	assert.Contains(t, content, `"User"`)
	assert.Contains(t, content, `"User documentation"`)
	assert.Contains(t, content, `"Status"`)
	assert.Contains(t, content, `"UserService"`)
	assert.Contains(t, content, `"GetUser"`)
	assert.Contains(t, content, `"WatchUsers"`)
	assert.Contains(t, content, `"VERSION"`)
	assert.Contains(t, content, `"1.0.0"`)
	assert.Contains(t, content, `"UserPath"`)
	assert.Contains(t, content, `"# Hello"`)
}

func TestGenerator_Generate_Indented(t *testing.T) {
	gen := New(&configtypes.IrTargetConfig{
		Output: "dist",
	})

	schema := &irtypes.IrSchema{
		Types: []irtypes.TypeDef{{Name: "User"}},
	}

	files, err := gen.Generate(context.Background(), schema)
	require.NoError(t, err)

	content := string(files[0].Content)

	// Verify output is indented (contains newlines and spaces)
	assert.Contains(t, content, "\n")
	assert.Contains(t, content, "  ") // 2-space indentation
}

func TestGenerator_Generate_EmptySchema(t *testing.T) {
	gen := New(&configtypes.IrTargetConfig{
		Output: "dist",
	})

	schema := &irtypes.IrSchema{}

	files, err := gen.Generate(context.Background(), schema)
	require.NoError(t, err)
	require.Len(t, files, 1)

	// Should still produce valid JSON with empty/null arrays
	var result map[string]any
	err = json.Unmarshal(files[0].Content, &result)
	require.NoError(t, err)
}

func TestGenerator_Generate_Minified(t *testing.T) {
	minify := true
	gen := New(&configtypes.IrTargetConfig{
		Output: "dist",
		Minify: &minify,
	})

	schema := &irtypes.IrSchema{
		Types: []irtypes.TypeDef{{Name: "User"}},
	}

	files, err := gen.Generate(context.Background(), schema)
	require.NoError(t, err)

	content := string(files[0].Content)

	// Minified output should NOT contain newlines or indentation
	assert.NotContains(t, content, "\n")
	assert.NotContains(t, content, "  ")

	// But should still be valid JSON
	var result map[string]any
	err = json.Unmarshal(files[0].Content, &result)
	require.NoError(t, err)
}

func TestGenerator_Generate_MinifyFalse(t *testing.T) {
	minify := false
	gen := New(&configtypes.IrTargetConfig{
		Output: "dist",
		Minify: &minify,
	})

	schema := &irtypes.IrSchema{
		Types: []irtypes.TypeDef{{Name: "User"}},
	}

	files, err := gen.Generate(context.Background(), schema)
	require.NoError(t, err)

	content := string(files[0].Content)

	// Non-minified output should contain newlines and indentation
	assert.Contains(t, content, "\n")
	assert.Contains(t, content, "  ")
}

func TestGenerator_Generate_DefaultPretty(t *testing.T) {
	// When minify is nil (not set), default should be pretty-printed
	gen := New(&configtypes.IrTargetConfig{
		Output: "dist",
		// Minify not set
	})

	schema := &irtypes.IrSchema{
		Types: []irtypes.TypeDef{{Name: "User"}},
	}

	files, err := gen.Generate(context.Background(), schema)
	require.NoError(t, err)

	content := string(files[0].Content)

	// Default should be pretty-printed
	assert.Contains(t, content, "\n")
	assert.Contains(t, content, "  ")
}

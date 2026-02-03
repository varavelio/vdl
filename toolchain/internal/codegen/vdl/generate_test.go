package vdl

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/varavelio/vdl/toolchain/internal/codegen/config/configtypes"
	"github.com/varavelio/vdl/toolchain/internal/core/ir/irtypes"
)

func TestGenerator_Name(t *testing.T) {
	gen := New(&configtypes.VdlTargetConfig{}, "")
	assert.Equal(t, "vdl", gen.Name())
}

func TestGenerator_Generate_DefaultFilename(t *testing.T) {
	gen := New(&configtypes.VdlTargetConfig{
		Output: "dist",
	}, "type User { id: string }")

	schema := &irtypes.IrSchema{}

	files, err := gen.Generate(context.Background(), schema)
	require.NoError(t, err)
	require.Len(t, files, 1)

	// Default filename should be schema.vdl
	assert.Equal(t, "schema.vdl", files[0].RelativePath)
}

func TestGenerator_Generate_CustomFilename(t *testing.T) {
	filename := "unified.vdl"
	gen := New(&configtypes.VdlTargetConfig{
		Output:   "dist",
		Filename: &filename,
	}, "type User { id: string }")

	schema := &irtypes.IrSchema{}

	files, err := gen.Generate(context.Background(), schema)
	require.NoError(t, err)
	require.Len(t, files, 1)

	assert.Equal(t, "unified.vdl", files[0].RelativePath)
}

func TestGenerator_Generate_EmptyFilenameUsesDefault(t *testing.T) {
	empty := ""
	gen := New(&configtypes.VdlTargetConfig{
		Output:   "dist",
		Filename: &empty,
	}, "type User { id: string }")

	schema := &irtypes.IrSchema{}

	files, err := gen.Generate(context.Background(), schema)
	require.NoError(t, err)
	require.Len(t, files, 1)

	// Empty filename should fallback to default
	assert.Equal(t, "schema.vdl", files[0].RelativePath)
}

func TestGenerator_Generate_PreservesFormattedSchema(t *testing.T) {
	formattedSchema := `// User type represents a system user
type User {
    id: string
    name: string
    email: string?
}

// UserService provides user operations
rpc UserService {
    proc GetUser {
        input { id: string }
        output { user: User }
    }
}
`

	gen := New(&configtypes.VdlTargetConfig{
		Output: "dist",
	}, formattedSchema)

	schema := &irtypes.IrSchema{}

	files, err := gen.Generate(context.Background(), schema)
	require.NoError(t, err)
	require.Len(t, files, 1)

	// Content should be exactly the formatted schema
	assert.Equal(t, formattedSchema, string(files[0].Content))
}

func TestGenerator_Generate_EmptySchema(t *testing.T) {
	gen := New(&configtypes.VdlTargetConfig{
		Output: "dist",
	}, "")

	schema := &irtypes.IrSchema{}

	files, err := gen.Generate(context.Background(), schema)
	require.NoError(t, err)
	require.Len(t, files, 1)

	// Content should be empty string
	assert.Equal(t, "", string(files[0].Content))
}

func TestGenerator_Generate_WithIncludes(t *testing.T) {
	// This test simulates what would be passed after merging includes
	mergedSchema := `// From common.vdl
type Email = string

// From user.vdl
type User {
    id: string
    email: Email
}

// From api.vdl
rpc UserAPI {
    proc GetUser {
        input { id: string }
        output { user: User }
    }
}
`

	gen := New(&configtypes.VdlTargetConfig{
		Output: "dist",
	}, mergedSchema)

	schema := &irtypes.IrSchema{}

	files, err := gen.Generate(context.Background(), schema)
	require.NoError(t, err)
	require.Len(t, files, 1)

	content := string(files[0].Content)

	// Should contain all merged content
	assert.Contains(t, content, "type Email = string")
	assert.Contains(t, content, "type User")
	assert.Contains(t, content, "email: Email")
	assert.Contains(t, content, "rpc UserAPI")
}

func TestGenerator_Generate_WithExternalDocs(t *testing.T) {
	// This test simulates schema with resolved external docs
	schemaWithDocs := `/// User represents a system user
/// 
/// Users can have different roles:
/// - admin: Full system access
/// - user: Standard access
/// - guest: Read-only access
type User {
    id: string
    role: string
}
`

	gen := New(&configtypes.VdlTargetConfig{
		Output: "dist",
	}, schemaWithDocs)

	schema := &irtypes.IrSchema{}

	files, err := gen.Generate(context.Background(), schema)
	require.NoError(t, err)
	require.Len(t, files, 1)

	content := string(files[0].Content)

	// Should preserve the inline documentation
	assert.Contains(t, content, "/// User represents a system user")
	assert.Contains(t, content, "/// - admin: Full system access")
}

func TestGenerator_Generate_PreservesWhitespace(t *testing.T) {
	schemaWithFormatting := `type User {
    id: string

    // Contact info
    email: string
    phone: string?
}


rpc API {
    proc Create {
        input { name: string }
        output { id: string }
    }
}
`

	gen := New(&configtypes.VdlTargetConfig{
		Output: "dist",
	}, schemaWithFormatting)

	schema := &irtypes.IrSchema{}

	files, err := gen.Generate(context.Background(), schema)
	require.NoError(t, err)

	// Should preserve exact whitespace/formatting
	assert.Equal(t, schemaWithFormatting, string(files[0].Content))
}

func TestGenerator_Generate_DoesNotModifyIR(t *testing.T) {
	// The VDL Schema generator should not use the IR parameter,
	// it should only use the pre-formatted schema string
	gen := New(&configtypes.VdlTargetConfig{
		Output: "dist",
	}, "type Simple { id: string }")

	// Pass a schema with different content - should be ignored
	primInt := irtypes.PrimitiveTypeInt
	schema := &irtypes.IrSchema{
		Types: []irtypes.TypeDef{
			{
				Name: "CompletelyDifferent",
				Fields: []irtypes.Field{
					{Name: "count", TypeRef: irtypes.TypeRef{Kind: irtypes.TypeKindPrimitive, PrimitiveName: &primInt}},
				},
			},
		},
	}

	files, err := gen.Generate(context.Background(), schema)
	require.NoError(t, err)

	content := string(files[0].Content)

	// Should use the formatted schema, not the IR
	assert.Contains(t, content, "type Simple")
	assert.NotContains(t, content, "CompletelyDifferent")
}

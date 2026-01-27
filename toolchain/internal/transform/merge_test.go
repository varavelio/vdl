package transform

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/varavelio/vdl/toolchain/internal/core/analysis"
	"github.com/varavelio/vdl/toolchain/internal/core/ast"
)

func TestMergeSchemas_NilProgram(t *testing.T) {
	result := MergeSchemas(nil)

	require.NotNil(t, result)
	require.Empty(t, result.Children)
}

func TestMergeSchemas_EmptyProgram(t *testing.T) {
	program := &analysis.Program{
		Files: map[string]*analysis.File{},
	}

	result := MergeSchemas(program)

	require.NotNil(t, result)
	require.Empty(t, result.Children)
}

func TestMergeSchemas_SingleFileNoIncludes(t *testing.T) {
	typeName := "User"
	typeDecl := &ast.TypeDecl{Name: typeName}

	program := &analysis.Program{
		EntryPoint: "/main.vdl",
		Files: map[string]*analysis.File{
			"/main.vdl": {
				Path:     "/main.vdl",
				Includes: []string{},
				AST: &ast.Schema{
					Children: []*ast.SchemaChild{
						{Type: typeDecl},
					},
				},
			},
		},
	}

	result := MergeSchemas(program)

	require.Len(t, result.Children, 1)
	require.NotNil(t, result.Children[0].Type)
	require.Equal(t, typeName, result.Children[0].Type.Name)
}

func TestMergeSchemas_RemovesIncludeStatements(t *testing.T) {
	typeDecl := &ast.TypeDecl{Name: "User"}
	includePath := ast.QuotedString("common.vdl")

	program := &analysis.Program{
		EntryPoint: "/main.vdl",
		Files: map[string]*analysis.File{
			"/main.vdl": {
				Path:     "/main.vdl",
				Includes: []string{},
				AST: &ast.Schema{
					Children: []*ast.SchemaChild{
						{Include: &ast.Include{Path: includePath}},
						{Type: typeDecl},
					},
				},
			},
		},
	}

	result := MergeSchemas(program)

	require.Len(t, result.Children, 1)
	require.NotNil(t, result.Children[0].Type)
	require.Equal(t, "User", result.Children[0].Type.Name)
}

func TestMergeSchemas_IncludedFilesAppearFirst(t *testing.T) {
	baseType := &ast.TypeDecl{Name: "BaseType"}
	mainType := &ast.TypeDecl{Name: "MainType"}
	includePath := ast.QuotedString("base.vdl")

	program := &analysis.Program{
		EntryPoint: "/main.vdl",
		Files: map[string]*analysis.File{
			"/main.vdl": {
				Path:     "/main.vdl",
				Includes: []string{"/base.vdl"},
				AST: &ast.Schema{
					Children: []*ast.SchemaChild{
						{Include: &ast.Include{Path: includePath}},
						{Type: mainType},
					},
				},
			},
			"/base.vdl": {
				Path:     "/base.vdl",
				Includes: []string{},
				AST: &ast.Schema{
					Children: []*ast.SchemaChild{
						{Type: baseType},
					},
				},
			},
		},
	}

	result := MergeSchemas(program)

	require.Len(t, result.Children, 2)
	require.Equal(t, "BaseType", result.Children[0].Type.Name, "included file content should appear first")
	require.Equal(t, "MainType", result.Children[1].Type.Name, "main file content should appear after included")
}

func TestMergeSchemas_DeepIncludes(t *testing.T) {
	// main includes mid, mid includes base
	baseType := &ast.TypeDecl{Name: "BaseType"}
	midType := &ast.TypeDecl{Name: "MidType"}
	mainType := &ast.TypeDecl{Name: "MainType"}

	program := &analysis.Program{
		EntryPoint: "/main.vdl",
		Files: map[string]*analysis.File{
			"/main.vdl": {
				Path:     "/main.vdl",
				Includes: []string{"/mid.vdl"},
				AST: &ast.Schema{
					Children: []*ast.SchemaChild{
						{Include: &ast.Include{Path: "mid.vdl"}},
						{Type: mainType},
					},
				},
			},
			"/mid.vdl": {
				Path:     "/mid.vdl",
				Includes: []string{"/base.vdl"},
				AST: &ast.Schema{
					Children: []*ast.SchemaChild{
						{Include: &ast.Include{Path: "base.vdl"}},
						{Type: midType},
					},
				},
			},
			"/base.vdl": {
				Path:     "/base.vdl",
				Includes: []string{},
				AST: &ast.Schema{
					Children: []*ast.SchemaChild{
						{Type: baseType},
					},
				},
			},
		},
	}

	result := MergeSchemas(program)

	require.Len(t, result.Children, 3)
	require.Equal(t, "BaseType", result.Children[0].Type.Name, "deepest include should appear first")
	require.Equal(t, "MidType", result.Children[1].Type.Name, "mid-level include should appear second")
	require.Equal(t, "MainType", result.Children[2].Type.Name, "main file content should appear last")
}

func TestMergeSchemas_CircularIncludesHandled(t *testing.T) {
	typeA := &ast.TypeDecl{Name: "TypeA"}
	typeB := &ast.TypeDecl{Name: "TypeB"}

	program := &analysis.Program{
		EntryPoint: "/a.vdl",
		Files: map[string]*analysis.File{
			"/a.vdl": {
				Path:     "/a.vdl",
				Includes: []string{"/b.vdl"},
				AST: &ast.Schema{
					Children: []*ast.SchemaChild{
						{Include: &ast.Include{Path: "b.vdl"}},
						{Type: typeA},
					},
				},
			},
			"/b.vdl": {
				Path:     "/b.vdl",
				Includes: []string{"/a.vdl"},
				AST: &ast.Schema{
					Children: []*ast.SchemaChild{
						{Include: &ast.Include{Path: "a.vdl"}},
						{Type: typeB},
					},
				},
			},
		},
	}

	result := MergeSchemas(program)

	// Should not infinite loop and should include both types exactly once
	require.Len(t, result.Children, 2)

	names := []string{result.Children[0].Type.Name, result.Children[1].Type.Name}
	require.Contains(t, names, "TypeA")
	require.Contains(t, names, "TypeB")
}

func TestMergeSchemas_MultipleIncludes(t *testing.T) {
	typeA := &ast.TypeDecl{Name: "TypeA"}
	typeB := &ast.TypeDecl{Name: "TypeB"}
	mainType := &ast.TypeDecl{Name: "MainType"}

	program := &analysis.Program{
		EntryPoint: "/main.vdl",
		Files: map[string]*analysis.File{
			"/main.vdl": {
				Path:     "/main.vdl",
				Includes: []string{"/a.vdl", "/b.vdl"},
				AST: &ast.Schema{
					Children: []*ast.SchemaChild{
						{Include: &ast.Include{Path: "a.vdl"}},
						{Include: &ast.Include{Path: "b.vdl"}},
						{Type: mainType},
					},
				},
			},
			"/a.vdl": {
				Path:     "/a.vdl",
				Includes: []string{},
				AST: &ast.Schema{
					Children: []*ast.SchemaChild{
						{Type: typeA},
					},
				},
			},
			"/b.vdl": {
				Path:     "/b.vdl",
				Includes: []string{},
				AST: &ast.Schema{
					Children: []*ast.SchemaChild{
						{Type: typeB},
					},
				},
			},
		},
	}

	result := MergeSchemas(program)

	require.Len(t, result.Children, 3)
	require.Equal(t, "TypeA", result.Children[0].Type.Name, "first include content should appear first")
	require.Equal(t, "TypeB", result.Children[1].Type.Name, "second include content should appear second")
	require.Equal(t, "MainType", result.Children[2].Type.Name, "main file content should appear last")
}

func TestMergeSchemas_MissingIncludedFile(t *testing.T) {
	mainType := &ast.TypeDecl{Name: "MainType"}

	program := &analysis.Program{
		EntryPoint: "/main.vdl",
		Files: map[string]*analysis.File{
			"/main.vdl": {
				Path:     "/main.vdl",
				Includes: []string{"/missing.vdl"},
				AST: &ast.Schema{
					Children: []*ast.SchemaChild{
						{Include: &ast.Include{Path: "missing.vdl"}},
						{Type: mainType},
					},
				},
			},
		},
	}

	result := MergeSchemas(program)

	// Should gracefully handle missing file and include main content
	require.Len(t, result.Children, 1)
	require.Equal(t, "MainType", result.Children[0].Type.Name)
}

func TestMergeSchemas_FileWithNilAST(t *testing.T) {
	mainType := &ast.TypeDecl{Name: "MainType"}

	program := &analysis.Program{
		EntryPoint: "/main.vdl",
		Files: map[string]*analysis.File{
			"/main.vdl": {
				Path:     "/main.vdl",
				Includes: []string{"/empty.vdl"},
				AST: &ast.Schema{
					Children: []*ast.SchemaChild{
						{Include: &ast.Include{Path: "empty.vdl"}},
						{Type: mainType},
					},
				},
			},
			"/empty.vdl": {
				Path:     "/empty.vdl",
				Includes: []string{},
				AST:      nil, // nil AST
			},
		},
	}

	result := MergeSchemas(program)

	require.Len(t, result.Children, 1)
	require.Equal(t, "MainType", result.Children[0].Type.Name)
}

func TestMergeSchemas_PreservesAllNodeTypes(t *testing.T) {
	typeDecl := &ast.TypeDecl{Name: "User"}
	enumDecl := &ast.EnumDecl{Name: "Status"}
	constDecl := &ast.ConstDecl{Name: "VERSION"}
	patternDecl := &ast.PatternDecl{Name: "email"}
	rpcDecl := &ast.RPCDecl{Name: "UserService"}
	simpleComment := "// comment"
	comment := &ast.Comment{Simple: &simpleComment}
	docstring := &ast.Docstring{Value: "docstring"}

	program := &analysis.Program{
		EntryPoint: "/main.vdl",
		Files: map[string]*analysis.File{
			"/main.vdl": {
				Path:     "/main.vdl",
				Includes: []string{},
				AST: &ast.Schema{
					Children: []*ast.SchemaChild{
						{Comment: comment},
						{Docstring: docstring},
						{Type: typeDecl},
						{Enum: enumDecl},
						{Const: constDecl},
						{Pattern: patternDecl},
						{RPC: rpcDecl},
					},
				},
			},
		},
	}

	result := MergeSchemas(program)

	require.Len(t, result.Children, 7)
	require.Equal(t, ast.SchemaChildKindComment, result.Children[0].Kind())
	require.Equal(t, ast.SchemaChildKindDocstring, result.Children[1].Kind())
	require.Equal(t, ast.SchemaChildKindType, result.Children[2].Kind())
	require.Equal(t, ast.SchemaChildKindEnum, result.Children[3].Kind())
	require.Equal(t, ast.SchemaChildKindConst, result.Children[4].Kind())
	require.Equal(t, ast.SchemaChildKindPattern, result.Children[5].Kind())
	require.Equal(t, ast.SchemaChildKindRPC, result.Children[6].Kind())
}

func TestMergeAndFormat_NilProgram(t *testing.T) {
	result := MergeAndFormat(nil)

	require.Empty(t, result)
}

func TestMergeAndFormat_EmptyProgram(t *testing.T) {
	program := &analysis.Program{
		Files: map[string]*analysis.File{},
	}

	result := MergeAndFormat(program)

	require.Empty(t, result)
}

func TestMergeAndFormat_ProducesFormattedOutput(t *testing.T) {
	typeDecl := &ast.TypeDecl{
		Name: "User",
		Children: []*ast.TypeDeclChild{
			{Field: &ast.Field{Name: "name", Type: ast.FieldType{Base: &ast.FieldTypeBase{Named: ptr("string")}}}},
		},
	}

	program := &analysis.Program{
		EntryPoint: "/main.vdl",
		Files: map[string]*analysis.File{
			"/main.vdl": {
				Path:     "/main.vdl",
				Includes: []string{},
				AST: &ast.Schema{
					Children: []*ast.SchemaChild{
						{Type: typeDecl},
					},
				},
			},
		},
	}

	result := MergeAndFormat(program)

	require.Contains(t, result, "type User")
	require.Contains(t, result, "name: string")
	require.True(t, result[len(result)-1] == '\n', "formatted output should end with newline")
}

func TestMergeAndFormat_MergesAndFormats(t *testing.T) {
	baseType := &ast.TypeDecl{
		Name: "Base",
		Children: []*ast.TypeDeclChild{
			{Field: &ast.Field{Name: "id", Type: ast.FieldType{Base: &ast.FieldTypeBase{Named: ptr("int")}}}},
		},
	}
	mainType := &ast.TypeDecl{
		Name: "Main",
		Children: []*ast.TypeDeclChild{
			{Field: &ast.Field{Name: "data", Type: ast.FieldType{Base: &ast.FieldTypeBase{Named: ptr("string")}}}},
		},
	}

	program := &analysis.Program{
		EntryPoint: "/main.vdl",
		Files: map[string]*analysis.File{
			"/main.vdl": {
				Path:     "/main.vdl",
				Includes: []string{"/base.vdl"},
				AST: &ast.Schema{
					Children: []*ast.SchemaChild{
						{Include: &ast.Include{Path: "base.vdl"}},
						{Type: mainType},
					},
				},
			},
			"/base.vdl": {
				Path:     "/base.vdl",
				Includes: []string{},
				AST: &ast.Schema{
					Children: []*ast.SchemaChild{
						{Type: baseType},
					},
				},
			},
		},
	}

	result := MergeAndFormat(program)

	// Base should appear before Main
	baseIdx := indexOf(result, "type Base")
	mainIdx := indexOf(result, "type Main")

	require.True(t, baseIdx < mainIdx, "Base type should appear before Main type in merged output")
}

func ptr(s string) *string {
	return &s
}

func indexOf(s, substr string) int {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

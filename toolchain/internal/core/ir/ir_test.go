package ir

import (
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/varavelio/vdl/toolchain/internal/core/analysis"
	"github.com/varavelio/vdl/toolchain/internal/core/ir/irtypes"
	"github.com/varavelio/vdl/toolchain/internal/core/vfs"
)

var update = flag.Bool("update", false, "update golden files")

// TestFromProgram_Golden runs golden file tests for the IR builder.
// Each .vdl file in testdata is parsed, analyzed, converted to IR,
// and compared against its corresponding .json file.
//
// Run with -update flag to regenerate golden files:
//
//	go test -run TestFromProgram_Golden -update
func TestFromProgram_Golden(t *testing.T) {
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
			schema := FromProgram(program)

			// 3. Serialize to JSON
			got, err := json.MarshalIndent(schema, "", "  ")
			require.NoError(t, err)

			// Golden file path (same name, .json extension)
			goldenPath := strings.TrimSuffix(input, ".vdl") + ".json"

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

			assert.JSONEq(t, string(want), string(got))
		})
	}
}

// TestNormalizeDoc tests the docstring normalization function.
func TestNormalizeDoc(t *testing.T) {
	tests := []struct {
		name     string
		input    *string
		expected string
	}{
		{
			name:     "nil input",
			input:    nil,
			expected: "",
		},
		{
			name:     "empty string",
			input:    ptr(""),
			expected: "",
		},
		{
			name:     "single line",
			input:    ptr("Hello World"),
			expected: "Hello World",
		},
		{
			name:     "single line with whitespace",
			input:    ptr("  Hello World  "),
			expected: "Hello World",
		},
		{
			name:     "multi-line no indent",
			input:    ptr("Line 1\nLine 2\nLine 3"),
			expected: "Line 1\nLine 2\nLine 3",
		},
		{
			name:     "multi-line with common indent",
			input:    ptr("    Line 1\n    Line 2\n    Line 3"),
			expected: "Line 1\nLine 2\nLine 3",
		},
		{
			name:     "multi-line with varying indent",
			input:    ptr("    Line 1\n      Line 2\n    Line 3"),
			expected: "Line 1\n  Line 2\nLine 3",
		},
		{
			name:     "multi-line with empty lines",
			input:    ptr("    Line 1\n\n    Line 2"),
			expected: "Line 1\n\nLine 2",
		},
		{
			name:     "leading and trailing newlines",
			input:    ptr("\n\n    Content\n\n"),
			expected: "Content",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeDoc(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestSorting verifies that all collections are sorted alphabetically.
func TestSorting(t *testing.T) {
	fs := vfs.New()

	// Parse a schema with multiple types in non-alphabetical order
	content := `
		type Zebra { name: string }
		type Alpha { name: string }
		type Middle { name: string }

		enum ZEnum { A }
		enum AEnum { B }

		const Z_CONST = 1
		const A_CONST = 2

		pattern ZPattern = "z:{id}"
		pattern APattern = "a:{id}"

		rpc ZService {
				proc ZProc { input {} output {} }
				proc AProc { input {} output {} }
		}
		rpc AService {
				proc Only { input {} output {} }
		}
	`
	absPath := "/test/sorting.vdl"
	fs.WriteFileCache(absPath, []byte(content))

	program, diags := analysis.Analyze(fs, absPath)
	require.Empty(t, diags)

	schema := FromProgram(program)

	// Check types are sorted
	require.Len(t, schema.Types, 3)
	assert.Equal(t, "Alpha", schema.Types[0].Name)
	assert.Equal(t, "Middle", schema.Types[1].Name)
	assert.Equal(t, "Zebra", schema.Types[2].Name)

	// Check enums are sorted
	require.Len(t, schema.Enums, 2)
	assert.Equal(t, "AEnum", schema.Enums[0].Name)
	assert.Equal(t, "ZEnum", schema.Enums[1].Name)

	// Check constants are sorted
	require.Len(t, schema.Constants, 2)
	assert.Equal(t, "A_CONST", schema.Constants[0].Name)
	assert.Equal(t, "Z_CONST", schema.Constants[1].Name)

	// Check patterns are sorted
	require.Len(t, schema.Patterns, 2)
	assert.Equal(t, "APattern", schema.Patterns[0].Name)
	assert.Equal(t, "ZPattern", schema.Patterns[1].Name)

	// Check RPCs are sorted
	require.Len(t, schema.Rpcs, 2)
	assert.Equal(t, "AService", schema.Rpcs[0].Name)
	assert.Equal(t, "ZService", schema.Rpcs[1].Name)

	// Check procedures are sorted (by RpcName, then Name)
	require.Len(t, schema.Procedures, 3)
	assert.Equal(t, "AService", schema.Procedures[0].RpcName)
	assert.Equal(t, "Only", schema.Procedures[0].Name)
	assert.Equal(t, "ZService", schema.Procedures[1].RpcName)
	assert.Equal(t, "AProc", schema.Procedures[1].Name)
	assert.Equal(t, "ZService", schema.Procedures[2].RpcName)
	assert.Equal(t, "ZProc", schema.Procedures[2].Name)
}

// TestSpreadFlattening verifies that spreads are properly expanded.
func TestSpreadFlattening(t *testing.T) {
	fs := vfs.New()

	content := `
		type Base {
				id: string
				createdAt: datetime
		}

		type Extended {
				...Base
				name: string
				active: bool
		}
	`
	absPath := "/test/spreads.vdl"
	fs.WriteFileCache(absPath, []byte(content))

	program, diags := analysis.Analyze(fs, absPath)
	require.Empty(t, diags)

	schema := FromProgram(program)

	// Find Extended type
	var extended *irtypes.TypeDef
	for i := range schema.Types {
		if schema.Types[i].Name == "Extended" {
			extended = &schema.Types[i]
			break
		}
	}
	require.NotNil(t, extended)

	// Extended should have 4 fields: id, createdAt (from Base), name, active
	require.Len(t, extended.Fields, 4)

	// Fields from spread should come first
	assert.Equal(t, "id", extended.Fields[0].Name)
	assert.Equal(t, "createdAt", extended.Fields[1].Name)
	assert.Equal(t, "name", extended.Fields[2].Name)
	assert.Equal(t, "active", extended.Fields[3].Name)
}

// TestEnumTypeInfo verifies that enum references include EnumType.
func TestEnumTypeInfo(t *testing.T) {
	fs := vfs.New()

	content := `
		enum Status {
				Active
				Inactive
		}

		enum Priority {
				Low = 1
				High = 2
		}

		type Item {
				status: Status
				priority: Priority
				name: string
		}
	`
	absPath := "/test/enuminfo.vdl"
	fs.WriteFileCache(absPath, []byte(content))

	program, diags := analysis.Analyze(fs, absPath)
	require.Empty(t, diags)

	schema := FromProgram(program)

	// Find Item type
	var item *irtypes.TypeDef
	for i := range schema.Types {
		if schema.Types[i].Name == "Item" {
			item = &schema.Types[i]
			break
		}
	}
	require.NotNil(t, item)
	require.Len(t, item.Fields, 3)

	// Find fields by name (they're in original order: status, priority, name)
	fieldsByName := make(map[string]*irtypes.Field)
	for i := range item.Fields {
		fieldsByName[item.Fields[i].Name] = &item.Fields[i]
	}

	// status field should have EnumType with string type
	statusField := fieldsByName["status"]
	require.NotNil(t, statusField)
	assert.Equal(t, irtypes.TypeKindEnum, statusField.TypeRef.Kind)
	require.NotNil(t, statusField.TypeRef.EnumName)
	assert.Equal(t, "Status", *statusField.TypeRef.EnumName)
	require.NotNil(t, statusField.TypeRef.EnumType)
	assert.Equal(t, irtypes.EnumTypeString, *statusField.TypeRef.EnumType)

	// priority field should have EnumType with int type
	priorityField := fieldsByName["priority"]
	require.NotNil(t, priorityField)
	assert.Equal(t, irtypes.TypeKindEnum, priorityField.TypeRef.Kind)
	require.NotNil(t, priorityField.TypeRef.EnumName)
	assert.Equal(t, "Priority", *priorityField.TypeRef.EnumName)
	require.NotNil(t, priorityField.TypeRef.EnumType)
	assert.Equal(t, irtypes.EnumTypeInt, *priorityField.TypeRef.EnumType)

	// name field should NOT have EnumType (it's a primitive)
	nameField := fieldsByName["name"]
	require.NotNil(t, nameField)
	assert.Nil(t, nameField.TypeRef.EnumType)
}

// TestArrayTypes verifies array type handling.
func TestArrayTypes(t *testing.T) {
	fs := vfs.New()

	content := `
		type WithArrays {
				simple: string[]
				nested: int[][]
		}
	`
	absPath := "/test/arrays.vdl"
	fs.WriteFileCache(absPath, []byte(content))

	program, diags := analysis.Analyze(fs, absPath)
	require.Empty(t, diags)

	schema := FromProgram(program)

	require.Len(t, schema.Types, 1)
	typ := schema.Types[0]
	require.Len(t, typ.Fields, 2)

	// Find simple field
	var simple, nested *irtypes.Field
	for i := range typ.Fields {
		if typ.Fields[i].Name == "simple" {
			simple = &typ.Fields[i]
		}
		if typ.Fields[i].Name == "nested" {
			nested = &typ.Fields[i]
		}
	}

	// simple: string[] -> array with 1 dimension of string
	require.NotNil(t, simple)
	assert.Equal(t, irtypes.TypeKindArray, simple.TypeRef.Kind)
	require.NotNil(t, simple.TypeRef.ArrayDims)
	assert.Equal(t, int64(1), *simple.TypeRef.ArrayDims)
	require.NotNil(t, simple.TypeRef.ArrayType)
	assert.Equal(t, irtypes.TypeKindPrimitive, simple.TypeRef.ArrayType.Kind)
	require.NotNil(t, simple.TypeRef.ArrayType.PrimitiveName)
	assert.Equal(t, irtypes.PrimitiveTypeString, *simple.TypeRef.ArrayType.PrimitiveName)

	// nested: int[][] -> array with 2 dimensions of int
	require.NotNil(t, nested)
	assert.Equal(t, irtypes.TypeKindArray, nested.TypeRef.Kind)
	require.NotNil(t, nested.TypeRef.ArrayDims)
	assert.Equal(t, int64(2), *nested.TypeRef.ArrayDims)
	require.NotNil(t, nested.TypeRef.ArrayType)
	assert.Equal(t, irtypes.TypeKindPrimitive, nested.TypeRef.ArrayType.Kind)
	require.NotNil(t, nested.TypeRef.ArrayType.PrimitiveName)
	assert.Equal(t, irtypes.PrimitiveTypeInt, *nested.TypeRef.ArrayType.PrimitiveName)
}

// TestIrSchemaJSONSerialization tests that IrSchema can be serialized to JSON.
func TestIrSchemaJSONSerialization(t *testing.T) {
	schema := &irtypes.IrSchema{
		Types: []irtypes.TypeDef{
			{
				Name: "User",
				Fields: []irtypes.Field{
					{
						Name:     "id",
						Optional: false,
						TypeRef: irtypes.TypeRef{
							Kind:          irtypes.TypeKindPrimitive,
							PrimitiveName: irtypes.Ptr(irtypes.PrimitiveTypeString),
						},
					},
				},
			},
		},
		Enums:      []irtypes.EnumDef{},
		Constants:  []irtypes.ConstantDef{},
		Patterns:   []irtypes.PatternDef{},
		Rpcs:       []irtypes.RpcDef{},
		Procedures: []irtypes.ProcedureDef{},
		Streams:    []irtypes.StreamDef{},
		Docs:       []irtypes.DocDef{},
	}

	jsonBytes, err := json.MarshalIndent(schema, "", "  ")
	require.NoError(t, err)

	// Parse back to verify it's valid JSON
	var parsed irtypes.IrSchema
	err = json.Unmarshal(jsonBytes, &parsed)
	require.NoError(t, err)

	assert.Equal(t, "User", parsed.Types[0].Name)
	assert.Equal(t, "id", parsed.Types[0].Fields[0].Name)
}

// Helper function
func ptr(s string) *string {
	return &s
}

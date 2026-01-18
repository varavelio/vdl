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

// TestToJSON tests the Schema.ToJSON method.
func TestToJSON(t *testing.T) {
	schema := &Schema{
		Types: []Type{
			{
				Name: "User",
				Fields: []Field{
					{
						Name: "id",
						Type: TypeRef{Kind: TypeKindPrimitive, Primitive: PrimitiveString},
					},
				},
			},
		},
		Enums:     []Enum{},
		Constants: []Constant{},
		Patterns:  []Pattern{},
		RPCs:      []RPC{},
	}

	jsonBytes, err := schema.ToJSON()
	require.NoError(t, err)

	// Parse back to verify it's valid JSON
	var parsed Schema
	err = json.Unmarshal(jsonBytes, &parsed)
	require.NoError(t, err)

	assert.Equal(t, "User", parsed.Types[0].Name)
	assert.Equal(t, "id", parsed.Types[0].Fields[0].Name)
}

// TestSorting verifies that all collections are sorted alphabetically.
func TestSorting(t *testing.T) {
	// Create a program-like structure that would result in unsorted output
	// The IR should sort everything alphabetically

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

	// Check services are sorted
	require.Len(t, schema.RPCs, 2)
	assert.Equal(t, "AService", schema.RPCs[0].Name)
	assert.Equal(t, "ZService", schema.RPCs[1].Name)

	// Check procs within service are sorted
	require.Len(t, schema.RPCs[1].Procs, 2)
	assert.Equal(t, "AProc", schema.RPCs[1].Procs[0].Name)
	assert.Equal(t, "ZProc", schema.RPCs[1].Procs[1].Name)
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
	var extended *Type
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

// TestEnumInfo verifies that enum references include EnumInfo.
func TestEnumInfo(t *testing.T) {
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
	var item *Type
	for i := range schema.Types {
		if schema.Types[i].Name == "Item" {
			item = &schema.Types[i]
			break
		}
	}
	require.NotNil(t, item)
	require.Len(t, item.Fields, 3)

	// Find fields by name (they're sorted alphabetically: name, priority, status)
	fieldsByName := make(map[string]*Field)
	for i := range item.Fields {
		fieldsByName[item.Fields[i].Name] = &item.Fields[i]
	}

	// status field should have EnumInfo with string type
	statusField := fieldsByName["status"]
	require.NotNil(t, statusField)
	assert.Equal(t, TypeKindEnum, statusField.Type.Kind)
	assert.Equal(t, "Status", statusField.Type.Enum)
	require.NotNil(t, statusField.Type.EnumInfo)
	assert.Equal(t, EnumValueTypeString, statusField.Type.EnumInfo.ValueType)

	// priority field should have EnumInfo with int type
	priorityField := fieldsByName["priority"]
	require.NotNil(t, priorityField)
	assert.Equal(t, TypeKindEnum, priorityField.Type.Kind)
	assert.Equal(t, "Priority", priorityField.Type.Enum)
	require.NotNil(t, priorityField.Type.EnumInfo)
	assert.Equal(t, EnumValueTypeInt, priorityField.Type.EnumInfo.ValueType)

	// name field should NOT have EnumInfo (it's a primitive)
	nameField := fieldsByName["name"]
	require.NotNil(t, nameField)
	assert.Nil(t, nameField.Type.EnumInfo)
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
	var simple, nested *Field
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
	assert.Equal(t, TypeKindArray, simple.Type.Kind)
	assert.Equal(t, 1, simple.Type.ArrayDimensions)
	require.NotNil(t, simple.Type.ArrayItem)
	assert.Equal(t, TypeKindPrimitive, simple.Type.ArrayItem.Kind)
	assert.Equal(t, PrimitiveString, simple.Type.ArrayItem.Primitive)

	// nested: int[][] -> array with 2 dimensions of int
	require.NotNil(t, nested)
	assert.Equal(t, TypeKindArray, nested.Type.Kind)
	assert.Equal(t, 2, nested.Type.ArrayDimensions)
	require.NotNil(t, nested.Type.ArrayItem)
	assert.Equal(t, TypeKindPrimitive, nested.Type.ArrayItem.Kind)
	assert.Equal(t, PrimitiveInt, nested.Type.ArrayItem.Primitive)
}

// Helper function
func ptr(s string) *string {
	return &s
}

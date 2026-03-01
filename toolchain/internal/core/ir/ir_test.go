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

// TestFromProgram_Golden validates IR output against golden JSON fixtures.
func TestFromProgram_Golden(t *testing.T) {
	inputs, err := filepath.Glob("testdata/*.vdl")
	require.NoError(t, err)
	require.NotEmpty(t, inputs, "no .vdl files found in testdata/")

	for _, input := range inputs {
		name := strings.TrimSuffix(filepath.Base(input), ".vdl")
		t.Run(name, func(t *testing.T) {
			absInput, err := filepath.Abs(input)
			require.NoError(t, err)

			fs := vfs.New()
			program, diags := analysis.Analyze(fs, absInput)
			require.Empty(t, diags, "analysis errors: %v", diags)

			schema := FromProgram(program)
			got, err := json.MarshalIndent(schema, "", "  ")
			require.NoError(t, err)

			goldenPath := strings.TrimSuffix(input, ".vdl") + ".json"
			if *update {
				err := os.WriteFile(goldenPath, got, 0644)
				require.NoError(t, err)
				t.Logf("updated golden file: %s", goldenPath)
				return
			}

			want, err := os.ReadFile(goldenPath)
			if os.IsNotExist(err) {
				t.Fatalf("golden file not found: %s (run with -update to create)", goldenPath)
			}
			require.NoError(t, err)

			assert.JSONEq(t, string(want), string(got))
		})
	}
}

func TestNormalizeDoc(t *testing.T) {
	tests := []struct {
		name     string
		input    *string
		expected string
	}{
		{name: "nil input", input: nil, expected: ""},
		{name: "empty string", input: ptr(""), expected: ""},
		{name: "single line", input: ptr("Hello World"), expected: "Hello World"},
		{name: "single line with whitespace", input: ptr("  Hello World  "), expected: "Hello World"},
		{name: "multi-line with indent", input: ptr("    Line 1\n    Line 2\n    Line 3"), expected: "Line 1\nLine 2\nLine 3"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, normalizeDoc(tt.input))
		})
	}
}

func TestSorting(t *testing.T) {
	fs := vfs.New()
	content := `
type Zebra {
  name string
}

type Alpha {
  name string
}

enum ZEnum {
  A
}

enum AEnum {
  B
}

const zConst = 1
const aConst = 2
`
	absPath := "/test/sorting.vdl"
	fs.WriteFileCache(absPath, []byte(content))

	program, diags := analysis.Analyze(fs, absPath)
	require.Empty(t, diags)

	schema := FromProgram(program)

	require.Len(t, schema.Types, 2)
	assert.Equal(t, "Alpha", schema.Types[0].Name)
	assert.Equal(t, "Zebra", schema.Types[1].Name)

	require.Len(t, schema.Enums, 2)
	assert.Equal(t, "AEnum", schema.Enums[0].Name)
	assert.Equal(t, "ZEnum", schema.Enums[1].Name)

	require.Len(t, schema.Constants, 2)
	assert.Equal(t, "aConst", schema.Constants[0].Name)
	assert.Equal(t, "zConst", schema.Constants[1].Name)
}

func TestSpreadFlattening(t *testing.T) {
	fs := vfs.New()
	content := `
type Base {
  id string
  createdAt datetime
}

type Extended {
  ...Base
  name string
  active bool
}
`
	absPath := "/test/spreads.vdl"
	fs.WriteFileCache(absPath, []byte(content))

	program, diags := analysis.Analyze(fs, absPath)
	require.Empty(t, diags)

	schema := FromProgram(program)

	var extended *irtypes.TypeDef
	for i := range schema.Types {
		if schema.Types[i].Name == "Extended" {
			extended = &schema.Types[i]
			break
		}
	}
	require.NotNil(t, extended)
	require.Len(t, extended.Fields, 4)

	assert.Equal(t, "id", extended.Fields[0].Name)
	assert.Equal(t, "createdAt", extended.Fields[1].Name)
	assert.Equal(t, "name", extended.Fields[2].Name)
	assert.Equal(t, "active", extended.Fields[3].Name)
}

func TestConstantResolution(t *testing.T) {
	fs := vfs.New()
	content := `
enum Status {
  Active
  Inactive = "inactive"
}

const base = {
  host "localhost"
  port 8080
}

const cfg = {
  ...base
  secure true
  status Status.Active
}

const cfgAlias = cfg
`
	absPath := "/test/constants.vdl"
	fs.WriteFileCache(absPath, []byte(content))

	program, diags := analysis.Analyze(fs, absPath)
	require.Empty(t, diags)

	schema := FromProgram(program)

	constants := map[string]irtypes.ConstantDef{}
	for _, c := range schema.Constants {
		constants[c.Name] = c
	}

	cfg, ok := constants["cfg"]
	require.True(t, ok)
	assert.Equal(t, irtypes.TypeKindObject, cfg.TypeRef.Kind)
	assert.Equal(t, irtypes.ValueKindObject, cfg.Value.Kind)

	entries := cfg.Value.GetObjectEntries()
	require.Len(t, entries, 4)
	assert.Equal(t, "host", entries[0].Key)
	assert.Equal(t, "localhost", entries[0].Value.GetStringValue())
	assert.Equal(t, "status", entries[3].Key)
	assert.Equal(t, "Active", entries[3].Value.GetStringValue())

	alias, ok := constants["cfgAlias"]
	require.True(t, ok)
	assert.Equal(t, cfg.Value, alias.Value)
}

func TestAnnotationsPreserved(t *testing.T) {
	fs := vfs.New()
	content := `
@entity
@meta({ owner "core" retries 2 })
type User {
  @id
  id string
}

@featureFlag
const maxRetries = 3

enum Status {
  @deprecated("Use Inactive")
  Legacy
  Active
}
`
	absPath := "/test/annotations.vdl"
	fs.WriteFileCache(absPath, []byte(content))

	program, diags := analysis.Analyze(fs, absPath)
	require.Empty(t, diags)

	schema := FromProgram(program)

	var user irtypes.TypeDef
	for _, typ := range schema.Types {
		if typ.Name == "User" {
			user = typ
			break
		}
	}
	require.Equal(t, "User", user.Name)
	require.Len(t, user.Annotations, 2)
	assert.Equal(t, "entity", user.Annotations[0].Name)
	assert.Equal(t, "meta", user.Annotations[1].Name)
	assert.Equal(t, irtypes.ValueKindObject, user.Annotations[1].Argument.GetKind())

	var status irtypes.EnumDef
	for _, enum := range schema.Enums {
		if enum.Name == "Status" {
			status = enum
			break
		}
	}
	require.Equal(t, "Status", status.Name)
	require.Len(t, status.Members, 2)
	require.Len(t, status.Members[0].Annotations, 1)
	assert.Equal(t, "deprecated", status.Members[0].Annotations[0].Name)
}

func TestIrSchemaJSONSerialization(t *testing.T) {
	schema := &irtypes.IrSchema{
		Types: []irtypes.TypeDef{
			{
				Name:        "User",
				Annotations: []irtypes.Annotation{},
				Fields: []irtypes.Field{
					{
						Name:        "id",
						Optional:    false,
						Annotations: []irtypes.Annotation{},
						TypeRef: irtypes.TypeRef{
							Kind:          irtypes.TypeKindPrimitive,
							PrimitiveName: irtypes.Ptr(irtypes.PrimitiveTypeString),
						},
					},
				},
			},
		},
		Enums: []irtypes.EnumDef{},
		Constants: []irtypes.ConstantDef{
			{
				Name:        "maxRetries",
				Annotations: []irtypes.Annotation{},
				TypeRef: irtypes.TypeRef{
					Kind:          irtypes.TypeKindPrimitive,
					PrimitiveName: irtypes.Ptr(irtypes.PrimitiveTypeInt),
				},
				Value: irtypes.Value{
					Kind:     irtypes.ValueKindInt,
					IntValue: irtypes.Ptr(int64(3)),
				},
			},
		},
		Docs: []irtypes.DocDef{},
	}

	jsonBytes, err := json.MarshalIndent(schema, "", "  ")
	require.NoError(t, err)

	var parsed irtypes.IrSchema
	err = json.Unmarshal(jsonBytes, &parsed)
	require.NoError(t, err)

	assert.Equal(t, "User", parsed.Types[0].Name)
	assert.Equal(t, "id", parsed.Types[0].Fields[0].Name)
	assert.Equal(t, "maxRetries", parsed.Constants[0].Name)
	assert.Equal(t, int64(3), parsed.Constants[0].Value.GetIntValue())
}

func ptr(s string) *string {
	return &s
}

package playground

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/varavelio/vdl/toolchain/internal/codegen/config/configtypes"
	"github.com/varavelio/vdl/toolchain/internal/core/ir/irtypes"
)

func TestGenerator_Name(t *testing.T) {
	gen := New(&configtypes.PlaygroundTargetConfig{}, "")
	assert.Equal(t, "playground", gen.Name())
}

func TestGenerator_Generate_BasicFiles(t *testing.T) {
	gen := New(&configtypes.PlaygroundTargetConfig{
		Output: "dist",
	}, "")

	schema := &irtypes.IrSchema{
		Types: []irtypes.TypeDef{},
		Enums: []irtypes.EnumDef{},
		Rpcs:  []irtypes.RpcDef{},
	}

	files, err := gen.Generate(context.Background(), schema)
	require.NoError(t, err)

	// Should have at least the embedded playground files
	require.NotEmpty(t, files)

	// Check that .gitkeep files are filtered out
	for _, f := range files {
		assert.NotEqual(t, ".gitkeep", f.RelativePath)
	}
}

func TestGenerator_Generate_WithFormattedSchema(t *testing.T) {
	schemaSource := `type User {
    id: string
    name: string
}

rpc Users {
    proc GetUser {
        input { id: string }
        output { user: User }
    }
}
`

	gen := New(&configtypes.PlaygroundTargetConfig{
		Output: "dist",
	}, schemaSource)

	schema := &irtypes.IrSchema{}

	files, err := gen.Generate(context.Background(), schema)
	require.NoError(t, err)

	// Find schema.vdl file
	var schemaFile *File
	for i := range files {
		if files[i].RelativePath == "schema.vdl" {
			schemaFile = &files[i]
			break
		}
	}

	require.NotNil(t, schemaFile, "schema.vdl should be included")
	assert.Equal(t, schemaSource, string(schemaFile.Content))
}

func TestGenerator_Generate_WithConfig(t *testing.T) {
	baseUrl := "https://api.example.com"
	headers := map[string]string{
		"Authorization": "Bearer token",
		"X-Custom":      "value",
	}
	gen := New(&configtypes.PlaygroundTargetConfig{
		Output:         "dist",
		DefaultBaseUrl: &baseUrl,
		DefaultHeaders: &headers,
	}, "")

	schema := &irtypes.IrSchema{}

	files, err := gen.Generate(context.Background(), schema)
	require.NoError(t, err)

	// Find config.json file
	var configFile *File
	for i := range files {
		if files[i].RelativePath == "config.json" {
			configFile = &files[i]
			break
		}
	}

	require.NotNil(t, configFile, "config.json should be included")

	// Check that config.json contains expected values
	content := string(configFile.Content)
	assert.Contains(t, content, "https://api.example.com")
	assert.Contains(t, content, "Authorization")
	assert.Contains(t, content, "Bearer token")
	assert.Contains(t, content, "X-Custom")
}

func TestGenerator_Generate_NoConfigWithoutValues(t *testing.T) {
	gen := New(&configtypes.PlaygroundTargetConfig{
		Output: "dist",
		// No base URL or headers
	}, "")

	schema := &irtypes.IrSchema{}

	files, err := gen.Generate(context.Background(), schema)
	require.NoError(t, err)

	// config.json should NOT be included when there's nothing to configure
	for _, f := range files {
		assert.NotEqual(t, "config.json", f.RelativePath)
	}
}

func TestGenerator_Generate_NoSchemaWithoutFormattedSchema(t *testing.T) {
	gen := New(&configtypes.PlaygroundTargetConfig{
		Output: "dist",
	}, "")

	schema := &irtypes.IrSchema{}

	files, err := gen.Generate(context.Background(), schema)
	require.NoError(t, err)

	// schema.vdl should NOT be included when FormattedSchema is empty
	for _, f := range files {
		assert.NotEqual(t, "schema.vdl", f.RelativePath)
	}
}

func TestGenerateConfigJSON(t *testing.T) {
	baseUrl := "https://api.test.com"
	headers := map[string]string{"Content-Type": "application/json"}
	gen := New(&configtypes.PlaygroundTargetConfig{
		DefaultBaseUrl: &baseUrl,
		DefaultHeaders: &headers,
	}, "")

	jsonBytes, err := gen.generateConfigJSON()
	require.NoError(t, err)

	expected := `{"baseUrl":"https://api.test.com","headers":[{"key":"Content-Type","value":"application/json"}]}`
	assert.JSONEq(t, expected, string(jsonBytes))
}

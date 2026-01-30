package playground

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/varavelio/vdl/toolchain/internal/codegen/config"
	"github.com/varavelio/vdl/toolchain/internal/core/ir"
)

func TestGenerator_Name(t *testing.T) {
	gen := New(&config.PlaygroundConfig{}, "")
	assert.Equal(t, "playground", gen.Name())
}

func TestGenerator_Generate_BasicFiles(t *testing.T) {
	gen := New(&config.PlaygroundConfig{
		CommonConfig: config.CommonConfig{
			Output: "dist",
		},
	}, "")

	schema := &ir.Schema{
		Types: []ir.Type{},
		Enums: []ir.Enum{},
		RPCs:  []ir.RPC{},
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

	gen := New(&config.PlaygroundConfig{
		CommonConfig: config.CommonConfig{
			Output: "dist",
		},
	}, schemaSource)

	schema := &ir.Schema{}

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
	gen := New(&config.PlaygroundConfig{
		CommonConfig: config.CommonConfig{
			Output: "dist",
		},
		DefaultBaseURL: "https://api.example.com",
		DefaultHeaders: []struct {
			Key   string `yaml:"key" json:"key"`
			Value string `yaml:"value" json:"value"`
		}{
			{Key: "Authorization", Value: "Bearer token"},
			{Key: "X-Custom", Value: "value"},
		},
	}, "")

	schema := &ir.Schema{}

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
	gen := New(&config.PlaygroundConfig{
		CommonConfig: config.CommonConfig{
			Output: "dist",
		},
		// No base URL or headers
	}, "")

	schema := &ir.Schema{}

	files, err := gen.Generate(context.Background(), schema)
	require.NoError(t, err)

	// config.json should NOT be included when there's nothing to configure
	for _, f := range files {
		assert.NotEqual(t, "config.json", f.RelativePath)
	}
}

func TestGenerator_Generate_NoSchemaWithoutFormattedSchema(t *testing.T) {
	gen := New(&config.PlaygroundConfig{
		CommonConfig: config.CommonConfig{
			Output: "dist",
		},
	}, "")

	schema := &ir.Schema{}

	files, err := gen.Generate(context.Background(), schema)
	require.NoError(t, err)

	// schema.vdl should NOT be included when FormattedSchema is empty
	for _, f := range files {
		assert.NotEqual(t, "schema.vdl", f.RelativePath)
	}
}

func TestGenerateConfigJSON(t *testing.T) {
	gen := New(&config.PlaygroundConfig{
		DefaultBaseURL: "https://api.test.com",
		DefaultHeaders: []struct {
			Key   string `yaml:"key" json:"key"`
			Value string `yaml:"value" json:"value"`
		}{
			{Key: "Content-Type", Value: "application/json"},
		},
	}, "")

	jsonBytes, err := gen.generateConfigJSON()
	require.NoError(t, err)

	expected := `{"baseUrl":"https://api.test.com","headers":[{"key":"Content-Type","value":"application/json"}]}`
	assert.JSONEq(t, expected, string(jsonBytes))
}

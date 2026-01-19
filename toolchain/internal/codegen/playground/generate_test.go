package playground

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/varavelio/vdl/toolchain/internal/core/ir"
)

func TestGenerator_Name(t *testing.T) {
	gen := New(Config{})
	assert.Equal(t, "playground", gen.Name())
}

func TestGenerator_Generate_BasicFiles(t *testing.T) {
	gen := New(Config{
		OutputDir: "dist",
	})

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

	gen := New(Config{
		OutputDir:       "dist",
		FormattedSchema: schemaSource,
	})

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
	gen := New(Config{
		OutputDir:      "dist",
		DefaultBaseURL: "https://api.example.com",
		DefaultHeaders: []Header{
			{Key: "Authorization", Value: "Bearer token"},
			{Key: "X-Custom", Value: "value"},
		},
	})

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
	gen := New(Config{
		OutputDir: "dist",
		// No base URL or headers
	})

	schema := &ir.Schema{}

	files, err := gen.Generate(context.Background(), schema)
	require.NoError(t, err)

	// config.json should NOT be included when there's nothing to configure
	for _, f := range files {
		assert.NotEqual(t, "config.json", f.RelativePath)
	}
}

func TestGenerator_Generate_NoSchemaWithoutFormattedSchema(t *testing.T) {
	gen := New(Config{
		OutputDir:       "dist",
		FormattedSchema: "", // Empty
	})

	schema := &ir.Schema{}

	files, err := gen.Generate(context.Background(), schema)
	require.NoError(t, err)

	// schema.vdl should NOT be included when FormattedSchema is empty
	for _, f := range files {
		assert.NotEqual(t, "schema.vdl", f.RelativePath)
	}
}

func TestConfig_Validate(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		cfg := Config{OutputDir: "dist"}
		assert.NoError(t, cfg.Validate())
	})

	t.Run("missing output_dir", func(t *testing.T) {
		cfg := Config{}
		err := cfg.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "output_dir")
	})
}

func TestGenerateConfigJSON(t *testing.T) {
	gen := New(Config{
		DefaultBaseURL: "https://api.test.com",
		DefaultHeaders: []Header{
			{Key: "Content-Type", Value: "application/json"},
		},
	})

	jsonBytes, err := gen.generateConfigJSON()
	require.NoError(t, err)

	expected := `{"baseUrl":"https://api.test.com","headers":[{"key":"Content-Type","value":"application/json"}]}`
	assert.JSONEq(t, expected, string(jsonBytes))
}

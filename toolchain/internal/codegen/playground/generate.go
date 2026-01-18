package playground

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/uforg/uforpc/embedplayground"
	"github.com/uforg/uforpc/urpc/internal/urpc/ast"
	"github.com/uforg/uforpc/urpc/internal/urpc/formatter"
)

// Generate takes a schema and a config and generates the playground for the schema.
func Generate(absConfigDir string, sch *ast.Schema, config Config) error {
	outputDir := filepath.Join(absConfigDir, config.OutputDir)

	if err := os.RemoveAll(outputDir); err != nil {
		return fmt.Errorf("error emptying output directory: %w", err)
	}

	err := extractEmbedFS(embedplayground.BuildFS, "build", outputDir)
	if err != nil {
		return fmt.Errorf("error extracting embedded filesystem: %w", err)
	}

	gitkeepPath := filepath.Join(outputDir, ".gitkeep")
	if err := os.Remove(gitkeepPath); err != nil {
		return fmt.Errorf("error deleting .gitkeep file: %w", err)
	}

	formattedSchema := formatter.FormatSchema(sch)
	formattedSchemaPath := filepath.Join(outputDir, "schema.urpc")
	if err := os.WriteFile(formattedSchemaPath, []byte(formattedSchema), 0644); err != nil {
		return fmt.Errorf("error writing formatted schema to %s: %w", formattedSchemaPath, err)
	}

	hasConfig := config.DefaultBaseURL != "" || len(config.DefaultHeaders) > 0
	if hasConfig {
		type jsonConfigHeader struct {
			Key   string `json:"key"`
			Value string `json:"value"`
		}

		type jsonConfig struct {
			BaseURL string             `json:"baseUrl,omitempty,omitzero"`
			Headers []jsonConfigHeader `json:"headers,omitempty,omitzero"`
		}

		jsonConfigHeaders := make([]jsonConfigHeader, len(config.DefaultHeaders))
		for i, header := range config.DefaultHeaders {
			jsonConfigHeaders[i] = jsonConfigHeader(header)
		}

		jsonConf := jsonConfig{
			BaseURL: config.DefaultBaseURL,
			Headers: jsonConfigHeaders,
		}

		jsonConfigBytes, err := json.Marshal(jsonConf)
		if err != nil {
			return fmt.Errorf("error marshalling config to JSON: %w", err)
		}

		configPath := filepath.Join(outputDir, "config.json")
		if err := os.WriteFile(configPath, jsonConfigBytes, 0644); err != nil {
			return fmt.Errorf("error writing config to %s: %w", configPath, err)
		}
	}

	return nil
}

func extractEmbedFS(embedFS embed.FS, rootDir string, destDir string) error {
	return fs.WalkDir(embedFS, rootDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(rootDir, path)
		if err != nil {
			return err
		}

		destPath := filepath.Join(destDir, relPath)

		if d.IsDir() {
			return os.MkdirAll(destPath, 0o700)
		}

		data, err := fs.ReadFile(embedFS, path)
		if err != nil {
			return err
		}

		return os.WriteFile(destPath, data, 0o644)
	})
}

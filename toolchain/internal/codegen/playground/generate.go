package playground

import (
	"context"
	"embed"
	"encoding/json"
	"io/fs"
	"path/filepath"
	"strings"

	embedplayground "github.com/varavelio/vdl/playground"
	"github.com/varavelio/vdl/toolchain/internal/codegen/config"
	"github.com/varavelio/vdl/toolchain/internal/core/ir/irtypes"
)

// File represents a generated file. This mirrors codegen.File to avoid import cycles.
type File struct {
	RelativePath string
	Content      []byte
}

// Generator implements the playground generator.
type Generator struct {
	config          *config.PlaygroundConfig
	formattedSchema string
}

// New creates a new playground generator with the given config.
func New(config *config.PlaygroundConfig, formattedSchema string) *Generator {
	return &Generator{
		config:          config,
		formattedSchema: formattedSchema,
	}
}

// Name returns the generator name.
func (g *Generator) Name() string {
	return "playground"
}

// Generate produces playground files from the IR schema.
// The playground consists of:
// - All static files from the embedded playground build
// - schema.vdl: The formatted VDL schema (from g.formattedSchema)
// - config.json: Optional configuration for base URL and headers
func (g *Generator) Generate(ctx context.Context, schema *irtypes.IrSchema) ([]File, error) {
	files := []File{}

	// Files to skip from embedded content (we'll generate our own)
	skipFiles := map[string]bool{
		".gitkeep":    true,
		"config.json": true,
		"schema.vdl":  true,
	}

	// 1. Extract all files from the embedded playground build
	embeddedFiles, err := extractEmbedFS(embedplayground.BuildFS, "build")
	if err != nil {
		return nil, err
	}

	// Filter out files we want to replace/skip
	for _, f := range embeddedFiles {
		baseName := filepath.Base(f.RelativePath)
		if skipFiles[baseName] {
			continue
		}
		files = append(files, f)
	}

	// 2. Add the formatted schema if provided
	if g.formattedSchema != "" {
		files = append(files, File{
			RelativePath: "schema.vdl",
			Content:      []byte(g.formattedSchema),
		})
	}

	// 3. Add config.json if there's configuration to include
	hasConfig := g.config.DefaultBaseURL != "" || len(g.config.DefaultHeaders) > 0
	if hasConfig {
		configJSON, err := g.generateConfigJSON()
		if err != nil {
			return nil, err
		}
		files = append(files, File{
			RelativePath: "config.json",
			Content:      configJSON,
		})
	}

	return files, nil
}

// generateConfigJSON creates the config.json content for the playground.
func (g *Generator) generateConfigJSON() ([]byte, error) {
	type jsonConfigHeader struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}

	type jsonConfig struct {
		BaseURL string             `json:"baseUrl,omitempty"`
		Headers []jsonConfigHeader `json:"headers,omitempty"`
	}

	headers := make([]jsonConfigHeader, len(g.config.DefaultHeaders))
	for i, header := range g.config.DefaultHeaders {
		headers[i] = jsonConfigHeader{
			Key:   header.Key,
			Value: header.Value,
		}
	}

	conf := jsonConfig{
		BaseURL: g.config.DefaultBaseURL,
		Headers: headers,
	}

	return json.Marshal(conf)
}

// extractEmbedFS extracts all files from an embedded filesystem.
// Returns files with paths relative to the rootDir.
func extractEmbedFS(embedFS embed.FS, rootDir string) ([]File, error) {
	files := []File{}

	err := fs.WalkDir(embedFS, rootDir, func(filePath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if d.IsDir() {
			return nil
		}

		// Calculate relative path from rootDir
		// For embedded FS, paths always use forward slashes
		relPath := strings.TrimPrefix(filePath, rootDir+"/")
		if relPath == filePath {
			// If no prefix was trimmed, try without trailing slash
			relPath = strings.TrimPrefix(filePath, rootDir)
		}

		// Read file content
		data, err := fs.ReadFile(embedFS, filePath)
		if err != nil {
			return err
		}

		files = append(files, File{
			RelativePath: relPath,
			Content:      data,
		})

		return nil
	})

	if err != nil {
		return nil, err
	}

	return files, nil
}

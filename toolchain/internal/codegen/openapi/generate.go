package openapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/varavelio/vdl/toolchain/internal/codegen/config"
	"github.com/varavelio/vdl/toolchain/internal/core/ir"
	"gopkg.in/yaml.v3"
)

// File represents a generated file. This mirrors codegen.File to avoid import cycles.
type File struct {
	RelativePath string
	Content      []byte
}

// Generator implements the OpenAPI generator.
type Generator struct {
	config *config.OpenAPIConfig
}

// New creates a new OpenAPI generator with the given config.
func New(config *config.OpenAPIConfig) *Generator {
	return &Generator{config: config}
}

// Name returns the generator name.
func (g *Generator) Name() string {
	return "openapi"
}

// Generate produces OpenAPI spec files from the IR schema.
func (g *Generator) Generate(ctx context.Context, schema *ir.Schema) ([]File, error) {
	cfg := g.config

	if cfg.Title == "" {
		cfg.Title = "VDL RPC API"
	}
	if cfg.Version == "" {
		cfg.Version = "1.0.0"
	}

	spec := Spec{
		OpenAPI: "3.0.0",
		Info: Info{
			Title:       cfg.Title,
			Version:     cfg.Version,
			Description: cfg.Description,
			Contact: InfoContact{
				Name:  cfg.ContactName,
				Email: cfg.ContactEmail,
			},
			License: InfoLicense{
				Name: cfg.LicenseName,
			},
		},
		Security: []map[string][]string{
			{
				"AuthToken": {},
			},
		},
	}

	if cfg.BaseURL != "" {
		spec.Servers = []Server{
			{
				URL: cfg.BaseURL,
			},
		}
	}

	// Generate tags from RPCs
	spec.Tags = generateTags(schema)

	// Generate paths
	spec.Paths = generatePaths(schema)

	// Generate components
	spec.Components = generateComponents(schema)

	// Encode spec
	code, err := encodeSpec(spec, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to generate spec file: %w", err)
	}

	filename := cfg.Filename
	if filename == "" {
		filename = "openapi.yaml"
	}

	return []File{
		{
			RelativePath: filename,
			Content:      []byte(code),
		},
	}, nil
}

// generateTags creates OpenAPI tags from the schema RPCs.
// Tags are generated in PascalCase format: {RPC}Procedures, {RPC}Streams
func generateTags(schema *ir.Schema) []Tag {
	tags := []Tag{}

	for _, rpc := range schema.RPCs {
		// Tag for procedures of this RPC
		if len(rpc.Procs) > 0 {
			desc := fmt.Sprintf("Procedures for %s", rpc.Name)
			if rpc.Doc != "" {
				desc = rpc.Doc
			}
			tags = append(tags, Tag{
				Name:        rpc.Name + "Procedures",
				Description: desc,
			})
		}

		// Tag for streams of this RPC
		if len(rpc.Streams) > 0 {
			desc := fmt.Sprintf("Streams for %s", rpc.Name)
			if rpc.Doc != "" {
				desc = rpc.Doc
			}
			tags = append(tags, Tag{
				Name:        rpc.Name + "Streams",
				Description: desc,
			})
		}
	}

	// Sort tags alphabetically for deterministic output
	sort.Slice(tags, func(i, j int) bool {
		return tags[i].Name < tags[j].Name
	})

	return tags
}

func encodeSpec(spec Spec, cfg *config.OpenAPIConfig) (string, error) {
	filename := cfg.Filename
	if filename == "" {
		filename = "openapi.yaml"
	}

	isYAML := strings.HasSuffix(filename, ".yaml") || strings.HasSuffix(filename, ".yml")
	var buf bytes.Buffer

	if isYAML {
		enc := yaml.NewEncoder(&buf)
		if err := enc.Encode(spec); err != nil {
			return "", fmt.Errorf("failed to encode yaml spec: %w", err)
		}
		return buf.String(), nil
	}

	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(spec); err != nil {
		return "", fmt.Errorf("failed to encode json spec: %w", err)
	}
	return buf.String(), nil
}

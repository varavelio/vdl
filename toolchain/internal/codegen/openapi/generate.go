package openapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/varavelio/vdl/toolchain/internal/codegen/config/configtypes"
	"github.com/varavelio/vdl/toolchain/internal/core/ir/irtypes"
	"gopkg.in/yaml.v3"
)

// File represents a generated file. This mirrors codegen.File to avoid import cycles.
type File struct {
	RelativePath string
	Content      []byte
}

// Generator implements the OpenAPI generator.
type Generator struct {
	config *configtypes.OpenApiConfig
}

// New creates a new OpenAPI generator with the given config.
func New(config *configtypes.OpenApiConfig) *Generator {
	return &Generator{config: config}
}

// Name returns the generator name.
func (g *Generator) Name() string {
	return "openapi"
}

// Generate produces OpenAPI spec files from the IR schema.
func (g *Generator) Generate(ctx context.Context, schema *irtypes.IrSchema) ([]File, error) {
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
			Title:   cfg.Title,
			Version: cfg.Version,
		},
		Security: []map[string][]string{
			{
				"AuthToken": {},
			},
		},
	}

	// Set optional Info fields
	if cfg.Description != nil {
		spec.Info.Description = *cfg.Description
	}
	if cfg.ContactName != nil {
		spec.Info.Contact.Name = *cfg.ContactName
	}
	if cfg.ContactEmail != nil {
		spec.Info.Contact.Email = *cfg.ContactEmail
	}
	if cfg.LicenseName != nil {
		spec.Info.License.Name = *cfg.LicenseName
	}
	if cfg.BaseUrl != nil && *cfg.BaseUrl != "" {
		spec.Servers = []Server{
			{
				URL: *cfg.BaseUrl,
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

	filename := cfg.GetFilenameOr("openapi.yaml")

	return []File{
		{
			RelativePath: filename,
			Content:      []byte(code),
		},
	}, nil
}

// generateTags creates OpenAPI tags from the schema RPCs.
// Tags are generated in PascalCase format: {RPC}Procedures, {RPC}Streams
func generateTags(schema *irtypes.IrSchema) []Tag {
	tags := []Tag{}

	// Build a map of RPC names to check which have procedures or streams
	rpcHasProcs := make(map[string]bool)
	rpcHasStreams := make(map[string]bool)
	rpcDocs := make(map[string]string)

	for _, rpc := range schema.Rpcs {
		rpcDocs[rpc.Name] = rpc.GetDoc()
	}

	for _, proc := range schema.Procedures {
		rpcHasProcs[proc.RpcName] = true
	}

	for _, stream := range schema.Streams {
		rpcHasStreams[stream.RpcName] = true
	}

	for _, rpc := range schema.Rpcs {
		// Tag for procedures of this RPC
		if rpcHasProcs[rpc.Name] {
			desc := fmt.Sprintf("Procedures for %s", rpc.Name)
			if rpcDocs[rpc.Name] != "" {
				desc = rpcDocs[rpc.Name]
			}
			tags = append(tags, Tag{
				Name:        rpc.Name + "Procedures",
				Description: desc,
			})
		}

		// Tag for streams of this RPC
		if rpcHasStreams[rpc.Name] {
			desc := fmt.Sprintf("Streams for %s", rpc.Name)
			if rpcDocs[rpc.Name] != "" {
				desc = rpcDocs[rpc.Name]
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

func encodeSpec(spec Spec, cfg *configtypes.OpenApiConfig) (string, error) {
	filename := cfg.GetFilenameOr("openapi.yaml")

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

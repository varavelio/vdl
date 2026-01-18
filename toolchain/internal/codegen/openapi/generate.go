package openapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/goccy/go-yaml"
	"github.com/uforg/uforpc/urpc/internal/schema"
)

func Generate(schema schema.Schema, config Config) (string, error) {
	if config.Title == "" {
		config.Title = "UFO RPC API"
	}
	if config.Version == "" {
		config.Version = "1.0.0"
	}

	spec := Spec{
		OpenAPI: "3.0.0",
		Info: Info{
			Title:       config.Title,
			Version:     config.Version,
			Description: config.Description,
			Contact: InfoContact{
				Name:  config.ContactName,
				Email: config.ContactEmail,
			},
			License: InfoLicense{
				Name: config.LicenseName,
			},
		},
		Security: []map[string][]string{
			{
				"AuthToken": {},
			},
		},
		Tags: []Tag{
			{
				Name:        "procedures",
				Description: "All procedures from the UFO RPC schema",
			},
			{
				Name:        "streams",
				Description: "All streams from the UFO RPC schema",
			},
		},
	}

	if config.BaseURL != "" {
		spec.Servers = []Server{
			{
				URL: config.BaseURL,
			},
		}
	}

	paths, err := generatePaths(schema)
	if err != nil {
		return "", fmt.Errorf("failed to generate paths: %w", err)
	}
	spec.Paths = paths

	components, err := generateComponents(schema)
	if err != nil {
		return "", fmt.Errorf("failed to generate components: %w", err)
	}
	spec.Components = components

	code, err := encodeSpec(spec, config)
	if err != nil {
		return "", fmt.Errorf("failed to generate spec file: %w", err)
	}

	return code, nil
}

func encodeSpec(spec Spec, config Config) (string, error) {
	isYAML := strings.HasSuffix(config.OutputFile, ".yaml") || strings.HasSuffix(config.OutputFile, ".yml")
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

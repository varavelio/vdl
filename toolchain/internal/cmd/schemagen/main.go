package main

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"

	"github.com/invopop/jsonschema"
	"github.com/varavelio/vdl/toolchain/internal/codegen/config"
	"github.com/varavelio/vdl/toolchain/internal/codegen/plugin"
	"github.com/varavelio/vdl/toolchain/internal/core/ir"
	"github.com/varavelio/vdl/toolchain/internal/version"
)

type schemaGenConfig struct {
	Struct      any
	ID          string
	Title       string
	Description string
	OutPath     []string
}

func main() {
	configs := []schemaGenConfig{
		{
			Struct:      &ir.Schema{},
			ID:          version.SchemaIRID,
			Title:       "VDL IR Schema",
			Description: "JSON Schema for the VDL Intermediate Representation",
			OutPath:     []string{"internal", "core", "ir", "ir.schema.json"},
		},
		{
			Struct:      &config.VDLConfig{},
			ID:          version.SchemaConfigID,
			Title:       "VDL Config Schema",
			Description: "JSON Schema for the VDL Config",
			OutPath:     []string{"internal", "codegen", "config", "config.schema.json"},
		},
		{
			Struct:      &plugin.Input{},
			ID:          version.SchemaPluginInputID,
			Title:       "VDL Plugin Input Schema",
			Description: "JSON Schema for the VDL Plugin Input Protocol",
			OutPath:     []string{"internal", "codegen", "plugin", "plugin_input.schema.json"},
		},
		{
			Struct:      &plugin.Output{},
			ID:          version.SchemaPluginOutputID,
			Title:       "VDL Plugin Output Schema",
			Description: "JSON Schema for the VDL Plugin Output Protocol",
			OutPath:     []string{"internal", "codegen", "plugin", "plugin_output.schema.json"},
		},
	}

	for _, cfg := range configs {
		if err := generate(cfg); err != nil {
			log.Fatalf("failed to generate schema for %s: %v", cfg.Title, err)
		}
	}
}

func generate(cfg schemaGenConfig) error {
	schema := (&jsonschema.Reflector{}).Reflect(cfg.Struct)
	schema.ID = jsonschema.ID(cfg.ID)
	schema.Title = cfg.Title
	schema.Description = cfg.Description

	data, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return err
	}

	outPath := filepath.Join(cfg.OutPath...)
	if err := os.WriteFile(outPath, data, 0644); err != nil {
		return err
	}

	log.Printf("Generated %s", outPath)
	return nil
}

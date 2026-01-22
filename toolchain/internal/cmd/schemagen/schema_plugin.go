package main

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"

	"github.com/invopop/jsonschema"
	"github.com/varavelio/vdl/toolchain/internal/codegen/plugin"
	"github.com/varavelio/vdl/toolchain/internal/version"
)

func generatePluginSchemas() {
	generateSchema(
		&plugin.Input{},
		version.SchemaPluginInputID,
		"VDL Plugin Input Schema",
		"JSON Schema for the VDL Plugin Input Protocol",
		"plugin_input.schema.json",
	)

	generateSchema(
		&plugin.Output{},
		version.SchemaPluginOutputID,
		"VDL Plugin Output Schema",
		"JSON Schema for the VDL Plugin Output Protocol",
		"plugin_output.schema.json",
	)
}

func generateSchema(v interface{}, id, title, description, filename string) {
	schema := (&jsonschema.Reflector{}).Reflect(v)
	schema.ID = jsonschema.ID(id)
	schema.Title = title
	schema.Description = description

	data, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		log.Fatalf("failed to marshal schema for %s: %v", filename, err)
	}

	outPath := filepath.Join("internal", "codegen", "plugin", filename)
	if err := os.WriteFile(outPath, data, 0644); err != nil {
		log.Fatalf("failed to write schema to %s: %v", outPath, err)
	}
	log.Printf("Generated %s", outPath)
}

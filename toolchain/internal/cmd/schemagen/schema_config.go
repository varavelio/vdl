package main

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"

	"github.com/invopop/jsonschema"
	"github.com/varavelio/vdl/toolchain/internal/codegen/config"
	"github.com/varavelio/vdl/toolchain/internal/version"
)

func generateConfigSchema() {
	schema := (&jsonschema.Reflector{}).Reflect(&config.VDLConfig{})
	schema.ID = jsonschema.ID(version.SchemaConfigID)
	schema.Title = "VDL Config Schema"
	schema.Description = "JSON Schema for the VDL Config"

	data, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		log.Fatalf("failed to marshal schema: %v", err)
	}

	outPath := filepath.Join("internal", "codegen", "config", "config.schema.json")
	if err := os.WriteFile(outPath, data, 0644); err != nil {
		log.Fatalf("failed to write schema to %s: %v", outPath, err)
	}

	log.Printf("Generated %s", outPath)
}

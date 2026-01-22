package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/invopop/jsonschema"
	"github.com/varavelio/vdl/toolchain/internal/codegen/config"
	"github.com/varavelio/vdl/toolchain/internal/version"
)

func generateConfigSchema() {
	// schemaID is the canonical URL for the IR JSON Schema.
	schemaID := fmt.Sprintf("https://vdl.varavel.com/schemas/v%s/config.schema.json", version.VersionMajor)

	schema := (&jsonschema.Reflector{}).Reflect(&config.VDLConfig{})
	schema.ID = jsonschema.ID(schemaID)
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

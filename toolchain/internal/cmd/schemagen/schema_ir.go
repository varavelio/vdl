package main

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"

	"github.com/invopop/jsonschema"
	"github.com/varavelio/vdl/toolchain/internal/core/ir"
	"github.com/varavelio/vdl/toolchain/internal/version"
)

func generateIRSchema() {
	schema := (&jsonschema.Reflector{}).Reflect(&ir.Schema{})
	schema.ID = jsonschema.ID(version.SchemaIRID)
	schema.Title = "VDL IR Schema"
	schema.Description = "JSON Schema for the VDL Intermediate Representation"

	data, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		log.Fatalf("failed to marshal schema: %v", err)
	}

	outPath := filepath.Join("internal", "core", "ir", "ir.schema.json")
	if err := os.WriteFile(outPath, data, 0644); err != nil {
		log.Fatalf("failed to write schema to %s: %v", outPath, err)
	}

	log.Printf("Generated %s", outPath)
}

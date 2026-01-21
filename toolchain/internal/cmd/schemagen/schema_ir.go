package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/invopop/jsonschema"
	"github.com/varavelio/vdl/toolchain/internal/core/ir"
	"github.com/varavelio/vdl/toolchain/internal/version"
)

func generateIRSchema() {
	// schemaID is the canonical URL for the IR JSON Schema.
	schemaID := fmt.Sprintf("https://vdl.varavel.com/schemas/v%s/ir.schema.json", version.VersionMajor)

	r := &jsonschema.Reflector{}

	schema := r.Reflect(&ir.Schema{})
	schema.ID = jsonschema.ID(schemaID)
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

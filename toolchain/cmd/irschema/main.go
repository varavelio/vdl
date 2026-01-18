// Command irschema generates the JSON Schema for the IR package.
//
// This command generates ir.schema.json from the ir.Schema Go struct,
// ensuring the schema stays in sync with the Go types.
//
// Usage:
//
//	go run ./cmd/irschema
package main

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"

	"github.com/invopop/jsonschema"
	"github.com/varavelio/vdl/toolchain/internal/core/ir"
)

// SchemaID is the canonical URL for the IR JSON Schema.
const SchemaID = "https://raw.githubusercontent.com/varavelio/vdl/main/toolchain/internal/core/ir/ir.schema.json"

func main() {
	r := &jsonschema.Reflector{}

	schema := r.Reflect(&ir.Schema{})
	schema.ID = jsonschema.ID(SchemaID)
	schema.Title = "VDL IR Schema"
	schema.Description = "JSON Schema for the VDL Intermediate Representation"

	data, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		log.Fatalf("failed to marshal schema: %v", err)
	}

	// Output path relative to toolchain root
	outPath := filepath.Join("internal", "core", "ir", "ir.schema.json")

	if err := os.WriteFile(outPath, data, 0644); err != nil {
		log.Fatalf("failed to write schema to %s: %v", outPath, err)
	}

	log.Printf("Generated %s", outPath)
}

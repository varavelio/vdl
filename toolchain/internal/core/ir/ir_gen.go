//go:build ignore

// This program generates ir.schema.json from the ir.Schema struct.
// It is invoked by running `go generate` in this package.
package main

import (
	"encoding/json"
	"log"
	"os"

	"github.com/invopop/jsonschema"
	"github.com/varavelio/vdl/toolchain/internal/core/ir"
)

func main() {
	r := &jsonschema.Reflector{}

	schema := r.Reflect(&ir.Schema{})
	schema.ID = "https://raw.githubusercontent.com/varavelio/vdl/main/toolchain/internal/core/ir/ir.schema.json"
	schema.Title = "VDL IR Schema"
	schema.Description = "JSON Schema for the VDL Intermediate Representation"

	data, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		log.Fatalf("failed to marshal schema: %v", err)
	}

	if err := os.WriteFile("ir.schema.json", data, 0644); err != nil {
		log.Fatalf("failed to write schema: %v", err)
	}

	log.Println("generated ir.schema.json")
}

// Command genschemas generates the JSON Schema for the IR package.
//
// This command generates ir.schema.json from the ir.Schema Go struct,
// ensuring the schema stays in sync with the Go types.
//
// Usage:
//
//	go run ./cmd/genschemas
package main

func main() {
	generateIRSchema()
	generateConfigSchema()
	generatePluginSchemas()
}

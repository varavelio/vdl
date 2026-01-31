package typescript

import (
	"fmt"

	"github.com/varavelio/gen"
	"github.com/varavelio/vdl/toolchain/internal/codegen/config"
	"github.com/varavelio/vdl/toolchain/internal/core/ir"
)

// formatImportPath returns a properly formatted import path based on the ImportExtension config.
// For example, if ImportExtension is ".js", "./core" becomes "./core.js".
func formatImportPath(path string, cfg *config.TypeScriptConfig) string {
	if cfg == nil || cfg.ImportExtension == "" || cfg.ImportExtension == "none" {
		return path
	}
	return path + cfg.ImportExtension
}

// generateImport generates an import statement for a list of items.
// If the list is empty, nothing is generated.
// items can be simple names like "Response" or renamed ones like "Response as CoreResponse".
func generateImport(g *gen.Generator, items []string, from string, isType bool, cfg *config.TypeScriptConfig) {
	if len(items) == 0 {
		return
	}

	importKeyword := "import"
	if isType {
		importKeyword = "import type"
	}

	path := formatImportPath(from, cfg)

	if len(items) == 1 {
		g.Linef("%s { %s } from %q;", importKeyword, items[0], path)
		return
	}

	g.Linef("%s {", importKeyword)
	for _, item := range items {
		g.Linef("  %s,", item)
	}
	g.Linef("} from %q;", path)
}

// formatExportAll returns a formatted export statement.
// Example: formatExportAll("./core", cfg) -> `export * from "./core.js";`
func formatExportAll(path string, cfg *config.TypeScriptConfig) string {
	return fmt.Sprintf("export * from \"%s\";", formatImportPath(path, cfg))
}

// collectImports returns lists of type names and value names that should be imported
// from the types file.
func collectImports(schema *ir.Schema) (types []string, values []string) {
	typeSet := make(map[string]bool)
	valueSet := make(map[string]bool)

	// Domain Types (with hydrate and validate)
	for _, t := range schema.Types {
		typeSet[t.Name] = true
		valueSet["hydrate"+t.Name] = true
		valueSet["validate"+t.Name] = true
	}

	// Enums (with isFn and list)
	for _, e := range schema.Enums {
		typeSet[e.Name] = true
		valueSet["is"+e.Name] = true
		valueSet[e.Name+"List"] = true
	}

	// Procedure Types (Input, Output, Response, Hydrate, Validate)
	for _, proc := range schema.Procedures {
		fullName := proc.FullName()
		typeSet[fullName+"Input"] = true
		typeSet[fullName+"Output"] = true
		typeSet[fullName+"Response"] = true
		valueSet["hydrate"+fullName+"Output"] = true
		valueSet["validate"+fullName+"Input"] = true
	}

	// Stream Types (Input, Output, Response, Hydrate, Validate)
	for _, stream := range schema.Streams {
		fullName := stream.FullName()
		typeSet[fullName+"Input"] = true
		typeSet[fullName+"Output"] = true
		typeSet[fullName+"Response"] = true
		valueSet["hydrate"+fullName+"Output"] = true
		valueSet["validate"+fullName+"Input"] = true
	}

	// Convert maps to slices
	for name := range typeSet {
		types = append(types, name)
	}
	for name := range valueSet {
		values = append(values, name)
	}
	return types, values
}

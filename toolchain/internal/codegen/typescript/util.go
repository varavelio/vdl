package typescript

import (
	"fmt"

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

// formatImport returns a formatted import statement.
// Example: formatImport("{ Response }", "./core", cfg) -> `import { Response } from "./core.js";`
func formatImport(imports, path string, cfg *config.TypeScriptConfig) string {
	return fmt.Sprintf("import %s from \"%s\";", imports, formatImportPath(path, cfg))
}

// formatExport returns a formatted export statement.
// Example: formatExport("./core", cfg) -> `export * from "./core.js";`
func formatExport(path string, cfg *config.TypeScriptConfig) string {
	return fmt.Sprintf("export * from \"%s\";", formatImportPath(path, cfg))
}

// collectAllTypeNames returns a list of all type names that should be imported
// from the types file.
func collectAllTypeNames(schema *ir.Schema) []string {
	names := make(map[string]bool)

	// Domain Types (with hydrate and validate)
	for _, t := range schema.Types {
		names[t.Name] = true
		names["hydrate"+t.Name] = true
		names["validate"+t.Name] = true
	}

	// Enums (with isFn and list)
	for _, e := range schema.Enums {
		names[e.Name] = true
		names["is"+e.Name] = true
		names[e.Name+"List"] = true
	}

	// Procedure Types (Input, Output, Response, Hydrate, Validate)
	for _, proc := range schema.Procedures {
		fullName := proc.FullName()
		names[fullName+"Input"] = true
		names[fullName+"Output"] = true
		names[fullName+"Response"] = true
		names["hydrate"+fullName+"Output"] = true
		names["validate"+fullName+"Input"] = true
	}

	// Stream Types (Input, Output, Response, Hydrate, Validate)
	for _, stream := range schema.Streams {
		fullName := stream.FullName()
		names[fullName+"Input"] = true
		names[fullName+"Output"] = true
		names[fullName+"Response"] = true
		names["hydrate"+fullName+"Output"] = true
		names["validate"+fullName+"Input"] = true
	}

	// Convert map to slice
	var result []string
	for name := range names {
		result = append(result, name)
	}
	return result
}

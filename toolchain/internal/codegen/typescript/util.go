package typescript

import (
	"fmt"
	"strings"

	"github.com/varavelio/gen"
	"github.com/varavelio/vdl/toolchain/internal/codegen/config"
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

	g.Linef("%s { %s } from %q;", importKeyword, strings.Join(items, ", "), path)
}

func generateImportAll(g *gen.Generator, as string, from string, cfg *config.TypeScriptConfig) {
	path := formatImportPath(from, cfg)
	g.Linef("import * as %s from %q;", as, path)
}

// generateExportAll returns a formatted export statement.
// Example: generateExportAll("./core", cfg) -> `export * from "./core.js";`
func generateExportAll(path string, cfg *config.TypeScriptConfig) string {
	return fmt.Sprintf("export * from \"%s\";", formatImportPath(path, cfg))
}

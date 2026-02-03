package typescript

import (
	"fmt"
	"strings"

	"github.com/varavelio/gen"
	"github.com/varavelio/vdl/toolchain/internal/codegen/config"
	"github.com/varavelio/vdl/toolchain/internal/codegen/config/configtypes"
)

// formatImportPath returns a properly formatted import path based on the ImportExtension config.
// For example, if ImportExtension is ".js", "./core" becomes "./core.js".
func formatImportPath(path string, cfg *configtypes.TypeScriptTargetConfig) string {
	ext := config.GetImportExtension(cfg.ImportExtension)
	if ext == "" || ext == configtypes.TypescriptImportExtensionNone {
		return path
	}
	return path + string(ext)
}

// generateImport generates an import statement for a list of items.
// If the list is empty, nothing is generated.
// items can be simple names like "Response" or renamed ones like "Response as CoreResponse".
func generateImport(g *gen.Generator, items []string, from string, isType bool, cfg *configtypes.TypeScriptTargetConfig) {
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

func generateImportAll(g *gen.Generator, as string, from string, cfg *configtypes.TypeScriptTargetConfig) {
	path := formatImportPath(from, cfg)
	g.Linef("import * as %s from %q;", as, path)
}

// generateExportAll returns a formatted export statement.
// Example: generateExportAll("./core", cfg) -> `export * from "./core.js";`
func generateExportAll(path string, cfg *configtypes.TypeScriptTargetConfig) string {
	return fmt.Sprintf("export * from \"%s\";", formatImportPath(path, cfg))
}

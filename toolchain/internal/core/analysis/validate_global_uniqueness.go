package analysis

import (
	"fmt"

	"github.com/varavelio/vdl/toolchain/internal/core/ast"
)

// symbolOrigin tracks where a name was first declared for collision detection.
type symbolOrigin struct {
	kind   string
	file   string       // File where declared
	pos    ast.Position // Position of declaration
	endPos ast.Position
}

// validateGlobalUniqueness checks that names are unique across declarations.
func validateGlobalUniqueness(symbols *symbolTable) []Diagnostic {
	var diagnostics []Diagnostic

	// Map of name -> first declaration
	seen := make(map[string]symbolOrigin)

	// Register all types
	for name, sym := range symbols.types {
		if orig, exists := seen[name]; exists {
			diagnostics = append(diagnostics, newDiagnostic(
				sym.File,
				sym.Pos,
				sym.EndPos,
				CodeDuplicateName,
				formatNameCollision("type", name, orig.kind, orig.file, orig.pos),
			))
		} else {
			seen[name] = symbolOrigin{
				kind:   "type",
				file:   sym.File,
				pos:    sym.Pos,
				endPos: sym.EndPos,
			}
		}
	}

	// Register all enums
	for name, sym := range symbols.enums {
		if orig, exists := seen[name]; exists {
			diagnostics = append(diagnostics, newDiagnostic(
				sym.File,
				sym.Pos,
				sym.EndPos,
				CodeDuplicateName,
				formatNameCollision("enum", name, orig.kind, orig.file, orig.pos),
			))
		} else {
			seen[name] = symbolOrigin{
				kind:   "enum",
				file:   sym.File,
				pos:    sym.Pos,
				endPos: sym.EndPos,
			}
		}
	}

	// Register all constants
	for name, sym := range symbols.consts {
		if orig, exists := seen[name]; exists {
			diagnostics = append(diagnostics, newDiagnostic(
				sym.File,
				sym.Pos,
				sym.EndPos,
				CodeDuplicateName,
				formatNameCollision("constant", name, orig.kind, orig.file, orig.pos),
			))
		} else {
			seen[name] = symbolOrigin{
				kind:   "constant",
				file:   sym.File,
				pos:    sym.Pos,
				endPos: sym.EndPos,
			}
		}
	}

	return diagnostics
}

// formatNameCollision creates a descriptive error message for name collisions.
func formatNameCollision(newKind, name, origKind, origFile string, origPos ast.Position) string {
	return fmt.Sprintf("%s %q conflicts with %s declared at %s:%d:%d",
		newKind, name, origKind, origFile, origPos.Line, origPos.Column)
}

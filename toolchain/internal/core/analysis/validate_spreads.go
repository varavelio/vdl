package analysis

import (
	"fmt"

	"github.com/varavelio/vdl/urpc/internal/core/ast"
)

// validateSpreads checks that all spread references are valid:
// - The referenced type must exist
// - No field name conflicts between spread sources and local fields
// - No circular spread dependencies
func validateSpreads(symbols *symbolTable) []Diagnostic {
	var diagnostics []Diagnostic

	// Validate type spreads
	for _, typ := range symbols.types {
		diagnostics = append(diagnostics, validateTypeSpreads(symbols, typ)...)
	}

	// Validate RPC proc/stream input/output spreads
	for _, rpc := range symbols.rpcs {
		for _, proc := range rpc.Procs {
			if proc.Input != nil {
				diagnostics = append(diagnostics, validateBlockSpreads(symbols, proc.Input, proc.File, "procedure input", proc.Name)...)
			}
			if proc.Output != nil {
				diagnostics = append(diagnostics, validateBlockSpreads(symbols, proc.Output, proc.File, "procedure output", proc.Name)...)
			}
		}
		for _, stream := range rpc.Streams {
			if stream.Input != nil {
				diagnostics = append(diagnostics, validateBlockSpreads(symbols, stream.Input, stream.File, "stream input", stream.Name)...)
			}
			if stream.Output != nil {
				diagnostics = append(diagnostics, validateBlockSpreads(symbols, stream.Output, stream.File, "stream output", stream.Name)...)
			}
		}
	}

	// Check for circular spread dependencies
	diagnostics = append(diagnostics, validateSpreadCycles(symbols)...)

	return diagnostics
}

// validateTypeSpreads validates spreads in a type declaration.
func validateTypeSpreads(symbols *symbolTable, typ *TypeSymbol) []Diagnostic {
	var diagnostics []Diagnostic

	// Collect all field names from direct fields
	fieldNames := make(map[string]FieldOrigin)
	for _, field := range typ.Fields {
		fieldNames[field.Name] = FieldOrigin{
			File:   field.File,
			Pos:    field.Pos,
			Source: "direct field",
		}
	}

	// Check each spread
	for _, spread := range typ.Spreads {
		// Check if the referenced type exists
		refType := symbols.lookupType(spread.TypeName)
		if refType == nil {
			diagnostics = append(diagnostics, newDiagnostic(
				typ.File,
				spread.Pos,
				spread.EndPos,
				CodeSpreadTypeNotFound,
				fmt.Sprintf("spread references undefined type %q", spread.TypeName),
			))
			continue
		}

		// Check for field name conflicts
		for _, spreadField := range refType.Fields {
			if existing, ok := fieldNames[spreadField.Name]; ok {
				diagnostics = append(diagnostics, newDiagnostic(
					typ.File,
					spread.Pos,
					spread.EndPos,
					CodeSpreadFieldConflict,
					fmt.Sprintf("field %q from spread %q conflicts with %s at %s:%d:%d",
						spreadField.Name, spread.TypeName, existing.Source,
						existing.File, existing.Pos.Line, existing.Pos.Column),
				))
			} else {
				fieldNames[spreadField.Name] = FieldOrigin{
					File:   refType.File,
					Pos:    spreadField.Pos,
					Source: fmt.Sprintf("spread ...%s", spread.TypeName),
				}
			}
		}
	}

	return diagnostics
}

// validateBlockSpreads validates spreads in an input/output block.
// parentFile is the file where the parent proc/stream is declared.
func validateBlockSpreads(symbols *symbolTable, block *BlockSymbol, parentFile, context, parentName string) []Diagnostic {
	var diagnostics []Diagnostic

	// Collect all field names from direct fields
	fieldNames := make(map[string]FieldOrigin)
	for _, field := range block.Fields {
		fieldNames[field.Name] = FieldOrigin{
			File:   field.File,
			Pos:    field.Pos,
			Source: "direct field",
		}
	}

	// Check each spread
	for _, spread := range block.Spreads {
		// Check if the referenced type exists
		refType := symbols.lookupType(spread.TypeName)
		if refType == nil {
			diagnostics = append(diagnostics, newDiagnostic(
				parentFile,
				spread.Pos,
				spread.EndPos,
				CodeSpreadTypeNotFound,
				fmt.Sprintf("spread in %s %q references undefined type %q",
					context, parentName, spread.TypeName),
			))
			continue
		}

		// Check for field name conflicts
		for _, spreadField := range refType.Fields {
			if existing, ok := fieldNames[spreadField.Name]; ok {
				diagnostics = append(diagnostics, newDiagnostic(
					parentFile,
					spread.Pos,
					spread.EndPos,
					CodeSpreadFieldConflict,
					fmt.Sprintf("field %q from spread %q in %s %q conflicts with %s",
						spreadField.Name, spread.TypeName, context, parentName, existing.Source),
				))
			} else {
				fieldNames[spreadField.Name] = FieldOrigin{
					File:   refType.File,
					Pos:    spreadField.Pos,
					Source: fmt.Sprintf("spread ...%s", spread.TypeName),
				}
			}
		}
	}

	return diagnostics
}

// FieldOrigin tracks where a field name originated from.
type FieldOrigin struct {
	File   string
	Pos    Position
	Source string
}

// Position is a simple position for tracking.
type Position = ast.Position

// validateSpreadCycles detects circular spread dependencies.
func validateSpreadCycles(symbols *symbolTable) []Diagnostic {
	var diagnostics []Diagnostic

	// Build spread dependency graph
	for typeName, typ := range symbols.types {
		if cycle := findSpreadCycle(symbols, typeName, []string{}); cycle != nil {
			diagnostics = append(diagnostics, newDiagnostic(
				typ.File,
				typ.Pos,
				typ.EndPos,
				CodeSpreadCycle,
				fmt.Sprintf("circular spread dependency detected: %s", formatCycle(cycle)),
			))
		}
	}

	return diagnostics
}

// findSpreadCycle detects if there's a cycle in spread dependencies.
func findSpreadCycle(symbols *symbolTable, typeName string, visited []string) []string {
	// Check if we've already visited this type
	for _, v := range visited {
		if v == typeName {
			return append(visited, typeName)
		}
	}

	typ := symbols.lookupType(typeName)
	if typ == nil {
		return nil
	}

	newVisited := append(visited, typeName)

	// Check all spreads
	for _, spread := range typ.Spreads {
		if cycle := findSpreadCycle(symbols, spread.TypeName, newVisited); cycle != nil {
			return cycle
		}
	}

	return nil
}

// formatCycle formats a cycle path for error messages.
func formatCycle(cycle []string) string {
	result := ""
	for i, name := range cycle {
		if i > 0 {
			result += " -> "
		}
		result += name
	}
	return result
}

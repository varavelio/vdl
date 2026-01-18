package analysis

import (
	"fmt"
	"slices"
)

// validateCycles detects circular type dependencies.
// A circular dependency occurs when type A references type B and type B references type A
// (either directly or through a chain of other types).
// All circular references are forbidden, including through arrays, maps, and optional fields.
func validateCycles(symbols *symbolTable) []Diagnostic {
	var diagnostics []Diagnostic

	// Check each type for cycles
	for typeName, typ := range symbols.types {
		if cycle := findTypeCycle(symbols, typeName, []string{}); cycle != nil {
			diagnostics = append(diagnostics, newDiagnostic(
				typ.File,
				typ.Pos,
				typ.EndPos,
				CodeCircularTypeDependency,
				fmt.Sprintf("circular type dependency detected: %s", formatCyclePath(cycle)),
			))
		}
	}

	return diagnostics
}

// findTypeCycle performs DFS to detect cycles in type dependencies.
// Returns the cycle path if found, nil otherwise.
func findTypeCycle(symbols *symbolTable, typeName string, visited []string) []string {
	// Check if we've already visited this type in the current path
	if slices.Contains(visited, typeName) {
		return append(visited, typeName)
	}

	typ := symbols.lookupType(typeName)
	if typ == nil {
		return nil
	}

	newVisited := append(visited, typeName)

	// Check all fields for type references
	for _, field := range typ.Fields {
		if cycle := findFieldTypeCycle(symbols, field.Type, newVisited); cycle != nil {
			return cycle
		}
	}

	return nil
}

// findFieldTypeCycle checks a field type for cycles.
// All type references are checked, including arrays and maps.
func findFieldTypeCycle(symbols *symbolTable, typeInfo *FieldTypeInfo, visited []string) []string {
	if typeInfo == nil {
		return nil
	}

	switch typeInfo.Kind {
	case FieldTypeKindPrimitive:
		return nil

	case FieldTypeKindCustom:
		return findTypeCycle(symbols, typeInfo.Name, visited)

	case FieldTypeKindMap:
		// Check the map value type for cycles
		if typeInfo.MapValue != nil {
			return findFieldTypeCycle(symbols, typeInfo.MapValue, visited)
		}

	case FieldTypeKindObject:
		// Check inline object fields
		if typeInfo.ObjectDef != nil {
			for _, field := range typeInfo.ObjectDef.Fields {
				if cycle := findFieldTypeCycle(symbols, field.Type, visited); cycle != nil {
					return cycle
				}
			}
		}
	}

	return nil
}

// formatCyclePath formats a cycle for error messages.
func formatCyclePath(cycle []string) string {
	result := ""
	for i, name := range cycle {
		if i > 0 {
			result += " -> "
		}
		result += name
	}
	return result
}

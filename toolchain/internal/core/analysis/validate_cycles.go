package analysis

import (
	"fmt"
	"slices"
	"strings"
)

// validateCycles detects circular type dependencies.
// A circular dependency occurs when type A references type B and type B references type A
// (either directly or through a chain of other types).
//
// Circular references are allowed if at least one field in the cycle is optional.
// This allows recursive types like: type Node { children?: []Node }
func validateCycles(symbols *symbolTable) []Diagnostic {
	var diagnostics []Diagnostic

	// Check each type for cycles
	for typeName, typ := range symbols.types {
		if cycle := findTypeCycle(symbols, typeName, []string{}, false); cycle != nil {
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
// The hasOptional parameter tracks whether we've encountered an optional field in the current path.
func findTypeCycle(symbols *symbolTable, typeName string, visited []string, hasOptional bool) []string {
	// Check if we've already visited this type in the current path
	if slices.Contains(visited, typeName) {
		// Cycle detected, but only report it if no optional field was found
		if hasOptional {
			return nil
		}
		return append(visited, typeName)
	}

	typ := symbols.lookupType(typeName)
	if typ == nil {
		return nil
	}

	newVisited := append(visited, typeName)

	// Check all fields for type references
	for _, field := range typ.Fields {
		fieldOptional := hasOptional || field.Optional
		if cycle := findFieldTypeCycle(symbols, field.Type, newVisited, fieldOptional); cycle != nil {
			return cycle
		}
	}

	return nil
}

// findFieldTypeCycle checks a field type for cycles.
// All type references are checked, including arrays and maps.
// The hasOptional parameter tracks whether the current field or any ancestor is optional.
func findFieldTypeCycle(symbols *symbolTable, typeInfo *FieldTypeInfo, visited []string, hasOptional bool) []string {
	if typeInfo == nil {
		return nil
	}

	switch typeInfo.Kind {
	case FieldTypeKindPrimitive:
		return nil

	case FieldTypeKindCustom:
		return findTypeCycle(symbols, typeInfo.Name, visited, hasOptional)

	case FieldTypeKindMap:
		// Check the map value type for cycles
		if typeInfo.MapValue != nil {
			return findFieldTypeCycle(symbols, typeInfo.MapValue, visited, hasOptional)
		}

	case FieldTypeKindObject:
		// Check inline object fields
		if typeInfo.ObjectDef != nil {
			for _, field := range typeInfo.ObjectDef.Fields {
				fieldOptional := hasOptional || field.Optional
				if cycle := findFieldTypeCycle(symbols, field.Type, visited, fieldOptional); cycle != nil {
					return cycle
				}
			}
		}
	}

	return nil
}

// formatCyclePath formats a cycle for error messages.
func formatCyclePath(cycle []string) string {
	var result strings.Builder
	for i, name := range cycle {
		if i > 0 {
			result.WriteString(" -> ")
		}
		result.WriteString(name)
	}
	return result.String()
}

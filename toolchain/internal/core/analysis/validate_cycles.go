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
// Spread edges are followed too: `type A { ...B }` copies B's fields into A,
// so a required field cycle through spreads (A spreads B, B requires A) is
// just as infinite as a direct one.
//
// Circular references are allowed if at least one field in the cycle is optional.
// This allows recursive types like: type Node { children?: []Node }.
//
// Pure spread cycles (no field edge, e.g. A spreads B spreads A) are not
// reported here: their field sets have a finite least fixed point and the
// spread validator owns them (E204).
func validateCycles(symbols *symbolTable) []Diagnostic {
	var diagnostics []Diagnostic

	// Check each type for cycles
	for typeName, typ := range symbols.types {
		if cycle := findTypeCycle(
			symbols,
			typeName,
			[]string{},
			false,
			false,
			false,
		); cycle != nil {
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
// sawSpread/sawField describe the path so far: a cycle reached only through
// spreads (no field edge) is a pure spread cycle owned by the spread
// validator (E204), never E601.
func findTypeCycle(
	symbols *symbolTable,
	typeName string,
	visited []string,
	hasOptional bool,
	sawSpread bool,
	sawField bool,
) []string {
	// Check if we've already visited this type in the current path
	if slices.Contains(visited, typeName) {
		// Cycle detected, but only report it if no optional field was found
		if hasOptional {
			return nil
		}
		if sawSpread && !sawField {
			return nil
		}
		return append(visited, typeName)
	}

	typ := symbols.lookupType(typeName)
	if typ == nil {
		return nil
	}

	newVisited := append(visited, typeName)

	// Follow the type's reference for cycles
	if typ.Type == nil {
		return nil
	}

	return findFieldTypeCycle(symbols, typ.Type, newVisited, hasOptional, sawSpread, sawField)
}

// findFieldTypeCycle checks a field type for cycles.
// All type references are checked, including arrays and maps.
// The hasOptional parameter tracks whether the current field or any ancestor is optional.
// Spread edges are followed through ResolveObjectType (aliases included): a
// spread copies fields, so the search continues into the resolved type with
// each copied field keeping its own optionality.
func findFieldTypeCycle(
	symbols *symbolTable,
	typeInfo *FieldTypeInfo,
	visited []string,
	hasOptional bool,
	sawSpread bool,
	sawField bool,
) []string {
	if typeInfo == nil {
		return nil
	}

	switch typeInfo.Kind {
	case FieldTypeKindPrimitive:
		return nil

	case FieldTypeKindCustom:
		return findTypeCycle(symbols, typeInfo.Name, visited, hasOptional, sawSpread, sawField)

	case FieldTypeKindMap:
		// Check the map value type for cycles
		if typeInfo.MapValue != nil {
			return findFieldTypeCycle(
				symbols,
				typeInfo.MapValue,
				visited,
				hasOptional,
				sawSpread,
				sawField,
			)
		}

	case FieldTypeKindObject:
		// Check inline object fields
		if typeInfo.ObjectDef != nil {
			for _, field := range typeInfo.ObjectDef.Fields {
				fieldOptional := hasOptional || field.Optional
				if cycle := findFieldTypeCycle(
					symbols,
					field.Type,
					visited,
					fieldOptional,
					sawSpread,
					true,
				); cycle != nil {
					return cycle
				}
			}
			for _, spread := range typeInfo.ObjectDef.Spreads {
				if spread == nil || spread.Member != nil {
					continue
				}
				resolved := ResolveObjectType(spread.Name, symbols.lookupType)
				if resolved == nil {
					continue
				}
				if cycle := findTypeCycle(
					symbols,
					resolved.Name,
					visited,
					hasOptional,
					true,
					sawField,
				); cycle != nil {
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

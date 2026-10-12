package analysis

import (
	"fmt"

	"github.com/varavelio/vdl/toolchain/internal/core/ast"
)

// validateStructure validates structural requirements:
//
// - Field names are unique within a type/block.
// - Object literal keys are unique within each object literal.
func validateStructure(symbols *symbolTable) []Diagnostic {
	var diagnostics []Diagnostic

	// Validate type field uniqueness. Each scope is visited once: nested
	// scopes (inline objects, including those inside maps and arrays) are
	// reached through the recursion in validateFieldUniqueness.
	for _, typ := range symbols.types {
		if typ.IsObject() && typ.Type.ObjectDef != nil {
			diagnostics = append(
				diagnostics,
				validateFieldUniqueness(
					typ.Type.ObjectDef.Fields, "type", typ.Name, typ.File,
				)...,
			)
		}
	}

	for _, cnst := range symbols.consts {
		if cnst.AST != nil && cnst.AST.Value != nil {
			diagnostics = append(
				diagnostics,
				validateObjectLiteralUniqueness(
					cnst.AST.Value, cnst.File,
				)...,
			)
		}
	}

	return diagnostics
}

// validateFieldUniqueness validates that all field names are unique within a scope.
func validateFieldUniqueness(fields []*FieldSymbol, context, parentName, file string) []Diagnostic {
	var diagnostics []Diagnostic

	fieldNames := make(map[string]ast.Position)
	for _, field := range fields {
		if existing, ok := fieldNames[field.Name]; ok {
			diagnostics = append(diagnostics, newDiagnostic(
				file,
				field.Pos,
				field.EndPos,
				CodeDuplicateField,
				fmt.Sprintf("field %q in %s %q is already declared at line %d",
					field.Name, context, parentName, existing.Line),
			))
		} else {
			fieldNames[field.Name] = field.Pos
		}

		// Recursively check nested scopes: inline objects directly or
		// nested inside maps and arrays.
		diagnostics = append(
			diagnostics,
			validateNestedFieldUniqueness(field.Type, field.Name, file)...,
		)
	}

	return diagnostics
}

// validateNestedFieldUniqueness descends into map and array element types to
// find inline object scopes. Arrays share their element type, so dimensions
// need no special handling.
func validateNestedFieldUniqueness(
	typeInfo *FieldTypeInfo,
	owner, file string,
) []Diagnostic {
	if typeInfo == nil {
		return nil
	}
	switch typeInfo.Kind {
	case FieldTypeKindMap:
		return validateNestedFieldUniqueness(typeInfo.MapValue, owner, file)
	case FieldTypeKindObject:
		if typeInfo.ObjectDef == nil {
			return nil
		}
		return validateFieldUniqueness(
			typeInfo.ObjectDef.Fields,
			"inline object in",
			owner,
			file,
		)
	}
	return nil
}

func validateObjectLiteralUniqueness(lit *ast.DataLiteral, file string) []Diagnostic {
	if lit == nil {
		return nil
	}

	var diagnostics []Diagnostic
	if lit.Object != nil {
		seen := map[string]ast.Position{}
		for _, entry := range lit.Object.Entries {
			if entry.Spread != nil {
				diagnostics = append(
					diagnostics,
					validateObjectLiteralUniqueness(
						entry.Value, file,
					)...,
				)
				continue
			}
			if existing, ok := seen[entry.Key]; ok {
				diagnostics = append(diagnostics, newDiagnostic(
					file,
					entry.Pos,
					entry.EndPos,
					CodeDuplicateField,
					fmt.Sprintf(
						"object key %q is already declared at line %d",
						entry.Key,
						existing.Line,
					),
				))
			} else {
				seen[entry.Key] = entry.Pos
			}
			diagnostics = append(diagnostics, validateObjectLiteralUniqueness(entry.Value, file)...)
		}
	}

	if lit.Array != nil {
		for _, elem := range lit.Array.Elements {
			diagnostics = append(diagnostics, validateObjectLiteralUniqueness(elem, file)...)
		}
	}

	return diagnostics
}

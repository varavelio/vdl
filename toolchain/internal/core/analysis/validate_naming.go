package analysis

import (
	"fmt"
	"slices"
	"unicode"
)

// validateNaming checks that all identifiers follow the naming conventions:
// - Types and Enums: PascalCase
// - Fields: camelCase
// - Constants: camelCase
// - Enum members: PascalCase
// - Annotations: camelCase
func validateNaming(symbols *symbolTable) []Diagnostic {
	var diagnostics []Diagnostic

	// Validate type names and their fields
	for _, typ := range symbols.types {
		if !isPascalCase(typ.Name) {
			diagnostics = append(diagnostics, newDiagnostic(
				typ.File,
				typ.Pos,
				typ.EndPos,
				CodeNotPascalCase,
				fmt.Sprintf("type name %q must be in PascalCase", typ.Name),
			))
		}
		diagnostics = append(diagnostics, validateFieldNaming(typ.Fields, "type", typ.Name)...)
		diagnostics = append(diagnostics, validateAnnotationNaming(typ.Annotations, typ.File, "type", typ.Name)...)
	}

	// Validate enum names and member names
	for _, enum := range symbols.enums {
		if !isPascalCase(enum.Name) {
			diagnostics = append(diagnostics, newDiagnostic(
				enum.File,
				enum.Pos,
				enum.EndPos,
				CodeNotPascalCase,
				fmt.Sprintf("enum name %q must be in PascalCase", enum.Name),
			))
		}
		for _, member := range enum.Members {
			if !isPascalCase(member.Name) {
				diagnostics = append(diagnostics, newDiagnostic(
					enum.File,
					member.Pos,
					member.EndPos,
					CodeEnumMemberNotPascal,
					fmt.Sprintf("enum member %q in enum %q must be in PascalCase", member.Name, enum.Name),
				))
			}
			diagnostics = append(diagnostics, validateAnnotationNaming(member.Annotations, enum.File, "enum member", member.Name)...)
		}
		diagnostics = append(diagnostics, validateAnnotationNaming(enum.Annotations, enum.File, "enum", enum.Name)...)
	}

	// Validate constant names
	for _, cnst := range symbols.consts {
		if !isCamelCase(cnst.Name) {
			diagnostics = append(diagnostics, newDiagnostic(
				cnst.File,
				cnst.Pos,
				cnst.EndPos,
				CodeNotCamelCase,
				fmt.Sprintf("constant name %q must be in camelCase", cnst.Name),
			))
		}
		diagnostics = append(diagnostics, validateAnnotationNaming(cnst.Annotations, cnst.File, "constant", cnst.Name)...)
	}

	return diagnostics
}

// validateFieldNaming validates that all fields in a list follow camelCase naming.
func validateFieldNaming(fields []*FieldSymbol, context, parentName string) []Diagnostic {
	var diagnostics []Diagnostic
	for _, field := range fields {
		if !isCamelCase(field.Name) {
			diagnostics = append(diagnostics, newDiagnostic(
				field.File,
				field.Pos,
				field.EndPos,
				CodeNotCamelCase,
				fmt.Sprintf("field %q in %s %q must be in camelCase", field.Name, context, parentName),
			))
		}
		// Recursively check inline object fields
		if field.Type != nil && field.Type.Kind == FieldTypeKindObject && field.Type.ObjectDef != nil {
			diagnostics = append(diagnostics, validateFieldNaming(field.Type.ObjectDef.Fields, "inline object in", field.Name)...)
		}
		diagnostics = append(diagnostics, validateAnnotationNaming(field.Annotations, field.File, "field", field.Name)...)
	}
	return diagnostics
}

func validateAnnotationNaming(annotations []*AnnotationRef, file, context, parentName string) []Diagnostic {
	var diagnostics []Diagnostic
	for _, ann := range annotations {
		if !isCamelCase(ann.Name) {
			diagnostics = append(diagnostics, newDiagnostic(
				file,
				ann.Pos,
				ann.EndPos,
				CodeNotCamelCase,
				fmt.Sprintf("annotation %q in %s %q must be in camelCase", ann.Name, context, parentName),
			))
		}
	}
	return diagnostics
}

// isPascalCase checks if a string follows PascalCase convention.
// PascalCase starts with an uppercase letter and has no underscores.
func isPascalCase(s string) bool {
	if len(s) == 0 {
		return false
	}
	runes := []rune(s)

	// First character must be uppercase
	if !unicode.IsUpper(runes[0]) {
		return false
	}

	// No underscores allowed
	return !slices.Contains(runes, '_')
}

// isCamelCase checks if a string follows camelCase convention.
// camelCase starts with a lowercase letter and has no underscores.
func isCamelCase(s string) bool {
	if len(s) == 0 {
		return false
	}
	runes := []rune(s)

	// First character must be lowercase
	if !unicode.IsLower(runes[0]) {
		return false
	}

	// No underscores allowed
	return !slices.Contains(runes, '_')
}

package analysis

import (
	"fmt"
	"unicode"

	"github.com/uforg/uforpc/urpc/internal/core/ast"
)

// validateNaming checks that all identifiers follow the naming conventions:
// - Types, Enums, RPCs, Procs, Streams, Patterns: PascalCase
// - Fields: camelCase
// - Constants: UPPER_SNAKE_CASE
// - Enum members: PascalCase
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
		}
	}

	// Validate constant names
	for _, cnst := range symbols.consts {
		if !isUpperSnakeCase(cnst.Name) {
			diagnostics = append(diagnostics, newDiagnostic(
				cnst.File,
				cnst.Pos,
				cnst.EndPos,
				CodeNotUpperSnakeCase,
				fmt.Sprintf("constant name %q must be in UPPER_SNAKE_CASE", cnst.Name),
			))
		}
	}

	// Validate pattern names
	for _, pattern := range symbols.patterns {
		if !isPascalCase(pattern.Name) {
			diagnostics = append(diagnostics, newDiagnostic(
				pattern.File,
				pattern.Pos,
				pattern.EndPos,
				CodeNotPascalCase,
				fmt.Sprintf("pattern name %q must be in PascalCase", pattern.Name),
			))
		}
	}

	// Validate RPC, proc, and stream names
	for _, rpc := range symbols.rpcs {
		if !isPascalCase(rpc.Name) {
			diagnostics = append(diagnostics, newDiagnostic(
				rpc.File,
				rpc.Pos,
				rpc.EndPos,
				CodeNotPascalCase,
				fmt.Sprintf("RPC name %q must be in PascalCase", rpc.Name),
			))
		}

		for _, proc := range rpc.Procs {
			if !isPascalCase(proc.Name) {
				diagnostics = append(diagnostics, newDiagnostic(
					proc.File,
					proc.Pos,
					proc.EndPos,
					CodeNotPascalCase,
					fmt.Sprintf("procedure name %q in RPC %q must be in PascalCase", proc.Name, rpc.Name),
				))
			}
			// Validate input/output fields
			if proc.Input != nil {
				diagnostics = append(diagnostics, validateFieldNaming(proc.Input.Fields, "procedure input", proc.Name)...)
			}
			if proc.Output != nil {
				diagnostics = append(diagnostics, validateFieldNaming(proc.Output.Fields, "procedure output", proc.Name)...)
			}
		}

		for _, stream := range rpc.Streams {
			if !isPascalCase(stream.Name) {
				diagnostics = append(diagnostics, newDiagnostic(
					stream.File,
					stream.Pos,
					stream.EndPos,
					CodeNotPascalCase,
					fmt.Sprintf("stream name %q in RPC %q must be in PascalCase", stream.Name, rpc.Name),
				))
			}
			// Validate input/output fields
			if stream.Input != nil {
				diagnostics = append(diagnostics, validateFieldNaming(stream.Input.Fields, "stream input", stream.Name)...)
			}
			if stream.Output != nil {
				diagnostics = append(diagnostics, validateFieldNaming(stream.Output.Fields, "stream output", stream.Name)...)
			}
		}
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
	for _, r := range runes {
		if r == '_' {
			return false
		}
	}

	return true
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
	for _, r := range runes {
		if r == '_' {
			return false
		}
	}

	return true
}

// isUpperSnakeCase checks if a string follows UPPER_SNAKE_CASE convention.
// All letters must be uppercase, words separated by underscores.
func isUpperSnakeCase(s string) bool {
	if len(s) == 0 {
		return false
	}
	runes := []rune(s)

	// First character must be uppercase letter
	if !unicode.IsUpper(runes[0]) {
		return false
	}

	// All letters must be uppercase, only underscores and digits allowed otherwise
	for i, r := range runes {
		if unicode.IsLetter(r) {
			if !unicode.IsUpper(r) {
				return false
			}
		} else if r == '_' {
			// Underscore cannot be first, last, or consecutive
			if i == 0 || i == len(runes)-1 {
				return false
			}
			if runes[i-1] == '_' {
				return false
			}
		} else if !unicode.IsDigit(r) {
			return false
		}
	}

	return true
}

// extractFieldsFromAST extracts field names from AST for validation.
func extractFieldsFromAST(children []*ast.TypeDeclChild) []string {
	var names []string
	for _, child := range children {
		if child.Field != nil {
			names = append(names, child.Field.Name)
		}
	}
	return names
}

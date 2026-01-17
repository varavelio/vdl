package analysis

import (
	"fmt"

	"github.com/uforg/uforpc/urpc/internal/core/ast"
)

// validateEnums validates all enum declarations:
// - All members must have consistent types (all string or all int)
// - For int enums, all members must have explicit values
// - All member names must be unique
// - All member values must be unique
func validateEnums(symbols *symbolTable) []Diagnostic {
	var diagnostics []Diagnostic

	for _, enum := range symbols.enums {
		diagnostics = append(diagnostics, validateEnum(enum)...)
	}

	return diagnostics
}

// validateEnum validates a single enum declaration.
func validateEnum(enum *EnumSymbol) []Diagnostic {
	var diagnostics []Diagnostic

	if len(enum.Members) == 0 {
		return diagnostics
	}

	// Check for duplicate member names
	memberNames := make(map[string]ast.Position)
	for _, member := range enum.Members {
		if existing, ok := memberNames[member.Name]; ok {
			diagnostics = append(diagnostics, newDiagnostic(
				enum.File,
				member.Pos,
				member.EndPos,
				CodeEnumDuplicateName,
				fmt.Sprintf("enum member %q in enum %q is already declared at line %d",
					member.Name, enum.Name, existing.Line),
			))
		} else {
			memberNames[member.Name] = member.Pos
		}
	}

	// Check for duplicate values
	memberValues := make(map[string]string) // value -> member name
	for _, member := range enum.Members {
		if existingName, ok := memberValues[member.Value]; ok {
			diagnostics = append(diagnostics, newDiagnostic(
				enum.File,
				member.Pos,
				member.EndPos,
				CodeEnumDuplicateValue,
				fmt.Sprintf("enum member %q has the same value %q as member %q",
					member.Name, member.Value, existingName),
			))
		} else {
			memberValues[member.Value] = member.Name
		}
	}

	// For int enums, all members must have explicit values
	if enum.ValueType == EnumValueTypeInt {
		for _, member := range enum.Members {
			if !member.HasExplicit {
				diagnostics = append(diagnostics, newDiagnostic(
					enum.File,
					member.Pos,
					member.EndPos,
					CodeEnumIntNeedsValues,
					fmt.Sprintf("int enum %q requires explicit values for all members, but %q has none",
						enum.Name, member.Name),
				))
			}
		}
	}

	return diagnostics
}

// buildEnumSymbol creates an EnumSymbol from an AST EnumDecl.
func buildEnumSymbol(decl *ast.EnumDecl, file string) *EnumSymbol {
	var docstring *string
	if decl.Docstring != nil {
		s := string(decl.Docstring.Value)
		docstring = &s
	}

	var deprecated *DeprecationInfo
	if decl.Deprecated != nil {
		msg := ""
		if decl.Deprecated.Message != nil {
			msg = string(*decl.Deprecated.Message)
		}
		deprecated = &DeprecationInfo{Message: msg}
	}

	enum := &EnumSymbol{
		Symbol: Symbol{
			Name:       decl.Name,
			File:       file,
			Pos:        decl.Pos,
			EndPos:     decl.EndPos,
			Docstring:  docstring,
			Deprecated: deprecated,
		},
		AST:       decl,
		ValueType: EnumValueTypeString, // Default to string
		Members:   []*EnumMemberSymbol{},
	}

	// Determine enum type and build members
	for i, member := range decl.Members {
		// Skip comments
		if member.Name == "" {
			continue
		}

		memberSym := &EnumMemberSymbol{
			Name:   member.Name,
			Pos:    member.Pos,
			EndPos: member.EndPos,
		}

		if member.Value != nil {
			memberSym.HasExplicit = true
			if member.Value.Int != nil {
				memberSym.Value = *member.Value.Int
				// First member with int value determines enum type
				if i == 0 || (enum.ValueType == EnumValueTypeString && len(enum.Members) == 0) {
					enum.ValueType = EnumValueTypeInt
				}
			} else if member.Value.Str != nil {
				memberSym.Value = string(*member.Value.Str)
			}
		} else {
			// No explicit value: use member name as value (string enum)
			memberSym.Value = member.Name
			memberSym.HasExplicit = false
		}

		enum.Members = append(enum.Members, memberSym)
	}

	// Infer type from first member if needed
	if len(enum.Members) > 0 && enum.Members[0].HasExplicit {
		// Check if it's an int by trying to see if the AST had an int value
		if decl.Members[0].Value != nil && decl.Members[0].Value.Int != nil {
			enum.ValueType = EnumValueTypeInt
		}
	}

	return enum
}

package analysis

import (
	"fmt"
	"slices"
	"strconv"

	"github.com/varavelio/vdl/toolchain/internal/core/ast"
	"github.com/varavelio/vdl/toolchain/internal/util/strutil"
)

// validateEnums validates all enum declarations:
// - All members must have consistent types (all string or all int)
// - For int enums, all members must have explicit values
// - All member names must be unique
// - All member values must be unique
func validateEnums(symbols *symbolTable) []Diagnostic {
	var diagnostics []Diagnostic

	for _, enum := range symbols.enums {
		diagnostics = append(diagnostics, validateEnum(symbols, enum)...)
	}

	return diagnostics
}

// validateEnum validates a single enum declaration.
func validateEnum(symbols *symbolTable, enum *EnumSymbol) []Diagnostic {
	var diagnostics []Diagnostic

	effectiveMembers, valueType, enumDiags := expandEnumMembers(symbols, enum)
	diagnostics = append(diagnostics, enumDiags...)

	if len(effectiveMembers) == 0 {
		return diagnostics
	}
	enum.ValueType = valueType

	// Check for duplicate member names
	memberNames := make(map[string]ast.Position)
	for _, member := range effectiveMembers {
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
	for _, member := range effectiveMembers {
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
		for _, member := range effectiveMembers {
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

	enum := &EnumSymbol{
		Symbol: Symbol{
			Name:        decl.Name,
			File:        file,
			Pos:         decl.Pos,
			EndPos:      decl.EndPos,
			Docstring:   docstring,
			Annotations: buildAnnotationRefs(decl.Annotations),
		},
		AST:       decl,
		ValueType: EnumValueTypeString, // Default to string
		Members:   []*EnumMemberSymbol{},
		Spreads:   []*SpreadRef{},
	}

	for _, member := range decl.Members {
		if member.Spread != nil {
			enum.Spreads = append(enum.Spreads, &SpreadRef{
				Name:   member.Spread.Ref.Name,
				Member: member.Spread.Ref.Member,
				Pos:    member.Spread.Pos,
				EndPos: member.Spread.EndPos,
			})
			continue
		}

		if member.Name == "" {
			continue
		}

		memberSym := &EnumMemberSymbol{
			Symbol: Symbol{
				Name:        member.Name,
				File:        file,
				Pos:         member.Pos,
				EndPos:      member.EndPos,
				Annotations: buildAnnotationRefs(member.Annotations),
			},
		}

		if member.Docstring != nil {
			d := string(member.Docstring.Value)
			memberSym.Docstring = &d
		}

		if member.Value != nil {
			memberSym.HasExplicit = true
			if member.Value.Int != nil {
				memberSym.Value = *member.Value.Int
			} else if member.Value.Str != nil {
				memberSym.Value = string(*member.Value.Str)
			}
		} else {
			memberSym.Value = member.Name
			memberSym.HasExplicit = false
		}

		enum.Members = append(enum.Members, memberSym)
	}

	return enum
}

func expandEnumMembers(symbols *symbolTable, enum *EnumSymbol) ([]*EnumMemberSymbol, EnumValueType, []Diagnostic) {
	type frame struct {
		enum *EnumSymbol
	}

	var diagnostics []Diagnostic
	stack := []string{}
	seen := map[string]bool{}

	var expand func(current *EnumSymbol) []*EnumMemberSymbol
	expand = func(current *EnumSymbol) []*EnumMemberSymbol {
		if seen[current.Name] {
			return nil
		}
		if slices.Contains(stack, current.Name) {
			cycle := append(append([]string{}, stack...), current.Name)
			diagnostics = append(diagnostics, newDiagnostic(
				current.File,
				current.Pos,
				current.EndPos,
				CodeSpreadCycle,
				fmt.Sprintf("circular enum spread dependency detected: %s", formatCycle(cycle)),
			))
			return nil
		}

		stack = append(stack, current.Name)
		result := make([]*EnumMemberSymbol, 0, len(current.Members))

		for _, spread := range current.Spreads {
			if spread.Member != nil {
				diagnostics = append(diagnostics, newDiagnostic(
					current.File,
					spread.Pos,
					spread.EndPos,
					CodeSpreadTypeNotFound,
					"spread references must use Name form; Name.Member is not allowed",
				))
				continue
			}

			refEnum := symbols.lookupEnum(spread.Name)
			if refEnum == nil {
				msg := fmt.Sprintf("spread references undefined enum %q", spread.Name)
				suggestions, _ := strutil.FuzzySearch(enumNames(symbols), spread.Name)
				if len(suggestions) > 0 {
					msg += fmt.Sprintf("; did you mean %s?", formatSuggestions(suggestions))
				}
				diagnostics = append(diagnostics, newDiagnostic(
					current.File,
					spread.Pos,
					spread.EndPos,
					CodeSpreadTypeNotFound,
					msg,
				))
				continue
			}

			result = append(result, expand(refEnum)...)
		}

		result = append(result, current.Members...)
		stack = stack[:len(stack)-1]
		seen[current.Name] = true
		return result
	}

	effectiveMembers := expand(enum)
	valueType := inferEnumValueType(effectiveMembers)

	for _, m := range effectiveMembers {
		if m.HasExplicit {
			if _, err := strconv.ParseInt(m.Value, 10, 64); err == nil {
				if valueType == EnumValueTypeString {
					diagnostics = append(diagnostics, newDiagnostic(
						enum.File,
						m.Pos,
						m.EndPos,
						CodeEnumMixedTypes,
						fmt.Sprintf("enum %q mixes string and integer values", enum.Name),
					))
				}
			} else if valueType == EnumValueTypeInt {
				diagnostics = append(diagnostics, newDiagnostic(
					enum.File,
					m.Pos,
					m.EndPos,
					CodeEnumMixedTypes,
					fmt.Sprintf("enum %q mixes string and integer values", enum.Name),
				))
			}
		}
	}

	_ = frame{}
	return effectiveMembers, valueType, diagnostics
}

func inferEnumValueType(members []*EnumMemberSymbol) EnumValueType {
	for _, m := range members {
		if !m.HasExplicit {
			return EnumValueTypeString
		}
		if _, err := strconv.ParseInt(m.Value, 10, 64); err == nil {
			return EnumValueTypeInt
		}
	}
	return EnumValueTypeString
}

func enumNames(symbols *symbolTable) []string {
	names := make([]string, 0, len(symbols.enums))
	for name := range symbols.enums {
		names = append(names, name)
	}
	return names
}

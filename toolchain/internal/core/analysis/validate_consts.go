package analysis

import (
	"fmt"
	"slices"
	"strings"

	"github.com/varavelio/vdl/toolchain/internal/core/ast"
	"github.com/varavelio/vdl/toolchain/internal/util/strutil"
)

func validateConsts(symbols *symbolTable) []Diagnostic {
	var diagnostics []Diagnostic

	diagnostics = append(diagnostics, validateConstCycles(symbols)...)

	for _, cnst := range symbols.consts {
		diagnostics = append(diagnostics, validateConst(symbols, cnst)...)
	}

	return diagnostics
}

// validateConstCycles reports circular constant references.
//
// Reference chains must terminate in a literal: const a = b with
// const b = a (or const loop = loop) has no value. The visiting guard in
// validateDataLiteral only prevents infinite recursion, so cycles need this
// explicit check to surface as diagnostics.
func validateConstCycles(symbols *symbolTable) []Diagnostic {
	var diagnostics []Diagnostic

	for _, cnst := range symbols.consts {
		if cycle := findConstCycle(symbols, cnst.Name, nil); cycle != nil {
			diagnostics = append(diagnostics, newDiagnostic(
				cnst.File,
				cnst.Pos,
				cnst.EndPos,
				CodeCircularConst,
				fmt.Sprintf(
					"circular constant dependency detected: %s",
					strings.Join(cycle, " -> "),
				),
			))
		}
	}

	return diagnostics
}

// findConstCycle follows constant references depth-first. It returns the
// cycle path when name reappears in the current path, nil otherwise.
func findConstCycle(
	symbols *symbolTable,
	name string,
	visited []string,
) []string {
	if slices.Contains(visited, name) {
		return append(visited, name)
	}

	cnst := symbols.lookupConst(name)
	if cnst == nil || cnst.AST == nil || cnst.AST.Value == nil {
		return nil
	}

	visited = append(visited, name)
	for _, ref := range collectConstRefs(cnst.AST.Value) {
		if cycle := findConstCycle(symbols, ref, visited); cycle != nil {
			return cycle
		}
	}

	return nil
}

// collectConstRefs lists every constant referenced by a literal, including
// references nested in objects, arrays, and spreads. Enum member references
// (Name.Member) are excluded: they point at enums, not constants.
func collectConstRefs(lit *ast.DataLiteral) []string {
	if lit == nil {
		return nil
	}

	if lit.Scalar != nil && lit.Scalar.Ref != nil && lit.Scalar.Ref.Member == nil {
		return []string{lit.Scalar.Ref.Name}
	}

	if lit.Object != nil {
		var refs []string
		for _, entry := range lit.Object.Entries {
			if entry == nil {
				continue
			}
			if entry.Spread != nil {
				if entry.Spread.Ref.Member == nil {
					refs = append(refs, entry.Spread.Ref.Name)
				}
				continue
			}
			refs = append(refs, collectConstRefs(entry.Value)...)
		}
		return refs
	}

	if lit.Array != nil {
		var refs []string
		for _, elem := range lit.Array.Elements {
			refs = append(refs, collectConstRefs(elem)...)
		}
		return refs
	}

	return nil
}

func validateConst(symbols *symbolTable, cnst *ConstSymbol) []Diagnostic {
	var diagnostics []Diagnostic

	if cnst.AST == nil || cnst.AST.Value == nil {
		return diagnostics
	}

	valueDiagnostics, inferred := validateDataLiteral(
		symbols,
		cnst.File,
		cnst.AST.Value,
		map[string]bool{cnst.Name: true},
	)
	diagnostics = append(diagnostics, valueDiagnostics...)
	if inferred != ConstValueTypeUnknown {
		cnst.ValueType = inferred
	}

	return diagnostics
}

func validateDataLiteral(
	symbols *symbolTable,
	file string,
	lit *ast.DataLiteral,
	visiting map[string]bool,
) ([]Diagnostic, ConstValueType) {
	if lit == nil {
		return nil, ConstValueTypeUnknown
	}

	var diagnostics []Diagnostic

	if lit.Scalar != nil {
		s := lit.Scalar
		switch {
		case s.Str != nil:
			return nil, ConstValueTypeString
		case s.Int != nil:
			return nil, ConstValueTypeInt
		case s.Float != nil:
			return nil, ConstValueTypeFloat
		case s.True || s.False:
			return nil, ConstValueTypeBool
		case s.Ref != nil:
			if s.Ref.Member == nil {
				refConst := symbols.lookupConst(s.Ref.Name)
				if refConst == nil {
					msg := fmt.Sprintf("undefined constant reference %q", s.Ref.Name)
					suggestions, _ := strutil.FuzzySearch(symbols.allConstNames(), s.Ref.Name)
					if len(suggestions) > 0 {
						msg += fmt.Sprintf("; did you mean %s?", formatSuggestions(suggestions))
					}
					diagnostics = append(
						diagnostics,
						newDiagnostic(file, s.Ref.Pos, s.Ref.EndPos, CodeInvalidReference, msg),
					)
					return diagnostics, ConstValueTypeUnknown
				}

				if visiting[s.Ref.Name] {
					return diagnostics, refConst.ValueType
				}

				if refConst.AST != nil && refConst.AST.Value != nil {
					visiting[s.Ref.Name] = true
					childDiags, kind := validateDataLiteral(
						symbols,
						refConst.File,
						refConst.AST.Value,
						visiting,
					)
					delete(visiting, s.Ref.Name)
					diagnostics = append(diagnostics, childDiags...)
					if kind != ConstValueTypeUnknown {
						return diagnostics, kind
					}
				}

				return diagnostics, refConst.ValueType
			}

			refEnum := symbols.lookupEnum(s.Ref.Name)
			if refEnum == nil {
				msg := fmt.Sprintf("undefined enum reference %q", s.Ref.Name)
				suggestions, _ := strutil.FuzzySearch(enumNames(symbols), s.Ref.Name)
				if len(suggestions) > 0 {
					msg += fmt.Sprintf("; did you mean %s?", formatSuggestions(suggestions))
				}
				diagnostics = append(
					diagnostics,
					newDiagnostic(file, s.Ref.Pos, s.Ref.EndPos, CodeInvalidReference, msg),
				)
				return diagnostics, ConstValueTypeUnknown
			}

			members, valueType, enumDiags := expandEnumMembers(symbols, refEnum)
			diagnostics = append(diagnostics, enumDiags...)

			found := false
			for _, m := range members {
				if m.Name == *s.Ref.Member {
					found = true
					break
				}
			}
			if !found {
				diagnostics = append(diagnostics, newDiagnostic(
					file,
					s.Ref.Pos,
					s.Ref.EndPos,
					CodeEnumMemberNotFound,
					fmt.Sprintf("enum member %q not found in enum %q", *s.Ref.Member, s.Ref.Name),
				))
				return diagnostics, ConstValueTypeUnknown
			}

			if valueType == EnumValueTypeInt {
				return diagnostics, ConstValueTypeInt
			}
			return diagnostics, ConstValueTypeString
		}
	}

	if lit.Object != nil {
		for _, entry := range lit.Object.Entries {
			if entry.Spread != nil {
				if entry.Spread.Ref.Member != nil {
					diagnostics = append(diagnostics, newDiagnostic(
						file,
						entry.Spread.Pos,
						entry.Spread.EndPos,
						CodeConstSpreadNotObject,
						"constant object spreads must use Name form; Name.Member is not allowed",
					))
					continue
				}

				refConst := symbols.lookupConst(entry.Spread.Ref.Name)
				if refConst == nil {
					diagnostics = append(diagnostics, newDiagnostic(
						file,
						entry.Spread.Pos,
						entry.Spread.EndPos,
						CodeInvalidReference,
						fmt.Sprintf(
							"undefined constant reference %q in object spread",
							entry.Spread.Ref.Name,
						),
					))
					continue
				}

				if refConst.AST == nil || refConst.AST.Value == nil ||
					refConst.AST.Value.Object == nil {
					diagnostics = append(diagnostics, newDiagnostic(
						file,
						entry.Spread.Pos,
						entry.Spread.EndPos,
						CodeConstSpreadNotObject,
						fmt.Sprintf(
							"constant %q used in object spread is not an object literal",
							entry.Spread.Ref.Name,
						),
					))
					continue
				}

				if !visiting[entry.Spread.Ref.Name] {
					visiting[entry.Spread.Ref.Name] = true
					childDiags, _ := validateDataLiteral(
						symbols,
						refConst.File,
						refConst.AST.Value,
						visiting,
					)
					delete(visiting, entry.Spread.Ref.Name)
					diagnostics = append(diagnostics, childDiags...)
				}
				continue
			}

			childDiags, _ := validateDataLiteral(symbols, file, entry.Value, visiting)
			diagnostics = append(diagnostics, childDiags...)
		}

		return diagnostics, ConstValueTypeObject
	}

	if lit.Array != nil {
		elementType := ConstValueTypeUnknown
		for _, elem := range lit.Array.Elements {
			childDiags, childType := validateDataLiteral(symbols, file, elem, visiting)
			diagnostics = append(diagnostics, childDiags...)

			if childType == ConstValueTypeUnknown {
				continue
			}
			if elementType == ConstValueTypeUnknown {
				elementType = childType
				continue
			}
			if childType != elementType {
				diagnostics = append(diagnostics, newDiagnostic(
					file,
					elem.Pos,
					elem.EndPos,
					CodeConstArrayMixedTypes,
					"array literals must contain elements of the same type",
				))
			}
		}

		return diagnostics, ConstValueTypeArray
	}

	return diagnostics, ConstValueTypeUnknown
}

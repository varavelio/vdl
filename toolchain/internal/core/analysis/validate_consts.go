package analysis

import (
	"fmt"

	"github.com/varavelio/vdl/toolchain/internal/core/ast"
	"github.com/varavelio/vdl/toolchain/internal/util/strutil"
)

func validateConsts(symbols *symbolTable) []Diagnostic {
	var diagnostics []Diagnostic

	for _, cnst := range symbols.consts {
		diagnostics = append(diagnostics, validateConst(symbols, cnst)...)
	}

	return diagnostics
}

func validateConst(symbols *symbolTable, cnst *ConstSymbol) []Diagnostic {
	var diagnostics []Diagnostic

	if cnst.AST == nil || cnst.AST.Value == nil {
		return diagnostics
	}

	valueDiagnostics, inferred := validateDataLiteral(symbols, cnst.File, cnst.AST.Value, map[string]bool{cnst.Name: true})
	diagnostics = append(diagnostics, valueDiagnostics...)
	if inferred != ConstValueTypeUnknown {
		cnst.ValueType = inferred
	}

	if cnst.ExplicitTypeName != nil {
		diagnostics = append(diagnostics, validateExplicitConstType(symbols, cnst, inferred)...)
	}

	return diagnostics
}

func validateExplicitConstType(symbols *symbolTable, cnst *ConstSymbol, inferred ConstValueType) []Diagnostic {
	typeName := *cnst.ExplicitTypeName

	if ast.IsPrimitiveType(typeName) {
		expected := primitiveNameToConstType(typeName)
		if expected != ConstValueTypeUnknown && inferred != ConstValueTypeUnknown && inferred != expected {
			return []Diagnostic{newDiagnostic(
				cnst.File,
				cnst.Pos,
				cnst.EndPos,
				CodeTypeNotDeclared,
				fmt.Sprintf("constant %q has explicit type %q but value is %s", cnst.Name, typeName, constValueTypeName(inferred)),
			)}
		}
		return nil
	}

	if symbols.lookupType(typeName) != nil {
		return nil
	}

	if symbols.lookupEnum(typeName) != nil {
		if inferred != ConstValueTypeString && inferred != ConstValueTypeInt && inferred != ConstValueTypeUnknown {
			return []Diagnostic{newDiagnostic(
				cnst.File,
				cnst.Pos,
				cnst.EndPos,
				CodeTypeNotDeclared,
				fmt.Sprintf("constant %q has explicit enum type %q but value is %s", cnst.Name, typeName, constValueTypeName(inferred)),
			)}
		}
		return nil
	}

	msg := fmt.Sprintf("undefined type %q in constant %q", typeName, cnst.Name)
	suggestions, _ := strutil.FuzzySearch(symbols.allFieldTypeNames(), typeName)
	if len(suggestions) > 0 {
		msg += fmt.Sprintf("; did you mean %s?", formatSuggestions(suggestions))
	}

	return []Diagnostic{newDiagnostic(cnst.File, cnst.Pos, cnst.EndPos, CodeTypeNotDeclared, msg)}
}

func validateDataLiteral(symbols *symbolTable, file string, lit *ast.DataLiteral, visiting map[string]bool) ([]Diagnostic, ConstValueType) {
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
					diagnostics = append(diagnostics, newDiagnostic(file, s.Ref.Pos, s.Ref.EndPos, CodeInvalidReference, msg))
					return diagnostics, ConstValueTypeUnknown
				}

				if visiting[s.Ref.Name] {
					return diagnostics, refConst.ValueType
				}

				if refConst.AST != nil && refConst.AST.Value != nil {
					visiting[s.Ref.Name] = true
					childDiags, kind := validateDataLiteral(symbols, refConst.File, refConst.AST.Value, visiting)
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
				diagnostics = append(diagnostics, newDiagnostic(file, s.Ref.Pos, s.Ref.EndPos, CodeInvalidReference, msg))
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
						fmt.Sprintf("undefined constant reference %q in object spread", entry.Spread.Ref.Name),
					))
					continue
				}

				if refConst.AST == nil || refConst.AST.Value == nil || refConst.AST.Value.Object == nil {
					diagnostics = append(diagnostics, newDiagnostic(
						file,
						entry.Spread.Pos,
						entry.Spread.EndPos,
						CodeConstSpreadNotObject,
						fmt.Sprintf("constant %q used in object spread is not an object literal", entry.Spread.Ref.Name),
					))
					continue
				}

				if !visiting[entry.Spread.Ref.Name] {
					visiting[entry.Spread.Ref.Name] = true
					childDiags, _ := validateDataLiteral(symbols, refConst.File, refConst.AST.Value, visiting)
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

func primitiveNameToConstType(name string) ConstValueType {
	switch name {
	case "string":
		return ConstValueTypeString
	case "int":
		return ConstValueTypeInt
	case "float":
		return ConstValueTypeFloat
	case "bool":
		return ConstValueTypeBool
	default:
		return ConstValueTypeUnknown
	}
}

func constValueTypeName(vt ConstValueType) string {
	switch vt {
	case ConstValueTypeString:
		return "string"
	case ConstValueTypeInt:
		return "int"
	case ConstValueTypeFloat:
		return "float"
	case ConstValueTypeBool:
		return "bool"
	case ConstValueTypeObject:
		return "object"
	case ConstValueTypeArray:
		return "array"
	case ConstValueTypeReference:
		return "reference"
	default:
		return "unknown"
	}
}

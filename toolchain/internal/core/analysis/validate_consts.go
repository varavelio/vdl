package analysis

import (
	"fmt"
	"maps"
	"slices"
	"strconv"
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
		var reference string
		hasReference := false
		for _, elem := range lit.Array.Elements {
			childDiags, _ := validateDataLiteral(symbols, file, elem, visiting)
			diagnostics = append(diagnostics, childDiags...)

			sig, ok := dataLiteralSig(symbols, elem, map[string]bool{})
			if !ok {
				continue
			}
			if !hasReference {
				reference = sig
				hasReference = true
				continue
			}
			if sig != reference {
				diagnostics = append(diagnostics, newDiagnostic(
					file,
					elem.Pos,
					elem.EndPos,
					CodeConstArrayMixedTypes,
					"array literals must contain elements of the same structure",
				))
			}
		}

		return diagnostics, ConstValueTypeArray
	}

	return diagnostics, ConstValueTypeUnknown
}

// dataLiteralSig returns the canonical structural signature of a literal.
//
// Scalar values are ignored, only their kind matters. Object keys are sorted,
// so key order is irrelevant. Spreads are expanded with override semantics.
// References resolve through constants and enum members.
//
// ok is false when any part cannot be resolved (undefined reference, invalid
// spread, cycle). Those cases are reported by their own validators, so the
// caller skips the structural comparison instead of cascading errors.
func dataLiteralSig(
	symbols *symbolTable,
	lit *ast.DataLiteral,
	visiting map[string]bool,
) (sig string, ok bool) {
	if lit == nil {
		return "", false
	}

	if lit.Scalar != nil {
		return scalarSig(symbols, lit.Scalar, visiting)
	}

	if lit.Object != nil {
		fields, ok := objectFieldsSig(symbols, lit.Object, visiting)
		if !ok {
			return "", false
		}
		return encodeObjectSig(fields), true
	}

	if lit.Array != nil {
		// The length is not part of the structure: [80 443] and [80] are
		// both "string arrays". A nested array whose own items diverge is
		// reported at its own level, so here it resolves to not-ok and the
		// outer comparison skips it instead of cascading.
		if len(lit.Array.Elements) == 0 {
			return "array[]", true
		}
		first, ok := dataLiteralSig(symbols, lit.Array.Elements[0], visiting)
		if !ok {
			return "", false
		}
		for _, elem := range lit.Array.Elements[1:] {
			s, ok := dataLiteralSig(symbols, elem, visiting)
			if !ok || s != first {
				return "", false
			}
		}
		return "array[" + first + "]", true
	}

	return "", false
}

// scalarSig returns the structural signature of a scalar literal.
func scalarSig(
	symbols *symbolTable,
	s *ast.ScalarLiteral,
	visiting map[string]bool,
) (string, bool) {
	switch {
	case s.Str != nil:
		return "string", true
	case s.Int != nil:
		return "int", true
	case s.Float != nil:
		return "float", true
	case s.True || s.False:
		return "bool", true
	case s.Ref != nil:
		if s.Ref.Member == nil {
			return constRefSig(symbols, s.Ref.Name, visiting)
		}
		kind, ok := enumMemberKind(symbols, s.Ref.Name, *s.Ref.Member, visiting)
		if !ok {
			return "", false
		}
		if kind == ConstValueTypeInt {
			return "int", true
		}
		return "string", true
	}

	return "", false
}

// constRefSig returns the signature of the value a constant reference points
// to. Cyclic and dangling references resolve to not-ok; their own validators
// report them.
func constRefSig(
	symbols *symbolTable,
	name string,
	visiting map[string]bool,
) (string, bool) {
	if visiting[name] {
		return "", false
	}
	refConst := symbols.lookupConst(name)
	if refConst == nil || refConst.AST == nil || refConst.AST.Value == nil {
		return "", false
	}
	visiting[name] = true
	defer delete(visiting, name)
	return dataLiteralSig(symbols, refConst.AST.Value, visiting)
}

// objectFieldsSig returns the merged key signatures of an object literal,
// expanding spreads with override semantics (explicit keys win).
func objectFieldsSig(
	symbols *symbolTable,
	obj *ast.DataLiteralObject,
	visiting map[string]bool,
) (map[string]string, bool) {
	fields := map[string]string{}
	for _, entry := range obj.Entries {
		if entry == nil {
			continue
		}
		if entry.Spread != nil {
			if entry.Spread.Ref.Member != nil {
				return nil, false
			}
			spread, ok := spreadFieldsSig(symbols, entry.Spread.Ref.Name, visiting)
			if !ok {
				return nil, false
			}
			maps.Copy(fields, spread)
			continue
		}
		s, ok := dataLiteralSig(symbols, entry.Value, visiting)
		if !ok {
			return nil, false
		}
		fields[entry.Key] = s
	}
	return fields, true
}

// spreadFieldsSig returns the key signatures of the object constant a spread
// points to. Non-object and dangling targets resolve to not-ok; their own
// validators report them.
func spreadFieldsSig(
	symbols *symbolTable,
	name string,
	visiting map[string]bool,
) (map[string]string, bool) {
	if visiting[name] {
		return nil, false
	}
	refConst := symbols.lookupConst(name)
	if refConst == nil || refConst.AST == nil || refConst.AST.Value == nil ||
		refConst.AST.Value.Object == nil {
		return nil, false
	}
	visiting[name] = true
	defer delete(visiting, name)
	return objectFieldsSig(symbols, refConst.AST.Value.Object, visiting)
}

// encodeObjectSig encodes key signatures with sorted keys, so key order is
// irrelevant to the comparison.
func encodeObjectSig(fields map[string]string) string {
	keys := make([]string, 0, len(fields))
	for k := range fields {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	var b strings.Builder
	b.WriteString("object{")
	for i, k := range keys {
		if i > 0 {
			b.WriteString(",")
		}
		b.WriteString(k)
		b.WriteString(":")
		b.WriteString(fields[k])
	}
	b.WriteString("}")
	return b.String()
}

// enumMemberKind returns the value kind of an enum member without emitting
// diagnostics. Real problems (missing enum, spread errors, mixed types) are
// reported by the enum and literal validators.
func enumMemberKind(
	symbols *symbolTable,
	enumName, memberName string,
	visiting map[string]bool,
) (ConstValueType, bool) {
	refEnum := symbols.lookupEnum(enumName)
	if refEnum == nil {
		return ConstValueTypeUnknown, false
	}
	return findEnumMemberKind(symbols, refEnum, memberName, visiting)
}

// findEnumMemberKind searches a member in an enum, following spreads quietly.
func findEnumMemberKind(
	symbols *symbolTable,
	enum *EnumSymbol,
	memberName string,
	visiting map[string]bool,
) (ConstValueType, bool) {
	for _, m := range enum.Members {
		if m.Name == memberName {
			if !m.HasExplicit {
				return ConstValueTypeString, true
			}
			if _, err := strconv.ParseInt(m.Value, 10, 64); err == nil {
				return ConstValueTypeInt, true
			}
			return ConstValueTypeString, true
		}
	}
	if visiting[enum.Name] {
		return ConstValueTypeUnknown, false
	}
	visiting[enum.Name] = true
	defer delete(visiting, enum.Name)
	for _, spread := range enum.Spreads {
		if spread == nil || spread.Member != nil {
			continue
		}
		refEnum := symbols.lookupEnum(spread.Name)
		if refEnum == nil {
			continue
		}
		if kind, ok := findEnumMemberKind(symbols, refEnum, memberName, visiting); ok {
			return kind, true
		}
	}
	return ConstValueTypeUnknown, false
}

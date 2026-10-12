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
			// The lexer guarantees digits with an optional minus, so a
			// parse failure here means the value exceeds 64 bits. Without
			// this check the IR would silently mistype the literal.
			if _, err := strconv.ParseInt(*s.Int, 10, 64); err != nil {
				diagnostics = append(diagnostics, newDiagnostic(
					file,
					lit.Pos,
					lit.EndPos,
					CodeNumericOutOfRange,
					"integer literal is out of range for 64-bit int",
				))
				return diagnostics, ConstValueTypeUnknown
			}
			return nil, ConstValueTypeInt
		case s.Float != nil:
			// Only true overflow (±Inf) fails here; ordinary precision
			// loss is standard float semantics, not an error.
			if _, err := strconv.ParseFloat(*s.Float, 64); err != nil {
				diagnostics = append(diagnostics, newDiagnostic(
					file,
					lit.Pos,
					lit.EndPos,
					CodeNumericOutOfRange,
					"float literal is out of range for 64-bit float",
				))
				return diagnostics, ConstValueTypeUnknown
			}
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
		// Items merge by union: shared object keys must have compatible
		// types, extra keys are allowed (they become optional fields for
		// plugins). Each item merges into the accumulated shape, so clashes
		// surface regardless of item order. Unknown items (unresolvable
		// references, cycles) are skipped: their own validators report them.
		var merged *literalShape
		for _, elem := range lit.Array.Elements {
			childDiags, _ := validateDataLiteral(symbols, file, elem, visiting)
			diagnostics = append(diagnostics, childDiags...)

			// Fresh guards per item: reusing one constant twice is not a cycle.
			shape := shapeOf(symbols, elem, map[string]bool{})
			if shape == nil {
				continue
			}
			if merged == nil {
				merged = shape
				continue
			}
			if next, ok := mergeShapes(merged, shape); ok {
				merged = next
			} else {
				diagnostics = append(diagnostics, newDiagnostic(
					file,
					elem.Pos,
					elem.EndPos,
					CodeConstArrayMixedTypes,
					"array literals must contain compatible elements: shared object keys must have the same type",
				))
			}
		}

		return diagnostics, ConstValueTypeArray
	}

	return diagnostics, ConstValueTypeUnknown
}

// literalShape describes canonical structure for union comparison of array
// items. Scalars carry their kind only; objects map keys to nested shapes;
// arrays carry the merged shape of their items. A nil shape (or nil map
// entry) means unknown: it merges with anything because its own validator
// already reports the underlying problem.
type literalShape struct {
	kind   string
	fields map[string]*literalShape
	items  *literalShape
}

// shapeOf returns the canonical structure of a literal.
func shapeOf(
	symbols *symbolTable,
	lit *ast.DataLiteral,
	visiting map[string]bool,
) *literalShape {
	if lit == nil {
		return nil
	}

	if lit.Scalar != nil {
		return scalarShape(symbols, lit.Scalar, visiting)
	}

	if lit.Object != nil {
		fields, ok := objectShapes(symbols, lit.Object, visiting)
		if !ok {
			return nil
		}
		return &literalShape{kind: "object", fields: fields}
	}

	if lit.Array != nil {
		// Length is not structure. A nested array whose own items clash is
		// reported at its own level, so here it stays unknown and the outer
		// comparison skips it instead of cascading.
		var merged *literalShape
		for _, elem := range lit.Array.Elements {
			shape := shapeOf(symbols, elem, visiting)
			if shape == nil {
				return nil
			}
			if merged == nil {
				merged = shape
				continue
			}
			next, ok := mergeShapes(merged, shape)
			if !ok {
				return nil
			}
			merged = next
		}
		return &literalShape{kind: "array", items: merged}
	}

	return nil
}

// mergeShapes returns the union of two shapes, or false when incompatible.
// Shared object keys merge recursively; extra keys are adopted as-is; an
// unknown (nil) side adopts the other. Inputs are never mutated.
func mergeShapes(a, b *literalShape) (*literalShape, bool) {
	if a == nil {
		return b, true
	}
	if b == nil {
		return a, true
	}
	if a.kind != b.kind {
		return nil, false
	}
	switch a.kind {
	case "object":
		fields := make(map[string]*literalShape, len(a.fields)+len(b.fields))
		maps.Copy(fields, a.fields)
		for k, v := range b.fields {
			existing, found := fields[k]
			if !found {
				fields[k] = v
				continue
			}
			merged, ok := mergeShapes(existing, v)
			if !ok {
				return nil, false
			}
			fields[k] = merged
		}
		return &literalShape{kind: "object", fields: fields}, true
	case "array":
		items, ok := mergeShapes(a.items, b.items)
		if !ok {
			return nil, false
		}
		return &literalShape{kind: "array", items: items}, true
	default:
		return &literalShape{kind: a.kind}, true
	}
}

// scalarShape returns the structure of a scalar literal: its kind, with
// references resolved through constants and enum members.
func scalarShape(
	symbols *symbolTable,
	s *ast.ScalarLiteral,
	visiting map[string]bool,
) *literalShape {
	switch {
	case s.Str != nil:
		return &literalShape{kind: "string"}
	case s.Int != nil:
		return &literalShape{kind: "int"}
	case s.Float != nil:
		return &literalShape{kind: "float"}
	case s.True || s.False:
		return &literalShape{kind: "bool"}
	case s.Ref != nil:
		if s.Ref.Member == nil {
			return constShape(symbols, s.Ref.Name, visiting)
		}
		kind, ok := enumMemberKind(symbols, s.Ref.Name, *s.Ref.Member, visiting)
		if !ok {
			return nil
		}
		if kind == ConstValueTypeInt {
			return &literalShape{kind: "int"}
		}
		return &literalShape{kind: "string"}
	}

	return nil
}

// constShape returns the shape of the value a constant reference points to.
// Cyclic and dangling references resolve to unknown; their own validators
// report them.
func constShape(
	symbols *symbolTable,
	name string,
	visiting map[string]bool,
) *literalShape {
	if visiting[name] {
		return nil
	}
	refConst := symbols.lookupConst(name)
	if refConst == nil || refConst.AST == nil || refConst.AST.Value == nil {
		return nil
	}
	visiting[name] = true
	defer delete(visiting, name)
	return shapeOf(symbols, refConst.AST.Value, visiting)
}

// objectShapes returns the merged key shapes of an object literal, expanding
// spreads with override semantics (explicit keys win). A bad spread resolves
// the whole object to unknown; its own validator reports it.
func objectShapes(
	symbols *symbolTable,
	obj *ast.DataLiteralObject,
	visiting map[string]bool,
) (map[string]*literalShape, bool) {
	fields := map[string]*literalShape{}
	for _, entry := range obj.Entries {
		if entry == nil {
			continue
		}
		if entry.Spread != nil {
			if entry.Spread.Ref.Member != nil {
				return nil, false
			}
			spread, ok := spreadShapes(symbols, entry.Spread.Ref.Name, visiting)
			if !ok {
				return nil, false
			}
			maps.Copy(fields, spread)
			continue
		}
		fields[entry.Key] = shapeOf(symbols, entry.Value, visiting)
	}
	return fields, true
}

// spreadShapes returns the key shapes of the object constant a spread points
// to. Non-object and dangling targets resolve to not-ok; their own validators
// report them.
func spreadShapes(
	symbols *symbolTable,
	name string,
	visiting map[string]bool,
) (map[string]*literalShape, bool) {
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
	return objectShapes(symbols, refConst.AST.Value.Object, visiting)
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

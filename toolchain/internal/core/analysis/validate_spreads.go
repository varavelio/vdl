package analysis

import (
	"fmt"
	"strings"

	"github.com/varavelio/vdl/toolchain/internal/core/ast"
	"github.com/varavelio/vdl/toolchain/internal/util/strutil"
)

// validateSpreads checks that spread references are valid for types and inline objects.
func validateSpreads(symbols *symbolTable) []Diagnostic {
	var diagnostics []Diagnostic

	for _, typ := range symbols.types {
		diagnostics = append(diagnostics, validateTypeSpreads(symbols, typ)...)
		for _, field := range typ.Fields {
			diagnostics = append(diagnostics, validateInlineObjectSpreads(symbols, field.Type, typ.File, field.Name)...)
		}
	}

	diagnostics = append(diagnostics, validateSpreadCycles(symbols)...)
	return diagnostics
}

func validateTypeSpreads(symbols *symbolTable, typ *TypeSymbol) []Diagnostic {
	var diagnostics []Diagnostic

	fieldNames := map[string]FieldOrigin{}
	for _, f := range typ.Fields {
		fieldNames[f.Name] = FieldOrigin{File: f.File, Pos: f.Pos, Source: "direct field"}
	}

	for _, spread := range typ.Spreads {
		if spread.Member != nil {
			diagnostics = append(diagnostics, newDiagnostic(
				typ.File,
				spread.Pos,
				spread.EndPos,
				CodeSpreadTypeNotFound,
				"spread references must use Name form; Name.Member is not allowed",
			))
			continue
		}

		refType := symbols.lookupType(spread.Name)
		if refType == nil {
			msg := fmt.Sprintf("spread references undefined type %q", spread.Name)
			suggestions, _ := strutil.FuzzySearch(symbols.allTypeNames(), spread.Name)
			if len(suggestions) > 0 {
				msg += fmt.Sprintf("; did you mean %s?", formatSuggestions(suggestions))
			}
			diagnostics = append(diagnostics, newDiagnostic(typ.File, spread.Pos, spread.EndPos, CodeSpreadTypeNotFound, msg))
			continue
		}

		for name, origin := range flattenTypeFieldOrigins(symbols, refType, map[string]bool{}) {
			if existing, ok := fieldNames[name]; ok {
				diagnostics = append(diagnostics, newDiagnostic(
					typ.File,
					spread.Pos,
					spread.EndPos,
					CodeSpreadFieldConflict,
					fmt.Sprintf("field %q from spread %q conflicts with %s at %s:%d:%d",
						name, spread.Name, existing.Source, existing.File, existing.Pos.Line, existing.Pos.Column),
				))
				continue
			}
			fieldNames[name] = origin
		}
	}

	return diagnostics
}

func validateInlineObjectSpreads(symbols *symbolTable, typeInfo *FieldTypeInfo, file, owner string) []Diagnostic {
	if typeInfo == nil {
		return nil
	}

	var diagnostics []Diagnostic

	switch typeInfo.Kind {
	case FieldTypeKindMap:
		return validateInlineObjectSpreads(symbols, typeInfo.MapValue, file, owner)
	case FieldTypeKindObject:
		if typeInfo.ObjectDef == nil {
			return nil
		}

		fieldNames := map[string]FieldOrigin{}
		for _, f := range typeInfo.ObjectDef.Fields {
			fieldNames[f.Name] = FieldOrigin{File: file, Pos: f.Pos, Source: "direct field"}
			diagnostics = append(diagnostics, validateInlineObjectSpreads(symbols, f.Type, file, f.Name)...)
		}

		for _, spread := range typeInfo.ObjectDef.Spreads {
			if spread.Member != nil {
				diagnostics = append(diagnostics, newDiagnostic(
					file,
					spread.Pos,
					spread.EndPos,
					CodeSpreadTypeNotFound,
					"spread references must use Name form; Name.Member is not allowed",
				))
				continue
			}

			refType := symbols.lookupType(spread.Name)
			if refType == nil {
				msg := fmt.Sprintf("spread in inline object %q references undefined type %q", owner, spread.Name)
				suggestions, _ := strutil.FuzzySearch(symbols.allTypeNames(), spread.Name)
				if len(suggestions) > 0 {
					msg += fmt.Sprintf("; did you mean %s?", formatSuggestions(suggestions))
				}
				diagnostics = append(diagnostics, newDiagnostic(file, spread.Pos, spread.EndPos, CodeSpreadTypeNotFound, msg))
				continue
			}

			for name, origin := range flattenTypeFieldOrigins(symbols, refType, map[string]bool{}) {
				if existing, ok := fieldNames[name]; ok {
					diagnostics = append(diagnostics, newDiagnostic(
						file,
						spread.Pos,
						spread.EndPos,
						CodeSpreadFieldConflict,
						fmt.Sprintf("field %q from spread %q in inline object %q conflicts with %s at %s:%d:%d",
							name, spread.Name, owner, existing.Source, existing.File, existing.Pos.Line, existing.Pos.Column),
					))
					continue
				}
				fieldNames[name] = origin
			}
		}
	}

	return diagnostics
}

type FieldOrigin struct {
	File   string
	Pos    Position
	Source string
}

type Position = ast.Position

func flattenTypeFieldOrigins(symbols *symbolTable, typ *TypeSymbol, visiting map[string]bool) map[string]FieldOrigin {
	result := map[string]FieldOrigin{}
	if typ == nil {
		return result
	}
	if visiting[typ.Name] {
		return result
	}
	visiting[typ.Name] = true

	for _, f := range typ.Fields {
		result[f.Name] = FieldOrigin{File: f.File, Pos: f.Pos, Source: fmt.Sprintf("spread ...%s", typ.Name)}
	}

	for _, spread := range typ.Spreads {
		if spread.Member != nil {
			continue
		}
		for name, origin := range flattenTypeFieldOrigins(symbols, symbols.lookupType(spread.Name), visiting) {
			if _, exists := result[name]; !exists {
				result[name] = origin
			}
		}
	}

	delete(visiting, typ.Name)
	return result
}

func validateSpreadCycles(symbols *symbolTable) []Diagnostic {
	var diagnostics []Diagnostic

	for typeName, typ := range symbols.types {
		if cycle := findSpreadCycle(symbols, typeName, []string{}); cycle != nil {
			diagnostics = append(diagnostics, newDiagnostic(
				typ.File,
				typ.Pos,
				typ.EndPos,
				CodeSpreadCycle,
				fmt.Sprintf("circular spread dependency detected: %s", formatCycle(cycle)),
			))
		}
	}

	return diagnostics
}

func findSpreadCycle(symbols *symbolTable, typeName string, visited []string) []string {
	for _, v := range visited {
		if v == typeName {
			return append(visited, typeName)
		}
	}

	typ := symbols.lookupType(typeName)
	if typ == nil {
		return nil
	}

	newVisited := append(visited, typeName)

	for _, spread := range typ.Spreads {
		if spread.Member != nil {
			continue
		}
		if cycle := findSpreadCycle(symbols, spread.Name, newVisited); cycle != nil {
			return cycle
		}
	}

	return nil
}

func formatCycle(cycle []string) string {
	var result strings.Builder
	for i, name := range cycle {
		if i > 0 {
			result.WriteString(" -> ")
		}
		result.WriteString(name)
	}
	return result.String()
}

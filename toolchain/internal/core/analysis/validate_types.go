package analysis

import (
	"fmt"

	"github.com/varavelio/vdl/toolchain/internal/core/ast"
	"github.com/varavelio/vdl/toolchain/internal/util/strutil"
)

// validateTypes checks that all type references point to existing types.
// This includes type expressions (non-object types), object type fields, inline objects, and map values.
func validateTypes(symbols *symbolTable) []Diagnostic {
	var diagnostics []Diagnostic

	for _, typ := range symbols.types {
		if typ.Type == nil {
			continue
		}
		diagnostics = append(diagnostics, validateTypeRef(symbols, typ)...)
	}

	return diagnostics
}

// validateTypeRef validates the type reference of a TypeSymbol.
func validateTypeRef(symbols *symbolTable, typ *TypeSymbol) []Diagnostic {
	typeInfo := typ.Type
	if typeInfo == nil {
		return nil
	}

	var diagnostics []Diagnostic

	switch typeInfo.Kind {
	case FieldTypeKindPrimitive:
		// Primitive types are always valid

	case FieldTypeKindCustom:
		typeSym := symbols.lookupType(typeInfo.Name)
		enumSym := symbols.lookupEnum(typeInfo.Name)

		if typeSym != nil {
			typeInfo.ResolvedType = typeSym
		} else if enumSym != nil {
			typeInfo.ResolvedEnum = enumSym
		} else {
			msg := fmt.Sprintf(
				"undefined type %q in type %q",
				typeInfo.Name, typ.Name,
			)
			suggestions, _ := strutil.FuzzySearch(symbols.allFieldTypeNames(), typeInfo.Name)
			if len(suggestions) > 0 {
				msg += fmt.Sprintf("; did you mean %s?", formatSuggestions(suggestions))
			}
			diagnostics = append(diagnostics, newDiagnostic(
				typ.File,
				typ.Pos,
				typ.EndPos,
				CodeTypeNotDeclared,
				msg,
			))
		}

	case FieldTypeKindMap:
		if typeInfo.MapValue != nil {
			diagnostics = append(diagnostics, validateNestedTypeRef(symbols, typeInfo.MapValue, typ)...)
		}

	case FieldTypeKindObject:
		// Validate object type field references
		if typeInfo.ObjectDef != nil {
			diagnostics = append(diagnostics, validateFieldTypes(symbols, typeInfo.ObjectDef.Fields, "type", typ.Name)...)
		}
	}

	return diagnostics
}

// validateNestedTypeRef validates type references within nested types (e.g., map values).
func validateNestedTypeRef(symbols *symbolTable, typeInfo *FieldTypeInfo, typ *TypeSymbol) []Diagnostic {
	if typeInfo == nil {
		return nil
	}

	var diagnostics []Diagnostic

	switch typeInfo.Kind {
	case FieldTypeKindPrimitive:
		// Always valid

	case FieldTypeKindCustom:
		typeSym := symbols.lookupType(typeInfo.Name)
		enumSym := symbols.lookupEnum(typeInfo.Name)

		if typeSym != nil {
			typeInfo.ResolvedType = typeSym
		} else if enumSym != nil {
			typeInfo.ResolvedEnum = enumSym
		} else {
			msg := fmt.Sprintf(
				"undefined type %q in type %q",
				typeInfo.Name, typ.Name,
			)
			suggestions, _ := strutil.FuzzySearch(symbols.allFieldTypeNames(), typeInfo.Name)
			if len(suggestions) > 0 {
				msg += fmt.Sprintf("; did you mean %s?", formatSuggestions(suggestions))
			}
			diagnostics = append(diagnostics, newDiagnostic(
				typ.File,
				typ.Pos,
				typ.EndPos,
				CodeTypeNotDeclared,
				msg,
			))
		}

	case FieldTypeKindMap:
		if typeInfo.MapValue != nil {
			diagnostics = append(diagnostics, validateNestedTypeRef(symbols, typeInfo.MapValue, typ)...)
		}

	case FieldTypeKindObject:
		if typeInfo.ObjectDef != nil {
			diagnostics = append(diagnostics, validateFieldTypes(symbols, typeInfo.ObjectDef.Fields, "inline object in type", typ.Name)...)
		}
	}

	return diagnostics
}

// validateFieldTypes validates that all field types reference existing types.
func validateFieldTypes(symbols *symbolTable, fields []*FieldSymbol, context, parentName string) []Diagnostic {
	var diagnostics []Diagnostic

	for _, field := range fields {
		diagnostics = append(diagnostics, validateFieldType(symbols, field.Type, field, context, parentName)...)
	}

	return diagnostics
}

// validateFieldType validates a single field type recursively.
func validateFieldType(symbols *symbolTable, typeInfo *FieldTypeInfo, field *FieldSymbol, context, parentName string) []Diagnostic {
	if typeInfo == nil {
		return nil
	}

	var diagnostics []Diagnostic

	switch typeInfo.Kind {
	case FieldTypeKindPrimitive:
		// Primitive types are always valid
		return nil

	case FieldTypeKindCustom:
		// Try to resolve the custom type (can be a type or enum)
		typeSym := symbols.lookupType(typeInfo.Name)
		enumSym := symbols.lookupEnum(typeInfo.Name)

		if typeSym != nil {
			// Link the resolved type for O(1) LSP navigation
			typeInfo.ResolvedType = typeSym
		} else if enumSym != nil {
			// Link the resolved enum for O(1) LSP navigation
			typeInfo.ResolvedEnum = enumSym
		} else {
			// Type not found - emit diagnostic with suggestions
			msg := fmt.Sprintf(
				"undefined type %q in field %q of %s %q",
				typeInfo.Name, field.Name, context, parentName,
			)

			// Find similar types to suggest
			suggestions, _ := strutil.FuzzySearch(symbols.allFieldTypeNames(), typeInfo.Name)
			if len(suggestions) > 0 {
				msg += fmt.Sprintf("; did you mean %s?", formatSuggestions(suggestions))
			}

			diagnostics = append(diagnostics, newDiagnostic(
				field.File,
				field.Pos,
				field.EndPos,
				CodeTypeNotDeclared,
				msg,
			))
		}

	case FieldTypeKindMap:
		// Validate map value type
		if typeInfo.MapValue != nil {
			diagnostics = append(diagnostics, validateFieldType(symbols, typeInfo.MapValue, field, context, parentName)...)
		}

	case FieldTypeKindObject:
		// Validate inline object fields recursively
		if typeInfo.ObjectDef != nil {
			diagnostics = append(diagnostics, validateFieldTypes(symbols, typeInfo.ObjectDef.Fields, "inline object in", field.Name)...)
		}
	}

	return diagnostics
}

// buildFieldTypeInfo converts an AST FieldType to a FieldTypeInfo.
func buildFieldTypeInfo(ft *ast.FieldType) *FieldTypeInfo {
	return buildFieldTypeInfoWithFile(ft, "")
}

// buildFieldTypeInfoWithFile converts an AST FieldType to a FieldTypeInfo with a file path for diagnostics.
func buildFieldTypeInfoWithFile(ft *ast.FieldType, file string) *FieldTypeInfo {
	if ft == nil || ft.Base == nil {
		return nil
	}

	info := &FieldTypeInfo{
		ArrayDims: int(ft.Dimensions),
	}

	if ft.Base.Named != nil {
		name := *ft.Base.Named
		if ast.IsPrimitiveType(name) {
			info.Kind = FieldTypeKindPrimitive
		} else {
			info.Kind = FieldTypeKindCustom
		}
		info.Name = name
	} else if ft.Base.Map != nil {
		info.Kind = FieldTypeKindMap
		info.MapValue = buildFieldTypeInfoWithFile(ft.Base.Map.ValueType, file)
	} else if ft.Base.Object != nil {
		info.Kind = FieldTypeKindObject
		info.ObjectDef = buildInlineObjectWithFile(ft.Base.Object, file)
	}

	return info
}

// buildInlineObjectWithFile converts an AST FieldTypeObject to an InlineObject with a file path for diagnostics.
func buildInlineObjectWithFile(obj *ast.FieldTypeObject, file string) *InlineObject {
	if obj == nil {
		return nil
	}

	inline := &InlineObject{
		Fields:  []*FieldSymbol{},
		Spreads: []*SpreadRef{},
	}

	for _, child := range obj.Members {
		if child.Field != nil {
			inline.Fields = append(inline.Fields, buildFieldSymbol(child.Field, file))
		}
		if child.Spread != nil {
			inline.Spreads = append(inline.Spreads, &SpreadRef{
				Name:   child.Spread.Ref.Name,
				Member: child.Spread.Ref.Member,
				Pos:    child.Spread.Pos,
				EndPos: child.Spread.EndPos,
			})
		}
	}

	return inline
}

// buildFieldSymbol creates a FieldSymbol from an AST Field.
func buildFieldSymbol(field *ast.Field, file string) *FieldSymbol {
	var docstring *string
	if field.Docstring != nil {
		s := string(field.Docstring.Value)
		docstring = &s
	}

	return &FieldSymbol{
		Symbol: Symbol{
			Name:        string(field.Name),
			File:        file,
			Pos:         field.Pos,
			EndPos:      field.EndPos,
			Docstring:   docstring,
			Annotations: buildAnnotationRefs(field.Annotations),
		},
		AST:      field,
		Optional: field.Optional,
		Type:     buildFieldTypeInfo(&field.Type),
	}
}

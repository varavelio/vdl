package analysis

import (
	"fmt"

	"github.com/varavelio/vdl/toolchain/internal/core/ast"
	"github.com/varavelio/vdl/toolchain/internal/util/strutil"
)

// validateTypes checks that all type references point to existing types.
// This includes:
// - Field types in type declarations
// - Field types in inline objects
// - Field types in input/output blocks
// - Map value types
func validateTypes(symbols *symbolTable) []Diagnostic {
	var diagnostics []Diagnostic

	// Check type fields
	for _, typ := range symbols.types {
		diagnostics = append(diagnostics, validateFieldTypes(symbols, typ.Fields, "type", typ.Name)...)
	}

	// Check RPC proc/stream fields
	for _, rpc := range symbols.rpcs {
		for _, proc := range rpc.Procs {
			if proc.Input != nil {
				diagnostics = append(diagnostics, validateFieldTypes(symbols, proc.Input.Fields, "procedure input", proc.Name)...)
			}
			if proc.Output != nil {
				diagnostics = append(diagnostics, validateFieldTypes(symbols, proc.Output.Fields, "procedure output", proc.Name)...)
			}
		}
		for _, stream := range rpc.Streams {
			if stream.Input != nil {
				diagnostics = append(diagnostics, validateFieldTypes(symbols, stream.Input.Fields, "stream input", stream.Name)...)
			}
			if stream.Output != nil {
				diagnostics = append(diagnostics, validateFieldTypes(symbols, stream.Output.Fields, "stream output", stream.Name)...)
			}
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
		info.MapValue = buildFieldTypeInfo(ft.Base.Map.ValueType)
	} else if ft.Base.Object != nil {
		info.Kind = FieldTypeKindObject
		info.ObjectDef = buildInlineObject(ft.Base.Object)
	}

	return info
}

// buildInlineObject converts an AST FieldTypeObject to an InlineObject.
func buildInlineObject(obj *ast.FieldTypeObject) *InlineObject {
	if obj == nil {
		return nil
	}

	inline := &InlineObject{
		Fields:  []*FieldSymbol{},
		Spreads: []*SpreadRef{},
	}

	for _, child := range obj.Children {
		if child.Field != nil {
			inline.Fields = append(inline.Fields, buildFieldSymbol(child.Field, ""))
		}
		if child.Spread != nil {
			inline.Spreads = append(inline.Spreads, &SpreadRef{
				TypeName: child.Spread.TypeName,
				Pos:      child.Spread.Pos,
				EndPos:   child.Spread.EndPos,
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
			Name:      string(field.Name),
			File:      file,
			Pos:       field.Pos,
			EndPos:    field.EndPos,
			Docstring: docstring,
		},
		AST:      field,
		Optional: field.Optional,
		Type:     buildFieldTypeInfo(&field.Type),
	}
}

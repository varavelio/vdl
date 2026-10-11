package analysis

import (
	"github.com/varavelio/vdl/toolchain/internal/core/ast"
)

// validateAnnotations checks annotation arguments on every declaration.
//
// Argument literals share const literal semantics, so they are validated with
// the same helpers: references must resolve, arrays must be homogeneous,
// spreads must target object constants, and object keys must be unique.
// Without this, unresolvable arguments would silently vanish from the IR.
func validateAnnotations(symbols *symbolTable) []Diagnostic {
	var diagnostics []Diagnostic

	for _, typ := range symbols.types {
		diagnostics = append(
			diagnostics,
			validateAnnotationArgs(symbols, typ.File, typ.Annotations)...,
		)
		if typ.IsObject() && typ.Type.ObjectDef != nil {
			diagnostics = append(
				diagnostics,
				validateFieldAnnotationArgs(
					symbols, typ.File, typ.Type.ObjectDef.Fields,
				)...,
			)
		}
	}

	for _, enum := range symbols.enums {
		diagnostics = append(
			diagnostics,
			validateAnnotationArgs(symbols, enum.File, enum.Annotations)...,
		)
		for _, member := range enum.Members {
			diagnostics = append(
				diagnostics,
				validateAnnotationArgs(
					symbols, enum.File, member.Annotations,
				)...,
			)
		}
	}

	for _, cnst := range symbols.consts {
		diagnostics = append(
			diagnostics,
			validateAnnotationArgs(symbols, cnst.File, cnst.Annotations)...,
		)
	}

	return diagnostics
}

// validateFieldAnnotationArgs checks annotations on fields, recursing into
// inline objects and map values.
func validateFieldAnnotationArgs(
	symbols *symbolTable,
	file string,
	fields []*FieldSymbol,
) []Diagnostic {
	var diagnostics []Diagnostic

	for _, field := range fields {
		diagnostics = append(
			diagnostics,
			validateAnnotationArgs(symbols, field.File, field.Annotations)...,
		)
		diagnostics = append(
			diagnostics,
			validateNestedAnnotationArgs(symbols, file, field.Type)...,
		)
	}

	return diagnostics
}

// validateNestedAnnotationArgs walks nested field types looking for annotated
// inline object fields.
func validateNestedAnnotationArgs(
	symbols *symbolTable,
	file string,
	typeInfo *FieldTypeInfo,
) []Diagnostic {
	if typeInfo == nil {
		return nil
	}

	switch typeInfo.Kind {
	case FieldTypeKindMap:
		return validateNestedAnnotationArgs(symbols, file, typeInfo.MapValue)
	case FieldTypeKindObject:
		if typeInfo.ObjectDef == nil {
			return nil
		}
		return validateFieldAnnotationArgs(
			symbols, file, typeInfo.ObjectDef.Fields,
		)
	}

	return nil
}

// validateAnnotationArgs validates the argument literal of each annotation.
func validateAnnotationArgs(
	symbols *symbolTable,
	file string,
	annotations []*AnnotationRef,
) []Diagnostic {
	var diagnostics []Diagnostic

	for _, ann := range annotations {
		if ann == nil || ann.Argument == nil {
			continue
		}
		diagnostics = append(
			diagnostics,
			validateAnnotationArg(symbols, file, ann.Argument)...,
		)
	}

	return diagnostics
}

// validateAnnotationArg validates a single annotation argument literal.
func validateAnnotationArg(
	symbols *symbolTable,
	file string,
	arg *ast.DataLiteral,
) []Diagnostic {
	var diagnostics []Diagnostic

	valueDiagnostics, _ := validateDataLiteral(
		symbols,
		file,
		arg,
		map[string]bool{},
	)
	diagnostics = append(diagnostics, valueDiagnostics...)
	diagnostics = append(
		diagnostics,
		validateObjectLiteralUniqueness(arg, file)...,
	)

	return diagnostics
}

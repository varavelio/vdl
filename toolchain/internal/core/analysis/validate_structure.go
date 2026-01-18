package analysis

import (
	"fmt"

	"github.com/varavelio/vdl/urpc/internal/core/ast"
)

// validateStructure validates structural requirements:
// - Procedures and streams have at most one input block
// - Procedures and streams have at most one output block
// - Field names are unique within a type/block
func validateStructure(symbols *symbolTable, files map[string]*File) []Diagnostic {
	var diagnostics []Diagnostic

	// Validate type field uniqueness
	for _, typ := range symbols.types {
		diagnostics = append(diagnostics, validateFieldUniqueness(typ.Fields, "type", typ.Name, typ.File)...)
	}

	// Validate proc/stream structure from AST (to check for multiple input/output blocks)
	for _, file := range files {
		for _, rpc := range file.AST.GetRPCs() {
			for _, proc := range rpc.GetProcs() {
				diagnostics = append(diagnostics, validateProcStreamStructure(proc.Children, "procedure", proc.Name, file.Path)...)
			}
			for _, stream := range rpc.GetStreams() {
				diagnostics = append(diagnostics, validateProcStreamStructure(stream.Children, "stream", stream.Name, file.Path)...)
			}
		}
	}

	// Validate proc/stream field uniqueness from symbols
	for _, rpc := range symbols.rpcs {
		for _, proc := range rpc.Procs {
			if proc.Input != nil {
				diagnostics = append(diagnostics, validateFieldUniqueness(proc.Input.Fields, "input of procedure", proc.Name, proc.File)...)
			}
			if proc.Output != nil {
				diagnostics = append(diagnostics, validateFieldUniqueness(proc.Output.Fields, "output of procedure", proc.Name, proc.File)...)
			}
		}
		for _, stream := range rpc.Streams {
			if stream.Input != nil {
				diagnostics = append(diagnostics, validateFieldUniqueness(stream.Input.Fields, "input of stream", stream.Name, stream.File)...)
			}
			if stream.Output != nil {
				diagnostics = append(diagnostics, validateFieldUniqueness(stream.Output.Fields, "output of stream", stream.Name, stream.File)...)
			}
		}
	}

	return diagnostics
}

// validateProcStreamStructure validates that a proc/stream has at most one input and one output block.
func validateProcStreamStructure(children []*ast.ProcOrStreamDeclChild, kind, name, file string) []Diagnostic {
	var diagnostics []Diagnostic

	inputCount := 0
	outputCount := 0
	var firstInputPos, firstOutputPos ast.Position

	for _, child := range children {
		if child.Input != nil {
			inputCount++
			if inputCount == 1 {
				firstInputPos = child.Input.Pos
			} else {
				diagnostics = append(diagnostics, newDiagnostic(
					file,
					child.Input.Pos,
					child.Input.EndPos,
					CodeMultipleInputBlocks,
					fmt.Sprintf("%s %q cannot have more than one input block (first at line %d)",
						kind, name, firstInputPos.Line),
				))
			}
		}
		if child.Output != nil {
			outputCount++
			if outputCount == 1 {
				firstOutputPos = child.Output.Pos
			} else {
				diagnostics = append(diagnostics, newDiagnostic(
					file,
					child.Output.Pos,
					child.Output.EndPos,
					CodeMultipleOutputBlocks,
					fmt.Sprintf("%s %q cannot have more than one output block (first at line %d)",
						kind, name, firstOutputPos.Line),
				))
			}
		}
	}

	return diagnostics
}

// validateFieldUniqueness validates that all field names are unique within a scope.
func validateFieldUniqueness(fields []*FieldSymbol, context, parentName, file string) []Diagnostic {
	var diagnostics []Diagnostic

	fieldNames := make(map[string]ast.Position)
	for _, field := range fields {
		if existing, ok := fieldNames[field.Name]; ok {
			diagnostics = append(diagnostics, newDiagnostic(
				file,
				field.Pos,
				field.EndPos,
				CodeDuplicateField,
				fmt.Sprintf("field %q in %s %q is already declared at line %d",
					field.Name, context, parentName, existing.Line),
			))
		} else {
			fieldNames[field.Name] = field.Pos
		}

		// Recursively check inline object fields
		if field.Type != nil && field.Type.Kind == FieldTypeKindObject && field.Type.ObjectDef != nil {
			diagnostics = append(diagnostics, validateFieldUniqueness(
				field.Type.ObjectDef.Fields,
				"inline object in",
				field.Name,
				file,
			)...)
		}
	}

	return diagnostics
}

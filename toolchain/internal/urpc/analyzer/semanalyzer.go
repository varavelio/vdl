package analyzer

import (
	"fmt"
	"strings"

	"slices"

	"github.com/uforg/uforpc/urpc/internal/urpc/ast"
	"github.com/uforg/uforpc/urpc/internal/util/strutil"
)

// semanalyzer is the semantic alyzer phase for the URPC schema analyzer.
//
// It performs the following checks:
//   - Custom type names are unique and valid.
//   - Custom procedure names are unique and valid.
//   - All referenced types exist.
type semanalyzer struct {
	astSchema   *ast.Schema
	diagnostics []Diagnostic
}

// newSemanalyzer creates a new Semanalyzer for the given URPC schema. See more
// details in the Semanalyzer struct.
func newSemanalyzer(astSchema *ast.Schema) *semanalyzer {
	return &semanalyzer{
		astSchema:   astSchema,
		diagnostics: []Diagnostic{},
	}
}

// Analyze analyzes the provided URPC schema.
//
// Returns:
//   - A list of diagnostics that occurred during the analysis.
//   - The first diagnostic converted to Error interface if any.
func (a *semanalyzer) analyze() ([]Diagnostic, error) {
	a.validateUniqueResourceNames()
	a.validateCustomTypeReferences()
	a.validateTypeFieldUniqueness()
	a.validateTypeCircularDependencies()
	a.validateProcStructure()
	a.validateStreamStructure()

	if len(a.diagnostics) > 0 {
		return a.diagnostics, a.diagnostics[0]
	}
	return nil, nil
}

// validateUniqueResourceNames validates the types, procedures and streams names and detects duplicates
// between them.
func (a *semanalyzer) validateUniqueResourceNames() {
	visited := map[string]Positions{}

	for _, typeDecl := range a.astSchema.GetTypes() {
		positions := Positions(typeDecl.Positions)
		typeName := typeDecl.Name

		if decl, isDecl := visited[typeName]; isDecl {
			a.diagnostics = append(a.diagnostics, Diagnostic{
				Positions: positions,
				Message:   fmt.Sprintf("type name \"%s\" is not unique, it is already declared at %s", typeName, decl.Pos.String()),
			})
			continue
		}
		visited[typeName] = positions

		if !strutil.IsPascalCase(typeName) {
			a.diagnostics = append(a.diagnostics, Diagnostic{
				Positions: positions,
				Message:   fmt.Sprintf("type name \"%s\" must be in PascalCase", typeName),
			})
			continue
		}
	}

	for _, procDecl := range a.astSchema.GetProcs() {
		positions := Positions(procDecl.Positions)
		procName := procDecl.Name

		if decl, isDecl := visited[procName]; isDecl {
			a.diagnostics = append(a.diagnostics, Diagnostic{
				Positions: positions,
				Message:   fmt.Sprintf("procedure name \"%s\" is not unique, it is already declared at %s", procName, decl.Pos.String()),
			})
			continue
		}
		visited[procName] = positions

		if !strutil.IsPascalCase(procName) {
			a.diagnostics = append(a.diagnostics, Diagnostic{
				Positions: positions,
				Message:   fmt.Sprintf("procedure name \"%s\" must be in PascalCase", procName),
			})
			continue
		}
	}

	for _, streamDecl := range a.astSchema.GetStreams() {
		positions := Positions(streamDecl.Positions)
		streamName := streamDecl.Name

		if decl, isDecl := visited[streamName]; isDecl {
			a.diagnostics = append(a.diagnostics, Diagnostic{
				Positions: positions,
				Message:   fmt.Sprintf("stream name \"%s\" is not unique, it is already declared at %s", streamName, decl.Pos.String()),
			})
			continue
		}
		visited[streamName] = positions

		if !strutil.IsPascalCase(streamName) {
			a.diagnostics = append(a.diagnostics, Diagnostic{
				Positions: positions,
				Message:   fmt.Sprintf("stream name \"%s\" must be in PascalCase", streamName),
			})
			continue
		}
	}
}

// validateCustomTypeReferences validates that all referenced custom types exist.
func (a *semanalyzer) validateCustomTypeReferences() {
	isValidType := func(typeName string) bool {
		if ast.IsPrimitiveType(typeName) {
			return true
		}

		for _, typeDecl := range a.astSchema.GetTypes() {
			if typeDecl.Name == typeName {
				return true
			}
		}

		return false
	}

	var checkFieldTypeReferences func([]*ast.Field, string)
	checkFieldTypeReferences = func(fields []*ast.Field, context string) {
		for _, field := range fields {
			if field.Type.Base.Named != nil {
				typeName := *field.Type.Base.Named

				if !isValidType(typeName) {
					a.diagnostics = append(a.diagnostics, Diagnostic{
						Positions: Positions{
							Pos:    field.Type.Pos,
							EndPos: field.Type.EndPos,
						},
						Message: fmt.Sprintf("type \"%s\" referenced %s is not declared", typeName, context),
					})
				}
			} else if field.Type.Base.Object != nil {
				// Extract fields from inline object and recursively check them
				inlineFields := extractFields(field.Type.Base.Object.Children)
				checkFieldTypeReferences(inlineFields, fmt.Sprintf("at inline object at field \"%s\"", field.Name))
			}
		}
	}

	// Check type declarations
	for _, typeDecl := range a.astSchema.GetTypes() {
		// Check fields
		typeFields := extractFields(typeDecl.Children)
		checkFieldTypeReferences(typeFields, fmt.Sprintf("at type \"%s\"", typeDecl.Name))
	}

	// Check procedure declarations
	for _, proc := range a.astSchema.GetProcs() {
		for _, child := range proc.Children {
			// Check input fields
			if child.Input != nil {
				inputFields := extractFields(child.Input.Children)
				checkFieldTypeReferences(inputFields, fmt.Sprintf("at input of procedure \"%s\"", proc.Name))
			}

			// Check output fields
			if child.Output != nil {
				outputFields := extractFields(child.Output.Children)
				checkFieldTypeReferences(outputFields, fmt.Sprintf("at output of procedure \"%s\"", proc.Name))
			}
		}
	}

	// Check stream declarations
	for _, stream := range a.astSchema.GetStreams() {
		for _, child := range stream.Children {
			// Check input fields
			if child.Input != nil {
				inputFields := extractFields(child.Input.Children)
				checkFieldTypeReferences(inputFields, fmt.Sprintf("at input of stream \"%s\"", stream.Name))
			}

			// Check output fields
			if child.Output != nil {
				outputFields := extractFields(child.Output.Children)
				checkFieldTypeReferences(outputFields, fmt.Sprintf("at output of stream \"%s\"", stream.Name))
			}
		}
	}
}

// Helper function to extract fields from FieldOrComment array
func extractFields(fieldOrComments []*ast.FieldOrComment) []*ast.Field {
	var fields []*ast.Field
	for _, foc := range fieldOrComments {
		if foc.Field != nil {
			fields = append(fields, foc.Field)
		}
	}
	return fields
}

// validateTypeFieldUniqueness validates that fields in a type (including extended types) are unique.
func (a *semanalyzer) validateTypeFieldUniqueness() {
	for _, typeDecl := range a.astSchema.GetTypes() {
		// Collect all fields from this type and its extensions
		allFields := make(map[string]Positions)

		// First collect fields from the type itself
		fields := extractFields(typeDecl.Children)
		for _, field := range fields {
			// Check if this field already exists
			if existingPos, exists := allFields[field.Name]; exists {
				a.diagnostics = append(a.diagnostics, Diagnostic{
					Positions: Positions{
						Pos:    typeDecl.Pos,
						EndPos: typeDecl.EndPos,
					},
					Message: fmt.Sprintf(
						"field \"%s\" in type \"%s\" is already defined at %s",
						field.Name, typeDecl.Name, existingPos.Pos.String(),
					),
				})
			} else {
				// Add the field to the map
				allFields[field.Name] = Positions{
					Pos:    field.Pos,
					EndPos: field.EndPos,
				}
			}
		}
	}
}

// validateTypeCircularDependencies validates that there are no circular dependencies between types.
func (a *semanalyzer) validateTypeCircularDependencies() {
	types := a.astSchema.GetTypesMap()
	for name, typeDecl := range types {
		if err := validateTypeCircularDependenciesCheckType(name, types, []string{}); err != nil {
			a.diagnostics = append(a.diagnostics, Diagnostic{
				Positions: Positions{
					Pos:    typeDecl.Pos,
					EndPos: typeDecl.EndPos,
				},
				Message: err.Error(),
			})
		}
	}
}

// validateTypeCircularDependenciesCheckType checks if a type has a circular dependency.
func validateTypeCircularDependenciesCheckType(name string, types map[string]*ast.TypeDecl, stack []string) error {
	// Is it already in the stack (cycle)?
	if slices.Contains(stack, name) {
		return fmt.Errorf("circular dependency detected between types: %s", strings.Join(stack, " -> "))
	}

	// Ensure the type exists before proceeding to avoid nil pointer dereference.
	typ, ok := types[name]
	if !ok || typ == nil {
		// The existence of the type is validated elsewhere, so we silently skip
		// it here to prevent a crash.
		return nil
	}

	// Add it to the stack
	stack = append(stack, name)

	// Check every field in the type (including nested types)
	for _, field := range extractFields(typ.Children) {
		if err := validateTypeCircularDependenciesCheckField(field.Type, types, stack); err != nil {
			return err
		}
	}

	return nil
}

// validateTypeCircularDependenciesCheckField checks if a field has a circular dependency.
func validateTypeCircularDependenciesCheckField(fieldType ast.FieldType, types map[string]*ast.TypeDecl, stack []string) error {
	// If it's a custom named type, check it
	if fieldType.Base.Named != nil {
		typeName := *fieldType.Base.Named
		if !ast.IsPrimitiveType(typeName) {
			return validateTypeCircularDependenciesCheckType(typeName, types, stack)
		}
	}

	// If it's an inline object, check all its fields
	if fieldType.Base.Object != nil {
		objectFields := extractFields(fieldType.Base.Object.Children)
		for _, field := range objectFields {
			if err := validateTypeCircularDependenciesCheckField(field.Type, types, stack); err != nil {
				return err
			}
		}
	}

	return nil
}

// validateProcStructure validates that procedure declarations have the correct structure:
// - At most one 'input' section
// - At most one 'output' section
func (a *semanalyzer) validateProcStructure() {
	for _, procDecl := range a.astSchema.GetProcs() {
		inputCount := 0
		outputCount := 0

		// Count the number of each section
		for _, child := range procDecl.Children {
			if child.Input != nil {
				inputCount++
			}
			if child.Output != nil {
				outputCount++
			}
		}

		// Validate 'input' section
		if inputCount > 1 {
			a.diagnostics = append(a.diagnostics, Diagnostic{
				Positions: Positions{
					Pos:    procDecl.Pos,
					EndPos: procDecl.EndPos,
				},
				Message: fmt.Sprintf("procedure \"%s\" cannot have more than one 'input' section", procDecl.Name),
			})
		}

		// Validate 'output' section
		if outputCount > 1 {
			a.diagnostics = append(a.diagnostics, Diagnostic{
				Positions: Positions{
					Pos:    procDecl.Pos,
					EndPos: procDecl.EndPos,
				},
				Message: fmt.Sprintf("procedure \"%s\" cannot have more than one 'output' section", procDecl.Name),
			})
		}
	}
}

// validateStreamStructure validates that stream declarations have the correct structure:
// - At most one 'input' section
// - At most one 'output' section
func (a *semanalyzer) validateStreamStructure() {
	for _, streamDecl := range a.astSchema.GetStreams() {
		inputCount := 0
		outputCount := 0

		// Count the number of each section
		for _, child := range streamDecl.Children {
			if child.Input != nil {
				inputCount++
			}
			if child.Output != nil {
				outputCount++
			}
		}

		// Validate 'input' section
		if inputCount > 1 {
			a.diagnostics = append(a.diagnostics, Diagnostic{
				Positions: Positions{
					Pos:    streamDecl.Pos,
					EndPos: streamDecl.EndPos,
				},
				Message: fmt.Sprintf("stream \"%s\" cannot have more than one 'input' section", streamDecl.Name),
			})
		}

		// Validate 'output' section
		if outputCount > 1 {
			a.diagnostics = append(a.diagnostics, Diagnostic{
				Positions: Positions{
					Pos:    streamDecl.Pos,
					EndPos: streamDecl.EndPos,
				},
				Message: fmt.Sprintf("stream \"%s\" cannot have more than one 'output' section", streamDecl.Name),
			})
		}
	}
}

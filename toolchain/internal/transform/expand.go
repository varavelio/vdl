package transform

import (
	"fmt"
	"strings"

	"github.com/varavelio/vdl/toolchain/internal/core/ast"
	"github.com/varavelio/vdl/toolchain/internal/core/parser"
	"github.com/varavelio/vdl/toolchain/internal/formatter"
)

// ExpandTypesStr expands all custom type references in the VDL schema to inline objects.
// It takes a VDL schema as a string and returns the expanded schema as a formatted string.
func ExpandTypesStr(filename, content string) (string, error) {
	if strings.TrimSpace(content) == "" {
		return "", nil
	}

	schema, err := parser.ParserInstance.ParseString(filename, content)
	if err != nil {
		return "", fmt.Errorf("error parsing VDL: %w", err)
	}

	expanded := ExpandTypes(schema)

	return formatter.FormatSchema(expanded), nil
}

// ExpandTypes returns a new schema with all custom type references expanded to inline objects.
// It traverses RPCs (containing procs and streams), and type declarations, replacing references
// to custom types with their full inline definitions.
// It also flattens spread operators (...TypeName) by inlining the fields from the referenced type.
// Circular references are handled by stopping expansion and keeping the type as a named reference.
func ExpandTypes(schema *ast.Schema) *ast.Schema {
	typesMap := schema.GetTypesMap()

	expanded := &ast.Schema{
		Positions: schema.Positions,
		Children:  make([]*ast.SchemaChild, 0, len(schema.Children)),
	}

	for _, child := range schema.Children {
		expandedChild := expandSchemaChild(child, typesMap)
		expanded.Children = append(expanded.Children, expandedChild)
	}

	return expanded
}

func expandSchemaChild(child *ast.SchemaChild, typesMap map[string]*ast.TypeDecl) *ast.SchemaChild {
	expanded := &ast.SchemaChild{
		Positions: child.Positions,
		Include:   child.Include,
		Comment:   child.Comment,
		Docstring: child.Docstring,
		Const:     child.Const,
		Enum:      child.Enum,
		Pattern:   child.Pattern,
	}

	if child.Type != nil {
		// Create a fresh visited set for each type declaration
		visited := make(map[string]bool)
		expanded.Type = expandTypeDecl(child.Type, typesMap, visited)
	}

	if child.RPC != nil {
		expanded.RPC = expandRPCDecl(child.RPC, typesMap)
	}

	return expanded
}

func expandTypeDecl(typeDecl *ast.TypeDecl, typesMap map[string]*ast.TypeDecl, visited map[string]bool) *ast.TypeDecl {
	// Mark this type as being visited to detect circular references
	visited[typeDecl.Name] = true
	defer func() { visited[typeDecl.Name] = false }()

	return &ast.TypeDecl{
		Positions:  typeDecl.Positions,
		Docstring:  typeDecl.Docstring,
		Deprecated: typeDecl.Deprecated,
		Name:       typeDecl.Name,
		Children:   expandTypeDeclChildren(typeDecl.Children, typesMap, visited),
	}
}

func expandRPCDecl(rpc *ast.RPCDecl, typesMap map[string]*ast.TypeDecl) *ast.RPCDecl {
	children := make([]*ast.RPCChild, 0, len(rpc.Children))

	for _, child := range rpc.Children {
		expanded := &ast.RPCChild{
			Positions: child.Positions,
			Comment:   child.Comment,
			Docstring: child.Docstring,
		}

		if child.Proc != nil {
			expanded.Proc = expandProcDecl(child.Proc, typesMap)
		}

		if child.Stream != nil {
			expanded.Stream = expandStreamDecl(child.Stream, typesMap)
		}

		children = append(children, expanded)
	}

	return &ast.RPCDecl{
		Positions:  rpc.Positions,
		Docstring:  rpc.Docstring,
		Deprecated: rpc.Deprecated,
		Name:       rpc.Name,
		Children:   children,
	}
}

func expandProcDecl(procDecl *ast.ProcDecl, typesMap map[string]*ast.TypeDecl) *ast.ProcDecl {
	children := make([]*ast.ProcOrStreamDeclChild, 0, len(procDecl.Children))

	for _, child := range procDecl.Children {
		expanded := &ast.ProcOrStreamDeclChild{
			Positions: child.Positions,
			Comment:   child.Comment,
		}

		if child.Input != nil {
			// Create a fresh visited set for input block
			visited := make(map[string]bool)
			expanded.Input = &ast.ProcOrStreamDeclChildInput{
				Positions: child.Input.Positions,
				Children:  expandInputOutputChildren(child.Input.Children, typesMap, visited),
			}
		}

		if child.Output != nil {
			// Create a fresh visited set for output block
			visited := make(map[string]bool)
			expanded.Output = &ast.ProcOrStreamDeclChildOutput{
				Positions: child.Output.Positions,
				Children:  expandInputOutputChildren(child.Output.Children, typesMap, visited),
			}
		}

		children = append(children, expanded)
	}

	return &ast.ProcDecl{
		Positions:  procDecl.Positions,
		Docstring:  procDecl.Docstring,
		Deprecated: procDecl.Deprecated,
		Name:       procDecl.Name,
		Children:   children,
	}
}

func expandStreamDecl(streamDecl *ast.StreamDecl, typesMap map[string]*ast.TypeDecl) *ast.StreamDecl {
	children := make([]*ast.ProcOrStreamDeclChild, 0, len(streamDecl.Children))

	for _, child := range streamDecl.Children {
		expanded := &ast.ProcOrStreamDeclChild{
			Positions: child.Positions,
			Comment:   child.Comment,
		}

		if child.Input != nil {
			// Create a fresh visited set for input block
			visited := make(map[string]bool)
			expanded.Input = &ast.ProcOrStreamDeclChildInput{
				Positions: child.Input.Positions,
				Children:  expandInputOutputChildren(child.Input.Children, typesMap, visited),
			}
		}

		if child.Output != nil {
			// Create a fresh visited set for output block
			visited := make(map[string]bool)
			expanded.Output = &ast.ProcOrStreamDeclChildOutput{
				Positions: child.Output.Positions,
				Children:  expandInputOutputChildren(child.Output.Children, typesMap, visited),
			}
		}

		children = append(children, expanded)
	}

	return &ast.StreamDecl{
		Positions:  streamDecl.Positions,
		Docstring:  streamDecl.Docstring,
		Deprecated: streamDecl.Deprecated,
		Name:       streamDecl.Name,
		Children:   children,
	}
}

// expandTypeDeclChildren expands children of a type declaration, including flattening spreads.
func expandTypeDeclChildren(children []*ast.TypeDeclChild, typesMap map[string]*ast.TypeDecl, visited map[string]bool) []*ast.TypeDeclChild {
	expanded := make([]*ast.TypeDeclChild, 0, len(children))

	for _, child := range children {
		// Handle spread: flatten the referenced type's fields into this type
		if child.Spread != nil {
			spreadTypeName := child.Spread.TypeName

			// Check for circular reference
			if visited[spreadTypeName] {
				// Circular reference detected - skip this spread to avoid infinite loop
				// We could optionally keep the spread as-is or log a warning
				continue
			}

			// Look up the spread type
			if spreadType, exists := typesMap[spreadTypeName]; exists {
				// Mark as visited before recursing
				visited[spreadTypeName] = true

				// Recursively expand the spread type's children
				spreadChildren := expandTypeDeclChildren(spreadType.Children, typesMap, visited)

				// Unmark after processing (backtracking)
				visited[spreadTypeName] = false

				// Add all the spread type's children to our expanded list
				expanded = append(expanded, spreadChildren...)
			}
			// If spread type not found, silently skip (semantic analyzer would catch this)
			continue
		}

		// Handle regular fields and comments
		newChild := &ast.TypeDeclChild{
			Positions: child.Positions,
			Comment:   child.Comment,
		}

		if child.Field != nil {
			newChild.Field = expandField(child.Field, typesMap, visited)
		}

		expanded = append(expanded, newChild)
	}

	return expanded
}

// expandInputOutputChildren expands children of input/output blocks, including flattening spreads.
func expandInputOutputChildren(children []*ast.InputOutputChild, typesMap map[string]*ast.TypeDecl, visited map[string]bool) []*ast.InputOutputChild {
	expanded := make([]*ast.InputOutputChild, 0, len(children))

	for _, child := range children {
		// Handle spread: flatten the referenced type's fields
		if child.Spread != nil {
			spreadTypeName := child.Spread.TypeName

			// Check for circular reference
			if visited[spreadTypeName] {
				// Circular reference detected - skip this spread
				continue
			}

			// Look up the spread type
			if spreadType, exists := typesMap[spreadTypeName]; exists {
				// Mark as visited before recursing
				visited[spreadTypeName] = true

				// Expand the spread type's children (TypeDeclChild -> InputOutputChild conversion)
				spreadTypeChildren := expandTypeDeclChildren(spreadType.Children, typesMap, visited)

				// Unmark after processing (backtracking)
				visited[spreadTypeName] = false

				// Convert TypeDeclChild to InputOutputChild
				for _, tc := range spreadTypeChildren {
					ioChild := &ast.InputOutputChild{
						Positions: tc.Positions,
						Comment:   tc.Comment,
						Field:     tc.Field,
						// Note: tc.Spread should already be nil after expandTypeDeclChildren
					}
					expanded = append(expanded, ioChild)
				}
			}
			continue
		}

		// Handle regular fields and comments
		newChild := &ast.InputOutputChild{
			Positions: child.Positions,
			Comment:   child.Comment,
		}

		if child.Field != nil {
			newChild.Field = expandField(child.Field, typesMap, visited)
		}

		expanded = append(expanded, newChild)
	}

	return expanded
}

func expandField(field *ast.Field, typesMap map[string]*ast.TypeDecl, visited map[string]bool) *ast.Field {
	return &ast.Field{
		Positions: field.Positions,
		Docstring: field.Docstring,
		Name:      field.Name,
		Optional:  field.Optional,
		Type:      expandFieldType(field.Type, typesMap, visited),
	}
}

func expandFieldType(fieldType ast.FieldType, typesMap map[string]*ast.TypeDecl, visited map[string]bool) ast.FieldType {
	return ast.FieldType{
		Positions:  fieldType.Positions,
		Base:       expandFieldTypeBase(fieldType.Base, typesMap, visited),
		Dimensions: fieldType.Dimensions,
	}
}

func expandFieldTypeBase(base *ast.FieldTypeBase, typesMap map[string]*ast.TypeDecl, visited map[string]bool) *ast.FieldTypeBase {
	if base.Named != nil {
		typeName := *base.Named

		// If it's a custom type (not primitive), try to expand it
		if !ast.IsPrimitiveType(typeName) {
			// Check for circular reference - if we're already expanding this type, keep it as named
			if visited[typeName] {
				// Circular reference detected - keep as named reference to break the cycle
				return &ast.FieldTypeBase{
					Positions: base.Positions,
					Named:     base.Named,
				}
			}

			// Type exists, expand it
			if typeDecl, exists := typesMap[typeName]; exists {
				// Mark as visited before recursing
				visited[typeName] = true

				// Convert to inline object with expanded children
				result := &ast.FieldTypeBase{
					Positions: base.Positions,
					Object:    typeToInlineObject(typeDecl, typesMap, visited),
				}

				// Unmark after processing (backtracking)
				visited[typeName] = false

				return result
			}
		}

		// Keep primitive types or unknown types as named
		return &ast.FieldTypeBase{
			Positions: base.Positions,
			Named:     base.Named,
		}
	}

	if base.Map != nil {
		// Recursively expand the map's value type
		return &ast.FieldTypeBase{
			Positions: base.Positions,
			Map: &ast.FieldTypeMap{
				Positions: base.Map.Positions,
				ValueType: expandFieldTypePtr(base.Map.ValueType, typesMap, visited),
			},
		}
	}

	if base.Object != nil {
		// Recursively expand fields within inline objects
		return &ast.FieldTypeBase{
			Positions: base.Positions,
			Object: &ast.FieldTypeObject{
				Positions: base.Object.Positions,
				Children:  expandTypeDeclChildren(base.Object.Children, typesMap, visited),
			},
		}
	}

	return base
}

func expandFieldTypePtr(fieldType *ast.FieldType, typesMap map[string]*ast.TypeDecl, visited map[string]bool) *ast.FieldType {
	if fieldType == nil {
		return nil
	}
	expanded := expandFieldType(*fieldType, typesMap, visited)
	return &expanded
}

func typeToInlineObject(typeDecl *ast.TypeDecl, typesMap map[string]*ast.TypeDecl, visited map[string]bool) *ast.FieldTypeObject {
	return &ast.FieldTypeObject{
		Positions: typeDecl.Positions,
		Children:  expandTypeDeclChildren(typeDecl.Children, typesMap, visited),
	}
}

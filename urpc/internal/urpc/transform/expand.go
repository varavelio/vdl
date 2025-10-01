package transform

import (
	"github.com/uforg/uforpc/urpc/internal/urpc/ast"
)

// ExpandTypes returns a new schema with all custom type references expanded to inline objects.
// It traverses procs, streams, and type declarations, replacing references to custom types
// with their full inline definitions.
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
		Version:   child.Version,
		Comment:   child.Comment,
		Docstring: child.Docstring,
	}

	if child.Type != nil {
		expanded.Type = expandTypeDecl(child.Type, typesMap)
	}

	if child.Proc != nil {
		expanded.Proc = expandProcDecl(child.Proc, typesMap)
	}

	if child.Stream != nil {
		expanded.Stream = expandStreamDecl(child.Stream, typesMap)
	}

	return expanded
}

func expandTypeDecl(typeDecl *ast.TypeDecl, typesMap map[string]*ast.TypeDecl) *ast.TypeDecl {
	return &ast.TypeDecl{
		Positions:  typeDecl.Positions,
		Docstring:  typeDecl.Docstring,
		Deprecated: typeDecl.Deprecated,
		Name:       typeDecl.Name,
		Children:   expandFieldOrComments(typeDecl.Children, typesMap),
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
			expanded.Input = &ast.ProcOrStreamDeclChildInput{
				Positions: child.Input.Positions,
				Children:  expandFieldOrComments(child.Input.Children, typesMap),
			}
		}

		if child.Output != nil {
			expanded.Output = &ast.ProcOrStreamDeclChildOutput{
				Positions: child.Output.Positions,
				Children:  expandFieldOrComments(child.Output.Children, typesMap),
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
			expanded.Input = &ast.ProcOrStreamDeclChildInput{
				Positions: child.Input.Positions,
				Children:  expandFieldOrComments(child.Input.Children, typesMap),
			}
		}

		if child.Output != nil {
			expanded.Output = &ast.ProcOrStreamDeclChildOutput{
				Positions: child.Output.Positions,
				Children:  expandFieldOrComments(child.Output.Children, typesMap),
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

func expandFieldOrComments(children []*ast.FieldOrComment, typesMap map[string]*ast.TypeDecl) []*ast.FieldOrComment {
	expanded := make([]*ast.FieldOrComment, 0, len(children))

	for _, child := range children {
		newChild := &ast.FieldOrComment{
			Positions: child.Positions,
			Comment:   child.Comment,
		}

		if child.Field != nil {
			newChild.Field = expandField(child.Field, typesMap)
		}

		expanded = append(expanded, newChild)
	}

	return expanded
}

func expandField(field *ast.Field, typesMap map[string]*ast.TypeDecl) *ast.Field {
	return &ast.Field{
		Positions: field.Positions,
		Docstring: field.Docstring,
		Name:      field.Name,
		Optional:  field.Optional,
		Type:      expandFieldType(field.Type, typesMap),
	}
}

func expandFieldType(fieldType ast.FieldType, typesMap map[string]*ast.TypeDecl) ast.FieldType {
	return ast.FieldType{
		Positions: fieldType.Positions,
		Base:      expandFieldTypeBase(fieldType.Base, typesMap),
		IsArray:   fieldType.IsArray,
	}
}

func expandFieldTypeBase(base *ast.FieldTypeBase, typesMap map[string]*ast.TypeDecl) *ast.FieldTypeBase {
	if base.Named != nil {
		typeName := *base.Named

		// If it's a custom type (not primitive), expand it to inline object
		if !ast.IsPrimitiveType(typeName) {
			if typeDecl, exists := typesMap[typeName]; exists {
				return &ast.FieldTypeBase{
					Positions: base.Positions,
					Object:    typeToInlineObject(typeDecl, typesMap),
				}
			}
		}

		// Keep primitive types as named
		return &ast.FieldTypeBase{
			Positions: base.Positions,
			Named:     base.Named,
		}
	}

	if base.Object != nil {
		// Recursively expand fields within inline objects
		return &ast.FieldTypeBase{
			Positions: base.Positions,
			Object: &ast.FieldTypeObject{
				Positions: base.Object.Positions,
				Children:  expandFieldOrComments(base.Object.Children, typesMap),
			},
		}
	}

	return base
}

func typeToInlineObject(typeDecl *ast.TypeDecl, typesMap map[string]*ast.TypeDecl) *ast.FieldTypeObject {
	return &ast.FieldTypeObject{
		Positions: typeDecl.Positions,
		Children:  expandFieldOrComments(typeDecl.Children, typesMap),
	}
}

package transpile

import (
	"fmt"

	"github.com/uforg/uforpc/urpc/internal/schema"
	"github.com/uforg/uforpc/urpc/internal/urpc/ast"
)

// ToURPC transpiles an UFO-RPC JSON schema to it's AST representation.
//
// The resulting AST Schema will not include any imports, extends, external
// docstrings, comments nor comment blocks.
//
// To get the string representation of the AST Schema, you can use the
// formatter package.
func ToURPC(jsonSchema schema.Schema) (ast.Schema, error) {
	result := ast.Schema{}

	// Add version declaration first
	if jsonSchema.Version > 0 {
		result.Children = append(result.Children, &ast.SchemaChild{
			Version: &ast.Version{
				Number: jsonSchema.Version,
			},
		})
	}

	// Process all nodes in the order they appear in the JSON schema
	for _, node := range jsonSchema.Nodes {
		switch n := node.(type) {
		case *schema.NodeDoc:
			// Convert standalone documentation to docstring
			result.Children = append(result.Children, &ast.SchemaChild{
				Docstring: &ast.Docstring{
					Value: n.Content,
				},
			})
		case *schema.NodeType:
			typeDecl, err := convertTypeToURPC(n)
			if err != nil {
				return ast.Schema{}, fmt.Errorf("error converting type '%s': %w", n.Name, err)
			}
			result.Children = append(result.Children, &ast.SchemaChild{
				Type: typeDecl,
			})
		case *schema.NodeProc:
			procDecl, err := convertProcToURPC(n)
			if err != nil {
				return ast.Schema{}, fmt.Errorf("error converting procedure '%s': %w", n.Name, err)
			}
			result.Children = append(result.Children, &ast.SchemaChild{
				Proc: procDecl,
			})
		case *schema.NodeStream:
			streamDecl, err := convertStreamToURPC(n)
			if err != nil {
				return ast.Schema{}, fmt.Errorf("error converting stream '%s': %w", n.Name, err)
			}
			result.Children = append(result.Children, &ast.SchemaChild{
				Stream: streamDecl,
			})
		}
	}

	return result, nil
}

// convertTypeToURPC converts a schema NodeType to an AST TypeDecl
func convertTypeToURPC(typeNode *schema.NodeType) (*ast.TypeDecl, error) {
	typeDecl := &ast.TypeDecl{
		Name: typeNode.Name,
	}

	// Add docstring if available
	if typeNode.Doc != nil && *typeNode.Doc != "" {
		typeDecl.Docstring = &ast.Docstring{
			Value: *typeNode.Doc,
		}
	}

	// Add deprecated if available
	if typeNode.Deprecated != nil {
		deprecated := &ast.Deprecated{}
		if *typeNode.Deprecated != "" {
			deprecated.Message = typeNode.Deprecated
		}
		typeDecl.Deprecated = deprecated
	}

	// Process fields
	for _, field := range typeNode.Fields {
		fieldNode, err := convertFieldToURPC(field)
		if err != nil {
			return nil, fmt.Errorf("error converting field '%s': %w", field.Name, err)
		}

		typeDecl.Children = append(typeDecl.Children, &ast.FieldOrComment{
			Field: fieldNode,
		})
	}

	return typeDecl, nil
}

// convertFieldToURPC converts a schema FieldDefinition to an AST Field
func convertFieldToURPC(fieldDef schema.FieldDefinition) (*ast.Field, error) {
	field := &ast.Field{
		Name:     fieldDef.Name,
		Optional: fieldDef.Optional,
	}

	// Add docstring if available
	if fieldDef.Doc != nil && *fieldDef.Doc != "" {
		field.Docstring = &ast.Docstring{
			Value: *fieldDef.Doc,
		}
	}

	// Process field type
	fieldType := ast.FieldType{
		IsArray: fieldDef.IsArray,
		Base:    &ast.FieldTypeBase{},
	}

	if fieldDef.IsNamed() {
		fieldType.Base.Named = fieldDef.TypeName
	}

	if fieldDef.IsInline() {
		object := &ast.FieldTypeObject{}

		// Process inline object fields
		for _, inlineField := range fieldDef.TypeInline.Fields {
			inlineFieldNode, err := convertFieldToURPC(inlineField)
			if err != nil {
				return nil, fmt.Errorf("error converting inline field '%s': %w", inlineField.Name, err)
			}

			object.Children = append(object.Children, &ast.FieldOrComment{
				Field: inlineFieldNode,
			})
		}

		fieldType.Base.Object = object
	}

	field.Type = fieldType

	return field, nil
}

// convertProcToURPC converts a schema NodeProc to an AST ProcDecl
func convertProcToURPC(procNode *schema.NodeProc) (*ast.ProcDecl, error) {
	procDecl := &ast.ProcDecl{
		Name: procNode.Name,
	}

	// Add docstring if available
	if procNode.Doc != nil && *procNode.Doc != "" {
		procDecl.Docstring = &ast.Docstring{
			Value: *procNode.Doc,
		}
	}

	// Add deprecated if available
	if procNode.Deprecated != nil {
		deprecated := &ast.Deprecated{}
		if *procNode.Deprecated != "" {
			deprecated.Message = procNode.Deprecated
		}
		procDecl.Deprecated = deprecated
	}

	// Process input fields if any
	if len(procNode.Input) > 0 {
		inputChild := &ast.ProcOrStreamDeclChildInput{}

		for _, field := range procNode.Input {
			fieldNode, err := convertFieldToURPC(field)
			if err != nil {
				return nil, fmt.Errorf("error converting input field '%s': %w", field.Name, err)
			}

			inputChild.Children = append(inputChild.Children, &ast.FieldOrComment{
				Field: fieldNode,
			})
		}

		procDecl.Children = append(procDecl.Children, &ast.ProcOrStreamDeclChild{
			Input: inputChild,
		})
	}

	// Process output fields if any
	if len(procNode.Output) > 0 {
		outputChild := &ast.ProcOrStreamDeclChildOutput{}

		for _, field := range procNode.Output {
			fieldNode, err := convertFieldToURPC(field)
			if err != nil {
				return nil, fmt.Errorf("error converting output field '%s': %w", field.Name, err)
			}

			outputChild.Children = append(outputChild.Children, &ast.FieldOrComment{
				Field: fieldNode,
			})
		}

		procDecl.Children = append(procDecl.Children, &ast.ProcOrStreamDeclChild{
			Output: outputChild,
		})
	}

	return procDecl, nil
}

// convertStreamToURPC converts a schema NodeStream to an AST StreamDecl
func convertStreamToURPC(streamNode *schema.NodeStream) (*ast.StreamDecl, error) {
	streamDecl := &ast.StreamDecl{
		Name: streamNode.Name,
	}

	// Add docstring if available
	if streamNode.Doc != nil && *streamNode.Doc != "" {
		streamDecl.Docstring = &ast.Docstring{
			Value: *streamNode.Doc,
		}
	}

	// Add deprecated if available
	if streamNode.Deprecated != nil {
		deprecated := &ast.Deprecated{}
		if *streamNode.Deprecated != "" {
			deprecated.Message = streamNode.Deprecated
		}
		streamDecl.Deprecated = deprecated
	}

	// Process input fields if any
	if len(streamNode.Input) > 0 {
		inputChild := &ast.ProcOrStreamDeclChildInput{}

		for _, field := range streamNode.Input {
			fieldNode, err := convertFieldToURPC(field)
			if err != nil {
				return nil, fmt.Errorf("error converting input field '%s': %w", field.Name, err)
			}

			inputChild.Children = append(inputChild.Children, &ast.FieldOrComment{
				Field: fieldNode,
			})
		}

		streamDecl.Children = append(streamDecl.Children, &ast.ProcOrStreamDeclChild{
			Input: inputChild,
		})
	}

	// Process output fields if any
	if len(streamNode.Output) > 0 {
		outputChild := &ast.ProcOrStreamDeclChildOutput{}

		for _, field := range streamNode.Output {
			fieldNode, err := convertFieldToURPC(field)
			if err != nil {
				return nil, fmt.Errorf("error converting output field '%s': %w", field.Name, err)
			}

			outputChild.Children = append(outputChild.Children, &ast.FieldOrComment{
				Field: fieldNode,
			})
		}

		streamDecl.Children = append(streamDecl.Children, &ast.ProcOrStreamDeclChild{
			Output: outputChild,
		})
	}

	return streamDecl, nil
}

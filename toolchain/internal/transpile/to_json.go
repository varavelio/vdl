package transpile

import (
	"fmt"

	"github.com/uforg/uforpc/urpc/internal/schema"
	"github.com/uforg/uforpc/urpc/internal/urpc/ast"
)

// ToJSON transpiles an UFO-RPC AST schema to it's JSON representation.
//
// The imports, extends and external docstrings of the AST Schema are expected
// to be already resolved.
//
// If there are any unresolved imports or extends, the transpiler
// will ignore them.
//
// If there are any unresolved external docstrings, the transpiler will
// treat them literally as strings.
//
// All comments and comment blocks will be ignored.
//
// To get the string representation of the JSON Schema, you can use the
// json.Marshal function.
func ToJSON(astSchema ast.Schema) (schema.Schema, error) {
	result := schema.Schema{
		Version: 1, // Default version is 1
		Nodes:   []schema.Node{},
	}

	// Process all nodes in the original order
	for _, child := range astSchema.Children {
		switch {
		case child.Version != nil:
			result.Version = child.Version.Number

		case child.Docstring != nil:
			docNode := &schema.NodeDoc{
				Kind:    "doc",
				Content: child.Docstring.Value,
			}
			result.Nodes = append(result.Nodes, docNode)

		case child.Type != nil:
			typeNode, err := convertTypeToJSON(child.Type)
			if err != nil {
				return schema.Schema{}, fmt.Errorf("error converting type '%s': %w", child.Type.Name, err)
			}
			result.Nodes = append(result.Nodes, typeNode)

		case child.Proc != nil:
			procNode, err := convertProcToJSON(child.Proc)
			if err != nil {
				return schema.Schema{}, fmt.Errorf("error converting procedure '%s': %w", child.Proc.Name, err)
			}
			result.Nodes = append(result.Nodes, procNode)

		case child.Stream != nil:
			streamNode, err := convertStreamToJSON(child.Stream)
			if err != nil {
				return schema.Schema{}, fmt.Errorf("error converting stream '%s': %w", child.Stream.Name, err)
			}
			result.Nodes = append(result.Nodes, streamNode)
		}
	}

	return result, nil
}

// convertTypeToJSON converts an AST TypeDecl to a schema NodeType
func convertTypeToJSON(typeDecl *ast.TypeDecl) (*schema.NodeType, error) {
	typeNode := &schema.NodeType{
		Kind: "type",
		Name: typeDecl.Name,
	}

	// Add docstring if available
	if typeDecl.Docstring != nil {
		docValue := typeDecl.Docstring.Value
		typeNode.Doc = &docValue
	}

	// Add deprecated if available
	if typeDecl.Deprecated != nil {
		if typeDecl.Deprecated.Message != nil {
			typeNode.Deprecated = typeDecl.Deprecated.Message
		} else {
			empty := ""
			typeNode.Deprecated = &empty
		}
	}

	// Process fields
	for _, child := range typeDecl.Children {
		if child.Field != nil {
			fieldDef, err := convertFieldToJSON(child.Field)
			if err != nil {
				return nil, fmt.Errorf("error converting field '%s': %w", child.Field.Name, err)
			}
			typeNode.Fields = append(typeNode.Fields, fieldDef)
		}
	}

	return typeNode, nil
}

// convertFieldToJSON converts an AST Field to a schema FieldDefinition
func convertFieldToJSON(field *ast.Field) (schema.FieldDefinition, error) {
	fieldDef := schema.FieldDefinition{
		Name:     field.Name,
		Optional: field.Optional,
		IsArray:  field.Type.IsArray,
	}

	// Add docstring if available
	if field.Docstring != nil {
		docValue := field.Docstring.Value
		fieldDef.Doc = &docValue
	}

	// Process field type
	if field.Type.Base.Named != nil {
		typeName := *field.Type.Base.Named
		fieldDef.TypeName = &typeName
	}

	if field.Type.Base.Object != nil {
		inlineType := &schema.InlineTypeDefinition{}

		// Process inline object fields
		for _, child := range field.Type.Base.Object.Children {
			if child.Field != nil {
				inlineField, err := convertFieldToJSON(child.Field)
				if err != nil {
					return schema.FieldDefinition{}, fmt.Errorf("error converting inline field '%s': %w", child.Field.Name, err)
				}
				inlineType.Fields = append(inlineType.Fields, inlineField)
			}
		}

		fieldDef.TypeInline = inlineType
	}

	return fieldDef, nil
}

// convertProcToJSON converts an AST ProcDecl to a schema NodeProc
func convertProcToJSON(procDecl *ast.ProcDecl) (*schema.NodeProc, error) {
	procNode := &schema.NodeProc{
		Kind: "proc",
		Name: procDecl.Name,
	}

	// Add docstring if available
	if procDecl.Docstring != nil {
		docValue := procDecl.Docstring.Value
		procNode.Doc = &docValue
	}

	// Add deprecated if available
	if procDecl.Deprecated != nil {
		if procDecl.Deprecated.Message != nil {
			procNode.Deprecated = procDecl.Deprecated.Message
		} else {
			empty := ""
			procNode.Deprecated = &empty
		}
	}

	// Process procedure children
	for _, child := range procDecl.Children {
		if child.Input != nil {
			for _, fieldOrComment := range child.Input.Children {
				if fieldOrComment.Field != nil {
					fieldDef, err := convertFieldToJSON(fieldOrComment.Field)
					if err != nil {
						return nil, fmt.Errorf("error converting input field '%s': %w", fieldOrComment.Field.Name, err)
					}
					procNode.Input = append(procNode.Input, fieldDef)
				}
			}
		}
		if child.Output != nil {
			for _, fieldOrComment := range child.Output.Children {
				if fieldOrComment.Field != nil {
					fieldDef, err := convertFieldToJSON(fieldOrComment.Field)
					if err != nil {
						return nil, fmt.Errorf("error converting output field '%s': %w", fieldOrComment.Field.Name, err)
					}
					procNode.Output = append(procNode.Output, fieldDef)
				}
			}
		}
	}

	return procNode, nil
}

// convertStreamToJSON converts an AST StreamDecl to a schema NodeStream
func convertStreamToJSON(streamDecl *ast.StreamDecl) (*schema.NodeStream, error) {
	streamNode := &schema.NodeStream{
		Kind: "stream",
		Name: streamDecl.Name,
	}

	// Add docstring if available
	if streamDecl.Docstring != nil {
		docValue := streamDecl.Docstring.Value
		streamNode.Doc = &docValue
	}

	// Add deprecated if available
	if streamDecl.Deprecated != nil {
		if streamDecl.Deprecated.Message != nil {
			streamNode.Deprecated = streamDecl.Deprecated.Message
		} else {
			empty := ""
			streamNode.Deprecated = &empty
		}
	}

	// Process procedure children
	for _, child := range streamDecl.Children {
		if child.Input != nil {
			for _, fieldOrComment := range child.Input.Children {
				if fieldOrComment.Field != nil {
					fieldDef, err := convertFieldToJSON(fieldOrComment.Field)
					if err != nil {
						return nil, fmt.Errorf("error converting input field '%s': %w", fieldOrComment.Field.Name, err)
					}
					streamNode.Input = append(streamNode.Input, fieldDef)
				}
			}
		}
		if child.Output != nil {
			for _, fieldOrComment := range child.Output.Children {
				if fieldOrComment.Field != nil {
					fieldDef, err := convertFieldToJSON(fieldOrComment.Field)
					if err != nil {
						return nil, fmt.Errorf("error converting output field '%s': %w", fieldOrComment.Field.Name, err)
					}
					streamNode.Output = append(streamNode.Output, fieldDef)
				}
			}
		}
	}

	return streamNode, nil
}

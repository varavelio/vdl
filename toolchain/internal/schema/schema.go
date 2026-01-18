package schema

import (
	"encoding/json"
	"fmt"
	"slices"

	"github.com/orsinium-labs/enum"
)

///////////
// ENUMS //
///////////

// PrimitiveType represents the primitive type names defined in the URPC specification.
type PrimitiveType enum.Member[string]

func (f PrimitiveType) MarshalJSON() ([]byte, error) {
	return json.Marshal(f.Value)
}

func (f *PrimitiveType) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &f.Value)
}

var (
	PrimitiveTypeString   = PrimitiveType{Value: "string"}
	PrimitiveTypeInt      = PrimitiveType{Value: "int"}
	PrimitiveTypeFloat    = PrimitiveType{Value: "float"}
	PrimitiveTypeBool     = PrimitiveType{Value: "bool"}
	PrimitiveTypeDatetime = PrimitiveType{Value: "datetime"}
)

////////////////////
// Main Structure //
////////////////////

// Node represents a generic node in the URPC schema structure.
// All specific node types (DocNode, RuleNode, etc.) implement this interface.
type Node interface {
	NodeKind() string
}

// Schema represents the root of the intermediate JSON structure.
type Schema struct {
	// Version is the URPC specification version (always 1 according to the schema).
	Version int `json:"version"`
	// Nodes contains the ordered list of declared elements in the schema.
	Nodes []Node `json:"nodes"`
}

// UnmarshalJSON implements custom JSON unmarshalling for Schema to handle the polymorphic Nodes array.
func (s *Schema) UnmarshalJSON(data []byte) error {
	// 1. Unmarshal into a temporary struct to get Version and raw Nodes data.
	var rawSchema struct {
		Version int               `json:"version"`
		Nodes   []json.RawMessage `json:"nodes"`
	}
	if err := json.Unmarshal(data, &rawSchema); err != nil {
		return fmt.Errorf("failed to unmarshal raw schema: %w", err)
	}

	// 2. Assign Version.
	s.Version = rawSchema.Version
	if s.Version != 1 {
		return fmt.Errorf("unsupported schema version: %d", s.Version)
	}

	// 3. Process each raw node message.
	s.Nodes = make([]Node, 0, len(rawSchema.Nodes))
	var nodeKind struct {
		Kind string `json:"kind"`
	}

	for i, rawNode := range rawSchema.Nodes {
		// 3.1. Peek at the kind field.
		if err := json.Unmarshal(rawNode, &nodeKind); err != nil {
			return fmt.Errorf("failed to determine kind of node at index %d: %w", i, err)
		}

		// 3.2. Unmarshal into the specific node type based on kind.
		var node Node
		var err error
		switch nodeKind.Kind {
		case "doc":
			var docNode NodeDoc
			err = json.Unmarshal(rawNode, &docNode)
			node = &docNode
		case "type":
			var typeNode NodeType
			err = json.Unmarshal(rawNode, &typeNode)
			node = &typeNode
		case "proc":
			var procNode NodeProc
			err = json.Unmarshal(rawNode, &procNode)
			node = &procNode
		case "stream":
			var streamNode NodeStream
			err = json.Unmarshal(rawNode, &streamNode)
			node = &streamNode
		default:
			return fmt.Errorf("unknown node kind '%s' at index %d", nodeKind.Kind, i)
		}

		if err != nil {
			return fmt.Errorf("failed to unmarshal node of kind '%s' at index %d: %w", nodeKind.Kind, i, err)
		}
		s.Nodes = append(s.Nodes, node)
	}

	return nil
}

// GetDocNodes returns all DocNode instances from the schema.
func (s *Schema) GetDocNodes() []*NodeDoc {
	docNodes := []*NodeDoc{}
	for _, node := range s.Nodes {
		if docNode, ok := node.(*NodeDoc); ok {
			docNodes = append(docNodes, docNode)
		}
	}
	return docNodes
}

// GetTypeNodes returns all TypeNode instances from the schema.
func (s *Schema) GetTypeNodes() []*NodeType {
	typeNodes := []*NodeType{}
	for _, node := range s.Nodes {
		if typeNode, ok := node.(*NodeType); ok {
			typeNodes = append(typeNodes, typeNode)
		}
	}
	return typeNodes
}

// GetTypeNodesMap returns a map of type nodes by name.
func (s *Schema) GetTypeNodesMap() map[string]*NodeType {
	typeNodes := s.GetTypeNodes()
	typeNodesMap := make(map[string]*NodeType)
	for _, node := range typeNodes {
		typeNodesMap[node.Name] = node
	}
	return typeNodesMap
}

// GetProcNodes returns all ProcNode instances from the schema.
func (s *Schema) GetProcNodes() []*NodeProc {
	procNodes := []*NodeProc{}
	for _, node := range s.Nodes {
		if procNode, ok := node.(*NodeProc); ok {
			procNodes = append(procNodes, procNode)
		}
	}
	return procNodes
}

// GetProcNodesMap returns a map of proc nodes by name.

func (s *Schema) GetProcNodesMap() map[string]*NodeProc {
	procNodes := s.GetProcNodes()
	procNodesMap := make(map[string]*NodeProc)
	for _, node := range procNodes {
		procNodesMap[node.Name] = node
	}
	return procNodesMap
}

// GetStreamNodes returns all StreamNode instances from the schema.
func (s *Schema) GetStreamNodes() []*NodeStream {
	streamNodes := []*NodeStream{}
	for _, node := range s.Nodes {
		if streamNode, ok := node.(*NodeStream); ok {
			streamNodes = append(streamNodes, streamNode)
		}
	}
	return streamNodes
}

// GetStreamNodesMap returns a map of stream nodes by name.
func (s *Schema) GetStreamNodesMap() map[string]*NodeStream {
	streamNodes := s.GetStreamNodes()
	streamNodesMap := make(map[string]*NodeStream)
	for _, node := range streamNodes {
		streamNodesMap[node.Name] = node
	}
	return streamNodesMap
}

////////////////
// Node Types //
////////////////

// NodeDoc represents a standalone documentation block.
type NodeDoc struct {
	Kind    string `json:"kind"` // Always "doc"
	Content string `json:"content"`
}

func (n *NodeDoc) NodeKind() string { return n.Kind }

// NodeType represents the definition of a custom data type.
type NodeType struct {
	Kind string `json:"kind"` // Always "type"
	Name string `json:"name"`
	// Doc is the associated documentation string (optional).
	Doc *string `json:"doc,omitempty"`
	// Deprecated indicates if the type is deprecated and contains the message
	// associated with the deprecation.
	Deprecated *string `json:"deprecated,omitempty"`
	// Fields is the ordered list of fields within the type.
	Fields []FieldDefinition `json:"fields"`
}

func (n *NodeType) NodeKind() string { return n.Kind }

// NodeProc represents the definition of an RPC procedure.
type NodeProc struct {
	Kind string `json:"kind"` // Always "proc"
	Name string `json:"name"`
	// Doc is the associated documentation string (optional).
	Doc *string `json:"doc,omitempty"`
	// Deprecated indicates if the procedure is deprecated and contains the message
	// associated with the deprecation.
	Deprecated *string `json:"deprecated,omitempty"`
	// Input is the ordered list of input fields for the procedure.
	Input []FieldDefinition `json:"input"`
	// Output is the ordered list of output fields for the procedure.
	Output []FieldDefinition `json:"output"`
}

func (n *NodeProc) NodeKind() string { return n.Kind }

// NodeStream represents the definition of an RPC stream.
type NodeStream struct {
	Kind string `json:"kind"` // Always "stream"
	Name string `json:"name"`
	// Doc is the associated documentation string (optional).
	Doc *string `json:"doc,omitempty"`
	// Deprecated indicates if the stream is deprecated and contains the message
	// associated with the deprecation.
	Deprecated *string `json:"deprecated,omitempty"`
	// Input is the ordered list of input fields for the stream.
	Input []FieldDefinition `json:"input"`
	// Output is the ordered list of output fields for the stream.
	Output []FieldDefinition `json:"output"`
}

func (n *NodeStream) NodeKind() string { return n.Kind }

//////////////////////////
// Auxiliary Structures //
//////////////////////////

// FieldDefinition defines a field within a type or procedure input/output.
type FieldDefinition struct {
	Name string `json:"name"`
	// Doc is the associated documentation string (optional).
	Doc *string `json:"doc,omitempty"`
	// TypeName holds the name if the type is named (primitive or custom). Mutually exclusive with TypeInline.
	TypeName *string `json:"typeName,omitempty"`
	// TypeInline holds the definition if the type is inline. Mutually exclusive with TypeName.
	TypeInline *InlineTypeDefinition `json:"typeInline,omitempty"`
	// IsArray indicates if the field is an array.
	IsArray bool `json:"isArray"`
	// Optional indicates if the field is optional.
	Optional bool `json:"optional"`
}

// IsNamed checks if the field definition uses a named type.
func (fd *FieldDefinition) IsNamed() bool {
	return fd.TypeName != nil
}

// IsInline checks if the field definition uses an inline type.
func (fd *FieldDefinition) IsInline() bool {
	return fd.TypeInline != nil
}

// IsBuiltInType checks if the field definition uses a built-in type.
func (fd *FieldDefinition) IsBuiltInType() bool {
	return fd.IsNamed() && slices.Contains([]string{"string", "int", "float", "bool", "datetime"}, *fd.TypeName)
}

// IsCustomType checks if the field definition uses a custom type.
func (fd *FieldDefinition) IsCustomType() bool {
	return fd.IsNamed() && !fd.IsBuiltInType()
}

// InlineTypeDefinition represents the structure of an anonymous inline object type.
// It's used within the FieldDefinition.TypeInline field.
type InlineTypeDefinition struct {
	// Fields is the ordered list of fields within the inline type.
	Fields []FieldDefinition `json:"fields"`
}

package schema

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/uforg/uforpc/urpc/internal/util/testutil"
)

func TestParseSchema(t *testing.T) {
	t.Run("Valid empty schema", func(t *testing.T) {
		input := `{
			"version": 1,
			"nodes": []
		}`

		schema, err := ParseSchema(input)
		require.NoError(t, err)
		require.Equal(t, 1, schema.Version)
		require.Empty(t, schema.Nodes)
	})

	t.Run("Schema with doc node", func(t *testing.T) {
		input := `{
			"version": 1,
			"nodes": [
				{
					"kind": "doc",
					"content": "This is documentation"
				}
			]
		}`

		schema, err := ParseSchema(input)
		require.NoError(t, err)
		require.Equal(t, 1, schema.Version)
		require.Len(t, schema.Nodes, 1)

		docNode, ok := schema.Nodes[0].(*NodeDoc)
		require.True(t, ok, "Node should be a NodeDoc")
		require.Equal(t, "doc", docNode.Kind)
		require.Equal(t, "This is documentation", docNode.Content)
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		input := `{
			"version": 1,
			"nodes": [
		}`

		_, err := ParseSchema(input)
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to unmarshal input schema")
	})

}

func TestPrimitiveType(t *testing.T) {
	t.Run("Marshal PrimitiveType", func(t *testing.T) {
		primitiveType := PrimitiveTypeString

		data, err := json.Marshal(primitiveType)
		require.NoError(t, err)
		require.Equal(t, `"string"`, string(data))
	})

	t.Run("Unmarshal PrimitiveType", func(t *testing.T) {
		input := `"string"`

		var primitiveType PrimitiveType
		err := json.Unmarshal([]byte(input), &primitiveType)
		require.NoError(t, err)
		require.Equal(t, PrimitiveTypeString, primitiveType)
	})
}

func TestNodeKind(t *testing.T) {
	t.Run("NodeDoc.NodeKind", func(t *testing.T) {
		node := NodeDoc{
			Kind:    "doc",
			Content: "Documentation",
		}
		require.Equal(t, "doc", node.NodeKind())
	})

	t.Run("NodeType.NodeKind", func(t *testing.T) {
		node := NodeType{
			Kind: "type",
			Name: "User",
		}
		require.Equal(t, "type", node.NodeKind())
	})

	t.Run("NodeProc.NodeKind", func(t *testing.T) {
		node := NodeProc{
			Kind: "proc",
			Name: "GetUser",
		}
		require.Equal(t, "proc", node.NodeKind())
	})
}

func TestBasicSchemaUnmarshal(t *testing.T) {
	t.Run("Valid empty schema", func(t *testing.T) {
		input := `{
			"version": 1,
			"nodes": []
		}`

		var schema Schema
		err := json.Unmarshal([]byte(input), &schema)
		require.NoError(t, err)
		require.Equal(t, 1, schema.Version)
		require.Empty(t, schema.Nodes)
	})

	t.Run("Schema with doc node", func(t *testing.T) {
		input := `{
			"version": 1,
			"nodes": [
				{
					"kind": "doc",
					"content": "This is documentation"
				}
			]
		}`

		var schema Schema
		err := json.Unmarshal([]byte(input), &schema)
		require.NoError(t, err)
		require.Equal(t, 1, schema.Version)
		require.Len(t, schema.Nodes, 1)

		docNode, ok := schema.Nodes[0].(*NodeDoc)
		require.True(t, ok, "Node should be a NodeDoc")
		require.Equal(t, "doc", docNode.Kind)
		require.Equal(t, "This is documentation", docNode.Content)
	})

	t.Run("Schema with type node", func(t *testing.T) {
		input := `{
			"version": 1,
			"nodes": [
				{
					"kind": "type",
					"name": "User",
					"doc": "User type",
					"fields": [
						{
							"name": "id",
							"doc": "User ID",
							"typeName": "string",
							"isArray": false,
							"optional": false
						},
						{
							"name": "name",
							"typeName": "string",
							"isArray": true,
							"optional": false
						}
					]
				}
			]
		}`

		var schema Schema
		err := json.Unmarshal([]byte(input), &schema)
		require.NoError(t, err)
		require.Equal(t, 1, schema.Version)
		require.Len(t, schema.Nodes, 1)

		// Check that the node is a type node
		typeNode, ok := schema.Nodes[0].(*NodeType)
		require.True(t, ok, "Node should be a NodeType")
		require.Equal(t, "type", typeNode.Kind)
		require.Equal(t, "User", typeNode.Name)
		require.NotNil(t, typeNode.Doc)
		require.Equal(t, "User type", *typeNode.Doc)
		require.Len(t, typeNode.Fields, 2)

		// Check the fields
		require.Equal(t, "id", typeNode.Fields[0].Name)
		require.NotNil(t, typeNode.Fields[0].Doc)
		require.Equal(t, "User ID", *typeNode.Fields[0].Doc)
		require.NotNil(t, typeNode.Fields[0].TypeName)
		require.Equal(t, "string", *typeNode.Fields[0].TypeName)
		require.False(t, typeNode.Fields[0].IsArray)
		require.False(t, typeNode.Fields[0].Optional)

		require.Equal(t, "name", typeNode.Fields[1].Name)
		require.NotNil(t, typeNode.Fields[1].TypeName)
		require.Equal(t, "string", *typeNode.Fields[1].TypeName)
		require.True(t, typeNode.Fields[1].IsArray)
		require.False(t, typeNode.Fields[1].Optional)
	})

	t.Run("Schema with proc node", func(t *testing.T) {
		input := `{
			"version": 1,
			"nodes": [
				{
					"kind": "proc",
					"name": "GetUser",
					"doc": "Get user by ID",
					"input": [
						{
							"name": "id",
							"doc": "User ID",
							"typeName": "string",
							"isArray": false,
							"optional": false
						}
					],
					"output": [
						{
							"name": "user",
							"typeName": "User",
							"isArray": false,
							"optional": false
						}
					]
				}
			]
		}`

		var schema Schema
		err := json.Unmarshal([]byte(input), &schema)
		require.NoError(t, err)
		require.Equal(t, 1, schema.Version)
		require.Len(t, schema.Nodes, 1)

		// Check that the node is a proc node
		procNode, ok := schema.Nodes[0].(*NodeProc)
		require.True(t, ok, "Node should be a NodeProc")
		require.Equal(t, "proc", procNode.Kind)
		require.Equal(t, "GetUser", procNode.Name)
		require.NotNil(t, procNode.Doc)
		require.Equal(t, "Get user by ID", *procNode.Doc)

		// Check input fields
		require.Len(t, procNode.Input, 1)
		require.Equal(t, "id", procNode.Input[0].Name)
		require.NotNil(t, procNode.Input[0].Doc)
		require.Equal(t, "User ID", *procNode.Input[0].Doc)
		require.NotNil(t, procNode.Input[0].TypeName)
		require.Equal(t, "string", *procNode.Input[0].TypeName)

		// Check output fields
		require.Len(t, procNode.Output, 1)
		require.Equal(t, "user", procNode.Output[0].Name)
		require.NotNil(t, procNode.Output[0].TypeName)
		require.Equal(t, "User", *procNode.Output[0].TypeName)
	})

	t.Run("Schema with stream node", func(t *testing.T) {
		input := `{
			"version": 1,
			"nodes": [
				{
					"kind": "stream",
					"name": "GetUser",
					"doc": "Get user by ID",
					"input": [
						{
							"name": "id",
							"doc": "User ID",
							"typeName": "string",
							"isArray": false,
							"optional": false
						}
					],
					"output": [
						{
							"name": "user",
							"typeName": "User",
							"isArray": false,
							"optional": false
						}
					]
				}
			]
		}`

		var schema Schema
		err := json.Unmarshal([]byte(input), &schema)
		require.NoError(t, err)
		require.Equal(t, 1, schema.Version)
		require.Len(t, schema.Nodes, 1)

		// Check that the node is a proc node
		streamNode, ok := schema.Nodes[0].(*NodeStream)
		require.True(t, ok, "Node should be a NodeStream")
		require.Equal(t, "stream", streamNode.Kind)
		require.Equal(t, "GetUser", streamNode.Name)
		require.NotNil(t, streamNode.Doc)
		require.Equal(t, "Get user by ID", *streamNode.Doc)

		// Check input fields
		require.Len(t, streamNode.Input, 1)
		require.Equal(t, "id", streamNode.Input[0].Name)
		require.NotNil(t, streamNode.Input[0].Doc)
		require.Equal(t, "User ID", *streamNode.Input[0].Doc)
		require.NotNil(t, streamNode.Input[0].TypeName)
		require.Equal(t, "string", *streamNode.Input[0].TypeName)

		// Check output fields
		require.Len(t, streamNode.Output, 1)
		require.Equal(t, "user", streamNode.Output[0].Name)
		require.NotNil(t, streamNode.Output[0].TypeName)
		require.Equal(t, "User", *streamNode.Output[0].TypeName)
	})
}

func TestGetNodeMethods(t *testing.T) {
	input := `{
		"version": 1,
		"nodes": [
			{
				"kind": "doc",
				"content": "Documentation"
			},
			{
				"kind": "type",
				"name": "User",
				"fields": [
					{
						"name": "id",
						"typeName": "string",
						"isArray": false,
						"optional": false
					}
				]
			},
			{
				"kind": "proc",
				"name": "GetUser",
				"input": [
					{
						"name": "id",
						"typeName": "string",
						"isArray": false,
						"optional": false
					}
				],
				"output": [
					{
						"name": "user",
						"typeName": "User",
						"isArray": false,
						"optional": false
					}
				]
			}
		]
	}`

	var schema Schema
	err := json.Unmarshal([]byte(input), &schema)
	require.NoError(t, err)

	// Test GetDocNodes
	docNodes := schema.GetDocNodes()
	require.Len(t, docNodes, 1)
	require.Equal(t, "Documentation", docNodes[0].Content)

	// Test GetTypeNodes
	typeNodes := schema.GetTypeNodes()
	require.Len(t, typeNodes, 1)
	require.Equal(t, "User", typeNodes[0].Name)

	// Test GetProcNodes
	procNodes := schema.GetProcNodes()
	require.Len(t, procNodes, 1)
	require.Equal(t, "GetUser", procNodes[0].Name)
}

func TestFieldDefinitionHelperMethods(t *testing.T) {
	// Test IsNamed
	namedField := FieldDefinition{
		Name:     "user",
		TypeName: testutil.Pointer("User"),
	}
	require.True(t, namedField.IsNamed())
	require.False(t, namedField.IsInline())

	// Test IsInline
	inlineField := FieldDefinition{
		Name: "address",
		TypeInline: &InlineTypeDefinition{
			Fields: []FieldDefinition{},
		},
	}
	require.False(t, inlineField.IsNamed())
	require.True(t, inlineField.IsInline())

	// Test neither named nor inline
	emptyField := FieldDefinition{
		Name: "empty",
	}
	require.False(t, emptyField.IsNamed())
	require.False(t, emptyField.IsInline())
}

func TestDeprecated(t *testing.T) {
	t.Run("Without message", func(t *testing.T) {
		input := `{
			"version": 1,
			"nodes": [
				{
					"kind": "type",
					"name": "User",
					"deprecated": ""
				},
				{
					"kind": "proc",
					"name": "GetUser",
					"deprecated": "",
					"input": [],
					"output": []
				}
			]
		}`

		var schema Schema
		err := json.Unmarshal([]byte(input), &schema)
		require.NoError(t, err)

		// Check type node
		typeNode, ok := schema.Nodes[0].(*NodeType)
		require.True(t, ok, "Node should be a NodeType")
		require.NotNil(t, typeNode.Deprecated)
		require.Empty(t, *typeNode.Deprecated)

		// Check proc node
		procNode, ok := schema.Nodes[1].(*NodeProc)
		require.True(t, ok, "Node should be a NodeProc")
		require.NotNil(t, procNode.Deprecated)
		require.Empty(t, *procNode.Deprecated)
	})

	t.Run("With message", func(t *testing.T) {
		input := `{
			"version": 1,
			"nodes": [
				{
					"kind": "type",
					"name": "User",
					"deprecated": "Deprecation message"
				},
				{
					"kind": "proc",
					"name": "GetUser",
					"deprecated": "Deprecation message",
					"input": [],
					"output": []
				}
			]
		}`

		var schema Schema
		err := json.Unmarshal([]byte(input), &schema)
		require.NoError(t, err)

		// Check type node
		typeNode, ok := schema.Nodes[0].(*NodeType)
		require.True(t, ok, "Node should be a NodeType")
		require.NotNil(t, typeNode.Deprecated)
		require.Equal(t, "Deprecation message", *typeNode.Deprecated)

		// Check proc node
		procNode, ok := schema.Nodes[1].(*NodeProc)
		require.True(t, ok, "Node should be a NodeProc")
		require.NotNil(t, procNode.Deprecated)
		require.Equal(t, "Deprecation message", *procNode.Deprecated)
	})
}

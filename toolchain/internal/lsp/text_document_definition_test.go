package lsp

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFindIdentifierAtPosition(t *testing.T) {
	content := `type FooType {
  firstField string
  secondField int[]
}

type Wrapper {
  foo FooType
}`

	tests := []struct {
		name     string
		position TextDocumentPosition
		want     string
	}{
		{
			name:     "Find type name",
			position: TextDocumentPosition{Line: 0, Character: 7},
			want:     "FooType",
		},
		{
			name:     "Find wrapper type name",
			position: TextDocumentPosition{Line: 5, Character: 7},
			want:     "Wrapper",
		},
		{
			name:     "Find type reference",
			position: TextDocumentPosition{Line: 6, Character: 8},
			want:     "FooType",
		},
		{
			name:     "Position out of range",
			position: TextDocumentPosition{Line: 100, Character: 1},
			want:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := findIdentifierAtPosition(content, tt.position)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestHandleTextDocumentDefinition(t *testing.T) {
	// Create a mock reader and writer for the LSP
	mockReader := &bytes.Buffer{}
	mockWriter := &bytes.Buffer{}

	// Create an LSP instance
	lsp := New(mockReader, mockWriter)

	// Create a test schema using new VDL syntax
	schemaContent := `type FooType {
  firstField string
  secondField int[]
}

type Wrapper {
  foo FooType
}`

	// Add the schema to the vfs
	filePath := "/test.vdl"
	lsp.fs.WriteFileCache(filePath, []byte(schemaContent))

	// Create a definition request - position is on "FooType" in Wrapper
	request := RequestMessageTextDocumentDefinition{
		RequestMessage: RequestMessage{
			Message: Message{
				JSONRPC: "2.0",
				Method:  "textDocument/definition",
				ID:      "1",
			},
		},
		Params: RequestMessageTextDocumentDefinitionParams{
			TextDocument: TextDocumentIdentifier{
				URI: "file://" + filePath,
			},
			Position: TextDocumentPosition{
				Line:      6, // 0-based, line with "foo FooType"
				Character: 8,
			},
		},
	}

	// Encode the request
	requestBytes, err := json.Marshal(request)
	require.NoError(t, err)

	// Handle the request
	response, err := lsp.handleTextDocumentDefinition(requestBytes)
	require.NoError(t, err)

	// Check the response
	defResponse, ok := response.(ResponseMessageTextDocumentDefinition)
	require.True(t, ok)
	require.NotNil(t, defResponse.Result)
	require.Len(t, defResponse.Result, 1)
	assert.Contains(t, defResponse.Result[0].URI, filePath)
	assert.Equal(t, 0, defResponse.Result[0].Range.Start.Line) // FooType is on line 0 (0-based)
}

package lsp

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtractCodeFromContent(t *testing.T) {
	content := `type FooType {
  firstField: string
  secondField: int[]
}

rpc Test {
  proc BarProc {
    input {
      foo: FooType
    }

    output {
      baz: bool
    }
  }
}`

	tests := []struct {
		name      string
		startLine int
		endLine   int
		want      string
		wantErr   bool
	}{
		{
			name:      "Extract type definition",
			startLine: 1,
			endLine:   4,
			want:      "type FooType {\n  firstField: string\n  secondField: int[]\n}",
			wantErr:   false,
		},
		{
			name:      "Extract single line",
			startLine: 2,
			endLine:   2,
			want:      "firstField: string",
			wantErr:   false,
		},
		{
			name:      "Start line out of range",
			startLine: 100,
			endLine:   101,
			want:      "",
			wantErr:   true,
		},
		{
			name:      "End line out of range",
			startLine: 1,
			endLine:   100,
			want:      "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := extractCodeFromContent(content, tt.startLine, tt.endLine)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestHandleTextDocumentHover(t *testing.T) {
	// Create a mock reader and writer for the LSP
	mockReader := &bytes.Buffer{}
	mockWriter := &bytes.Buffer{}

	// Create an LSP instance
	lsp := New(mockReader, mockWriter)

	// Create a test schema using the new VDL syntax
	schemaContent := `type FooType {
  firstField: string
  secondField: int[]
}

rpc Test {
  proc BarProc {
    input {
      foo: FooType
    }

    output {
      baz: bool
    }
  }
}`

	// Add the schema to the vfs
	filePath := "/test.vdl"
	lsp.fs.WriteFileCache(filePath, []byte(schemaContent))

	// Create a hover request for a type reference (line 9 is "foo: FooType")
	request := RequestMessageTextDocumentHover{
		RequestMessage: RequestMessage{
			Message: Message{
				JSONRPC: "2.0",
				Method:  "textDocument/hover",
				ID:      "1",
			},
		},
		Params: RequestMessageTextDocumentHoverParams{
			TextDocument: TextDocumentIdentifier{
				URI: "file://" + filePath,
			},
			Position: TextDocumentPosition{
				Line:      8, // 0-based, line with "foo: FooType"
				Character: 11,
			},
		},
	}

	// Encode the request
	requestBytes, err := json.Marshal(request)
	require.NoError(t, err)

	// Handle the request
	response, err := lsp.handleTextDocumentHover(requestBytes)
	require.NoError(t, err)

	// Check the response
	hoverResponse, ok := response.(ResponseMessageTextDocumentHover)
	require.True(t, ok)
	require.NotNil(t, hoverResponse.Result)
	assert.Equal(t, "markdown", hoverResponse.Result.Contents.Kind)
	assert.Contains(t, hoverResponse.Result.Contents.Value, "```vdl")
	assert.Contains(t, hoverResponse.Result.Contents.Value, "type FooType")
}

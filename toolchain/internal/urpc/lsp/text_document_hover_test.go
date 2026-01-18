package lsp

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uforg/uforpc/urpc/internal/urpc/analyzer"
)

func TestExtractCodeFromContent(t *testing.T) {
	content := `version 1

type FooType {
  firstField: string
  secondField: int[]
}

proc BarProc {
  input {
    foo: FooType
  }

  output {
    baz: bool
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
			startLine: 3,
			endLine:   6,
			want:      "type FooType {\n  firstField: string\n  secondField: int[]\n}",
			wantErr:   false,
		},
		{
			name:      "Extract proc definition",
			startLine: 7,
			endLine:   15,
			want:      "\nproc BarProc {\n  input {\n    foo: FooType\n  }\n\n  output {\n    baz: bool\n  }",
			wantErr:   false,
		},
		{
			name:      "Extract single line",
			startLine: 4,
			endLine:   4,
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

	// Create a test schema
	schemaContent := `version 1

type FooType {
  firstField: string
  secondField: int[]
}

proc BarProc {
  input {
    foo: FooType
  }

  output {
    baz: bool
  }
}`

	// Add the schema to the docstore
	filePath := "file:///test.urpc"
	err := lsp.docstore.OpenInMem(filePath, schemaContent)
	require.NoError(t, err)

	// Create a hover request for a type reference
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
				URI: filePath,
			},
			Position: TextDocumentPosition{
				Line:      9,
				Character: 10,
			},
		},
	}

	// Encode the request
	requestBytes, err := json.Marshal(request)
	require.NoError(t, err)

	// Create a mock analyzer
	mockAnalyzer, err := analyzer.NewAnalyzer(lsp.docstore)
	require.NoError(t, err)
	lsp.analyzer = mockAnalyzer

	// Analyze the file to populate the combined schema
	_, _, err = lsp.analyzer.Analyze(filePath)
	require.NoError(t, err)

	// Handle the request
	response, err := lsp.handleTextDocumentHover(requestBytes)
	require.NoError(t, err)

	// Check the response
	hoverResponse, ok := response.(ResponseMessageTextDocumentHover)
	require.True(t, ok)
	require.NotNil(t, hoverResponse.Result)
	assert.Equal(t, "markdown", hoverResponse.Result.Contents.Kind)
	assert.Contains(t, hoverResponse.Result.Contents.Value, "```urpc")
	assert.Contains(t, hoverResponse.Result.Contents.Value, "type FooType")
}

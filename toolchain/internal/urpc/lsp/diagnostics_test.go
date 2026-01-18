package lsp

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uforg/uforpc/urpc/internal/urpc/analyzer"
	"github.com/uforg/uforpc/urpc/internal/urpc/ast"
)

func TestDiagnostics(t *testing.T) {
	// Create a mock reader and writer for the LSP
	mockReader := &bytes.Buffer{}
	mockWriter := &bytes.Buffer{}

	// Create an LSP instance
	lsp := New(mockReader, mockWriter)

	// Test conversion from analyzer diagnostic to LSP diagnostic
	t.Run("ConvertAnalyzerDiagnosticToLSPDiagnostic", func(t *testing.T) {
		analyzerDiag := analyzer.Diagnostic{
			Positions: analyzer.Positions{
				Pos:    ast.Position{Filename: "test.urpc", Line: 10, Column: 5, Offset: 100},
				EndPos: ast.Position{Filename: "test.urpc", Line: 10, Column: 15, Offset: 110},
			},
			Message: "Test diagnostic message",
		}

		lspDiag := ConvertAnalyzerDiagnosticToLSPDiagnostic(analyzerDiag)

		assert.Equal(t, 9, lspDiag.Range.Start.Line, "Line should be converted to 0-based")
		assert.Equal(t, 4, lspDiag.Range.Start.Character, "Column should be converted to 0-based")
		assert.Equal(t, 9, lspDiag.Range.End.Line, "Line should be converted to 0-based")
		assert.Equal(t, 14, lspDiag.Range.End.Character, "Column should be converted to 0-based")
		assert.Equal(t, "Test diagnostic message", lspDiag.Message)
		assert.Equal(t, DiagnosticSeverityError, lspDiag.Severity)
		assert.Equal(t, "urpc", lspDiag.Source)
	})

	// Test publishing diagnostics
	t.Run("PublishDiagnostics", func(t *testing.T) {
		// Clear the writer buffer
		mockWriter.Reset()

		// Create some diagnostics
		diagnostics := []Diagnostic{
			{
				Range: TextDocumentRange{
					Start: TextDocumentPosition{Line: 1, Character: 2},
					End:   TextDocumentPosition{Line: 1, Character: 10},
				},
				Severity: DiagnosticSeverityError,
				Source:   "urpc",
				Message:  "Test error message",
			},
		}

		// Publish diagnostics
		lsp.publishDiagnostics("file:///test.urpc", diagnostics)

		// Read the response from the writer
		response := mockWriter.String()
		assert.Contains(t, response, "textDocument/publishDiagnostics")
		assert.Contains(t, response, "file:///test.urpc")
		assert.Contains(t, response, "Test error message")

		// Parse the response to verify the structure
		var parsedResponse map[string]interface{}
		headerEnd := strings.Index(response, "\r\n\r\n")
		require.NotEqual(t, -1, headerEnd, "Header end not found")
		jsonContent := response[headerEnd+4:]
		err := json.Unmarshal([]byte(jsonContent), &parsedResponse)
		require.NoError(t, err, "Failed to parse JSON response")

		// Verify the response structure
		assert.Equal(t, "2.0", parsedResponse["jsonrpc"])
		assert.Equal(t, "textDocument/publishDiagnostics", parsedResponse["method"])
		params, ok := parsedResponse["params"].(map[string]interface{})
		require.True(t, ok, "Params not found or not an object")
		assert.Equal(t, "file:///test.urpc", params["uri"])
		diagArray, ok := params["diagnostics"].([]interface{})
		require.True(t, ok, "Diagnostics not found or not an array")
		require.Len(t, diagArray, 1, "Expected 1 diagnostic")
		diag, ok := diagArray[0].(map[string]interface{})
		require.True(t, ok, "Diagnostic not an object")
		assert.Equal(t, "Test error message", diag["message"])
	})

	// Test clearing diagnostics
	t.Run("ClearDiagnostics", func(t *testing.T) {
		// Clear the writer buffer
		mockWriter.Reset()

		// Clear diagnostics
		lsp.clearDiagnostics("file:///test.urpc")

		// Read the response from the writer
		response := mockWriter.String()
		assert.Contains(t, response, "textDocument/publishDiagnostics")
		assert.Contains(t, response, "file:///test.urpc")

		// Parse the response to verify the structure
		var parsedResponse map[string]interface{}
		headerEnd := strings.Index(response, "\r\n\r\n")
		require.NotEqual(t, -1, headerEnd, "Header end not found")
		jsonContent := response[headerEnd+4:]
		err := json.Unmarshal([]byte(jsonContent), &parsedResponse)
		require.NoError(t, err, "Failed to parse JSON response")

		// Verify the response structure
		params, ok := parsedResponse["params"].(map[string]interface{})
		require.True(t, ok, "Params not found or not an object")
		diagArray, ok := params["diagnostics"].([]interface{})
		require.True(t, ok, "Diagnostics not found or not an array")
		assert.Empty(t, diagArray, "Expected empty diagnostics array")
	})
}

// MockFileProvider is a mock implementation of the analyzer.FileProvider interface
type MockFileProvider struct {
	files map[string]string
}

func NewMockFileProvider() *MockFileProvider {
	return &MockFileProvider{
		files: make(map[string]string),
	}
}

func (m *MockFileProvider) GetFileAndHash(relativeTo string, path string) (string, string, error) {
	content, ok := m.files[path]
	if !ok {
		return "", "", io.ErrUnexpectedEOF
	}
	return content, "mock-hash", nil
}

func (m *MockFileProvider) AddFile(path string, content string) {
	m.files[path] = content
}

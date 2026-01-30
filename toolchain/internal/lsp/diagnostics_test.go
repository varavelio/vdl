package lsp

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/varavelio/vdl/toolchain/internal/core/analysis"
	"github.com/varavelio/vdl/toolchain/internal/core/ast"
)

func TestDiagnostics(t *testing.T) {
	// Create a mock reader and writer for the LSP
	mockReader := &bytes.Buffer{}
	mockWriter := &bytes.Buffer{}

	// Create an LSP instance
	lsp := New(mockReader, mockWriter)

	// Test conversion from analysis diagnostic to LSP diagnostic
	t.Run("ConvertAnalysisDiagnosticToLSPDiagnostic", func(t *testing.T) {
		analysisDiag := analysis.Diagnostic{
			File:    "test.vdl",
			Pos:     ast.Position{Filename: "test.vdl", Line: 10, Column: 5, Offset: 100},
			EndPos:  ast.Position{Filename: "test.vdl", Line: 10, Column: 15, Offset: 110},
			Code:    "E001",
			Message: "Test diagnostic message",
		}

		lspDiag := ConvertAnalysisDiagnosticToLSPDiagnostic(analysisDiag)

		assert.Equal(t, 9, lspDiag.Range.Start.Line, "Line should be converted to 0-based")
		assert.Equal(t, 4, lspDiag.Range.Start.Character, "Column should be converted to 0-based")
		assert.Equal(t, 9, lspDiag.Range.End.Line, "Line should be converted to 0-based")
		assert.Equal(t, 14, lspDiag.Range.End.Character, "Column should be converted to 0-based")
		assert.Equal(t, "Test diagnostic message", lspDiag.Message)
		assert.Equal(t, DiagnosticSeverityError, lspDiag.Severity)
		assert.Equal(t, "vdl", lspDiag.Source)
		assert.Equal(t, "E001", lspDiag.Code)
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
				Source:   "vdl",
				Message:  "Test error message",
			},
		}

		// Publish diagnostics
		lsp.publishDiagnostics("file:///test.vdl", diagnostics)

		// Read the response from the writer
		response := mockWriter.String()
		assert.Contains(t, response, "textDocument/publishDiagnostics")
		assert.Contains(t, response, "file:///test.vdl")
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
		assert.Equal(t, "file:///test.vdl", params["uri"])
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
		lsp.clearDiagnostics("file:///test.vdl")

		// Read the response from the writer
		response := mockWriter.String()
		assert.Contains(t, response, "textDocument/publishDiagnostics")
		assert.Contains(t, response, "file:///test.vdl")

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

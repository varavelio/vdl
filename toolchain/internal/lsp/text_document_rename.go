package lsp

import (
	"fmt"

	"github.com/varavelio/vdl/toolchain/internal/core/parser"
)

// RequestMessageTextDocumentRename represents a request for renaming a symbol.
type RequestMessageTextDocumentRename struct {
	RequestMessage
	Params RequestMessageTextDocumentRenameParams `json:"params"`
}

// RequestMessageTextDocumentRenameParams represents the parameters for a rename request.
type RequestMessageTextDocumentRenameParams struct {
	// The text document.
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	// The position inside the text document.
	Position TextDocumentPosition `json:"position"`
	// The new name of the symbol.
	NewName string `json:"newName"`
}

// ResponseMessageTextDocumentRename represents a response to a rename request.
type ResponseMessageTextDocumentRename struct {
	ResponseMessage
	// The result of the request – the edits to apply.
	Result WorkspaceEdit `json:"result"`
}

// WorkspaceEdit represents a collection of changes to multiple resources.
type WorkspaceEdit struct {
	// Holds changes to existing resources.
	Changes map[string][]TextDocumentTextEdit `json:"changes"`
}

// handleTextDocumentRename handles a textDocument/rename request.
func (l *LSP) handleTextDocumentRename(rawMessage []byte) (any, error) {
	var request RequestMessageTextDocumentRename
	if err := decode(rawMessage, &request); err != nil {
		return nil, fmt.Errorf("failed to decode rename request: %w", err)
	}

	filePath := UriToPath(request.Params.TextDocument.URI)
	position := request.Params.Position
	newName := request.Params.NewName

	// Get the current document content
	content, err := l.fs.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file from vfs: %w", err)
	}

	// Determine the identifier to be renamed
	oldName := findIdentifierAtPosition(string(content), position)
	if oldName == "" {
		resp := ResponseMessageTextDocumentRename{
			ResponseMessage: ResponseMessage{
				Message: DefaultMessage,
				ID:      request.ID,
				Error: ResponseError{
					Code:    ErrorCodeInvalidParams,
					Message: "symbol not found at position",
				},
			},
			Result: WorkspaceEdit{},
		}
		return resp, nil
	}

	// If the new name is the same as the old one, return empty edit.
	if oldName == newName {
		resp := ResponseMessageTextDocumentRename{
			ResponseMessage: ResponseMessage{
				Message: DefaultMessage,
				ID:      request.ID,
			},
			Result: WorkspaceEdit{},
		}
		return resp, nil
	}

	// Collect all occurrences of the identifier in the document
	edits := collectRenameEdits(string(content), filePath, oldName, newName)

	workspaceEdit := WorkspaceEdit{
		Changes: map[string][]TextDocumentTextEdit{
			request.Params.TextDocument.URI: edits,
		},
	}

	response := ResponseMessageTextDocumentRename{
		ResponseMessage: ResponseMessage{
			Message: DefaultMessage,
			ID:      request.ID,
		},
		Result: workspaceEdit,
	}

	return response, nil
}

// collectRenameEdits scans the given content using the AST and returns text edits that replace every
// occurrence of oldName (as an identifier) with newName.
func collectRenameEdits(content, filePath, oldName, newName string) []TextDocumentTextEdit {
	// Parse the content to get the AST
	schema, err := parser.ParserInstance.ParseString(filePath, content)
	if err != nil || schema == nil {
		return nil
	}

	// Find all references in the schema
	references := findReferencesInSchema(schema, oldName, content)

	// Convert to text edits
	var edits []TextDocumentTextEdit
	for _, ref := range references {
		edits = append(edits, TextDocumentTextEdit{
			Range: TextDocumentRange{
				Start: convertASTPositionToLSPPosition(ref.Pos),
				End:   convertASTPositionToLSPPosition(ref.EndPos),
			},
			NewText: newName,
		})
	}

	return edits
}

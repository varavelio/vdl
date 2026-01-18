package lsp

import (
	"fmt"
	"strings"

	"github.com/uforg/uforpc/urpc/internal/urpc/ast"
	"github.com/uforg/uforpc/urpc/internal/urpc/lexer"
	"github.com/uforg/uforpc/urpc/internal/urpc/token"
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

	filePath := request.Params.TextDocument.URI
	position := request.Params.Position
	newName := request.Params.NewName

	// Get the current document content
	content, _, found, err := l.docstore.GetInMemory("", filePath)
	if !found {
		return nil, fmt.Errorf("text document not found in memory: %s", filePath)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get in memory text document: %w", err)
	}

	// Convert LSP position (0-based) to 1-based indices expected by lexer util
	astPosLine := position.Line + 1
	astPosCol := position.Character + 1

	// Determine the identifier to be renamed
	oldName, err := findTokenAtPosition(content, ast.Position{Filename: filePath, Line: astPosLine, Column: astPosCol})
	if err != nil {
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
	edits := l.collectRenameEditsInDocument(content, oldName, newName)

	workspaceEdit := WorkspaceEdit{
		Changes: map[string][]TextDocumentTextEdit{
			filePath: edits,
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

// collectRenameEditsInDocument scans the given content and returns text edits that replace every
// occurrence of oldName (as an identifier) with newName.
func (l *LSP) collectRenameEditsInDocument(content, oldName, newName string) []TextDocumentTextEdit {
	var edits []TextDocumentTextEdit

	lex := lexer.NewLexer("", content)

	for {
		tok := lex.NextToken()
		if tok.Type == token.Eof {
			break
		}

		// We only care about identifiers that exactly match oldName
		if tok.Type != token.Ident {
			continue
		}
		if strings.TrimSpace(tok.Literal) != oldName {
			continue
		}

		// Build range (convert 1-based token positions to 0-based)
		rng := TextDocumentRange{
			Start: TextDocumentPosition{Line: tok.LineStart - 1, Character: tok.ColumnStart - 1},
			End:   TextDocumentPosition{Line: tok.LineEnd - 1, Character: tok.ColumnEnd},
		}

		edits = append(edits, TextDocumentTextEdit{
			Range:   rng,
			NewText: newName,
		})
	}

	return edits
}

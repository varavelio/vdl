package lsp

import (
	"fmt"

	"github.com/varavelio/vdl/toolchain/internal/core/parser"
)

// RequestMessageTextDocumentReferences represents a request for references of a symbol.
type RequestMessageTextDocumentReferences struct {
	RequestMessage
	Params RequestMessageTextDocumentReferencesParams `json:"params"`
}

// RequestMessageTextDocumentReferencesParams are the params for references request.
type RequestMessageTextDocumentReferencesParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     TextDocumentPosition   `json:"position"`
	// We ignore the context.includeDeclaration for simplicity.
}

// ResponseMessageTextDocumentReferences is the response for references request.
type ResponseMessageTextDocumentReferences struct {
	ResponseMessage
	Result []Location `json:"result"`
}

// handleTextDocumentReferences handles textDocument/references.
func (l *LSP) handleTextDocumentReferences(rawMessage []byte) (any, error) {
	var request RequestMessageTextDocumentReferences
	if err := decode(rawMessage, &request); err != nil {
		return nil, fmt.Errorf("failed to decode references request: %w", err)
	}

	filePath := UriToPath(request.Params.TextDocument.URI)
	pos := request.Params.Position

	// Get content
	content, err := l.fs.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file from vfs: %w", err)
	}

	// Determine symbol under cursor
	symbol := findIdentifierAtPosition(string(content), pos)
	if symbol == "" {
		response := ResponseMessageTextDocumentReferences{
			ResponseMessage: ResponseMessage{Message: DefaultMessage, ID: request.ID},
			Result:          nil,
		}
		return response, nil
	}

	locations := collectReferences(string(content), filePath, symbol)
	response := ResponseMessageTextDocumentReferences{
		ResponseMessage: ResponseMessage{Message: DefaultMessage, ID: request.ID},
		Result:          locations,
	}

	return response, nil
}

// collectReferences scans the given content using the AST and returns a list of Locations where symbol appears.
func collectReferences(content, filePath, symbol string) []Location {
	// Parse the content to get the AST
	schema, err := parser.ParserInstance.ParseString(filePath, content)
	if err != nil || schema == nil {
		return nil
	}

	// Find all references in the schema
	references := findReferencesInSchema(schema, symbol)

	// Convert to LSP locations
	var locs []Location
	for _, ref := range references {
		locs = append(locs, Location{
			URI: PathToUri(filePath),
			Range: TextDocumentRange{
				Start: convertASTPositionToLSPPosition(ref.Pos),
				End:   convertASTPositionToLSPPosition(ref.EndPos),
			},
		})
	}

	return locs
}

package lsp

import (
	"fmt"
	"path/filepath"
)

// RequestMessageTextDocumentDefinition represents a request for the definition of a symbol.
type RequestMessageTextDocumentDefinition struct {
	RequestMessage
	Params RequestMessageTextDocumentDefinitionParams `json:"params"`
}

// RequestMessageTextDocumentDefinitionParams represents the parameters for a definition request.
type RequestMessageTextDocumentDefinitionParams struct {
	// The text document.
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	// The position inside the text document.
	Position TextDocumentPosition `json:"position"`
}

// ResponseMessageTextDocumentDefinition represents a response to a definition request.
type ResponseMessageTextDocumentDefinition struct {
	ResponseMessage
	// The result of the request. Can be a single location or an array of locations.
	Result []Location `json:"result"`
}

// Location represents a location inside a resource, such as a line inside a text file.
type Location struct {
	// The URI of the document.
	URI string `json:"uri"`
	// The range inside the document.
	Range TextDocumentRange `json:"range"`
}

// handleTextDocumentDefinition handles a textDocument/definition request.
func (l *LSP) handleTextDocumentDefinition(rawMessage []byte) (any, error) {
	var request RequestMessageTextDocumentDefinition
	if err := decode(rawMessage, &request); err != nil {
		return nil, fmt.Errorf("failed to decode definition request: %w", err)
	}

	filePath := UriToPath(request.Params.TextDocument.URI)
	position := request.Params.Position

	// Get the document content
	content, err := l.fs.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file from vfs: %w", err)
	}

	// Run the analysis to get the program
	program, _ := l.analyze(filePath)

	// First check if we're on an external docstring (.md file reference)
	if mdPath, ok := findDocstringPathAtPosition(string(content), position); ok {
		location := l.findExternalDocstringDefinition(filePath, mdPath)
		if location != nil {
			return ResponseMessageTextDocumentDefinition{
				ResponseMessage: ResponseMessage{
					Message: DefaultMessage,
					ID:      request.ID,
				},
				Result: []Location{*location},
			}, nil
		}
	}

	// Find the identifier at the cursor position
	identifier := findIdentifierAtPosition(string(content), position)
	if identifier == "" {
		return ResponseMessageTextDocumentDefinition{
			ResponseMessage: ResponseMessage{
				Message: DefaultMessage,
				ID:      request.ID,
			},
			Result: nil,
		}, nil
	}

	// Find the definition
	location := resolveSymbolDefinition(program, identifier)

	var result []Location
	if location != nil {
		result = []Location{*location}
	}

	response := ResponseMessageTextDocumentDefinition{
		ResponseMessage: ResponseMessage{
			Message: DefaultMessage,
			ID:      request.ID,
		},
		Result: result,
	}

	return response, nil
}

// findExternalDocstringDefinition finds the location of an external docstring file.
func (l *LSP) findExternalDocstringDefinition(currentFile, mdPath string) *Location {
	// Resolve the path relative to the current file
	baseDir := filepath.Dir(currentFile)
	resolvedPath := filepath.Join(baseDir, mdPath)
	resolvedPath = filepath.Clean(resolvedPath)

	// Check if the file exists by trying to read it
	_, err := l.fs.ReadFile(resolvedPath)
	if err != nil {
		return nil
	}

	return &Location{
		URI: PathToUri(resolvedPath),
		Range: TextDocumentRange{
			Start: TextDocumentPosition{Line: 0, Character: 0},
			End:   TextDocumentPosition{Line: 0, Character: 0},
		},
	}
}

package lsp

import (
	"fmt"
	"strings"

	"github.com/uforg/uforpc/urpc/internal/urpc/ast"
	"github.com/uforg/uforpc/urpc/internal/urpc/docstore"
	"github.com/uforg/uforpc/urpc/internal/urpc/lexer"
	"github.com/uforg/uforpc/urpc/internal/urpc/token"
	"github.com/uforg/uforpc/urpc/internal/util/filepathutil"
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

	filePath := request.Params.TextDocument.URI
	position := request.Params.Position

	// Get the document content
	content, _, found, err := l.docstore.GetInMemory("", filePath)
	if !found {
		return nil, fmt.Errorf("text document not found in memory: %s", filePath)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get in memory text document: %w", err)
	}

	// Run the analyzer to get the combined schema
	astSchema, _, err := l.analyzer.Analyze(filePath)
	if err != nil {
		return ResponseMessageTextDocumentDefinition{
			ResponseMessage: ResponseMessage{
				Message: DefaultMessage,
				ID:      request.ID,
			},
			Result: nil,
		}, nil
	}

	// Convert LSP position (0-based) to AST position (1-based)
	astPosition := ast.Position{
		Filename: filePath,
		Line:     position.Line + 1,
		Column:   position.Character + 1,
	}

	// Find the definition
	locations := l.findDefinition(content, astPosition, astSchema)

	response := ResponseMessageTextDocumentDefinition{
		ResponseMessage: ResponseMessage{
			Message: DefaultMessage,
			ID:      request.ID,
		},
		Result: locations,
	}

	return response, nil
}

// findDefinition finds the definition of a symbol at the given position.
func (l *LSP) findDefinition(content string, position ast.Position, astSchema *ast.Schema) []Location {
	// We don't need to parse the document here since we're using the token finder
	// to extract the token at the position and then look it up in the astSchema

	// Find the tokenLiteral at the position
	tokenLiteral, err := findTokenAtPosition(content, position)
	if err != nil {
		return nil
	}

	if strings.HasSuffix(tokenLiteral, ".md") {
		if location := findExternalDocstringDefinition(l.docstore, position, tokenLiteral); location != nil {
			return []Location{*location}
		}

		return nil
	}

	// Check if the tokenLiteral is a reference to a type
	if location := findTypeDefinition(tokenLiteral, astSchema); location != nil {
		return []Location{*location}
	}

	return nil
}

// findTokenAtPosition finds the token at the given position in the content.
func findTokenAtPosition(content string, position ast.Position) (string, error) {
	lex := lexer.NewLexer("", content)

	for {
		tok := lex.NextToken()
		if tok.Type == token.Eof {
			break
		}

		// Skip non-identifier and non-docstring tokens
		if tok.Type != token.Ident && tok.Type != token.Docstring {
			continue
		}

		// Skip docstrings that are not markdown files
		if tok.Type == token.Docstring {
			_, isExternal := ast.DocstringIsExternal(tok.Literal)
			if !isExternal {
				continue
			}
		}

		matchLine := tok.LineStart <= position.Line && tok.LineEnd >= position.Line
		matchColumn := tok.ColumnStart <= position.Column && tok.ColumnEnd >= position.Column
		match := matchLine && matchColumn

		if match {
			return strings.TrimSpace(tok.Literal), nil
		}
	}

	return "", fmt.Errorf("no token at position")
}

// findExternalDocstringDefinition finds the definition of an external docstring.
func findExternalDocstringDefinition(docstore *docstore.Docstore, position ast.Position, tokenLiteral string) *Location {
	// Check if the file exists in the docstore
	_, _, err := docstore.GetFileAndHash(position.Filename, tokenLiteral)
	if err != nil {
		return nil
	}

	normFilePath, err := filepathutil.Normalize(position.Filename, tokenLiteral)
	if err != nil {
		return nil
	}

	return &Location{
		URI: "file://" + strings.TrimPrefix(normFilePath, "file://"),
		Range: TextDocumentRange{
			Start: TextDocumentPosition{Line: 1, Character: 1},
			End:   TextDocumentPosition{Line: 1, Character: 1},
		},
	}
}

// findTypeDefinition finds the definition of a type.
func findTypeDefinition(tokenLiteral string, astSchema *ast.Schema) *Location {
	// Check if the token is a type name
	typeDecl, exists := astSchema.GetTypesMap()[tokenLiteral]
	if !exists {
		return nil
	}

	// Create a location for the type definition
	// Ensure the URI has the file:// prefix
	uri := typeDecl.Pos.Filename
	if !strings.HasPrefix(uri, "file://") {
		uri = "file://" + uri
	}

	return &Location{
		URI: uri,
		Range: TextDocumentRange{
			Start: convertASTPositionToLSPPosition(typeDecl.Pos),
			End:   convertASTPositionToLSPPosition(typeDecl.EndPos),
		},
	}
}

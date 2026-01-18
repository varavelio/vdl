package lsp

import (
	"fmt"
	"strings"

	"github.com/uforg/uforpc/urpc/internal/urpc/ast"
	"github.com/uforg/uforpc/urpc/internal/urpc/lexer"
	"github.com/uforg/uforpc/urpc/internal/urpc/token"
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

	uri := request.Params.TextDocument.URI
	pos := request.Params.Position

	// Get content
	content, _, found, err := l.docstore.GetInMemory("", uri)
	if !found {
		return nil, fmt.Errorf("text document not found in memory: %s", uri)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get document content: %w", err)
	}

	// Determine symbol under cursor
	symbol, err := findTokenAtPosition(content, ast.Position{Filename: uri, Line: pos.Line + 1, Column: pos.Character + 1})
	if err != nil {
		response := ResponseMessageTextDocumentReferences{
			ResponseMessage: ResponseMessage{Message: DefaultMessage, ID: request.ID},
			Result:          nil,
		}
		return response, nil
	}

	locations := l.collectReferences(content, uri, symbol)
	response := ResponseMessageTextDocumentReferences{
		ResponseMessage: ResponseMessage{Message: DefaultMessage, ID: request.ID},
		Result:          locations,
	}

	return response, nil
}

// collectReferences scans the given content and returns a list of Locations where symbol appears.
func (l *LSP) collectReferences(content, uri, symbol string) []Location {
	var locs []Location

	lex := lexer.NewLexer(uri, content)

	for {
		tok := lex.NextToken()
		if tok.Type == token.Eof {
			break
		}
		if tok.Type != token.Ident {
			continue
		}
		if strings.TrimSpace(tok.Literal) != symbol {
			continue
		}

		uriWithScheme := uri
		if !strings.HasPrefix(uriWithScheme, "file://") {
			uriWithScheme = "file://" + uriWithScheme
		}

		locs = append(locs, Location{
			URI: uriWithScheme,
			Range: TextDocumentRange{
				Start: TextDocumentPosition{Line: tok.LineStart - 1, Character: tok.ColumnStart - 1},
				End:   TextDocumentPosition{Line: tok.LineEnd - 1, Character: tok.ColumnEnd},
			},
		})
	}

	return locs
}

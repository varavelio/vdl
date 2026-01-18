package lsp

import (
	"fmt"
	"strings"

	"github.com/uforg/uforpc/urpc/internal/urpc/ast"
	"github.com/uforg/uforpc/urpc/internal/urpc/lexer"
	"github.com/uforg/uforpc/urpc/internal/urpc/token"
	"github.com/uforg/uforpc/urpc/internal/util/filepathutil"
)

// RequestMessageTextDocumentDocumentLink represents a request for document links inside a text document.
type RequestMessageTextDocumentDocumentLink struct {
	RequestMessage
	Params RequestMessageTextDocumentDocumentLinkParams `json:"params"`
}

// RequestMessageTextDocumentDocumentLinkParams are the params for the documentLink request.
type RequestMessageTextDocumentDocumentLinkParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

// ResponseMessageTextDocumentDocumentLink represents the response containing document links.
type ResponseMessageTextDocumentDocumentLink struct {
	ResponseMessage
	Result []DocumentLink `json:"result"`
}

// DocumentLink represents a link inside the document.
type DocumentLink struct {
	Range   TextDocumentRange `json:"range"`
	Target  string            `json:"target"`
	Tooltip string            `json:"tooltip,omitempty"`
}

// handleTextDocumentDocumentLink handles a textDocument/documentLink request.
func (l *LSP) handleTextDocumentDocumentLink(rawMessage []byte) (any, error) {
	var request RequestMessageTextDocumentDocumentLink
	if err := decode(rawMessage, &request); err != nil {
		return nil, fmt.Errorf("failed to decode documentLink request: %w", err)
	}

	uri := request.Params.TextDocument.URI

	// Fetch document content
	content, _, found, err := l.docstore.GetInMemory("", uri)
	if !found {
		return nil, fmt.Errorf("text document not found in memory: %s", uri)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get document content: %w", err)
	}

	links := l.collectDocumentLinks(content, uri)

	response := ResponseMessageTextDocumentDocumentLink{
		ResponseMessage: ResponseMessage{
			Message: DefaultMessage,
			ID:      request.ID,
		},
		Result: links,
	}

	return response, nil
}

// collectDocumentLinks scans the document content for external docstring references and returns DocumentLink slices.
func (l *LSP) collectDocumentLinks(content string, docURI string) []DocumentLink {
	var links []DocumentLink

	lex := lexer.NewLexer(docURI, content)

	for {
		tok := lex.NextToken()
		if tok.Type == token.Eof {
			break
		}
		if tok.Type != token.Docstring {
			continue
		}
		trimmed, isExternal := ast.DocstringIsExternal(tok.Literal)
		if !isExternal {
			continue
		}

		// Resolve path relative to document
		normPath, err := filepathutil.Normalize(docURI, trimmed)
		if err != nil {
			continue // skip invalid paths
		}
		if !strings.HasPrefix(normPath, "file://") {
			normPath = "file://" + normPath
		}

		rng := TextDocumentRange{
			Start: TextDocumentPosition{Line: tok.LineStart - 1, Character: tok.ColumnStart - 1},
			End:   TextDocumentPosition{Line: tok.LineEnd - 1, Character: tok.ColumnEnd},
		}

		links = append(links, DocumentLink{
			Range:   rng,
			Target:  normPath,
			Tooltip: "Open markdown file",
		})
	}

	return links
}

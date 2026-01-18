package lsp

import (
	"fmt"
	"sort"
	"strings"

	"github.com/uforg/uforpc/urpc/internal/urpc/ast"
	"github.com/uforg/uforpc/urpc/internal/urpc/lexer"
	"github.com/uforg/uforpc/urpc/internal/urpc/token"
)

const (
	CompletionItemKindText          = 1
	CompletionItemKindMethod        = 2
	CompletionItemKindFunction      = 3
	CompletionItemKindConstructor   = 4
	CompletionItemKindField         = 5
	CompletionItemKindVariable      = 6
	CompletionItemKindClass         = 7
	CompletionItemKindInterface     = 8
	CompletionItemKindModule        = 9
	CompletionItemKindProperty      = 10
	CompletionItemKindUnit          = 11
	CompletionItemKindValue         = 12
	CompletionItemKindEnum          = 13
	CompletionItemKindKeyword       = 14
	CompletionItemKindSnippet       = 15
	CompletionItemKindColor         = 16
	CompletionItemKindFile          = 17
	CompletionItemKindReference     = 18
	CompletionItemKindFolder        = 19
	CompletionItemKindEnumMember    = 20
	CompletionItemKindConstant      = 21
	CompletionItemKindStruct        = 22
	CompletionItemKindEvent         = 23
	CompletionItemKindOperator      = 24
	CompletionItemKindTypeParameter = 25
)

// Request/Response structures

type RequestMessageTextDocumentCompletion struct {
	RequestMessage
	Params RequestMessageTextDocumentCompletionParams `json:"params"`
}

type RequestMessageTextDocumentCompletionParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     TextDocumentPosition   `json:"position"`
}

type ResponseMessageTextDocumentCompletion struct {
	ResponseMessage
	Result []CompletionItem `json:"result"`
}

type CompletionItem struct {
	Label string `json:"label"`
	Kind  int    `json:"kind,omitempty"`
}

// handleTextDocumentCompletion provides completion after ": " in a field definition.
func (l *LSP) handleTextDocumentCompletion(rawMessage []byte) (any, error) {
	var request RequestMessageTextDocumentCompletion
	if err := decode(rawMessage, &request); err != nil {
		return nil, fmt.Errorf("failed to decode completion request: %w", err)
	}

	uri := request.Params.TextDocument.URI
	pos := request.Params.Position

	content, _, found, err := l.docstore.GetInMemory("", uri)
	if !found {
		return nil, fmt.Errorf("text document not found in memory: %s", uri)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get document content: %w", err)
	}

	prefix, ok := getFieldTypePrefix(content, pos)
	if !ok {
		// Not appropriate context; return empty result
		resp := ResponseMessageTextDocumentCompletion{
			ResponseMessage: ResponseMessage{Message: DefaultMessage, ID: request.ID},
			Result:          nil,
		}
		return resp, nil
	}

	// Collect primitive types
	itemsMap := map[string]int{}
	for _, prim := range ast.PrimitiveTypes {
		itemsMap[prim] = CompletionItemKindValue
	}

	// Collect custom types via lexer to avoid parse errors in incomplete schemas
	for _, tName := range collectCustomTypes(content, uri) {
		itemsMap[tName] = CompletionItemKindValue
	}

	// Convert to slice and sort alphabetically, applying prefix filter
	labels := make([]string, 0, len(itemsMap))
	for label := range itemsMap {
		if prefix == "" || strings.HasPrefix(strings.ToLower(label), strings.ToLower(prefix)) {
			labels = append(labels, label)
		}
	}
	sort.Strings(labels)

	var completions []CompletionItem
	for _, label := range labels {
		completions = append(completions, CompletionItem{Label: label, Kind: itemsMap[label]})
	}

	response := ResponseMessageTextDocumentCompletion{
		ResponseMessage: ResponseMessage{Message: DefaultMessage, ID: request.ID},
		Result:          completions,
	}
	return response, nil
}

// getFieldTypePrefix returns the current prefix typed after a field's ": " and whether the context matches.
func getFieldTypePrefix(content string, pos TextDocumentPosition) (string, bool) {
	lines := strings.Split(content, "\n")
	if pos.Line >= len(lines) {
		return "", false
	}
	line := lines[pos.Line]
	if pos.Character > len(line) {
		return "", false
	}
	beforeCursor := line[:pos.Character]
	idx := strings.LastIndex(beforeCursor, ":")
	if idx == -1 {
		return "", false
	}
	// Ensure only whitespace between ':' and cursor
	segment := beforeCursor[idx+1:]
	trimmed := strings.TrimLeft(segment, " \t")
	if strings.Contains(trimmed, " ") || strings.Contains(trimmed, "\t") {
		return "", false
	}
	return trimmed, true
}

// collectCustomTypes scans the document with the lexer and returns type names defined via "Type Ident".
func collectCustomTypes(content, uri string) []string {
	lex := lexer.NewLexer(uri, content)
	var types []string

	for {
		tok := lex.NextToken()
		if tok.Type == token.Eof {
			break
		}

		if tok.Type != token.Type {
			continue
		}

		// Skip whitespace/newline tokens to find the next meaningful token
		nextTok := lex.NextToken()
		for nextTok.Type == token.Newline || nextTok.Type == token.Whitespace {
			nextTok = lex.NextToken()
		}

		if nextTok.Type == token.Ident {
			types = append(types, strings.TrimSpace(nextTok.Literal))
		}
	}

	// Deduplicate
	unique := make(map[string]struct{})
	for _, t := range types {
		unique[t] = struct{}{}
	}
	res := make([]string, 0, len(unique))
	for t := range unique {
		res = append(res, t)
	}
	return res
}

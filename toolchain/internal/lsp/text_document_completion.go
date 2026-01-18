package lsp

import (
	"fmt"
	"sort"
	"strings"

	"github.com/varavelio/vdl/toolchain/internal/core/ast"
	"github.com/varavelio/vdl/toolchain/internal/core/parser"
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

	filePath := uriToPath(request.Params.TextDocument.URI)
	pos := request.Params.Position

	content, err := l.fs.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file from vfs: %w", err)
	}

	prefix, ok := getFieldTypePrefix(string(content), pos)
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

	// Collect custom types from the parsed schema
	customTypes := collectCustomTypesFromContent(string(content), filePath)
	for _, tName := range customTypes {
		itemsMap[tName] = CompletionItemKindStruct
	}

	// Collect enums from the parsed schema
	customEnums := collectEnumsFromContent(string(content), filePath)
	for _, eName := range customEnums {
		itemsMap[eName] = CompletionItemKindEnum
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

// collectCustomTypesFromContent parses the content and returns type names.
// Note: We ignore parse errors because the parser returns partial results
// which are useful for completion even in incomplete/invalid schemas.
func collectCustomTypesFromContent(content, uri string) []string {
	schema, _ := parser.ParserInstance.ParseString(uri, content)
	if schema == nil {
		return nil
	}

	var types []string
	for _, t := range schema.GetTypes() {
		if t.Name != "" {
			types = append(types, t.Name)
		}
	}

	return types
}

// collectEnumsFromContent parses the content and returns enum names.
// Note: We ignore parse errors because the parser returns partial results
// which are useful for completion even in incomplete/invalid schemas.
func collectEnumsFromContent(content, uri string) []string {
	schema, _ := parser.ParserInstance.ParseString(uri, content)
	if schema == nil {
		return nil
	}

	var enums []string
	for _, e := range schema.GetEnums() {
		if e.Name != "" {
			enums = append(enums, e.Name)
		}
	}

	return enums
}

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

	filePath := UriToPath(request.Params.TextDocument.URI)
	pos := request.Params.Position

	content, err := l.fs.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file from vfs: %w", err)
	}

	// Determine completion context
	prefix, ctxKind, ok := getCompletionContext(string(content), pos)
	if !ok {
		// Not appropriate context; return empty result
		resp := ResponseMessageTextDocumentCompletion{
			ResponseMessage: ResponseMessage{Message: DefaultMessage, ID: request.ID},
			Result:          nil,
		}
		return resp, nil
	}

	// Determine which types to include
	includePrimitives := false
	includeEnums := false
	includeTypes := false

	switch ctxKind {
	case CompletionContextFieldType:
		includePrimitives = true
		includeEnums = true
		includeTypes = true
	case CompletionContextSpread:
		includePrimitives = false
		includeEnums = false
		includeTypes = true
	}

	itemsMap := map[string]int{}

	// Collect primitive types
	if includePrimitives {
		for _, prim := range ast.PrimitiveTypes {
			itemsMap[prim] = CompletionItemKindValue
		}
	}

	// Identify files to scan: current file + all dependencies
	filesToScan := []string{filePath}
	dependencies := l.depGraph.GetAllDependencies(filePath)
	filesToScan = append(filesToScan, dependencies...)

	for _, fPath := range filesToScan {
		fContentBytes, err := l.fs.ReadFile(fPath)
		if err != nil {
			continue
		}
		fContent := string(fContentBytes)

		// Collect custom types
		if includeTypes {
			customTypes := collectCustomTypesFromContent(fContent, fPath)
			for _, tName := range customTypes {
				itemsMap[tName] = CompletionItemKindStruct
			}
		}

		// Collect enums
		if includeEnums {
			customEnums := collectEnumsFromContent(fContent, fPath)
			for _, eName := range customEnums {
				itemsMap[eName] = CompletionItemKindEnum
			}
		}
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

type CompletionContextKind int

const (
	CompletionContextFieldType CompletionContextKind = iota
	CompletionContextSpread
)

// getCompletionContext returns the completion prefix and kind.
func getCompletionContext(content string, pos TextDocumentPosition) (string, CompletionContextKind, bool) {
	lines := strings.Split(content, "\n")
	if pos.Line >= len(lines) {
		return "", 0, false
	}
	line := lines[pos.Line]
	if pos.Character > len(line) {
		return "", 0, false
	}
	beforeCursor := line[:pos.Character]

	idx := len(beforeCursor) - 1

	// Consume Prefix (Identifier chars)
	for idx >= 0 && isIdentifierChar(beforeCursor[idx]) {
		idx--
	}
	prefix := beforeCursor[idx+1:]

	// Consume Whitespace
	for idx >= 0 && (beforeCursor[idx] == ' ' || beforeCursor[idx] == '\t') {
		idx--
	}

	if idx < 0 {
		return "", 0, false
	}

	// Check Delimiter
	lastChar := beforeCursor[idx]

	// Case A: Field Definition "Name: Type"
	if lastChar == ':' {
		return prefix, CompletionContextFieldType, true
	}

	// Case B: Map "map<Type"
	if lastChar == '<' {
		// Check for "map" keyword before '<'
		// We need to check text BEFORE '<'.

		mapEndIdx := idx - 1
		// Consume whitespace between 'map' and '<'
		for mapEndIdx >= 0 && (beforeCursor[mapEndIdx] == ' ' || beforeCursor[mapEndIdx] == '\t') {
			mapEndIdx--
		}

		// Check for "map"
		if mapEndIdx >= 2 &&
			beforeCursor[mapEndIdx] == 'p' &&
			beforeCursor[mapEndIdx-1] == 'a' &&
			beforeCursor[mapEndIdx-2] == 'm' {

			// Verify word boundary
			if mapEndIdx-3 < 0 || !isIdentifierChar(beforeCursor[mapEndIdx-3]) {
				return prefix, CompletionContextFieldType, true
			}
		}
		return "", 0, false
	}

	// Case C: Spread "...Type"
	if lastChar == '.' {
		// Check for "..."
		if idx >= 2 && beforeCursor[idx-1] == '.' && beforeCursor[idx-2] == '.' {
			return prefix, CompletionContextSpread, true
		}
	}

	return "", 0, false
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

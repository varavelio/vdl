package lsp

import (
	"fmt"
	"strings"

	"github.com/uforg/uforpc/urpc/internal/urpc/ast"
)

// SymbolKind values (subset of LSP specification)
const (
	SymbolKindFile          = 1
	SymbolKindModule        = 2
	SymbolKindNamespace     = 3
	SymbolKindPackage       = 4
	SymbolKindClass         = 5
	SymbolKindMethod        = 6
	SymbolKindProperty      = 7
	SymbolKindField         = 8
	SymbolKindConstructor   = 9
	SymbolKindEnum          = 10
	SymbolKindInterface     = 11
	SymbolKindFunction      = 12
	SymbolKindVariable      = 13
	SymbolKindConstant      = 14
	SymbolKindString        = 15
	SymbolKindNumber        = 16
	SymbolKindBoolean       = 17
	SymbolKindArray         = 18
	SymbolKindObject        = 19
	SymbolKindKey           = 20
	SymbolKindNull          = 21
	SymbolKindEnumMember    = 22
	SymbolKindStruct        = 23
	SymbolKindEvent         = 24
	SymbolKindOperator      = 25
	SymbolKindTypeParameter = 26
)

// Request / response types

type RequestMessageTextDocumentDocumentSymbol struct {
	RequestMessage
	Params RequestMessageTextDocumentDocumentSymbolParams `json:"params"`
}

type RequestMessageTextDocumentDocumentSymbolParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

type ResponseMessageTextDocumentDocumentSymbol struct {
	ResponseMessage
	Result []DocumentSymbol `json:"result"`
}

// DocumentSymbol represents a symbol information with hierarchy.

type DocumentSymbol struct {
	Name           string            `json:"name"`
	Detail         string            `json:"detail,omitempty"`
	Kind           int               `json:"kind"`
	Range          TextDocumentRange `json:"range"`
	SelectionRange TextDocumentRange `json:"selectionRange"`
	Children       []DocumentSymbol  `json:"children,omitempty"`
}

// handleTextDocumentDocumentSymbol handles textDocument/documentSymbol request.
func (l *LSP) handleTextDocumentDocumentSymbol(rawMessage []byte) (any, error) {
	var request RequestMessageTextDocumentDocumentSymbol
	if err := decode(rawMessage, &request); err != nil {
		return nil, fmt.Errorf("failed to decode documentSymbol request: %w", err)
	}

	uri := request.Params.TextDocument.URI

	// Run analyzer to get AST schema
	astSchema, _, err := l.analyzer.Analyze(uri)
	if err != nil {
		// Return empty result but no error (so client still gets response)
		resp := ResponseMessageTextDocumentDocumentSymbol{
			ResponseMessage: ResponseMessage{Message: DefaultMessage, ID: request.ID},
			Result:          nil,
		}
		return resp, nil
	}

	symbols := l.buildDocumentSymbols(astSchema)
	response := ResponseMessageTextDocumentDocumentSymbol{
		ResponseMessage: ResponseMessage{Message: DefaultMessage, ID: request.ID},
		Result:          symbols,
	}
	return response, nil
}

// buildDocumentSymbols converts the AST schema to LSP document symbols.
func (l *LSP) buildDocumentSymbols(schema *ast.Schema) []DocumentSymbol {
	var symbols []DocumentSymbol

	for _, ds := range schema.GetDocstrings() {
		name := strings.TrimSpace(ds.Value)
		name = strings.Split(name, "\n")[0]
		name = strings.ReplaceAll(name, "#", "")
		name = strings.TrimSpace(name)
		if len(name) > 30 {
			name = name[:30] + "..."
		}
		if name == "" {
			continue
		}

		sym := DocumentSymbol{
			Name:           name,
			Kind:           SymbolKindString,
			Range:          TextDocumentRange{Start: convertASTPositionToLSPPosition(ds.Pos), End: convertASTPositionToLSPPosition(ds.EndPos)},
			SelectionRange: TextDocumentRange{Start: convertASTPositionToLSPPosition(ds.Pos), End: convertASTPositionToLSPPosition(ds.Pos)},
		}
		symbols = append(symbols, sym)
	}

	for _, t := range schema.GetTypes() {
		sym := DocumentSymbol{
			Name:           t.Name,
			Kind:           SymbolKindStruct,
			Range:          TextDocumentRange{Start: convertASTPositionToLSPPosition(t.Pos), End: convertASTPositionToLSPPosition(t.EndPos)},
			SelectionRange: TextDocumentRange{Start: convertASTPositionToLSPPosition(t.Pos), End: convertASTPositionToLSPPosition(t.Pos)},
		}
		symbols = append(symbols, sym)
	}

	for _, p := range schema.GetProcs() {
		procSym := DocumentSymbol{
			Name:           p.Name,
			Kind:           SymbolKindFunction,
			Range:          TextDocumentRange{Start: convertASTPositionToLSPPosition(p.Pos), End: convertASTPositionToLSPPosition(p.EndPos)},
			SelectionRange: TextDocumentRange{Start: convertASTPositionToLSPPosition(p.Pos), End: convertASTPositionToLSPPosition(p.Pos)},
		}

		// Build children (input/output)
		for _, child := range p.Children {
			if child.Input != nil {
				c := DocumentSymbol{
					Name:           "input",
					Kind:           SymbolKindObject,
					Range:          TextDocumentRange{Start: convertASTPositionToLSPPosition(child.Input.Pos), End: convertASTPositionToLSPPosition(child.Input.EndPos)},
					SelectionRange: TextDocumentRange{Start: convertASTPositionToLSPPosition(child.Input.Pos), End: convertASTPositionToLSPPosition(child.Input.Pos)},
				}
				procSym.Children = append(procSym.Children, c)
			}
			if child.Output != nil {
				c := DocumentSymbol{
					Name:           "output",
					Kind:           SymbolKindObject,
					Range:          TextDocumentRange{Start: convertASTPositionToLSPPosition(child.Output.Pos), End: convertASTPositionToLSPPosition(child.Output.EndPos)},
					SelectionRange: TextDocumentRange{Start: convertASTPositionToLSPPosition(child.Output.Pos), End: convertASTPositionToLSPPosition(child.Output.Pos)},
				}
				procSym.Children = append(procSym.Children, c)
			}
		}

		symbols = append(symbols, procSym)
	}

	for _, s := range schema.GetStreams() {
		streamSym := DocumentSymbol{
			Name:           s.Name,
			Kind:           SymbolKindEvent,
			Range:          TextDocumentRange{Start: convertASTPositionToLSPPosition(s.Pos), End: convertASTPositionToLSPPosition(s.EndPos)},
			SelectionRange: TextDocumentRange{Start: convertASTPositionToLSPPosition(s.Pos), End: convertASTPositionToLSPPosition(s.Pos)},
		}

		// Children (input/output)
		for _, child := range s.Children {
			if child.Input != nil {
				c := DocumentSymbol{
					Name:           "input",
					Kind:           SymbolKindObject,
					Range:          TextDocumentRange{Start: convertASTPositionToLSPPosition(child.Input.Pos), End: convertASTPositionToLSPPosition(child.Input.EndPos)},
					SelectionRange: TextDocumentRange{Start: convertASTPositionToLSPPosition(child.Input.Pos), End: convertASTPositionToLSPPosition(child.Input.Pos)},
				}
				streamSym.Children = append(streamSym.Children, c)
			}
			if child.Output != nil {
				c := DocumentSymbol{
					Name:           "output",
					Kind:           SymbolKindObject,
					Range:          TextDocumentRange{Start: convertASTPositionToLSPPosition(child.Output.Pos), End: convertASTPositionToLSPPosition(child.Output.EndPos)},
					SelectionRange: TextDocumentRange{Start: convertASTPositionToLSPPosition(child.Output.Pos), End: convertASTPositionToLSPPosition(child.Output.Pos)},
				}
				streamSym.Children = append(streamSym.Children, c)
			}
		}

		symbols = append(symbols, streamSym)
	}

	return symbols
}

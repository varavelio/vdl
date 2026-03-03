package lsp

import (
	"context"
	"fmt"
	"strings"

	"github.com/varavelio/vdl/toolchain/internal/core/analysis"
)

// SymbolKind values (subset of LSP specification).
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

	filePath := UriToPath(request.Params.TextDocument.URI)

	// Run analysis to get the program
	program, _ := l.analyze(context.Background(), filePath)

	symbols := buildDocumentSymbols(program, filePath)
	response := ResponseMessageTextDocumentDocumentSymbol{
		ResponseMessage: ResponseMessage{Message: DefaultMessage, ID: request.ID},
		Result:          symbols,
	}
	return response, nil
}

// buildDocumentSymbols converts the program to LSP document symbols.
// Only includes symbols that are defined in the given file.
func buildDocumentSymbols(program *analysis.Program, filePath string) []DocumentSymbol {
	var symbols []DocumentSymbol

	// Get the AST for this file to access standalone docstrings
	if file, ok := program.Files[filePath]; ok && file.AST != nil {
		// Standalone docstrings
		for _, ds := range file.AST.GetDocstrings() {
			name := strings.TrimSpace(string(ds.Value))
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
	}

	// Types defined in this file
	for _, t := range program.Types {
		if t.File != filePath {
			continue
		}
		sym := DocumentSymbol{
			Name:           t.Name,
			Kind:           SymbolKindStruct,
			Range:          TextDocumentRange{Start: convertASTPositionToLSPPosition(t.Pos), End: convertASTPositionToLSPPosition(t.EndPos)},
			SelectionRange: TextDocumentRange{Start: convertASTPositionToLSPPosition(t.Pos), End: convertASTPositionToLSPPosition(t.Pos)},
		}
		symbols = append(symbols, sym)
	}

	// Enums defined in this file
	for _, e := range program.Enums {
		if e.File != filePath {
			continue
		}
		sym := DocumentSymbol{
			Name:           e.Name,
			Kind:           SymbolKindEnum,
			Range:          TextDocumentRange{Start: convertASTPositionToLSPPosition(e.Pos), End: convertASTPositionToLSPPosition(e.EndPos)},
			SelectionRange: TextDocumentRange{Start: convertASTPositionToLSPPosition(e.Pos), End: convertASTPositionToLSPPosition(e.Pos)},
		}
		symbols = append(symbols, sym)
	}

	// Constants defined in this file
	for _, c := range program.Consts {
		if c.File != filePath {
			continue
		}
		sym := DocumentSymbol{
			Name:           c.Name,
			Kind:           SymbolKindConstant,
			Range:          TextDocumentRange{Start: convertASTPositionToLSPPosition(c.Pos), End: convertASTPositionToLSPPosition(c.EndPos)},
			SelectionRange: TextDocumentRange{Start: convertASTPositionToLSPPosition(c.Pos), End: convertASTPositionToLSPPosition(c.Pos)},
		}
		symbols = append(symbols, sym)
	}

	for _, e := range program.Enums {
		if e.File != filePath {
			continue
		}

		for _, m := range e.Members {
			sym := DocumentSymbol{
				Name:           m.Name,
				Kind:           SymbolKindEnumMember,
				Range:          TextDocumentRange{Start: convertASTPositionToLSPPosition(m.Pos), End: convertASTPositionToLSPPosition(m.EndPos)},
				SelectionRange: TextDocumentRange{Start: convertASTPositionToLSPPosition(m.Pos), End: convertASTPositionToLSPPosition(m.Pos)},
			}
			symbols = append(symbols, sym)
		}
	}

	return symbols
}

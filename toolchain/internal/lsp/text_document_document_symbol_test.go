package lsp

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHandleTextDocumentDocumentSymbol(t *testing.T) {
	schema := `""" Title """

type Person {}

enum Status {
  ONLINE
}
`
	uri := "file:///symbols.vdl"
	l := newTestLSP(t, schema, uri)

	req := RequestMessageTextDocumentDocumentSymbol{
		RequestMessage: RequestMessage{Message: Message{JSONRPC: "2.0", Method: "textDocument/documentSymbol", ID: "1"}},
		Params:         RequestMessageTextDocumentDocumentSymbolParams{TextDocument: TextDocumentIdentifier{URI: uri}},
	}
	b, _ := json.Marshal(req)
	anyResp, err := l.handleTextDocumentDocumentSymbol(b)
	require.NoError(t, err)
	resp := anyResp.(ResponseMessageTextDocumentDocumentSymbol)
	// We expect at least: top-level docstring, type, enum, and enum member symbols.
	require.GreaterOrEqual(t, len(resp.Result), 3)
}

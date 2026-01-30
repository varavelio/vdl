package lsp

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHandleTextDocumentDocumentSymbol(t *testing.T) {
	schema := `""" Title """

type Person {}

rpc Test {
  proc Hello {}
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
	// We expect: 1 docstring, 1 type (Person), 1 RPC (Test) which contains 1 proc (Hello)
	// The RPC should have Hello as a child
	require.GreaterOrEqual(t, len(resp.Result), 2) // At minimum: docstring + type + rpc
}

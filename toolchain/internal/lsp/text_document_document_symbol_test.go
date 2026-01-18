package lsp

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHandleTextDocumentDocumentSymbol(t *testing.T) {
	schema := `version 1

""" Title """

type Person {}
proc Hello {}
`
	uri := "file:///symbols.urpc"
	l := newTestLSP(t, schema, uri)
	// analyze so no parse error
	_, _, _ = l.analyzer.Analyze(uri)

	req := RequestMessageTextDocumentDocumentSymbol{
		RequestMessage: RequestMessage{Message: Message{JSONRPC: "2.0", Method: "textDocument/documentSymbol", ID: "1"}},
		Params:         RequestMessageTextDocumentDocumentSymbolParams{TextDocument: TextDocumentIdentifier{URI: uri}},
	}
	b, _ := json.Marshal(req)
	anyResp, err := l.handleTextDocumentDocumentSymbol(b)
	require.NoError(t, err)
	resp := anyResp.(ResponseMessageTextDocumentDocumentSymbol)
	require.Equal(t, len(resp.Result), 3)
}

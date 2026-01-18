package lsp

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHandleTextDocumentDocumentLink(t *testing.T) {
	schema := `version 1

""" ./doc.md """`
	uri := "file:///links.urpc"
	l := newTestLSP(t, schema, uri)

	req := RequestMessageTextDocumentDocumentLink{
		RequestMessage: RequestMessage{Message: Message{JSONRPC: "2.0", Method: "textDocument/documentLink", ID: "1"}},
		Params:         RequestMessageTextDocumentDocumentLinkParams{TextDocument: TextDocumentIdentifier{URI: uri}},
	}
	b, _ := json.Marshal(req)
	anyResp, err := l.handleTextDocumentDocumentLink(b)
	require.NoError(t, err)
	resp := anyResp.(ResponseMessageTextDocumentDocumentLink)
	require.Len(t, resp.Result, 1)
	require.Contains(t, resp.Result[0].Target, "doc.md")
}

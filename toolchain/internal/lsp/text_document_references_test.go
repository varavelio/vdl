package lsp

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHandleTextDocumentReferences(t *testing.T) {
	schema := `version 1

type Foo {}

proc Bar {
  input { foo: Foo }
  output { ok: bool }
}
`
	uri := "file:///refs.urpc"
	l := newTestLSP(t, schema, uri)

	req := RequestMessageTextDocumentReferences{
		RequestMessage: RequestMessage{Message: Message{JSONRPC: "2.0", Method: "textDocument/references", ID: "1"}},
		Params: RequestMessageTextDocumentReferencesParams{
			TextDocument: TextDocumentIdentifier{URI: uri},
			Position:     TextDocumentPosition{Line: 5, Character: 15}, // Foo reference
		},
	}
	b, _ := json.Marshal(req)
	anyResp, err := l.handleTextDocumentReferences(b)
	require.NoError(t, err)
	resp := anyResp.(ResponseMessageTextDocumentReferences)
	require.Equal(t, len(resp.Result), 2)
}

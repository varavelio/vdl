package lsp

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHandleTextDocumentReferences(t *testing.T) {
	schema := `type Foo {}

rpc Test {
  proc Bar {
    input { foo: Foo }
    output { ok: bool }
  }
}
`
	uri := "file:///refs.vdl"
	l := newTestLSP(t, schema, uri)

	req := RequestMessageTextDocumentReferences{
		RequestMessage: RequestMessage{Message: Message{JSONRPC: "2.0", Method: "textDocument/references", ID: "1"}},
		Params: RequestMessageTextDocumentReferencesParams{
			TextDocument: TextDocumentIdentifier{URI: uri},
			Position:     TextDocumentPosition{Line: 4, Character: 17}, // Foo reference
		},
	}
	b, _ := json.Marshal(req)
	anyResp, err := l.handleTextDocumentReferences(b)
	require.NoError(t, err)
	resp := anyResp.(ResponseMessageTextDocumentReferences)
	// Should find at least the type definition and the usage in input
	require.GreaterOrEqual(t, len(resp.Result), 2)
}

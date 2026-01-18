package lsp

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHandleTextDocumentRename(t *testing.T) {
	schema := `version 1

type Foo {}

proc Bar {
  input { foo: Foo }
  output { ok: bool }
}
`
	uri := "file:///rename.urpc"
	l := newTestLSP(t, schema, uri)

	req := RequestMessageTextDocumentRename{
		RequestMessage: RequestMessage{Message: Message{JSONRPC: "2.0", Method: "textDocument/rename", ID: "1"}},
		Params: RequestMessageTextDocumentRenameParams{
			TextDocument: TextDocumentIdentifier{URI: uri},
			Position:     TextDocumentPosition{Line: 5, Character: 15}, // Foo reference
			NewName:      "BarType",
		},
	}
	b, _ := json.Marshal(req)
	anyResp, err := l.handleTextDocumentRename(b)
	require.NoError(t, err)
	resp := anyResp.(ResponseMessageTextDocumentRename)
	edits := resp.Result.Changes[uri]
	require.Greater(t, len(edits), 1)
	for _, e := range edits {
		require.Equal(t, "BarType", e.NewText)
	}
}

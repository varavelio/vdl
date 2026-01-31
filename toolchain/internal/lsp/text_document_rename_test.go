package lsp

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHandleTextDocumentRename(t *testing.T) {
	schema := `type Foo {}

rpc Test {
  proc Bar {
    input { foo: Foo }
    output { ok: bool }
  }
}
`
	uri := "file:///rename.vdl"
	l := newTestLSP(t, schema, uri)

	req := RequestMessageTextDocumentRename{
		RequestMessage: RequestMessage{Message: Message{JSONRPC: "2.0", Method: "textDocument/rename", ID: "1"}},
		Params: RequestMessageTextDocumentRenameParams{
			TextDocument: TextDocumentIdentifier{URI: uri},
			Position:     TextDocumentPosition{Line: 4, Character: 17}, // Foo reference
			NewName:      "BarType",
		},
	}
	b, _ := json.Marshal(req)
	anyResp, err := l.handleTextDocumentRename(b)
	require.NoError(t, err)
	resp := anyResp.(ResponseMessageTextDocumentRename)
	edits := resp.Result.Changes[uri]
	require.Len(t, edits, 2)

	// Verify edits ranges
	for _, e := range edits {
		require.Equal(t, "BarType", e.NewText)
		// Length of "Foo" is 3
		require.Equal(t, 3, e.Range.End.Character-e.Range.Start.Character)
		require.Equal(t, e.Range.Start.Line, e.Range.End.Line)
	}
}

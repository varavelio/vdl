package lsp

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHandleTextDocumentCompletion(t *testing.T) {
	schema := `version 1

type User {}

proc Foo {
  input { user:  }
}`
	uri := "file:///comp.urpc"
	l := newTestLSP(t, schema, uri)

	req := RequestMessageTextDocumentCompletion{
		RequestMessage: RequestMessage{Message: Message{JSONRPC: "2.0", Method: "textDocument/completion", ID: "1"}},
		Params: RequestMessageTextDocumentCompletionParams{
			TextDocument: TextDocumentIdentifier{URI: uri},
			Position:     TextDocumentPosition{Line: 5, Character: 15}, // after colon and space
		},
	}
	b, _ := json.Marshal(req)
	anyResp, err := l.handleTextDocumentCompletion(b)
	require.NoError(t, err)
	resp := anyResp.(ResponseMessageTextDocumentCompletion)
	require.NotEmpty(t, resp.Result)
	hasInt, hasUser := false, false
	for _, item := range resp.Result {
		if item.Label == "int" {
			hasInt = true
		}
		if item.Label == "User" {
			hasUser = true
		}
	}
	require.True(t, hasInt)
	require.True(t, hasUser)
}

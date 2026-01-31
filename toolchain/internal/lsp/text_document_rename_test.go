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

func TestHandleTextDocumentRename_CrossFile(t *testing.T) {
	// a.vdl defines Foo
	schemaA := `type Foo {}`
	// b.vdl uses Foo
	schemaB := `type Bar { f: Foo }`

	uriA := "file:///a.vdl"
	uriB := "file:///b.vdl"
	pathA := UriToPath(uriA)
	pathB := UriToPath(uriB)

	l := newTestLSP(t, schemaA, uriA)
	// Add b.vdl to VFS
	l.fs.WriteFileCache(pathB, []byte(schemaB))

	// Register dependency: B imports A
	l.depGraph.UpdateDependencies(pathB, []string{pathA})

	// Rename Foo in A
	req := RequestMessageTextDocumentRename{
		RequestMessage: RequestMessage{Message: Message{JSONRPC: "2.0", Method: "textDocument/rename", ID: "1"}},
		Params: RequestMessageTextDocumentRenameParams{
			TextDocument: TextDocumentIdentifier{URI: uriA},
			Position:     TextDocumentPosition{Line: 0, Character: 6}, // "Foo"
			NewName:      "NewFoo",
		},
	}
	b, _ := json.Marshal(req)
	anyResp, err := l.handleTextDocumentRename(b)
	require.NoError(t, err)
	resp := anyResp.(ResponseMessageTextDocumentRename)

	changes := resp.Result.Changes
	require.Len(t, changes, 2)

	// Check A
	editsA := changes[uriA]
	require.Len(t, editsA, 1)
	require.Equal(t, "NewFoo", editsA[0].NewText)

	// Check B
	editsB := changes[uriB]
	require.Len(t, editsB, 1)
	require.Equal(t, "NewFoo", editsB[0].NewText)
}

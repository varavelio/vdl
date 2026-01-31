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

func TestHandleTextDocumentRename_TransitiveDependency(t *testing.T) {
	// shared.vdl defines MyType
	schemaShared := `type MyType {}`
	// middle.vdl imports shared.vdl
	schemaMiddle := `
include "shared.vdl"
type Middle { t: MyType }
`
	// main.vdl imports middle.vdl and uses MyType (implicitly available?)
	// Assuming VDL allows using transitive types if they are in scope or if they are exported.
	// Even if not strictly valid VDL without direct import, the LSP should probably try to rename it if it finds the reference.
	// But `collectRenameEdits` uses the AST. If `MyType` is used in `main.vdl`, the parser must parse it.
	// If it's valid syntax, `findReferencesInSchema` will find `MyType`.
	schemaMain := `
include "middle.vdl"
type Main { t: MyType }
`

	uriShared := "file:///shared.vdl"
	uriMiddle := "file:///middle.vdl"
	uriMain := "file:///main.vdl"

	pathShared := UriToPath(uriShared)
	pathMiddle := UriToPath(uriMiddle)
	pathMain := UriToPath(uriMain)

	l := newTestLSP(t, schemaShared, uriShared)
	l.fs.WriteFileCache(pathMiddle, []byte(schemaMiddle))
	l.fs.WriteFileCache(pathMain, []byte(schemaMain))

	// Register dependencies
	// middle imports shared
	l.depGraph.UpdateDependencies(pathMiddle, []string{pathShared})
	// main imports middle
	l.depGraph.UpdateDependencies(pathMain, []string{pathMiddle})

	// Rename MyType in shared.vdl
	req := RequestMessageTextDocumentRename{
		RequestMessage: RequestMessage{Message: Message{JSONRPC: "2.0", Method: "textDocument/rename", ID: "1"}},
		Params: RequestMessageTextDocumentRenameParams{
			TextDocument: TextDocumentIdentifier{URI: uriShared},
			Position:     TextDocumentPosition{Line: 0, Character: 6}, // "MyType"
			NewName:      "YourType",
		},
	}
	b, _ := json.Marshal(req)
	anyResp, err := l.handleTextDocumentRename(b)
	require.NoError(t, err)
	resp := anyResp.(ResponseMessageTextDocumentRename)

	changes := resp.Result.Changes
	require.Len(t, changes, 3)

	// Check Shared
	require.Equal(t, "YourType", changes[uriShared][0].NewText)
	// Check Middle
	require.Equal(t, "YourType", changes[uriMiddle][0].NewText)
	// Check Main
	require.Equal(t, "YourType", changes[uriMain][0].NewText)
}

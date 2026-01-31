package lsp

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHandleTextDocumentCompletion(t *testing.T) {
	schema := `type User {}

rpc Test {
  proc Foo {
    input { user:  }
  }
}`
	uri := "file:///comp.vdl"
	l := newTestLSP(t, schema, uri)

	req := RequestMessageTextDocumentCompletion{
		RequestMessage: RequestMessage{Message: Message{JSONRPC: "2.0", Method: "textDocument/completion", ID: "1"}},
		Params: RequestMessageTextDocumentCompletionParams{
			TextDocument: TextDocumentIdentifier{URI: uri},
			Position:     TextDocumentPosition{Line: 4, Character: 18}, // after colon and space
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

func TestHandleTextDocumentCompletion_CrossFile(t *testing.T) {
	schemaFoo := `type FooType {}
enum FooEnum { A }`
	schemaMain := `
include "foo.vdl"
type Main {
  f: 
}`

	uriFoo := "file:///foo.vdl"
	uriMain := "file:///main.vdl"
	pathFoo := UriToPath(uriFoo)
	pathMain := UriToPath(uriMain)

	l := newTestLSP(t, schemaFoo, uriFoo)
	l.fs.WriteFileCache(pathMain, []byte(schemaMain))
	// Register dependency explicitly as parser isn't running automatically in test setup
	l.depGraph.UpdateDependencies(pathMain, []string{pathFoo})

	// Trigger completion in Main after "f: "
	req := RequestMessageTextDocumentCompletion{
		RequestMessage: RequestMessage{Message: Message{JSONRPC: "2.0", Method: "textDocument/completion", ID: "1"}},
		Params: RequestMessageTextDocumentCompletionParams{
			TextDocument: TextDocumentIdentifier{URI: uriMain},
			Position:     TextDocumentPosition{Line: 3, Character: 5},
		},
	}
	b, _ := json.Marshal(req)
	anyResp, err := l.handleTextDocumentCompletion(b)
	require.NoError(t, err)
	resp := anyResp.(ResponseMessageTextDocumentCompletion)

	hasFooType, hasFooEnum := false, false
	for _, item := range resp.Result {
		if item.Label == "FooType" {
			hasFooType = true
		}
		if item.Label == "FooEnum" {
			hasFooEnum = true
		}
	}
	require.True(t, hasFooType, "Should suggest FooType from included file")
	require.True(t, hasFooEnum, "Should suggest FooEnum from included file")
}

func TestHandleTextDocumentCompletion_Spread(t *testing.T) {
	schema := `type Base {}
type Derived {
  ...
}`
	uri := "file:///spread.vdl"
	l := newTestLSP(t, schema, uri)

	req := RequestMessageTextDocumentCompletion{
		RequestMessage: RequestMessage{Message: Message{JSONRPC: "2.0", Method: "textDocument/completion", ID: "1"}},
		Params: RequestMessageTextDocumentCompletionParams{
			TextDocument: TextDocumentIdentifier{URI: uri},
			Position:     TextDocumentPosition{Line: 2, Character: 5}, // after "..."
		},
	}
	b, _ := json.Marshal(req)
	anyResp, err := l.handleTextDocumentCompletion(b)
	require.NoError(t, err)
	resp := anyResp.(ResponseMessageTextDocumentCompletion)

	hasBase := false
	hasInt := false
	for _, item := range resp.Result {
		if item.Label == "Base" {
			hasBase = true
		}
		if item.Label == "int" {
			hasInt = true
		}
	}
	require.True(t, hasBase, "Should suggest Base for spread")
	require.False(t, hasInt, "Should NOT suggest int for spread")
}

func TestHandleTextDocumentCompletion_Map(t *testing.T) {
	schema := `type Foo {}
type Bar {
  m: map<
}`
	uri := "file:///map.vdl"
	l := newTestLSP(t, schema, uri)

	// Test case 1: "map<"
	req := RequestMessageTextDocumentCompletion{
		RequestMessage: RequestMessage{Message: Message{JSONRPC: "2.0", Method: "textDocument/completion", ID: "1"}},
		Params: RequestMessageTextDocumentCompletionParams{
			TextDocument: TextDocumentIdentifier{URI: uri},
			Position:     TextDocumentPosition{Line: 2, Character: 9}, // after "map<"
		},
	}
	b, _ := json.Marshal(req)
	anyResp, err := l.handleTextDocumentCompletion(b)
	require.NoError(t, err)
	resp := anyResp.(ResponseMessageTextDocumentCompletion)

	hasFoo := false
	hasInt := false
	for _, item := range resp.Result {
		if item.Label == "Foo" {
			hasFoo = true
		}
		if item.Label == "int" {
			hasInt = true
		}
	}
	require.True(t, hasFoo, "Should suggest Foo inside map")
	require.True(t, hasInt, "Should suggest int inside map")

	// Test case 2: "map < F" (with spaces and prefix)
	l.fs.WriteFileCache(UriToPath(uri), []byte(`type Foo {}
type Bar {
  m: map < F
}`))
	req = RequestMessageTextDocumentCompletion{
		RequestMessage: RequestMessage{Message: Message{JSONRPC: "2.0", Method: "textDocument/completion", ID: "2"}},
		Params: RequestMessageTextDocumentCompletionParams{
			TextDocument: TextDocumentIdentifier{URI: uri},
			Position:     TextDocumentPosition{Line: 2, Character: 12}, // after "map < F"
		},
	}
	b, _ = json.Marshal(req)
	anyResp, err = l.handleTextDocumentCompletion(b)
	require.NoError(t, err)
	resp = anyResp.(ResponseMessageTextDocumentCompletion)

	hasFoo = false
	hasInt = false
	for _, item := range resp.Result {
		if item.Label == "Foo" {
			hasFoo = true
		}
		if item.Label == "int" {
			hasInt = true
		}
	}
	require.True(t, hasFoo, "Should suggest Foo inside map with prefix 'F'")
	require.False(t, hasInt, "Should NOT suggest int inside map with prefix 'F'")
}

func TestHandleTextDocumentCompletion_ForwardReference(t *testing.T) {
	// The cursor is inside TypeA, but TypeB is defined *after* it.
	// Because TypeA is incomplete/invalid syntax at the moment of typing,
	// the parser might stop before reaching TypeB.
	// We want to ensure TypeB is still suggested.
	schema := `type TypeA {
  f: 
}

type TypeB {}
enum EnumC { X }
`
	uri := "file:///forward.vdl"
	l := newTestLSP(t, schema, uri)

	req := RequestMessageTextDocumentCompletion{
		RequestMessage: RequestMessage{Message: Message{JSONRPC: "2.0", Method: "textDocument/completion", ID: "1"}},
		Params: RequestMessageTextDocumentCompletionParams{
			TextDocument: TextDocumentIdentifier{URI: uri},
			Position:     TextDocumentPosition{Line: 1, Character: 5}, // after "f: "
		},
	}
	b, _ := json.Marshal(req)
	anyResp, err := l.handleTextDocumentCompletion(b)
	require.NoError(t, err)
	resp := anyResp.(ResponseMessageTextDocumentCompletion)

	hasTypeB := false
	hasEnumC := false
	for _, item := range resp.Result {
		if item.Label == "TypeB" {
			hasTypeB = true
		}
		if item.Label == "EnumC" {
			hasEnumC = true
		}
	}
	require.True(t, hasTypeB, "Should suggest TypeB defined below")
	require.True(t, hasEnumC, "Should suggest EnumC defined below")
}

package lsp

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHandleTextDocumentDocumentLink(t *testing.T) {
	schema := `""" ./doc.md """`
	uri := "file:///links.vdl"
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

func TestHandleTextDocumentDocumentLinkRanges(t *testing.T) {
	t.Run("docstring path range only", func(t *testing.T) {
		// The docstring is `""" ./doc.md """`, which spans:
		// Positions: 0-15 (includes quotes and spaces)
		// Characters 4-11 are "./doc.md" (9 characters)
		// Range should be [4, 12) or in LSP terms: start=4, end=13 (exclusive)
		schema := `""" ./doc.md """`
		uri := "file:///links.vdl"
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

		// Verify the range covers only the path, not the quotes or spaces
		link := resp.Result[0]
		require.Equal(t, 0, link.Range.Start.Line)
		require.Equal(t, 4, link.Range.Start.Character) // After """ and space
		require.Equal(t, 0, link.Range.End.Line)
		require.Equal(t, 12, link.Range.End.Character) // After "./doc.md" (exclusive end)
	})

	t.Run("include path range only", func(t *testing.T) {
		// The include is `include "foo.vdl"`, which spans positions 0-18
		// We want the range to only cover `foo.vdl` at positions 9-16
		schema := `include "foo.vdl"`
		uri := "file:///links.vdl"
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

		// Verify the range covers only the path, not the keyword or quotes
		link := resp.Result[0]
		require.Equal(t, 0, link.Range.Start.Line)
		require.Equal(t, 9, link.Range.Start.Character) // After 'include "'
		require.Equal(t, 0, link.Range.End.Line)
		require.Equal(t, 16, link.Range.End.Character) // After "foo.vdl"
	})

	t.Run("multiple links with correct ranges", func(t *testing.T) {
		schema := `""" ./doc.md """
include "foo.vdl"
type User {
	name: string
}`
		uri := "file:///links.vdl"
		l := newTestLSP(t, schema, uri)

		req := RequestMessageTextDocumentDocumentLink{
			RequestMessage: RequestMessage{Message: Message{JSONRPC: "2.0", Method: "textDocument/documentLink", ID: "1"}},
			Params:         RequestMessageTextDocumentDocumentLinkParams{TextDocument: TextDocumentIdentifier{URI: uri}},
		}
		b, _ := json.Marshal(req)
		anyResp, err := l.handleTextDocumentDocumentLink(b)
		require.NoError(t, err)
		resp := anyResp.(ResponseMessageTextDocumentDocumentLink)
		require.Len(t, resp.Result, 2)

		// First link: docstring on line 0
		docLink := resp.Result[0]
		require.Equal(t, 0, docLink.Range.Start.Line)
		require.Equal(t, 4, docLink.Range.Start.Character)
		require.Equal(t, 0, docLink.Range.End.Line)
		require.Equal(t, 12, docLink.Range.End.Character)

		// Second link: include on line 1
		includeLink := resp.Result[1]
		require.Equal(t, 1, includeLink.Range.Start.Line)
		require.Equal(t, 9, includeLink.Range.Start.Character)
		require.Equal(t, 1, includeLink.Range.End.Line)
		require.Equal(t, 16, includeLink.Range.End.Character)
	})
}

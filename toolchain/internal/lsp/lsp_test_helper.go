package lsp

import (
	"bytes"

	"github.com/stretchr/testify/require"
)

// newTestLSP returns an initialized LSP with the given schema opened in memory.
func newTestLSP(t require.TestingT, schema, uri string) *LSP {
	reader := &bytes.Buffer{}
	writer := &bytes.Buffer{}
	l := New(reader, writer)

	// Convert URI to path and store in vfs
	filePath := uriToPath(uri)
	l.fs.WriteFileCache(filePath, []byte(schema))

	return l
}

package lsp

type NotificationMessageTextDocumentDidClose struct {
	NotificationMessage
	Params NotificationMessageTextDocumentDidCloseParams `json:"params"`
}

type NotificationMessageTextDocumentDidCloseParams struct {
	// The document that did close.
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

func (l *LSP) handleTextDocumentDidClose(rawMessage []byte) (any, error) {
	var notification NotificationMessageTextDocumentDidClose
	if err := decode(rawMessage, &notification); err != nil {
		return nil, err
	}

	filePath := UriToPath(notification.Params.TextDocument.URI)

	// Remove the file from the virtual file system cache
	l.fs.RemoveFileCache(filePath)

	// Unregister this document and clean up its dependencies
	l.unregisterOpenDoc(filePath)

	l.logger.Info("text document did close", "uri", notification.Params.TextDocument.URI)

	// Clear diagnostics when a document is closed
	l.clearDiagnostics(notification.Params.TextDocument.URI)

	return nil, nil
}

package lsp

type NotificationMessageTextDocumentDidOpen struct {
	NotificationMessage
	Params NotificationMessageTextDocumentDidOpenParams `json:"params"`
}

type NotificationMessageTextDocumentDidOpenParams struct {
	// The document that was opened.
	TextDocument TextDocumentItem `json:"textDocument"`
}

func (l *LSP) handleTextDocumentDidOpen(rawMessage []byte) (any, error) {
	var notification NotificationMessageTextDocumentDidOpen
	if err := decode(rawMessage, &notification); err != nil {
		return nil, err
	}

	filePath := UriToPath(notification.Params.TextDocument.URI)
	content := notification.Params.TextDocument.Text
	uri := notification.Params.TextDocument.URI

	// Store the content in the virtual file system
	l.fs.WriteFileCache(filePath, []byte(content))

	// Register this document as open
	l.registerOpenDoc(filePath, uri)

	l.logger.Info("text document did open", "uri", uri)

	// Trigger immediate analysis for newly opened documents
	l.analyzeAndPublishDiagnosticsImmediate(filePath, uri)

	return nil, nil
}

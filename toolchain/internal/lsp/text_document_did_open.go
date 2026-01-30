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

	// Store the content in the virtual file system
	l.fs.WriteFileCache(filePath, []byte(content))

	l.logger.Info("text document did open", "uri", notification.Params.TextDocument.URI)

	// Trigger immediate analysis for newly opened documents
	l.analyzeAndPublishDiagnostics(filePath, notification.Params.TextDocument.URI)

	return nil, nil
}

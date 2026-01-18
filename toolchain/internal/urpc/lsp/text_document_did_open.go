package lsp

import "fmt"

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

	filePath := notification.Params.TextDocument.URI
	content := notification.Params.TextDocument.Text
	if err := l.docstore.OpenInMem(filePath, content); err != nil {
		return nil, fmt.Errorf("failed to open in memory file: %w", err)
	}

	l.logger.Info("text document did open", "uri", notification.Params.TextDocument.URI)

	// Trigger immediate analysis for newly opened documents
	if l.analyzer != nil {
		l.analyzeAndPublishDiagnostics(filePath)
	}

	return nil, nil
}

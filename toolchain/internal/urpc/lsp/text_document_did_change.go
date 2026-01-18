package lsp

import "fmt"

type NotificationMessageTextDocumentDidChange struct {
	NotificationMessage
	Params NotificationMessageTextDocumentDidChangeParams `json:"params"`
}

type NotificationMessageTextDocumentDidChangeParams struct {
	// The document that did change.
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	// The content of the document.
	ContentChanges []TextDocumentContentChangeEvent `json:"contentChanges"`
}

func (l *LSP) handleTextDocumentDidChange(rawMessage []byte) (any, error) {
	var notification NotificationMessageTextDocumentDidChange
	if err := decode(rawMessage, &notification); err != nil {
		return nil, err
	}

	if len(notification.Params.ContentChanges) == 0 {
		return nil, fmt.Errorf("no content changes provided")
	}

	filePath := notification.Params.TextDocument.URI
	content := notification.Params.ContentChanges[0].Text
	if err := l.docstore.ChangeInMem(filePath, content); err != nil {
		return nil, fmt.Errorf("failed to change in memory file: %w", err)
	}

	l.logger.Info("text document did change", "uri", notification.Params.TextDocument.URI)

	// Schedule analysis with debouncing
	if l.analyzer != nil {
		l.analyzeAndPublishDiagnosticsDebounced(filePath)
	}

	return nil, nil
}

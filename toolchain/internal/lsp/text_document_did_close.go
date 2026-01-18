package lsp

import "fmt"

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

	filePath := notification.Params.TextDocument.URI
	if err := l.docstore.CloseInMem(filePath); err != nil {
		return nil, fmt.Errorf("failed to close in memory file: %w", err)
	}

	l.logger.Info("text document did close", "uri", notification.Params.TextDocument.URI)

	// Clear diagnostics when a document is closed
	if l.analyzer != nil {
		l.clearDiagnostics(filePath)
	}

	return nil, nil
}

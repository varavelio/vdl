package lsp

import "fmt"

func (l *LSP) handleInitialized(rawMessage []byte) (any, error) {
	var message NotificationMessage
	if err := decode(rawMessage, &message); err != nil {
		return nil, fmt.Errorf("failed to decode initialized message: %w", err)
	}

	l.logger.Info("initialized message received", "method", message.Method)
	return nil, nil
}

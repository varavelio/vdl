package lsp

import "fmt"

type ResponseMessageShutdown struct {
	ResponseMessage
	Result any `json:"result"`
}

func (l *LSP) handleShutdown(rawMessage []byte) (any, error) {
	var message RequestMessage
	if err := decode(rawMessage, &message); err != nil {
		return nil, fmt.Errorf("failed to decode shutdown message: %w", err)
	}

	l.logger.Info("shutdown message received", "id", message.ID, "method", message.Method)

	response := ResponseMessageShutdown{
		ResponseMessage: ResponseMessage{
			Message: DefaultMessage,
			ID:      message.ID,
		},
		Result: nil,
	}

	return response, nil
}

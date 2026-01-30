package lsp

import (
	"fmt"

	"github.com/varavelio/vdl/toolchain/internal/version"
)

type RequestMessageInitialize struct {
	RequestMessage
	Params RequestMessageInitializeParams `json:"params"`
}

type RequestMessageInitializeParams struct {
	ClientInfo struct {
		Name    string `json:"name"`
		Version string `json:"version,omitzero,omitempty"`
	} `json:"clientInfo,omitzero"`
}

type ResponseMessageInitialize struct {
	ResponseMessage
	Result ResponseMessageInitializeResult `json:"result"`
}

type ResponseMessageInitializeResult struct {
	ServerInfo   ResponseMessageInitializeResultServerInfo   `json:"serverInfo"`
	Capabilities ResponseMessageInitializeResultCapabilities `json:"capabilities"`
}

type ResponseMessageInitializeResultServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type ResponseMessageInitializeResultCapabilities struct {
	DocumentFormattingProvider bool `json:"documentFormattingProvider"`
	TextDocumentSync           int  `json:"textDocumentSync"`
	// Advertise diagnostic capabilities
	DiagnosticProvider bool `json:"diagnosticProvider,omitempty"`
	// Advertise definition capabilities
	DefinitionProvider bool `json:"definitionProvider,omitempty"`
	// Advertise hover capabilities
	HoverProvider bool `json:"hoverProvider,omitempty"`
	// Advertise rename capabilities
	RenameProvider bool `json:"renameProvider,omitempty"`
	// Advertise document link capabilities
	DocumentLinkProvider bool `json:"documentLinkProvider,omitempty"`
	// Advertise references capabilities
	ReferencesProvider bool `json:"referencesProvider,omitempty"`
	// Advertise document symbol capabilities
	DocumentSymbolProvider bool `json:"documentSymbolProvider,omitempty"`
	// Advertise completion capabilities
	CompletionProvider bool `json:"completionProvider,omitempty"`
}

func (l *LSP) handleInitialize(rawMessage []byte) (any, error) {
	var message RequestMessageInitialize
	if err := decode(rawMessage, &message); err != nil {
		return nil, fmt.Errorf("failed to decode initialize message: %w", err)
	}

	l.logger.Info(
		"initialize message received",
		"id", message.ID,
		"method", message.Method,
		"clientName", message.Params.ClientInfo.Name,
		"clientVersion", message.Params.ClientInfo.Version,
	)

	response := ResponseMessageInitialize{
		ResponseMessage: ResponseMessage{
			Message: DefaultMessage,
			ID:      message.ID,
		},
		Result: ResponseMessageInitializeResult{
			ServerInfo: ResponseMessageInitializeResultServerInfo{
				Name:    "VDL Language Server",
				Version: "v" + version.Version,
			},
			Capabilities: ResponseMessageInitializeResultCapabilities{
				// Documents are synced by always sending the full content of the document.
				TextDocumentSync: 1,
				// Document formatting is supported.
				DocumentFormattingProvider: true,
				// Diagnostics are supported
				DiagnosticProvider: true,
				// Definition (go to definition) is supported
				DefinitionProvider: true,
				// Hover is supported
				HoverProvider: true,
				// Rename is supported
				RenameProvider: true,
				// Document link is supported
				DocumentLinkProvider: true,
				// References are supported
				ReferencesProvider: true,
				// Document symbol capabilities are supported
				DocumentSymbolProvider: true,
				// Completion capabilities are supported
				CompletionProvider: true,
			},
		},
	}

	return response, nil
}

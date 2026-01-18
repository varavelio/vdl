package lsp

import (
	"fmt"
	"strings"

	"github.com/varavelio/vdl/toolchain/internal/formatter"
)

type RequestMessageTextDocumentFormatting struct {
	RequestMessage
	Params RequestMessageTextDocumentFormattingParams `json:"params"`
}

type RequestMessageTextDocumentFormattingParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	// Options are not used because the formatting rules are not configurable
}

type ResponseMessageTextDocumentFormatting struct {
	ResponseMessage
	Result *[]TextDocumentTextEdit `json:"result"`
}

func (l *LSP) handleTextDocumentFormatting(rawMessage []byte) (any, error) {
	var request RequestMessageTextDocumentFormatting
	if err := decode(rawMessage, &request); err != nil {
		return nil, fmt.Errorf("failed to decode text document formatting request: %w", err)
	}

	filePath := uriToPath(request.Params.TextDocument.URI)
	content, err := l.fs.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file from vfs: %w", err)
	}

	response := ResponseMessageTextDocumentFormatting{
		ResponseMessage: ResponseMessage{
			Message: DefaultMessage,
			ID:      request.ID,
		},
		Result: nil,
	}

	formattedText, err := formatter.Format(filePath, string(content))
	if err != nil {
		// If formatting fails, return no edits.
		return response, nil
	}

	lines := strings.Split(string(content), "\n")
	lastLine := max(len(lines)-1, 0)
	lastLineChar := len(lines[lastLine])

	response.Result = &[]TextDocumentTextEdit{
		{
			Range: TextDocumentRange{
				Start: TextDocumentPosition{Line: 0, Character: 0},
				End:   TextDocumentPosition{Line: lastLine, Character: lastLineChar},
			},
			NewText: formattedText,
		},
	}

	return response, nil
}

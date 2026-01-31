package lsp

import (
	"context"
	"fmt"
	"strings"

	"github.com/varavelio/vdl/toolchain/internal/core/analysis"
)

// RequestMessageTextDocumentHover represents a request for hover information.
type RequestMessageTextDocumentHover struct {
	RequestMessage
	Params RequestMessageTextDocumentHoverParams `json:"params"`
}

// RequestMessageTextDocumentHoverParams represents the parameters for a hover request.
type RequestMessageTextDocumentHoverParams struct {
	// The text document.
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	// The position inside the text document.
	Position TextDocumentPosition `json:"position"`
}

// ResponseMessageTextDocumentHover represents a response to a hover request.
type ResponseMessageTextDocumentHover struct {
	ResponseMessage
	// The result of the request.
	Result *HoverResult `json:"result"`
}

// HoverResult represents the result of a hover request.
type HoverResult struct {
	// The hover's content.
	Contents MarkupContent `json:"contents"`
	// An optional range that is used to visualize the hover.
	Range *TextDocumentRange `json:"range,omitempty"`
}

// MarkupContent represents a hover content with a specific kind of markup.
type MarkupContent struct {
	// The type of the markup content. Currently only "markdown" is supported.
	Kind string `json:"kind"`
	// The content itself.
	Value string `json:"value"`
}

// handleTextDocumentHover handles a textDocument/hover request.
func (l *LSP) handleTextDocumentHover(rawMessage []byte) (any, error) {
	var request RequestMessageTextDocumentHover
	if err := decode(rawMessage, &request); err != nil {
		return nil, fmt.Errorf("failed to decode hover request: %w", err)
	}

	filePath := UriToPath(request.Params.TextDocument.URI)
	position := request.Params.Position

	// Get the document content
	content, err := l.fs.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file from vfs: %w", err)
	}

	// Run the analysis to get the program
	program, _ := l.analyze(context.Background(), filePath)

	// Find the identifier at the cursor position
	identifier := findIdentifierAtPosition(string(content), position)
	if identifier == "" {
		return ResponseMessageTextDocumentHover{
			ResponseMessage: ResponseMessage{
				Message: DefaultMessage,
				ID:      request.ID,
			},
			Result: nil,
		}, nil
	}

	// Find the hover information
	hoverResult := findHoverInfo(l.fs, identifier, program)

	response := ResponseMessageTextDocumentHover{
		ResponseMessage: ResponseMessage{
			Message: DefaultMessage,
			ID:      request.ID,
		},
		Result: hoverResult,
	}

	return response, nil
}

// findHoverInfo finds hover information for a symbol in the program.
func findHoverInfo(fs interface{ ReadFile(string) ([]byte, error) }, identifier string, program *analysis.Program) *HoverResult {
	// Check if the identifier is a type
	if t, ok := program.Types[identifier]; ok {
		sourceCode, err := extractSourceCode(fs, t.File, t.Pos.Line, t.EndPos.Line)
		if err != nil {
			return nil
		}
		return &HoverResult{
			Contents: MarkupContent{
				Kind:  "markdown",
				Value: fmt.Sprintf("```vdl\n%s\n```", sourceCode),
			},
		}
	}

	// Check if the identifier is an enum
	if e, ok := program.Enums[identifier]; ok {
		sourceCode, err := extractSourceCode(fs, e.File, e.Pos.Line, e.EndPos.Line)
		if err != nil {
			return nil
		}
		return &HoverResult{
			Contents: MarkupContent{
				Kind:  "markdown",
				Value: fmt.Sprintf("```vdl\n%s\n```", sourceCode),
			},
		}
	}

	// Check if the identifier is a constant
	if c, ok := program.Consts[identifier]; ok {
		sourceCode, err := extractSourceCode(fs, c.File, c.Pos.Line, c.EndPos.Line)
		if err != nil {
			return nil
		}
		return &HoverResult{
			Contents: MarkupContent{
				Kind:  "markdown",
				Value: fmt.Sprintf("```vdl\n%s\n```", sourceCode),
			},
		}
	}

	// Check if the identifier is a pattern
	if p, ok := program.Patterns[identifier]; ok {
		sourceCode, err := extractSourceCode(fs, p.File, p.Pos.Line, p.EndPos.Line)
		if err != nil {
			return nil
		}
		return &HoverResult{
			Contents: MarkupContent{
				Kind:  "markdown",
				Value: fmt.Sprintf("```vdl\n%s\n```", sourceCode),
			},
		}
	}

	// Check if the identifier is an RPC
	if r, ok := program.RPCs[identifier]; ok {
		sourceCode, err := extractSourceCode(fs, r.File, r.Pos.Line, r.EndPos.Line)
		if err != nil {
			return nil
		}
		return &HoverResult{
			Contents: MarkupContent{
				Kind:  "markdown",
				Value: fmt.Sprintf("```vdl\n%s\n```", sourceCode),
			},
		}
	}

	// Check procs and streams within RPCs
	for _, rpc := range program.RPCs {
		if proc, ok := rpc.Procs[identifier]; ok {
			sourceCode, err := extractSourceCode(fs, proc.File, proc.Pos.Line, proc.EndPos.Line)
			if err != nil {
				return nil
			}
			return &HoverResult{
				Contents: MarkupContent{
					Kind:  "markdown",
					Value: fmt.Sprintf("```vdl\n%s\n```", sourceCode),
				},
			}
		}
		if stream, ok := rpc.Streams[identifier]; ok {
			sourceCode, err := extractSourceCode(fs, stream.File, stream.Pos.Line, stream.EndPos.Line)
			if err != nil {
				return nil
			}
			return &HoverResult{
				Contents: MarkupContent{
					Kind:  "markdown",
					Value: fmt.Sprintf("```vdl\n%s\n```", sourceCode),
				},
			}
		}
	}

	return nil
}

// extractSourceCode extracts a range of lines from a file.
func extractSourceCode(fs interface{ ReadFile(string) ([]byte, error) }, filePath string, startLine, endLine int) (string, error) {
	content, err := fs.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	return extractCodeFromContent(string(content), startLine, endLine)
}

// extractCodeFromContent extracts a range of lines from the content.
func extractCodeFromContent(content string, startLine, endLine int) (string, error) {
	if startLine <= 0 {
		return "", fmt.Errorf("start line out of range: %d", startLine)
	}

	if endLine < startLine {
		return "", fmt.Errorf("end line before start line: %d < %d", endLine, startLine)
	}

	// Find the start offset
	currentLine := 1
	startOffset := 0
	for currentLine < startLine && startOffset < len(content) {
		idx := strings.IndexByte(content[startOffset:], '\n')
		if idx == -1 {
			break
		}
		startOffset += idx + 1
		currentLine++
	}

	if currentLine < startLine {
		return "", fmt.Errorf("start line out of range: %d", startLine)
	}

	// Find the end offset
	endOffset := startOffset
	for currentLine <= endLine && endOffset < len(content) {
		idx := strings.IndexByte(content[endOffset:], '\n')
		if idx == -1 {
			endOffset = len(content)
			break
		}
		endOffset += idx + 1
		currentLine++
	}

	// If startOffset >= len(content), it's empty or out of bounds.
	if startOffset >= len(content) {
		// If startLine was exactly len(lines)+1 ? No.
		return "", fmt.Errorf("start line out of range")
	}

	// If endOffset > len(content), clamp it.
	if endOffset > len(content) {
		endOffset = len(content)
	}

	chunk := content[startOffset:endOffset]
	// Remove trailing newline of the chunk to avoid an empty last line in split
	if strings.HasSuffix(chunk, "\n") {
		chunk = chunk[:len(chunk)-1]
	} else if strings.HasSuffix(chunk, "\r\n") {
		chunk = chunk[:len(chunk)-2]
	}

	extractedLines := strings.Split(chunk, "\n")

	// Find the minimum indentation
	minIndent := -1
	for _, line := range extractedLines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		indent := 0
		for i := 0; i < len(line); i++ {
			if line[i] == ' ' || line[i] == '\t' {
				indent++
			} else {
				break
			}
		}

		if minIndent == -1 || indent < minIndent {
			minIndent = indent
		}
	}

	// Remove the minimum indentation from each line
	if minIndent > 0 {
		for i, line := range extractedLines {
			if len(line) >= minIndent {
				extractedLines[i] = line[minIndent:]
			}
		}
	}

	// Join the lines
	return strings.Join(extractedLines, "\n"), nil
}

package lsp

import (
	"fmt"
	"strings"

	"github.com/uforg/uforpc/urpc/internal/urpc/ast"
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

	filePath := request.Params.TextDocument.URI
	position := request.Params.Position

	// Get the document content
	content, _, found, err := l.docstore.GetInMemory("", filePath)
	if !found {
		return nil, fmt.Errorf("text document not found in memory: %s", filePath)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get in memory text document: %w", err)
	}

	// Run the analyzer to get the combined schema
	astSchema, _, err := l.analyzer.Analyze(filePath)
	if err != nil {
		l.logger.Error("failed to analyze document", "uri", filePath, "error", err)
	}

	// Convert LSP position (0-based) to AST position (1-based)
	astPosition := ast.Position{
		Filename: filePath,
		Line:     position.Line + 1,
		Column:   position.Character + 1,
	}

	// Find the hover information
	hoverResult := l.findHoverInfo(content, astPosition, astSchema)

	response := ResponseMessageTextDocumentHover{
		ResponseMessage: ResponseMessage{
			Message: DefaultMessage,
			ID:      request.ID,
		},
		Result: hoverResult,
	}

	return response, nil
}

// findHoverInfo finds hover information for a symbol at the given position.
func (l *LSP) findHoverInfo(content string, position ast.Position, astSchema *ast.Schema) *HoverResult {
	// Find the tokenLiteral at the position
	tokenLiteral, err := findTokenAtPosition(content, position)
	if err != nil {
		return nil
	}

	// Check if the token is a reference to a type
	if hoverInfo := l.findTypeHoverInfo(tokenLiteral, astSchema); hoverInfo != nil {
		return hoverInfo
	}

	return nil
}

// findTypeHoverInfo finds hover information for a type.
func (l *LSP) findTypeHoverInfo(tokenLiteral string, astSchema *ast.Schema) *HoverResult {
	// Check if the token is a type name
	typeDecl, exists := astSchema.GetTypesMap()[tokenLiteral]
	if !exists {
		return nil
	}

	// Get the source code of the type definition
	sourceCode, err := l.getTypeSourceCode(typeDecl)
	if err != nil {
		return nil
	}

	// Create a hover result with the source code
	return &HoverResult{
		Contents: MarkupContent{
			Kind:  "markdown",
			Value: fmt.Sprintf("```urpc\n%s\n```", sourceCode),
		},
	}
}

// getTypeSourceCode extracts the source code of a type definition.
func (l *LSP) getTypeSourceCode(typeDecl *ast.TypeDecl) (string, error) {
	content, _, err := l.docstore.GetFileAndHash("", typeDecl.Pos.Filename)
	if err != nil {
		return "", fmt.Errorf("failed to get file content: %w", err)
	}

	// Extract the type definition from the content
	return extractCodeFromContent(content, typeDecl.Pos.Line, typeDecl.EndPos.Line)
}

// extractCodeFromContent extracts a range of lines from the content.
func extractCodeFromContent(content string, startLine, endLine int) (string, error) {
	lines := strings.Split(content, "\n")

	if startLine <= 0 || startLine > len(lines) {
		return "", fmt.Errorf("start line out of range: %d", startLine)
	}

	if endLine <= 0 || endLine > len(lines) {
		return "", fmt.Errorf("end line out of range: %d", endLine)
	}

	// Extract the lines
	extractedLines := lines[startLine-1 : endLine]

	// Find the minimum indentation
	minIndent := -1
	for _, line := range extractedLines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		indent := len(line) - len(strings.TrimLeft(line, " \t"))
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

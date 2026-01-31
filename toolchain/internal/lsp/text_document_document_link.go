package lsp

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/varavelio/vdl/toolchain/internal/core/ast"
	"github.com/varavelio/vdl/toolchain/internal/core/parser"
)

// RequestMessageTextDocumentDocumentLink represents a request for document links inside a text document.
type RequestMessageTextDocumentDocumentLink struct {
	RequestMessage
	Params RequestMessageTextDocumentDocumentLinkParams `json:"params"`
}

// RequestMessageTextDocumentDocumentLinkParams are the params for the documentLink request.
type RequestMessageTextDocumentDocumentLinkParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

// ResponseMessageTextDocumentDocumentLink represents the response containing document links.
type ResponseMessageTextDocumentDocumentLink struct {
	ResponseMessage
	Result []DocumentLink `json:"result"`
}

// DocumentLink represents a link inside the document.
type DocumentLink struct {
	Range   TextDocumentRange `json:"range"`
	Target  string            `json:"target"`
	Tooltip string            `json:"tooltip,omitempty"`
}

// handleTextDocumentDocumentLink handles a textDocument/documentLink request.
func (l *LSP) handleTextDocumentDocumentLink(rawMessage []byte) (any, error) {
	var request RequestMessageTextDocumentDocumentLink
	if err := decode(rawMessage, &request); err != nil {
		return nil, fmt.Errorf("failed to decode documentLink request: %w", err)
	}

	filePath := UriToPath(request.Params.TextDocument.URI)

	// Fetch document content
	content, err := l.fs.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file from vfs: %w", err)
	}

	links := collectDocumentLinks(string(content), filePath)

	response := ResponseMessageTextDocumentDocumentLink{
		ResponseMessage: ResponseMessage{
			Message: DefaultMessage,
			ID:      request.ID,
		},
		Result: links,
	}

	return response, nil
}

// collectDocumentLinks scans the document content for external docstring references and include paths.
func collectDocumentLinks(content string, docPath string) []DocumentLink {
	var links []DocumentLink

	// Parse the content to get the AST
	schema, err := parser.ParserInstance.ParseString(docPath, content)
	if err != nil || schema == nil {
		return nil
	}

	baseDir := filepath.Dir(docPath)

	// Find all docstrings that reference external files
	for _, child := range schema.Children {
		if child.Docstring != nil {
			if path, isExternal := child.Docstring.GetExternal(); isExternal {
				normPath := filepath.Join(baseDir, path)
				normPath = filepath.Clean(normPath)

				links = append(links, DocumentLink{
					Range:   calculateDocstringPathRange(child.Docstring, path),
					Target:  PathToUri(normPath),
					Tooltip: "Open markdown file",
				})
			}
		}

		// Check docstrings in types, enums, etc.
		if child.Type != nil && child.Type.Docstring != nil {
			addExternalDocstringLink(&links, child.Type.Docstring, baseDir)
		}
		if child.Enum != nil && child.Enum.Docstring != nil {
			addExternalDocstringLink(&links, child.Enum.Docstring, baseDir)
		}
		if child.Const != nil && child.Const.Docstring != nil {
			addExternalDocstringLink(&links, child.Const.Docstring, baseDir)
		}
		if child.Pattern != nil && child.Pattern.Docstring != nil {
			addExternalDocstringLink(&links, child.Pattern.Docstring, baseDir)
		}
		if child.RPC != nil {
			if child.RPC.Docstring != nil {
				addExternalDocstringLink(&links, child.RPC.Docstring, baseDir)
			}
			// Check procs and streams within RPC
			for _, rpcChild := range child.RPC.Children {
				if rpcChild.Docstring != nil {
					addExternalDocstringLink(&links, rpcChild.Docstring, baseDir)
				}
				if rpcChild.Proc != nil && rpcChild.Proc.Docstring != nil {
					addExternalDocstringLink(&links, rpcChild.Proc.Docstring, baseDir)
				}
				if rpcChild.Stream != nil && rpcChild.Stream.Docstring != nil {
					addExternalDocstringLink(&links, rpcChild.Stream.Docstring, baseDir)
				}
			}
		}

		// Include statements also become links
		if child.Include != nil {
			normPath := filepath.Join(baseDir, string(child.Include.Path))
			normPath = filepath.Clean(normPath)

			links = append(links, DocumentLink{
				Range:   calculateIncludePathRange(child.Include),
				Target:  PathToUri(normPath),
				Tooltip: "Open included file",
			})
		}
	}

	return links
}

// addExternalDocstringLink adds a document link for an external docstring if applicable.
func addExternalDocstringLink(links *[]DocumentLink, docstring *ast.Docstring, baseDir string) {
	path, isExternal := docstring.GetExternal()
	if !isExternal {
		return
	}

	normPath := filepath.Join(baseDir, path)
	normPath = filepath.Clean(normPath)

	*links = append(*links, DocumentLink{
		Range:   calculateDocstringPathRange(docstring, path),
		Target:  PathToUri(normPath),
		Tooltip: "Open markdown file",
	})
}

// calculateDocstringPathRange calculates the exact range for the path inside a docstring.
// For example, in `""" ./doc.md """`, this returns the range for just `./doc.md`.
func calculateDocstringPathRange(docstring *ast.Docstring, path string) TextDocumentRange {
	// The docstring starts at Pos, which includes the opening """
	startPos := docstring.Pos

	// The path is inside the """ ... """ markers
	// We need to find where the path starts within the docstring
	// The docstring.Value already has the """ stripped, so we need to calculate from the raw position

	// Starting from the opening """, we add 3 characters for the opening quotes
	pathStartLine := startPos.Line
	pathStartColumn := startPos.Column + 3

	// Find leading whitespace by trimming the docstring value
	trimmedValue := docstring.Value.String()
	originalLength := len(trimmedValue)
	trimmedValue = strings.TrimLeft(trimmedValue, " \t")
	leadingWhitespace := originalLength - len(trimmedValue)

	pathStartColumn += leadingWhitespace

	// The path length determines the end position
	pathEndColumn := pathStartColumn + len(path)

	return TextDocumentRange{
		Start: TextDocumentPosition{
			Line:      pathStartLine - 1,   // Convert to 0-based
			Character: pathStartColumn - 1, // Convert to 0-based
		},
		End: TextDocumentPosition{
			Line:      pathStartLine - 1, // Convert to 0-based
			Character: pathEndColumn - 1, // Convert to 0-based
		},
	}
}

// calculateIncludePathRange calculates the exact range for the path inside an include statement.
// For example, in `include "foo.vdl"`, this returns the range for just `foo.vdl`.
func calculateIncludePathRange(include *ast.Include) TextDocumentRange {
	// The include starts at Pos, which includes the "include" keyword
	startPos := include.Pos

	// Since we don't know exactly how many spaces it has, we'll calculate from the end
	// The path is already stripped of quotes in include.Path
	path := include.Path.String()

	// The EndPos is at the end of the closing quote
	// So the closing quote is at EndPos.Column - 1
	// The path ends at EndPos.Column - 2 (before the closing quote)
	// The path starts at (path end - path length)

	pathEndColumn := include.EndPos.Column - 1 // Before the closing quote
	pathStartColumn := pathEndColumn - len(path)

	return TextDocumentRange{
		Start: TextDocumentPosition{
			Line:      startPos.Line - 1,   // Convert to 0-based
			Character: pathStartColumn - 1, // Convert to 0-based
		},
		End: TextDocumentPosition{
			Line:      include.EndPos.Line - 1, // Convert to 0-based
			Character: pathEndColumn - 1,       // Convert to 0-based
		},
	}
}

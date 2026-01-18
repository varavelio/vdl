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

	filePath := uriToPath(request.Params.TextDocument.URI)

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
				if !strings.HasPrefix(normPath, "file://") {
					normPath = "file://" + normPath
				}

				links = append(links, DocumentLink{
					Range: TextDocumentRange{
						Start: convertASTPositionToLSPPosition(child.Docstring.Pos),
						End:   convertASTPositionToLSPPosition(child.Docstring.EndPos),
					},
					Target:  normPath,
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
			if !strings.HasPrefix(normPath, "file://") {
				normPath = "file://" + normPath
			}

			links = append(links, DocumentLink{
				Range: TextDocumentRange{
					Start: convertASTPositionToLSPPosition(child.Include.Pos),
					End:   convertASTPositionToLSPPosition(child.Include.EndPos),
				},
				Target:  normPath,
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
	if !strings.HasPrefix(normPath, "file://") {
		normPath = "file://" + normPath
	}

	*links = append(*links, DocumentLink{
		Range: TextDocumentRange{
			Start: convertASTPositionToLSPPosition(docstring.Pos),
			End:   convertASTPositionToLSPPosition(docstring.EndPos),
		},
		Target:  normPath,
		Tooltip: "Open markdown file",
	})
}

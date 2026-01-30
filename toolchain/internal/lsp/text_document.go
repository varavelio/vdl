package lsp

// TextDocumentIdentifier: Text documents are identified using a URI.
type TextDocumentIdentifier struct {
	// The text document's URI.
	URI string `json:"uri"`
}

// TextDocumentItem: An item to transfer a text document from the client to the server.
type TextDocumentItem struct {
	// The text document's URI.
	URI string `json:"uri"`
	// The text document's language identifier.
	LanguageID string `json:"languageId"`
	// The version number of this document (it will increase after each change, including undo/redo).
	Version int `json:"version"`
	// The content of the opened text document.
	Text string `json:"text"`
}

// TextDocumentContentChangeEvent: An event describing a change to a text document. If only
// a text is provided it is considered to be the full content of the document.
type TextDocumentContentChangeEvent struct {
	// The new text of the whole document.
	Text string `json:"text"`
}

// TextDocumentPosition: Position in a text document expressed as zero-based line and
// zero-based character offset.
type TextDocumentPosition struct {
	// The zero-based line number from the start of the document.
	Line int `json:"line"`
	// The zero-based character offset from the start of the line.
	Character int `json:"character"`
}

// TextDocumentRange: A range in a text document expressed as (zero-based) start and end positions.
type TextDocumentRange struct {
	// The start position of the range.
	Start TextDocumentPosition `json:"start"`
	// The end position of the range.
	End TextDocumentPosition `json:"end"`
}

// TextDocumentTextEdit: A text edit represents a change to a document.
type TextDocumentTextEdit struct {
	// The range of the text document to change.
	Range TextDocumentRange `json:"range"`
	// The string to replace the given range.
	NewText string `json:"newText"`
}

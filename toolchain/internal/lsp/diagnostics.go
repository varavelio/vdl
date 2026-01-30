package lsp

import (
	"runtime/debug"
	"time"

	"github.com/varavelio/vdl/toolchain/internal/core/analysis"
	"github.com/varavelio/vdl/toolchain/internal/core/ast"
)

// DiagnosticSeverity defines the severity level of a diagnostic.
type DiagnosticSeverity int

const (
	// Error severity level.
	DiagnosticSeverityError DiagnosticSeverity = 1
	// Warning severity level.
	DiagnosticSeverityWarning DiagnosticSeverity = 2
	// Information severity level.
	DiagnosticSeverityInformation DiagnosticSeverity = 3
	// Hint severity level.
	DiagnosticSeverityHint DiagnosticSeverity = 4
)

// Diagnostic represents a diagnostic, such as a compiler error or warning.
type Diagnostic struct {
	// The range at which the message applies.
	Range TextDocumentRange `json:"range"`
	// The diagnostic's severity. If omitted, client should treat it as error.
	Severity DiagnosticSeverity `json:"severity,omitempty"`
	// The diagnostic's code, which might appear in the user interface.
	Code string `json:"code,omitempty"`
	// A human-readable string describing the source of this diagnostic.
	Source string `json:"source,omitempty"`
	// The diagnostic's message.
	Message string `json:"message"`
}

// NotificationMessagePublishDiagnostics represents a notification message for publishing diagnostics.
type NotificationMessagePublishDiagnostics struct {
	NotificationMessage
	Params NotificationMessagePublishDiagnosticsParams `json:"params"`
}

// NotificationMessagePublishDiagnosticsParams represents the parameters for a publish diagnostics notification.
type NotificationMessagePublishDiagnosticsParams struct {
	// The URI for which diagnostic information is reported.
	URI string `json:"uri"`
	// An array of diagnostic information items.
	Diagnostics []Diagnostic `json:"diagnostics"`
}

// ConvertAnalysisDiagnosticToLSPDiagnostic converts an analysis diagnostic to an LSP diagnostic.
func ConvertAnalysisDiagnosticToLSPDiagnostic(diag analysis.Diagnostic) Diagnostic {
	return Diagnostic{
		Range: TextDocumentRange{
			Start: convertASTPositionToLSPPosition(diag.Pos),
			End:   convertASTPositionToLSPPosition(diag.EndPos),
		},
		Severity: DiagnosticSeverityError, // All analysis diagnostics are treated as errors for now
		Code:     diag.Code,
		Source:   "vdl",
		Message:  diag.Message,
	}
}

// convertASTPositionToLSPPosition converts an AST position to an LSP position.
func convertASTPositionToLSPPosition(pos ast.Position) TextDocumentPosition {
	// LSP positions are zero-based, but AST positions are one-based
	line := max(pos.Line-1, 0)
	char := max(pos.Column-1, 0)
	return TextDocumentPosition{
		Line:      line,
		Character: char,
	}
}

// publishDiagnostics sends diagnostics to the client.
func (l *LSP) publishDiagnostics(uri string, diagnostics []Diagnostic) {
	notification := NotificationMessagePublishDiagnostics{
		NotificationMessage: NotificationMessage{
			Message: Message{
				JSONRPC: "2.0",
				Method:  "textDocument/publishDiagnostics",
			},
		},
		Params: NotificationMessagePublishDiagnosticsParams{
			URI:         uri,
			Diagnostics: diagnostics,
		},
	}

	if err := l.sendMessage(notification); err != nil {
		l.logger.Error("failed to publish diagnostics", "uri", uri, "error", err)
	}
}

// clearDiagnostics clears diagnostics for the given URI.
func (l *LSP) clearDiagnostics(uri string) {
	l.publishDiagnostics(uri, []Diagnostic{})
}

// analyzeAndPublishDiagnostics analyzes the document at the given file path and publishes diagnostics.
// filePath is the native OS path used for analysis, uri is the LSP URI to send to the client.
func (l *LSP) analyzeAndPublishDiagnostics(filePath, uri string) {
	// Run the analysis
	_, diagnostics := l.analyze(filePath)

	// Convert analysis diagnostics to LSP diagnostics
	lspDiagnostics := make([]Diagnostic, 0, len(diagnostics))
	for _, diag := range diagnostics {
		lspDiagnostics = append(lspDiagnostics, ConvertAnalysisDiagnosticToLSPDiagnostic(diag))
	}

	// Publish diagnostics
	l.publishDiagnostics(uri, lspDiagnostics)
}

// analyzeAndPublishDiagnosticsDebounced schedules an analysis for the given file with debouncing.
// If another analysis is scheduled within the debounce time, the previous one is cancelled.
// filePath is the native OS path used for analysis, uri is the LSP URI to send to the client.
func (l *LSP) analyzeAndPublishDiagnosticsDebounced(filePath, uri string) {
	// debounceTime is the time to wait before running the analyzer after a document change.
	const debounceTime = 500 * time.Millisecond

	l.analysisTimerMu.Lock()
	defer l.analysisTimerMu.Unlock()

	// Cancel any existing timer
	if l.analysisTimer != nil {
		l.analysisTimer.Stop()
	}

	// Schedule a new analysis
	l.analysisTimer = time.AfterFunc(debounceTime, func() {
		// Recover from any panic inside the goroutine so the server keeps running
		defer func() {
			if r := recover(); r != nil {
				l.logger.Error("panic during diagnostics analysis", "panic", r, "stack", string(debug.Stack()))
			}
		}()

		// Check if another analysis is already in progress
		l.analysisInProgressMu.Lock()
		if l.analysisInProgress {
			l.analysisInProgressMu.Unlock()
			// If an analysis is already in progress, schedule another one
			l.analyzeAndPublishDiagnosticsDebounced(filePath, uri)
			return
		}
		l.analysisInProgress = true
		l.analysisInProgressMu.Unlock()

		// Run the analysis
		l.analyzeAndPublishDiagnostics(filePath, uri)

		// Mark analysis as complete
		l.analysisInProgressMu.Lock()
		l.analysisInProgress = false
		l.analysisInProgressMu.Unlock()
	})
}

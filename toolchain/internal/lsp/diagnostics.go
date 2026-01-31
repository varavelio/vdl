package lsp

import (
	"context"
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
// It uses the provided context for cancellation support.
// filePath is the native OS path used for analysis, uri is the LSP URI to send to the client.
func (l *LSP) analyzeAndPublishDiagnostics(ctx context.Context, filePath, uri string) {
	// Check for cancellation before starting
	if ctx.Err() != nil {
		return
	}

	// Run the analysis
	_, diagnostics := l.analyze(ctx, filePath)

	// Check for cancellation - don't publish if cancelled
	if ctx.Err() != nil {
		return
	}

	// Convert analysis diagnostics to LSP diagnostics
	lspDiagnostics := make([]Diagnostic, 0, len(diagnostics))
	for _, diag := range diagnostics {
		lspDiagnostics = append(lspDiagnostics, ConvertAnalysisDiagnosticToLSPDiagnostic(diag))
	}

	// Publish diagnostics
	l.publishDiagnostics(uri, lspDiagnostics)
}

// analyzeAndPublishDiagnosticsImmediate runs immediate analysis without debouncing.
// This is used for didOpen events where we want instant feedback.
func (l *LSP) analyzeAndPublishDiagnosticsImmediate(filePath, uri string) {
	ctx := context.Background()
	l.analyzeAndPublishDiagnostics(ctx, filePath, uri)
}

// analyzeAndPublishDiagnosticsDebounced schedules an analysis for the given file with debouncing.
// If another analysis is scheduled within the debounce time, the previous one is cancelled.
// This also triggers re-analysis of dependent files (files that import the changed file).
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

	// Cancel any in-progress analysis
	l.analysisCtxMu.Lock()
	if l.analysisCancel != nil {
		l.analysisCancel()
	}
	l.analysisCtxMu.Unlock()

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

		// Create a new cancellable context for this analysis run
		l.analysisCtxMu.Lock()
		ctx, cancel := context.WithCancel(context.Background())
		l.analysisCtx = ctx
		l.analysisCancel = cancel
		l.analysisCtxMu.Unlock()

		// Run the analysis for the changed file
		l.analyzeAndPublishDiagnostics(ctx, filePath, uri)

		// Propagate changes to dependent files
		if ctx.Err() == nil {
			l.reanalyzeDependents(ctx, filePath)
		}

		// Mark analysis as complete
		l.analysisInProgressMu.Lock()
		l.analysisInProgress = false
		l.analysisInProgressMu.Unlock()
	})
}

// reanalyzeDependents re-analyzes all open files that depend on the given file.
// This ensures that when an imported file is modified, the diagnostics in
// importing files are updated without requiring manual edits to those files.
func (l *LSP) reanalyzeDependents(ctx context.Context, filePath string) {
	// Get all files that depend on this file
	dependents := l.depGraph.GetDependents(filePath)
	if len(dependents) == 0 {
		return
	}

	l.logger.Info("propagating changes to dependents", "file", filePath, "dependents", dependents)

	// Re-analyze each dependent that is currently open
	for _, depPath := range dependents {
		// Check for cancellation
		if ctx.Err() != nil {
			return
		}

		// Only re-analyze if the file is open in the editor
		uri := l.getOpenDocURI(depPath)
		if uri == "" {
			continue
		}

		l.logger.Info("re-analyzing dependent", "file", depPath)
		l.analyzeAndPublishDiagnostics(ctx, depPath, uri)

		// Recursively propagate to transitive dependents
		l.reanalyzeDependents(ctx, depPath)
	}
}

// registerOpenDoc registers a document as open with its file path and URI.
func (l *LSP) registerOpenDoc(filePath, uri string) {
	l.openDocsMu.Lock()
	defer l.openDocsMu.Unlock()
	l.openDocs[filePath] = uri
}

// unregisterOpenDoc removes a document from the open documents registry.
func (l *LSP) unregisterOpenDoc(filePath string) {
	l.openDocsMu.Lock()
	defer l.openDocsMu.Unlock()
	delete(l.openDocs, filePath)

	// Also clean up the dependency graph for this file
	l.depGraph.RemoveFile(filePath)
}

// getOpenDocURI returns the URI for an open document, or empty string if not open.
func (l *LSP) getOpenDocURI(filePath string) string {
	l.openDocsMu.RLock()
	defer l.openDocsMu.RUnlock()
	return l.openDocs[filePath]
}

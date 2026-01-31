package lsp

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"runtime/debug"
	"sync"
	"time"

	"github.com/varavelio/vdl/toolchain/internal/core/analysis"
	"github.com/varavelio/vdl/toolchain/internal/core/vfs"
)

type LSP struct {
	reader               io.Reader
	writer               io.Writer
	handlerMu            sync.Mutex
	logger               *LSPLogger
	fs                   *vfs.FileSystem
	analysisTimer        *time.Timer
	analysisTimerMu      sync.Mutex
	analysisInProgress   bool
	analysisInProgressMu sync.Mutex

	// Dependency tracking for reactive updates
	depGraph *DependencyGraph

	// Open documents tracking: maps absolute file path to its LSP URI
	openDocs   map[string]string
	openDocsMu sync.RWMutex

	// Context cancellation for analysis pipeline
	analysisCtx    context.Context
	analysisCancel context.CancelFunc
	analysisCtxMu  sync.Mutex
}

// New creates a new LSP instance. It uses the given reader and writer to read and write
// messages to the LSP server.
func New(reader io.Reader, writer io.Writer) *LSP {
	return &LSP{
		reader:               reader,
		writer:               writer,
		handlerMu:            sync.Mutex{},
		logger:               NewLSPLogger(),
		fs:                   vfs.New(),
		analysisTimer:        nil,
		analysisTimerMu:      sync.Mutex{},
		analysisInProgress:   false,
		analysisInProgressMu: sync.Mutex{},
		depGraph:             NewDependencyGraph(),
		openDocs:             make(map[string]string),
		openDocsMu:           sync.RWMutex{},
		analysisCtx:          nil,
		analysisCancel:       nil,
		analysisCtxMu:        sync.Mutex{},
	}
}

// analyze runs the analysis pipeline on the given file path and returns the program and diagnostics.
// It also updates the dependency graph based on the resolved imports.
func (l *LSP) analyze(ctx context.Context, filePath string) (*analysis.Program, []analysis.Diagnostic) {
	// Check for cancellation before starting
	if ctx.Err() != nil {
		return nil, nil
	}

	program, diagnostics := analysis.AnalyzeWithContext(ctx, l.fs, filePath)

	// Check for cancellation after analysis
	if ctx.Err() != nil {
		return nil, nil
	}

	// Update dependency graph based on the program's file imports
	if program != nil {
		for path, file := range program.Files {
			l.depGraph.UpdateDependencies(path, file.Includes)
		}
	}

	return program, diagnostics
}

// Run starts the LSP server. It will read messages from the reader and write responses
// to the writer.
func (l *LSP) Run() error {
	if l.reader == nil || l.writer == nil {
		return fmt.Errorf("reader and writer are required")
	}

	scanner := bufio.NewScanner(l.reader)
	scanner.Split(scannerSplitFunc)

	for scanner.Scan() {
		shouldExit, err := l.handleMessage(scanner.Bytes())
		if err != nil {
			l.logger.Error(err.Error())
			return err
		}

		if shouldExit {
			return nil
		}
	}

	return nil
}

func (l *LSP) handleMessage(rawBytes []byte) (bool, error) {
	// Add panic recovery to prevent crashes. Instead of crashing, log the panic.
	defer func() {
		if r := recover(); r != nil {
			l.logger.Error("panic while handling message", "panic", r, "stack", string(debug.Stack()))
		}
	}()

	l.handlerMu.Lock()
	defer l.handlerMu.Unlock()

	rawMessage, err := decodeToMap(rawBytes)
	if err != nil {
		return false, fmt.Errorf("failed to decode message: %w", err)
	}

	messageID, messageHasID := rawMessage["id"]
	messageMethod, messageHasMethod := rawMessage["method"]
	if !messageHasMethod {
		return false, nil
	}

	if messageHasID {
		l.logger.Info("message received", "id", messageID, "method", messageMethod, "raw", rawMessage)
	} else {
		l.logger.Info("notification received", "method", messageMethod, "raw", rawMessage)
	}

	var response any
	var shouldExit bool

	switch messageMethod {
	// Lifecycle operations
	case "initialize":
		response, err = l.handleInitialize(rawBytes)
	case "initialized":
		response, err = l.handleInitialized(rawBytes)
	case "shutdown":
		response, err = l.handleShutdown(rawBytes)
	case "exit":
		shouldExit = true

	// Text document operations
	case "textDocument/didOpen":
		response, err = l.handleTextDocumentDidOpen(rawBytes)
	case "textDocument/didChange":
		response, err = l.handleTextDocumentDidChange(rawBytes)
	case "textDocument/didClose":
		response, err = l.handleTextDocumentDidClose(rawBytes)
	case "textDocument/formatting":
		response, err = l.handleTextDocumentFormatting(rawBytes)
	case "textDocument/definition":
		response, err = l.handleTextDocumentDefinition(rawBytes)
	case "textDocument/hover":
		response, err = l.handleTextDocumentHover(rawBytes)
	case "textDocument/rename":
		response, err = l.handleTextDocumentRename(rawBytes)
	case "textDocument/documentLink":
		response, err = l.handleTextDocumentDocumentLink(rawBytes)
	case "textDocument/references":
		response, err = l.handleTextDocumentReferences(rawBytes)
	case "textDocument/documentSymbol":
		response, err = l.handleTextDocumentDocumentSymbol(rawBytes)
	case "textDocument/completion":
		response, err = l.handleTextDocumentCompletion(rawBytes)
	}

	if err != nil {
		return false, fmt.Errorf("failed to handle message: %w", err)
	}

	if response != nil {
		if err := l.sendMessage(response); err != nil {
			return false, fmt.Errorf("failed to send message: %w", err)
		}
	}

	return shouldExit, nil
}

func (l *LSP) sendMessage(message any) error {
	messageBytes, err := encode(message)
	if err != nil {
		return fmt.Errorf("failed to encode message: %w", err)
	}

	_, err = l.writer.Write(messageBytes)
	if err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	rawResp, err := decodeToMap(messageBytes)
	if err != nil {
		return fmt.Errorf("failed to decode sent message: %w", err)
	}

	respID, respHasID := rawResp["id"]
	respMethod := rawResp["method"]

	if respHasID {
		l.logger.Info("message response sent", "id", respID, "method", respMethod, "raw", rawResp)
	} else {
		l.logger.Info("notification response sent", "method", respMethod, "raw", rawResp)
	}

	return nil
}

package lsp

import (
	"bytes"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
)

// LSPLogger is a LSPLogger for the LSP.
type LSPLogger struct {
	slogger  *slog.Logger
	filePath string
	writeMu  sync.Mutex
}

// NewLSPLogger creates a new logger. It will log to a file named .urpc-lsp.log in the
// user's home directory and if that fails, it will store the log in the system's temp
// directory.
//
// The logs are written in JSON format and the latest logs are always at the top of the file.
func NewLSPLogger() *LSPLogger {
	dir, err := os.UserHomeDir()
	if err != nil {
		dir = os.TempDir()
	}

	filePath := filepath.Join(dir, ".urpc-lsp.log")

	lgr := &LSPLogger{
		filePath: filePath,
		writeMu:  sync.Mutex{},
	}
	lgr.slogger = slog.New(slog.NewJSONHandler(lgr, nil))

	return lgr
}

func (l *LSPLogger) ensureLogFile() error {
	_, err := os.Stat(l.filePath)

	if os.IsNotExist(err) {
		file, err := os.Create(l.filePath)
		if err != nil {
			return err
		}
		return file.Close()
	}

	if err != nil {
		return err
	}

	return nil
}

func (l *LSPLogger) cleanOldLogs() error {
	const maxLogLines = 10_000

	content, err := os.ReadFile(l.filePath)
	if err != nil {
		return err
	}

	lines := bytes.Split(content, []byte("\n"))
	if len(lines) <= maxLogLines {
		return nil
	}

	lines = lines[len(lines)-maxLogLines:]
	err = os.WriteFile(l.filePath, bytes.Join(lines, []byte("\n")), 0644)
	if err != nil {
		return err
	}

	return nil
}

func (l *LSPLogger) Write(content []byte) (int, error) {
	l.writeMu.Lock()
	defer l.writeMu.Unlock()

	if err := l.ensureLogFile(); err != nil {
		return 0, fmt.Errorf("failed to ensure log file: %w", err)
	}

	if err := l.cleanOldLogs(); err != nil {
		return 0, fmt.Errorf("failed to clean old logs: %w", err)
	}

	existingContent, err := os.ReadFile(l.filePath)
	if err != nil {
		return 0, fmt.Errorf("failed to read log file: %w", err)
	}

	if err := os.WriteFile(l.filePath, append(content, existingContent...), 0644); err != nil {
		return 0, fmt.Errorf("failed to write log file: %w", err)
	}

	return len(content), nil
}

// Info logs a message with severity INFO.
func (l *LSPLogger) Info(msg string, args ...any) {
	l.slogger.Info(msg, args...)
}

// Error logs a message with severity ERROR.
func (l *LSPLogger) Error(msg string, args ...any) {
	l.slogger.Error(msg, args...)
}

// Debug logs a message with severity DEBUG.
func (l *LSPLogger) Debug(msg string, args ...any) {
	l.slogger.Debug(msg, args...)
}

// Warn logs a message with severity WARNING.
func (l *LSPLogger) Warn(msg string, args ...any) {
	l.slogger.Warn(msg, args...)
}

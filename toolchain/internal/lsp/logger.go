package lsp

import (
	"log/slog"
	"os"
	"path/filepath"
	"sync"
)

const (
	maxLogSizeBytes = 3 * 1024 * 1024 // 3MB
	logDirName      = "vdl"
	logFileName     = "lsp.log"
	oldLogFileName  = "lsp.old.log"
)

type LSPLogger struct {
	logger *slog.Logger
	file   *os.File
	mu     sync.Mutex
	path   string
	size   int64
}

// GetLogFilePath returns the absolute path to the log file.
// Used by the CLI (vdl lsp --log-path) and the Logger.
func GetLogFilePath() string {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		cacheDir = os.TempDir()
	}

	// Ensure directory exists
	dir := filepath.Join(cacheDir, logDirName)
	_ = os.MkdirAll(dir, 0755)

	return filepath.Join(dir, logFileName)
}

func NewLSPLogger() *LSPLogger {
	l := &LSPLogger{
		path: GetLogFilePath(),
	}

	// Ensure file is open and initial rotation check
	l.ensureFileOpen()
	l.rotate()

	l.logger = slog.New(slog.NewJSONHandler(l, nil))
	return l
}

// ensureFileOpen opens the file if not already open and updates the size counter.
// Caller must verify error.
func (l *LSPLogger) ensureFileOpen() {
	if l.file != nil {
		return
	}

	// Open in append mode for performance
	file, err := os.OpenFile(l.path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return
	}
	l.file = file

	// Initialize size from disk
	info, err := file.Stat()
	if err == nil {
		l.size = info.Size()
	} else {
		l.size = 0
	}
}

func (l *LSPLogger) rotate() {
	if l.file == nil || l.size <= maxLogSizeBytes {
		return
	}

	if l.file != nil {
		_ = l.file.Close()
		l.file = nil
	}

	oldPath := filepath.Join(filepath.Dir(l.path), oldLogFileName)
	_ = os.Rename(l.path, oldPath)

	l.ensureFileOpen()
}

// Write implements io.Writer with optimized in-memory size tracking.
func (l *LSPLogger) Write(p []byte) (n int, err error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Ensure the file is open and rotate file if necessary
	l.ensureFileOpen()
	l.rotate()

	// Protect writes to a nil file
	if l.file == nil {
		return 0, os.ErrPermission
	}

	// Write file and update counter
	n, err = l.file.Write(p)
	if err == nil {
		l.size += int64(n)
	}

	return n, err
}

func (l *LSPLogger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.file != nil {
		err := l.file.Close()
		l.file = nil
		return err
	}
	return nil
}

// Info logs at [slog.LevelInfo].
func (l *LSPLogger) Info(msg string, args ...any) {
	l.logger.Info(msg, args...)
}

// Error logs at [slog.LevelError].
func (l *LSPLogger) Error(msg string, args ...any) {
	l.logger.Error(msg, args...)
}

// Warn logs at [slog.LevelWarn].
func (l *LSPLogger) Warn(msg string, args ...any) {
	l.logger.Warn(msg, args...)
}

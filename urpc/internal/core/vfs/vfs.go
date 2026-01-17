// Package vfs provides a virtual file system with in-memory caching.
// It is designed to support scenarios where files may exist both on disk
// and in memory (e.g., unsaved editor buffers in an LSP context).
package vfs

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// FileSystem provides a thread-safe virtual file system with read-through caching.
// Files written via WriteFile are stored in memory and take precedence over disk.
type FileSystem struct {
	mu    sync.RWMutex
	files map[string][]byte
}

// New creates a new FileSystem instance with an empty cache.
func New() *FileSystem {
	return &FileSystem{
		files: make(map[string][]byte),
	}
}

// Resolve computes the canonical absolute path for a given path.
//
//   - If path is already absolute, it returns the cleaned version.
//   - If baseFile is empty (entry point scenario), path is resolved relative to the current working directory.
//   - Otherwise, path is resolved relative to the directory containing baseFile.
func (fs *FileSystem) Resolve(baseFile string, path string) string {
	if filepath.IsAbs(path) {
		return filepath.Clean(path)
	}

	if baseFile == "" {
		abs, err := filepath.Abs(path)
		if err != nil {
			return filepath.Clean(path)
		}
		return abs
	}

	baseDir := filepath.Dir(baseFile)
	fullPath := filepath.Join(baseDir, path)

	return filepath.Clean(fullPath)
}

// ReadFile returns the content of the file at the given absolute path.
//
// It first checks the in-memory cache; if not found, it reads from disk
// and caches the result.
//
// The path is normalized to an absolute path internally.
//
// Use Resolve if you need to compute the absolute path before reading.
func (fs *FileSystem) ReadFile(absolutePath string) ([]byte, error) {
	absPath, err := filepath.Abs(absolutePath)
	if err != nil {
		return nil, fmt.Errorf("absolute path check: %w", err)
	}

	fs.mu.RLock()
	content, hit := fs.files[absPath]
	fs.mu.RUnlock()

	if hit {
		return content, nil
	}

	content, err = os.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("read file from disk: %w", err)
	}

	fs.mu.Lock()
	fs.files[absPath] = content
	fs.mu.Unlock()

	return content, nil
}

// WriteFileCache stores content in the virtual file system's memory cache.
//
// This is useful for tracking unsaved changes (dirty buffers) without
// writing to disk.
//
// The path is normalized to an absolute path internally.
func (fs *FileSystem) WriteFileCache(path string, content []byte) {
	absPath, _ := filepath.Abs(path)

	fs.mu.Lock()
	fs.files[absPath] = content
	fs.mu.Unlock()
}

// RemoveFileCache removes a file from the memory cache.
//
// This is useful when a file is closed or saved, and the cached version
// should no longer take precedence over the disk version.
//
// The path is normalized to an absolute path internally.
// Returns true if the file was in the cache and removed, false otherwise.
func (fs *FileSystem) RemoveFileCache(path string) bool {
	absPath, _ := filepath.Abs(path)

	fs.mu.Lock()
	_, exists := fs.files[absPath]
	if exists {
		delete(fs.files, absPath)
	}
	fs.mu.Unlock()

	return exists
}

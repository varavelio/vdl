// Package docstore provides a file caching system with memory and disk tiers.
// It's designed for efficient file access in language servers and analyzers.
package docstore

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/uforg/uforpc/urpc/internal/util/filepathutil"
)

var (
	// ErrFileNotFound is returned when a requested file cannot be found.
	// It's an alias to os.ErrNotExist for compatibility with standard libraries.
	ErrFileNotFound = os.ErrNotExist
)

// MemCacheFile represents a file stored in memory.
// It contains the file content and its SHA-256 hash.
type MemCacheFile struct {
	Content string // File content
	Hash    string // SHA-256 hash of content in hex format
}

// DiskCacheFile represents a file from disk cached in memory.
// It includes content, hash, and modification time for cache invalidation.
type DiskCacheFile struct {
	Content string    // File content
	Hash    string    // SHA-256 hash of content in hex format
	Mtime   time.Time // Last modification time for cache invalidation
}

// Docstore provides file access with a two-tier caching system.
// It implements analyzer.FileProvider and manages file content with thread safety.
//
// The two caches are:
//   - memCache: Primary cache for files opened or modified in memory
//   - diskCache: Secondary cache for files read from disk (used as fallback when not in memCache)
type Docstore struct {
	memCache  map[string]MemCacheFile  // Normalized Path -> MemCacheFile
	diskCache map[string]DiskCacheFile // Normalized Path -> DiskCacheFile
	mu        sync.RWMutex             // Protects concurrent access to memCache and diskCache
}

// NewDocstore creates a new Docstore with initialized caches and mutex.
func NewDocstore() *Docstore {
	return &Docstore{
		memCache:  make(map[string]MemCacheFile),
		diskCache: make(map[string]DiskCacheFile),
		mu:        sync.RWMutex{},
	}
}

// OpenInMem stores file content in memory cache with its hash.
// It normalizes the path and removes any existing disk cache entry for the same path.
func (d *Docstore) OpenInMem(filePath string, content string) error {
	normFilePath, err := filepathutil.Normalize("", filePath)
	if err != nil {
		return fmt.Errorf("error normalizing file path: %w", err)
	}

	sum := sha256.Sum256([]byte(content))
	hash := fmt.Sprintf("%x", sum)

	d.mu.Lock()
	defer d.mu.Unlock()

	d.memCache[normFilePath] = MemCacheFile{
		Content: content,
		Hash:    hash,
	}

	// If exists in diskCache then delete it
	// to prioritize the in-memory version
	delete(d.diskCache, normFilePath)

	return nil
}

// ChangeInMem updates file content in memory cache.
// It's equivalent to OpenInMem and provided for semantic clarity.
func (d *Docstore) ChangeInMem(filePath string, content string) error {
	return d.OpenInMem(filePath, content)
}

// CloseInMem removes a file from memory cache.
// The file will still be accessible from disk if it exists there.
func (d *Docstore) CloseInMem(filePath string) error {
	normFilePath, err := filepathutil.Normalize("", filePath)
	if err != nil {
		return fmt.Errorf("error normalizing file path: %w", err)
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	delete(d.memCache, normFilePath)

	return nil
}

// GetInMemory retrieves file content from memory cache.
// Returns content, hash, existence flag, and any error encountered.
//
// Parameters:
//   - relativeToFilePath: Optional base path for resolving relative paths
//   - filePath: Path to retrieve (absolute or relative to relativeToFilePath)
func (d *Docstore) GetInMemory(relativeToFilePath string, filePath string) (string, string, bool, error) {
	normFilePath, err := filepathutil.Normalize(relativeToFilePath, filePath)
	if err != nil {
		return "", "", false, fmt.Errorf("error normalizing file path: %w", err)
	}

	d.mu.RLock()
	defer d.mu.RUnlock()

	cachedFile, ok := d.memCache[normFilePath]
	if !ok {
		return "", "", false, nil
	}

	return cachedFile.Content, cachedFile.Hash, true, nil
}

// GetFromDisk retrieves file content from disk with caching.
// It checks disk cache first, validates freshness via mtime, and reads from disk if needed.
// Thread-safe with proper lock handling for concurrent access.
//
// Parameters:
//   - relativeToFilePath: Optional base path for resolving relative paths
//   - filePath: Path to retrieve (absolute or relative to relativeToFilePath)
func (d *Docstore) GetFromDisk(relativeToFilePath string, filePath string) (string, string, bool, error) {
	// 1. Normalize the file path
	normFilePath, err := filepathutil.Normalize(relativeToFilePath, filePath)
	if err != nil {
		return "", "", false, fmt.Errorf("error normalizing file path: %w", err)
	}

	// Use a loop instead of recursion to avoid potential stack overflow
	for {
		// Start with a read lock
		d.mu.RLock()

		// 2. Check if the file exists in diskCache
		cachedFile, ok := d.diskCache[normFilePath]

		if !ok {
			// Not in cache, release read lock
			d.mu.RUnlock()

			// Check if file exists and get info
			fileInfo, err := os.Stat(normFilePath)
			if errors.Is(err, os.ErrNotExist) {
				return "", "", false, nil
			}
			if err != nil {
				return "", "", false, fmt.Errorf("error getting file info: %w", err)
			}
			if fileInfo.IsDir() {
				return "", "", false, fmt.Errorf("file path is a directory: %s", normFilePath)
			}

			// Read file content
			content, err := os.ReadFile(normFilePath)
			if err != nil {
				return "", "", false, fmt.Errorf("error reading file: %w", err)
			}

			// Calculate hash
			sum := sha256.Sum256(content)
			hash := fmt.Sprintf("%x", sum)

			// Acquire write lock to update cache
			d.mu.Lock()
			d.diskCache[normFilePath] = DiskCacheFile{
				Content: string(content),
				Hash:    hash,
				Mtime:   fileInfo.ModTime(),
			}
			d.mu.Unlock()

			return string(content), hash, true, nil
		}

		// File is in cache, get file info to check if it's stale
		// We can release the read lock while checking the file
		d.mu.RUnlock()

		fileInfo, err := os.Stat(normFilePath)
		if errors.Is(err, os.ErrNotExist) {
			// File no longer exists, acquire write lock to remove from cache
			d.mu.Lock()
			delete(d.diskCache, normFilePath)
			d.mu.Unlock()
			return "", "", false, nil
		}
		if err != nil {
			return "", "", false, fmt.Errorf("error getting file info: %w", err)
		}
		if fileInfo.IsDir() {
			return "", "", false, fmt.Errorf("file path is a directory: %s", normFilePath)
		}

		// Check if file has been modified
		mtime := fileInfo.ModTime()
		if mtime != cachedFile.Mtime {
			// File has changed, acquire write lock to remove from cache
			d.mu.Lock()
			delete(d.diskCache, normFilePath)
			d.mu.Unlock()
			// Continue the loop to read the updated file
			continue
		}

		// File is in cache and not stale, return cached content
		return cachedFile.Content, cachedFile.Hash, true, nil
	}
}

// GetFileAndHash retrieves file content with a two-tier cache strategy.
// It implements analyzer.FileProvider.GetFileAndHash by checking memory cache first,
// then falling back to disk cache. Returns standard os.ErrNotExist if file not found.
func (d *Docstore) GetFileAndHash(relativeTo string, path string) (string, string, error) {
	content, hash, exists, err := d.GetInMemory(relativeTo, path)
	if err != nil {
		return "", "", fmt.Errorf("error getting file from memCache: %w", err)
	}
	if exists {
		return content, hash, nil
	}

	content, hash, exists, err = d.GetFromDisk(relativeTo, path)
	if err != nil {
		return "", "", fmt.Errorf("error getting file from diskCache: %w", err)
	}
	if exists {
		return content, hash, nil
	}

	// Return a standard error that can be checked with errors.Is
	return "", "", fmt.Errorf("file not found: %s: %w", path, os.ErrNotExist)
}

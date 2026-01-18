package vfs

import (
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	t.Run("creates a FileSystem with empty cache", func(t *testing.T) {
		fs := New()

		require.NotNil(t, fs)
		require.NotNil(t, fs.files)
		require.Empty(t, fs.files)
	})
}

func TestFileSystem_Resolve(t *testing.T) {
	fs := New()

	t.Run("returns cleaned path when path is already absolute", func(t *testing.T) {
		result := fs.Resolve("/some/base/file.txt", "/etc/hosts")

		require.Equal(t, "/etc/hosts", result)
	})

	t.Run("cleans absolute path with redundant elements", func(t *testing.T) {
		result := fs.Resolve("/some/base/file.txt", "/foo/../bar/./baz")

		require.Equal(t, "/bar/baz", result)
	})

	t.Run("resolves relative path against CWD when baseFile is empty", func(t *testing.T) {
		cwd, err := os.Getwd()
		require.NoError(t, err)

		result := fs.Resolve("", "relative/path.txt")

		expected := filepath.Join(cwd, "relative/path.txt")
		require.Equal(t, expected, result)
	})

	t.Run("resolves relative path against baseFile directory", func(t *testing.T) {
		result := fs.Resolve("/project/src/main.urpc", "types/user.urpc")

		require.Equal(t, "/project/src/types/user.urpc", result)
	})

	t.Run("handles parent directory traversal in relative path", func(t *testing.T) {
		result := fs.Resolve("/project/src/nested/file.urpc", "../common/types.urpc")

		require.Equal(t, "/project/src/common/types.urpc", result)
	})

	t.Run("handles current directory reference in relative path", func(t *testing.T) {
		result := fs.Resolve("/project/src/file.urpc", "./types.urpc")

		require.Equal(t, "/project/src/types.urpc", result)
	})

	t.Run("handles multiple parent traversals", func(t *testing.T) {
		result := fs.Resolve("/a/b/c/d/file.txt", "../../../x/y.txt")

		require.Equal(t, "/a/x/y.txt", result)
	})

	t.Run("handles baseFile in root directory", func(t *testing.T) {
		result := fs.Resolve("/root.txt", "subdir/file.txt")

		require.Equal(t, "/subdir/file.txt", result)
	})
}

func TestFileSystem_ReadFile(t *testing.T) {
	t.Run("reads file from disk and caches it", func(t *testing.T) {
		fs := New()
		tempDir := t.TempDir()
		filePath := filepath.Join(tempDir, "test.txt")
		expectedContent := []byte("hello world")

		err := os.WriteFile(filePath, expectedContent, 0644)
		require.NoError(t, err)

		content, err := fs.ReadFile(filePath)

		require.NoError(t, err)
		require.Equal(t, expectedContent, content)

		// Verify it was cached
		absPath, _ := filepath.Abs(filePath)
		fs.mu.RLock()
		cachedContent, exists := fs.files[absPath]
		fs.mu.RUnlock()
		require.True(t, exists)
		require.Equal(t, expectedContent, cachedContent)
	})

	t.Run("returns cached content on subsequent reads", func(t *testing.T) {
		fs := New()
		tempDir := t.TempDir()
		filePath := filepath.Join(tempDir, "test.txt")
		originalContent := []byte("original content")

		err := os.WriteFile(filePath, originalContent, 0644)
		require.NoError(t, err)

		// First read - caches the content
		content1, err := fs.ReadFile(filePath)
		require.NoError(t, err)
		require.Equal(t, originalContent, content1)

		// Modify the file on disk
		modifiedContent := []byte("modified content")
		err = os.WriteFile(filePath, modifiedContent, 0644)
		require.NoError(t, err)

		// Second read - should return cached content
		content2, err := fs.ReadFile(filePath)
		require.NoError(t, err)
		require.Equal(t, originalContent, content2, "should return cached content, not modified file")
	})

	t.Run("returns error for non-existent file", func(t *testing.T) {
		fs := New()
		tempDir := t.TempDir()
		nonExistentPath := filepath.Join(tempDir, "does_not_exist.txt")

		content, err := fs.ReadFile(nonExistentPath)

		require.Error(t, err)
		require.Nil(t, content)
		require.True(t, errors.Is(err, os.ErrNotExist))
	})

	t.Run("normalizes path before caching", func(t *testing.T) {
		fs := New()
		tempDir := t.TempDir()
		filePath := filepath.Join(tempDir, "test.txt")
		expectedContent := []byte("content")

		err := os.WriteFile(filePath, expectedContent, 0644)
		require.NoError(t, err)

		// Read using a path with redundant elements
		dirtyPath := filepath.Join(tempDir, "subdir", "..", "test.txt")
		content, err := fs.ReadFile(dirtyPath)

		require.NoError(t, err)
		require.Equal(t, expectedContent, content)

		// Verify it was cached with the clean absolute path
		absPath, _ := filepath.Abs(filePath)
		fs.mu.RLock()
		_, exists := fs.files[absPath]
		fs.mu.RUnlock()
		require.True(t, exists)
	})

	t.Run("reads empty file correctly", func(t *testing.T) {
		fs := New()
		tempDir := t.TempDir()
		filePath := filepath.Join(tempDir, "empty.txt")

		err := os.WriteFile(filePath, []byte{}, 0644)
		require.NoError(t, err)

		content, err := fs.ReadFile(filePath)

		require.NoError(t, err)
		require.Empty(t, content)
	})
}

func TestFileSystem_WriteFileCache(t *testing.T) {
	t.Run("stores content in memory cache", func(t *testing.T) {
		fs := New()
		tempDir := t.TempDir()
		filePath := filepath.Join(tempDir, "virtual.txt")
		content := []byte("virtual content")

		fs.WriteFileCache(filePath, content)

		absPath, _ := filepath.Abs(filePath)
		fs.mu.RLock()
		storedContent, exists := fs.files[absPath]
		fs.mu.RUnlock()

		require.True(t, exists)
		require.Equal(t, content, storedContent)
	})

	t.Run("overwrites existing cached content", func(t *testing.T) {
		fs := New()
		tempDir := t.TempDir()
		filePath := filepath.Join(tempDir, "file.txt")

		fs.WriteFileCache(filePath, []byte("first"))
		fs.WriteFileCache(filePath, []byte("second"))

		absPath, _ := filepath.Abs(filePath)
		fs.mu.RLock()
		storedContent := fs.files[absPath]
		fs.mu.RUnlock()

		require.Equal(t, []byte("second"), storedContent)
	})

	t.Run("does not write to disk", func(t *testing.T) {
		fs := New()
		tempDir := t.TempDir()
		filePath := filepath.Join(tempDir, "no_disk.txt")
		content := []byte("only in memory")

		fs.WriteFileCache(filePath, content)

		_, err := os.Stat(filePath)
		require.True(t, os.IsNotExist(err), "file should not exist on disk")
	})

	t.Run("ReadFile returns WriteFileCache content over disk", func(t *testing.T) {
		fs := New()
		tempDir := t.TempDir()
		filePath := filepath.Join(tempDir, "test.txt")
		diskContent := []byte("disk content")
		memoryContent := []byte("memory content")

		err := os.WriteFile(filePath, diskContent, 0644)
		require.NoError(t, err)

		fs.WriteFileCache(filePath, memoryContent)

		content, err := fs.ReadFile(filePath)
		require.NoError(t, err)
		require.Equal(t, memoryContent, content)
	})

	t.Run("handles empty content", func(t *testing.T) {
		fs := New()
		tempDir := t.TempDir()
		filePath := filepath.Join(tempDir, "empty.txt")

		fs.WriteFileCache(filePath, []byte{})

		absPath, _ := filepath.Abs(filePath)
		fs.mu.RLock()
		storedContent, exists := fs.files[absPath]
		fs.mu.RUnlock()

		require.True(t, exists)
		require.Empty(t, storedContent)
	})

	t.Run("handles nil content", func(t *testing.T) {
		fs := New()
		tempDir := t.TempDir()
		filePath := filepath.Join(tempDir, "nil.txt")

		fs.WriteFileCache(filePath, nil)

		absPath, _ := filepath.Abs(filePath)
		fs.mu.RLock()
		storedContent, exists := fs.files[absPath]
		fs.mu.RUnlock()

		require.True(t, exists)
		require.Nil(t, storedContent)
	})
}

func TestFileSystem_RemoveFileCache(t *testing.T) {
	t.Run("removes cached file and returns true", func(t *testing.T) {
		fs := New()
		tempDir := t.TempDir()
		filePath := filepath.Join(tempDir, "cached.txt")

		fs.WriteFileCache(filePath, []byte("cached content"))

		removed := fs.RemoveFileCache(filePath)

		require.True(t, removed)

		absPath, _ := filepath.Abs(filePath)
		fs.mu.RLock()
		_, exists := fs.files[absPath]
		fs.mu.RUnlock()
		require.False(t, exists)
	})

	t.Run("returns false when file is not in cache", func(t *testing.T) {
		fs := New()
		tempDir := t.TempDir()
		filePath := filepath.Join(tempDir, "not_cached.txt")

		removed := fs.RemoveFileCache(filePath)

		require.False(t, removed)
	})

	t.Run("ReadFile falls back to disk after cache removal", func(t *testing.T) {
		fs := New()
		tempDir := t.TempDir()
		filePath := filepath.Join(tempDir, "test.txt")
		diskContent := []byte("disk content")
		cacheContent := []byte("cache content")

		err := os.WriteFile(filePath, diskContent, 0644)
		require.NoError(t, err)

		fs.WriteFileCache(filePath, cacheContent)

		// Verify cache is used
		content, err := fs.ReadFile(filePath)
		require.NoError(t, err)
		require.Equal(t, cacheContent, content)

		// Remove from cache
		removed := fs.RemoveFileCache(filePath)
		require.True(t, removed)

		// Should now read from disk
		content, err = fs.ReadFile(filePath)
		require.NoError(t, err)
		require.Equal(t, diskContent, content)
	})

	t.Run("only removes specified file from cache", func(t *testing.T) {
		fs := New()
		tempDir := t.TempDir()
		filePath1 := filepath.Join(tempDir, "file1.txt")
		filePath2 := filepath.Join(tempDir, "file2.txt")

		fs.WriteFileCache(filePath1, []byte("content1"))
		fs.WriteFileCache(filePath2, []byte("content2"))

		fs.RemoveFileCache(filePath1)

		absPath1, _ := filepath.Abs(filePath1)
		absPath2, _ := filepath.Abs(filePath2)

		fs.mu.RLock()
		_, exists1 := fs.files[absPath1]
		content2, exists2 := fs.files[absPath2]
		fs.mu.RUnlock()

		require.False(t, exists1, "file1 should be removed")
		require.True(t, exists2, "file2 should still exist")
		require.Equal(t, []byte("content2"), content2)
	})

	t.Run("handles path normalization", func(t *testing.T) {
		fs := New()
		tempDir := t.TempDir()
		filePath := filepath.Join(tempDir, "normalized.txt")
		dirtyPath := filepath.Join(tempDir, "subdir", "..", "normalized.txt")

		fs.WriteFileCache(filePath, []byte("content"))

		// Remove using dirty path
		removed := fs.RemoveFileCache(dirtyPath)

		require.True(t, removed)

		absPath, _ := filepath.Abs(filePath)
		fs.mu.RLock()
		_, exists := fs.files[absPath]
		fs.mu.RUnlock()
		require.False(t, exists)
	})
}

func TestFileSystem_Concurrency(t *testing.T) {
	t.Run("handles concurrent reads safely", func(t *testing.T) {
		fs := New()
		tempDir := t.TempDir()
		filePath := filepath.Join(tempDir, "concurrent.txt")
		content := []byte("concurrent content")

		err := os.WriteFile(filePath, content, 0644)
		require.NoError(t, err)

		var wg sync.WaitGroup
		const goroutines = 100

		for range goroutines {
			wg.Go(func() {
				result, err := fs.ReadFile(filePath)
				require.NoError(t, err)
				require.Equal(t, content, result)
			})
		}

		wg.Wait()
	})

	t.Run("handles concurrent writes safely", func(t *testing.T) {
		fs := New()
		tempDir := t.TempDir()
		filePath := filepath.Join(tempDir, "concurrent_write.txt")

		var wg sync.WaitGroup
		const goroutines = 100

		for i := range goroutines {
			wg.Go(func() {
				content := []byte{byte(i)}
				fs.WriteFileCache(filePath, content)
			})
		}

		wg.Wait()

		// Verify that a value was written (any of the concurrent writes)
		absPath, _ := filepath.Abs(filePath)
		fs.mu.RLock()
		_, exists := fs.files[absPath]
		fs.mu.RUnlock()
		require.True(t, exists)
	})

	t.Run("handles concurrent read and write safely", func(t *testing.T) {
		fs := New()
		tempDir := t.TempDir()
		filePath := filepath.Join(tempDir, "mixed.txt")
		initialContent := []byte("initial")

		err := os.WriteFile(filePath, initialContent, 0644)
		require.NoError(t, err)

		var wg sync.WaitGroup
		const goroutines = 50

		// Concurrent reads
		for range goroutines {
			wg.Go(func() {
				_, err := fs.ReadFile(filePath)
				require.NoError(t, err)
			})
		}

		// Concurrent writes
		for i := range goroutines {
			wg.Go(func() {
				content := []byte{byte(i)}
				fs.WriteFileCache(filePath, content)
			})
		}

		wg.Wait()
	})

	t.Run("handles concurrent remove operations safely", func(t *testing.T) {
		fs := New()
		tempDir := t.TempDir()

		var wg sync.WaitGroup
		const goroutines = 100

		// Pre-populate cache with files
		for i := range goroutines {
			filePath := filepath.Join(tempDir, "file"+string(rune('A'+i%26))+".txt")
			fs.WriteFileCache(filePath, []byte{byte(i)})
		}

		// Concurrent removes
		for i := range goroutines {
			wg.Go(func() {
				filePath := filepath.Join(tempDir, "file"+string(rune('A'+i%26))+".txt")
				fs.RemoveFileCache(filePath)
			})
		}

		wg.Wait()
	})
}

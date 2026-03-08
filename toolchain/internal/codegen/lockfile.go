package codegen

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/varavelio/vdl/toolchain/internal/codegen/locktypes"
	"github.com/varavelio/vdl/toolchain/internal/dirs"
)

const (
	defaultLockFileName = "vdl.lock"
	lockFileVersion     = int64(1)
	cacheLockDirName    = "lock"
	generatedFileMode   = 0o644
	generatedDirMode    = 0o755
)

var remoteHTTPClient = &http.Client{}

func loadLockFile(path string) (locktypes.VdlLockFileSchema, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return newLockFile(), nil
		}
		return locktypes.VdlLockFileSchema{}, fmt.Errorf("failed to read lock file %q: %w", path, err)
	}

	var lockFile locktypes.VdlLockFileSchema
	if err := json.Unmarshal(data, &lockFile); err != nil {
		return locktypes.VdlLockFileSchema{}, fmt.Errorf("failed to parse lock file %q: %w", path, err)
	}
	if lockFile.GetVersion() != lockFileVersion {
		return locktypes.VdlLockFileSchema{}, fmt.Errorf(
			"unsupported lock file version %d in %q: expected %d",
			lockFile.GetVersion(),
			path,
			lockFileVersion,
		)
	}

	hashes := cloneStringMap(lockFile.GetHashesOr(map[string]string{}))
	lockFile.Hashes = &hashes
	return lockFile, nil
}

func writeLockFile(path string, lockFile locktypes.VdlLockFileSchema) error {
	hashes := cloneStringMap(lockFile.GetHashesOr(map[string]string{}))
	lockFile.Version = lockFileVersion
	lockFile.Hashes = &hashes

	data, err := json.MarshalIndent(lockFile, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal lock file %q: %w", path, err)
	}
	data = append(data, '\n')

	if err := os.WriteFile(path, data, generatedFileMode); err != nil {
		return fmt.Errorf("failed to write lock file %q: %w", path, err)
	}

	return nil
}

func materializeRemotePlugins(plugins []runtimePlugin, lockFile *locktypes.VdlLockFileSchema) error {
	cacheDir, err := dirs.GetCacheDir()
	if err != nil {
		return err
	}

	cacheLockDir := filepath.Join(cacheDir, cacheLockDirName)
	if err := os.MkdirAll(cacheLockDir, generatedDirMode); err != nil {
		return fmt.Errorf("failed to create cache directory %q: %w", cacheLockDir, err)
	}

	hashes := cloneStringMap(lockFile.GetHashesOr(map[string]string{}))
	for i := range plugins {
		if plugins[i].Source.Kind != pluginSourceKindRemote {
			continue
		}

		expectedHash := hashes[plugins[i].Source.CanonicalURL]
		cachePath, actualHash, err := materializeRemoteDependency(
			plugins[i].Source.CanonicalURL,
			plugins[i].Source.Headers,
			expectedHash,
			cacheLockDir,
		)
		if err != nil {
			return fmt.Errorf("failed to fetch plugin %q: %w", plugins[i].Source.DisplayName, err)
		}

		plugins[i].Source.CachePath = cachePath
		hashes[plugins[i].Source.CanonicalURL] = actualHash
	}

	lockFile.Version = lockFileVersion
	lockFile.Hashes = &hashes
	return nil
}

func materializeRemoteDependency(rawURL string, headers http.Header, expectedHash, cacheDir string) (string, string, error) {
	cachePath := filepath.Join(cacheDir, hashRemoteCacheKey(rawURL)+".js")

	if data, err := os.ReadFile(cachePath); err == nil {
		actualHash := sha256Digest(data)
		if expectedHash != "" && actualHash != expectedHash {
			return "", "", fmt.Errorf(
				"cached dependency %q has hash %q but lock file requires %q",
				rawURL,
				actualHash,
				expectedHash,
			)
		}
		return cachePath, actualHash, nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return "", "", fmt.Errorf("failed to read cached dependency %q: %w", cachePath, err)
	}

	data, err := downloadRemoteDependency(rawURL, headers)
	if err != nil {
		return "", "", err
	}

	actualHash := sha256Digest(data)
	if expectedHash != "" && actualHash != expectedHash {
		return "", "", fmt.Errorf(
			"CRITICAL: remote dependency %q changed hash from %q to %q; the remote server may have been compromised",
			rawURL,
			expectedHash,
			actualHash,
		)
	}

	if err := os.WriteFile(cachePath, data, generatedFileMode); err != nil {
		return "", "", fmt.Errorf("failed to cache dependency %q: %w", rawURL, err)
	}

	return cachePath, actualHash, nil
}

func downloadRemoteDependency(rawURL string, headers http.Header) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request for %q: %w", rawURL, err)
	}
	req.Header = cloneHTTPHeader(headers)

	resp, err := remoteHTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download %q: %w", rawURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("failed to download %q: unexpected HTTP status %s", rawURL, resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body for %q: %w", rawURL, err)
	}

	return data, nil
}

func newLockFile() locktypes.VdlLockFileSchema {
	hashes := map[string]string{}
	return locktypes.VdlLockFileSchema{
		Version: lockFileVersion,
		Hashes:  &hashes,
	}
}

func hashRemoteCacheKey(rawURL string) string {
	sum := sha256.Sum256([]byte(rawURL))
	return hex.EncodeToString(sum[:])
}

func sha256Digest(data []byte) string {
	sum := sha256.Sum256(data)
	return "sha256-" + hex.EncodeToString(sum[:])
}

func cloneHTTPHeader(header http.Header) http.Header {
	if len(header) == 0 {
		return make(http.Header)
	}
	return header.Clone()
}

package dirs

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	defaultHomeDirName = ".vdl"
	cacheDirName       = "cache"
	logsDirName        = "logs"
	directoryMode      = 0o755
	logFileMode        = 0o666
)

var (
	lookupEnv = os.LookupEnv
	userHome  = os.UserHomeDir
	tempDir   = os.TempDir
	absPath   = filepath.Abs
	makeDir   = os.MkdirAll
	openFile  = os.OpenFile
)

// GetVDLHome returns the root VDL directory.
//
// Resolution order:
//  1. Uses VDL_HOME when defined and non-empty.
//  2. Falls back to ~/.vdl.
//  3. Uses the system temp directory when no user home is available.
func GetVDLHome() string {
	if customHome, ok := lookupEnv("VDL_HOME"); ok {
		customHome = strings.TrimSpace(customHome)
		if customHome != "" {
			return normalizePath(customHome)
		}
	}

	homeDir, err := userHome()
	if err == nil {
		homeDir = strings.TrimSpace(homeDir)
		if homeDir != "" {
			return filepath.Join(homeDir, defaultHomeDirName)
		}
	}

	return filepath.Join(tempDir(), defaultHomeDirName)
}

// GetCacheDir returns the absolute directory used for plugins and downloads.
func GetCacheDir() (string, error) {
	return ensureDir(filepath.Join(GetVDLHome(), cacheDirName))
}

// GetLogsDir returns the absolute logs directory.
func GetLogsDir() (string, error) {
	return ensureDir(filepath.Join(GetVDLHome(), logsDirName))
}

// OpenLog creates or opens a log file under the VDL logs directory.
func OpenLog(name string) (*os.File, string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, "", errors.New("log file name cannot be empty")
	}
	if filepath.Base(name) != name {
		return nil, "", fmt.Errorf("log file name must not contain path separators: %q", name)
	}

	logsDir, err := GetLogsDir()
	if err != nil {
		return nil, "", err
	}

	path := filepath.Join(logsDir, name)
	file, err := openFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, logFileMode)
	if err != nil {
		return nil, "", fmt.Errorf("open log file %q: %w", path, err)
	}

	return file, path, nil
}

// ensureDir creates path when necessary and returns the normalized directory
// path on success.
func ensureDir(path string) (string, error) {
	if err := makeDir(path, directoryMode); err != nil {
		return "", fmt.Errorf("create directory %q: %w", path, err)
	}
	return path, nil
}

// normalizePath converts path to an absolute clean path when possible.
func normalizePath(path string) string {
	if filepath.IsAbs(path) {
		return filepath.Clean(path)
	}

	absolutePath, err := absPath(path)
	if err != nil {
		return filepath.Clean(path)
	}

	return filepath.Clean(absolutePath)
}

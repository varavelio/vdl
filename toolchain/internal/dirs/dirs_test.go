package dirs

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetVDLHome(t *testing.T) {
	t.Run("uses VDL_HOME when it is configured", func(t *testing.T) {
		reset := setTestHooks(t)
		defer reset()

		customHome := filepath.Join(string(filepath.Separator), "custom", "vdl")

		lookupEnv = func(string) (string, bool) {
			return customHome, true
		}
		userHome = func() (string, error) {
			t.Fatal("user home should not be queried when VDL_HOME is configured")
			return "", nil
		}

		require.Equal(t, filepath.Clean(customHome), GetVDLHome())
	})

	t.Run("normalizes relative VDL_HOME paths", func(t *testing.T) {
		reset := setTestHooks(t)
		defer reset()

		lookupEnv = func(string) (string, bool) {
			return "relative-home", true
		}
		expectedHome := filepath.Join(string(filepath.Separator), "absolute", "relative-home")

		absPath = func(path string) (string, error) {
			require.Equal(t, "relative-home", path)
			return expectedHome, nil
		}

		require.Equal(t, expectedHome, GetVDLHome())
	})

	t.Run("falls back to user home", func(t *testing.T) {
		reset := setTestHooks(t)
		defer reset()

		home := filepath.Join(string(filepath.Separator), "home", "dev")

		lookupEnv = func(string) (string, bool) {
			return "", false
		}
		userHome = func() (string, error) {
			return home, nil
		}

		require.Equal(t, filepath.Join(home, ".vdl"), GetVDLHome())
	})

	t.Run("falls back to temp directory when user home is unavailable", func(t *testing.T) {
		reset := setTestHooks(t)
		defer reset()

		fallbackTemp := filepath.Join(string(filepath.Separator), "tmp", "fallback")

		lookupEnv = func(string) (string, bool) {
			return "", false
		}
		userHome = func() (string, error) {
			return "", errors.New("no home")
		}
		tempDir = func() string {
			return fallbackTemp
		}

		require.Equal(t, filepath.Join(fallbackTemp, ".vdl"), GetVDLHome())
	})
}

func TestGetCacheDir(t *testing.T) {
	t.Run("returns primary cache directory when creation succeeds", func(t *testing.T) {
		reset := setTestHooks(t)
		defer reset()

		customHome := filepath.Join(string(filepath.Separator), "custom", "vdl")

		lookupEnv = func(string) (string, bool) {
			return customHome, true
		}

		var created []string
		makeDir = func(path string, perm os.FileMode) error {
			require.Equal(t, os.FileMode(directoryMode), perm)
			created = append(created, path)
			return nil
		}

		cacheDir, err := GetCacheDir()
		require.NoError(t, err)
		require.Equal(t, filepath.Join(customHome, "cache"), cacheDir)
		require.Equal(t, []string{filepath.Join(customHome, "cache")}, created)
	})

	t.Run("returns error when creation fails", func(t *testing.T) {
		reset := setTestHooks(t)
		defer reset()

		customHome := filepath.Join(string(filepath.Separator), "custom", "vdl")

		lookupEnv = func(string) (string, bool) {
			return customHome, true
		}

		primary := filepath.Join(customHome, "cache")

		var created []string
		makeDir = func(path string, perm os.FileMode) error {
			require.Equal(t, os.FileMode(directoryMode), perm)
			created = append(created, path)
			return errors.New("permission denied")
		}

		cacheDir, err := GetCacheDir()
		require.Error(t, err)
		require.Empty(t, cacheDir)
		require.Contains(t, err.Error(), primary)
		require.Equal(t, []string{primary}, created)
	})
}

func TestGetLogsDir(t *testing.T) {
	t.Run("returns primary logs directory when creation succeeds", func(t *testing.T) {
		reset := setTestHooks(t)
		defer reset()

		customHome := filepath.Join(string(filepath.Separator), "custom", "vdl")

		lookupEnv = func(string) (string, bool) {
			return customHome, true
		}

		var created []string
		makeDir = func(path string, perm os.FileMode) error {
			require.Equal(t, os.FileMode(directoryMode), perm)
			created = append(created, path)
			return nil
		}

		logsDir, err := GetLogsDir()
		require.NoError(t, err)
		require.Equal(t, filepath.Join(customHome, "logs"), logsDir)
		require.Equal(t, []string{filepath.Join(customHome, "logs")}, created)
	})

	t.Run("returns error when creation fails", func(t *testing.T) {
		reset := setTestHooks(t)
		defer reset()

		customHome := filepath.Join(string(filepath.Separator), "custom", "vdl")

		lookupEnv = func(string) (string, bool) {
			return customHome, true
		}

		primary := filepath.Join(customHome, "logs")

		var created []string
		makeDir = func(path string, perm os.FileMode) error {
			require.Equal(t, os.FileMode(directoryMode), perm)
			created = append(created, path)
			return errors.New("permission denied")
		}

		logsDir, err := GetLogsDir()
		require.Error(t, err)
		require.Empty(t, logsDir)
		require.Contains(t, err.Error(), primary)
		require.Equal(t, []string{primary}, created)
	})
}

func TestOpenLog(t *testing.T) {
	t.Run("opens log file under logs directory", func(t *testing.T) {
		reset := setTestHooks(t)
		defer reset()

		customHome := filepath.Join(string(filepath.Separator), "custom", "vdl")
		lookupEnv = func(string) (string, bool) {
			return customHome, true
		}

		expectedPath := filepath.Join(customHome, "logs", "lsp.log")
		openFile = func(path string, flag int, perm os.FileMode) (*os.File, error) {
			require.Equal(t, expectedPath, path)
			require.Equal(t, os.O_CREATE|os.O_WRONLY|os.O_APPEND, flag)
			require.Equal(t, os.FileMode(logFileMode), perm)

			file, err := os.CreateTemp(t.TempDir(), "log-*.log")
			require.NoError(t, err)
			return file, nil
		}

		file, path, err := OpenLog("lsp.log")
		require.NoError(t, err)
		require.NotNil(t, file)
		require.Equal(t, expectedPath, path)
		require.NoError(t, file.Close())
	})

	t.Run("returns error when file name is empty", func(t *testing.T) {
		reset := setTestHooks(t)
		defer reset()

		file, path, err := OpenLog("   ")
		require.Error(t, err)
		require.Nil(t, file)
		require.Empty(t, path)
		require.Contains(t, err.Error(), "cannot be empty")
	})

	t.Run("returns error when file name contains separators", func(t *testing.T) {
		reset := setTestHooks(t)
		defer reset()

		file, path, err := OpenLog(filepath.Join("nested", "lsp.log"))
		require.Error(t, err)
		require.Nil(t, file)
		require.Empty(t, path)
		require.Contains(t, err.Error(), "must not contain path separators")
	})
}

func setTestHooks(t *testing.T) func() {
	t.Helper()

	originalLookupEnv := lookupEnv
	originalUserHome := userHome
	originalTempDir := tempDir
	originalAbsPath := absPath
	originalMakeDir := makeDir
	originalOpenFile := openFile

	return func() {
		lookupEnv = originalLookupEnv
		userHome = originalUserHome
		tempDir = originalTempDir
		absPath = originalAbsPath
		makeDir = originalMakeDir
		openFile = originalOpenFile
	}
}

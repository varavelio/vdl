package codegen

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRunWithLocalPlugin(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, filepath.Join(dir, "schema.vdl"), "type User {\n  name string\n}\n")
	writeTestFile(t, filepath.Join(dir, "plugin.js"), `exports.generate = () => ({ files: [{ path: "nested/generated.txt", content: "hello" }] })`)
	writeTestFile(t, filepath.Join(dir, defaultConfigFileName), `
		const config = {
			version 1
			plugins [
				{
					src "./plugin.js"
					schema "./schema.vdl"
					outDir "./gen"
				}
			]
		}
	`)
	writeTestFile(t, filepath.Join(dir, "gen", "stale.txt"), "old")

	fileCount, err := Run(dir)
	require.NoError(t, err)
	require.Equal(t, 1, fileCount)
	require.NoFileExists(t, filepath.Join(dir, "gen", "stale.txt"))
	require.FileExists(t, filepath.Join(dir, "gen", "nested", "generated.txt"))

	info, err := os.Stat(filepath.Join(dir, "gen", "nested", "generated.txt"))
	require.NoError(t, err)
	require.Zero(t, info.Mode().Perm()&0o111)
	require.FileExists(t, filepath.Join(dir, defaultLockFileName))
}

func TestRunRejectsTraversalBeforeCleaningOutputs(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, filepath.Join(dir, "schema.vdl"), "type User {\n  name string\n}\n")
	writeTestFile(t, filepath.Join(dir, "plugin.js"), `exports.generate = () => ({ files: [{ path: "../escape.txt", content: "bad" }] })`)
	writeTestFile(t, filepath.Join(dir, defaultConfigFileName), `
		const config = {
			version 1
			plugins [
				{
					src "./plugin.js"
					schema "./schema.vdl"
					outDir "./gen"
				}
			]
		}
	`)
	writeTestFile(t, filepath.Join(dir, "gen", "stale.txt"), "old")

	_, err := Run(dir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "outDir")
	require.FileExists(t, filepath.Join(dir, "gen", "stale.txt"))
	require.NoFileExists(t, filepath.Join(dir, "escape.txt"))
}

func TestRunWithRemotePluginUsesCacheAndWritesLockFile(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("VDL_HOME", filepath.Join(dir, ".vdl-home"))
	writeTestFile(t, filepath.Join(dir, "schema.vdl"), "type User {\n  name string\n}\n")

	requestCount := 0
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		require.Equal(t, "/plugins/remote.js", r.URL.Path)
		require.Equal(t, "secret-value", r.Header.Get("X-Plugin-Token"))
		_, _ = w.Write([]byte(`exports.generate = () => ({ files: [{ path: "remote.txt", content: "remote" }] })`))
	}))
	defer server.Close()

	originalClient := remoteHTTPClient
	remoteHTTPClient = server.Client()
	defer func() { remoteHTTPClient = originalClient }()

	host := strings.TrimPrefix(server.URL, "https://")
	t.Setenv("REMOTE_HEADER_NAME", "X-Plugin-Token")
	t.Setenv("REMOTE_HEADER_VALUE", "secret-value")

	writeTestFile(t, filepath.Join(dir, defaultConfigFileName), fmt.Sprintf(`
		const config = {
			version 1
			remotes [
				{
					host %q
					auth {
						header {
							nameEnv "REMOTE_HEADER_NAME"
							valueEnv "REMOTE_HEADER_VALUE"
						}
					}
				}
			]
			plugins [
				{
					src %q
					schema "./schema.vdl"
					outDir "./gen"
				}
			]
		}
	`, host, server.URL+"/plugins/remote.js"))

	fileCount, err := Run(dir)
	require.NoError(t, err)
	require.Equal(t, 1, fileCount)
	require.Equal(t, 1, requestCount)
	require.FileExists(t, filepath.Join(dir, "gen", "remote.txt"))

	lockContents, err := os.ReadFile(filepath.Join(dir, defaultLockFileName))
	require.NoError(t, err)
	require.Contains(t, string(lockContents), server.URL+"/plugins/remote.js")
	server.Close()

	fileCount, err = Run(dir)
	require.NoError(t, err)
	require.Equal(t, 1, fileCount)
	require.Equal(t, 1, requestCount)
	writeBytes, err := os.ReadFile(filepath.Join(dir, "gen", "remote.txt"))
	require.NoError(t, err)
	require.Equal(t, "remote", string(writeBytes))
}

func TestMaterializeRemoteDependencyRejectsHashChanges(t *testing.T) {
	content := []byte("console.log('v2')")
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(content)
	}))
	defer server.Close()

	originalClient := remoteHTTPClient
	remoteHTTPClient = server.Client()
	defer func() { remoteHTTPClient = originalClient }()

	cacheDir := filepath.Join(t.TempDir(), "lock")
	require.NoError(t, os.MkdirAll(cacheDir, generatedDirMode))

	_, _, err := materializeRemoteDependency(server.URL+"/plugin.js", http.Header{}, sha256Digest([]byte("console.log('v1')")), cacheDir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "may have been compromised")
}

func writeTestFile(t *testing.T, path, contents string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(filepath.Dir(path), generatedDirMode))
	require.NoError(t, os.WriteFile(path, []byte(contents), generatedFileMode))
}

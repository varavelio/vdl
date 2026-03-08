package codegen

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadRuntimeConfig(t *testing.T) {
	t.Run("loads config from directory path", func(t *testing.T) {
		dir := t.TempDir()
		writeConfigFile(t, dir, `
const config = {
  version 1
  cleanOutDir true
  plugins [
    {
      src "./plugin.js"
      schema "./schema.vdl"
      outDir "./gen"
      options {
        target "typescript"
      }
    }
  ]
}
`)

		runtimeConfig, err := loadRuntimeConfig(dir)
		require.NoError(t, err)
		require.Equal(t, int64(1), runtimeConfig.Config.Version)
		require.Equal(t, true, runtimeConfig.Config.GetCleanOutDir())
		require.Len(t, runtimeConfig.Config.GetPlugins(), 1)
		require.Equal(t, "./plugin.js", runtimeConfig.Config.GetPlugins()[0].Src)
		require.Equal(
			t,
			filepath.Join(dir, defaultConfigFileName),
			runtimeConfig.Path,
		)
	})

	t.Run("loads config from explicit file path", func(t *testing.T) {
		dir := t.TempDir()
		configPath := writeConfigFile(t, dir, `
const config = {
  version 1
}
`)

		runtimeConfig, err := loadRuntimeConfig(configPath)
		require.NoError(t, err)
		require.Equal(t, int64(1), runtimeConfig.Config.Version)
		require.Equal(t, configPath, runtimeConfig.Path)
	})

	t.Run("fails when const config is missing", func(t *testing.T) {
		dir := t.TempDir()
		writeConfigFile(t, dir, `
const other = {
  version 1
}
`)

		_, err := loadRuntimeConfig(dir)
		require.Error(t, err)
		require.Contains(t, err.Error(), `must declare const "config"`)
	})

	t.Run("fails when config file does not exist in directory", func(t *testing.T) {
		dir := t.TempDir()

		_, err := loadRuntimeConfig(dir)
		require.Error(t, err)
		require.Contains(t, err.Error(), `could not find "vdl.config.vdl"`)
	})

	t.Run("fails when explicit file name is not vdl.config.vdl", func(t *testing.T) {
		dir := t.TempDir()
		invalidPath := filepath.Join(dir, "custom.vdl")
		require.NoError(t, os.WriteFile(invalidPath, []byte("const config = { version 1 }"), 0o644))

		_, err := loadRuntimeConfig(invalidPath)
		require.Error(t, err)
		require.Contains(t, err.Error(), `expected "vdl.config.vdl"`)
	})
}

func writeConfigFile(t *testing.T, dir, contents string) string {
	t.Helper()

	configPath := filepath.Join(dir, defaultConfigFileName)
	require.NoError(t, os.WriteFile(configPath, []byte(contents), 0o644))
	return configPath
}

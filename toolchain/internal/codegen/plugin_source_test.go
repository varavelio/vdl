package codegen

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/varavelio/vdl/toolchain/internal/codegen/configtypes"
)

func TestResolveRuntimePlugins(t *testing.T) {
	t.Run("resolves github shorthand to canonical raw url", func(t *testing.T) {
		dir := t.TempDir()
		config := runtimeConfig{
			Path: filepath.Join(dir, defaultConfigFileName),
			Dir:  dir,
			Config: configtypes.VdlConfig{
				Version: 1,
				Plugins: &[]configtypes.VdlConfigPlugin{{
					Src:    "varavelio/vdl-plugin-demo@v1.2.3",
					Schema: "./schema.vdl",
					OutDir: "./gen",
				}},
			},
		}

		plugins, err := resolveRuntimePlugins(config)
		require.NoError(t, err)
		require.Len(t, plugins, 1)
		require.Equal(
			t,
			"https://raw.githubusercontent.com/varavelio/vdl-plugin-demo/v1.2.3/plugin.js",
			plugins[0].Source.CanonicalURL,
		)
	})

	t.Run("rejects github repositories without required prefix", func(t *testing.T) {
		dir := t.TempDir()
		config := runtimeConfig{
			Path: filepath.Join(dir, defaultConfigFileName),
			Dir:  dir,
			Config: configtypes.VdlConfig{
				Version: 1,
				Plugins: &[]configtypes.VdlConfigPlugin{{
					Src:    "varavelio/demo@v1.2.3",
					Schema: "./schema.vdl",
					OutDir: "./gen",
				}},
			},
		}

		_, err := resolveRuntimePlugins(config)
		require.Error(t, err)
		require.Contains(t, err.Error(), `must start with "vdl-plugin-"`)
	})

	t.Run("rejects insecure http plugins", func(t *testing.T) {
		dir := t.TempDir()
		config := runtimeConfig{
			Path: filepath.Join(dir, defaultConfigFileName),
			Dir:  dir,
			Config: configtypes.VdlConfig{
				Version: 1,
				Plugins: &[]configtypes.VdlConfigPlugin{{
					Src:    "http://example.com/plugin.js",
					Schema: "./schema.vdl",
					OutDir: "./gen",
				}},
			},
		}

		_, err := resolveRuntimePlugins(config)
		require.Error(t, err)
		require.Contains(t, err.Error(), "only HTTPS URLs are allowed")
	})

	t.Run("allows insecure http plugins when enabled by environment", func(t *testing.T) {
		dir := t.TempDir()
		t.Setenv("VDL_INSECURE_ALLOW_HTTP", "true")
		config := runtimeConfig{
			Path: filepath.Join(dir, defaultConfigFileName),
			Dir:  dir,
			Config: configtypes.VdlConfig{
				Version: 1,
				Plugins: &[]configtypes.VdlConfigPlugin{{
					Src:    "http://example.com/plugin.js",
					Schema: "./schema.vdl",
					OutDir: "./gen",
				}},
			},
		}

		plugins, err := resolveRuntimePlugins(config)
		require.NoError(t, err)
		require.Len(t, plugins, 1)
		require.Equal(t, "http://example.com/plugin.js", plugins[0].Source.CanonicalURL)
	})

	t.Run("applies the most specific remote auth headers", func(t *testing.T) {
		dir := t.TempDir()
		t.Setenv("REMOTE_HEADER_NAME", "X-Token")
		t.Setenv("REMOTE_HEADER_VALUE", "specific-value")
		t.Setenv("REMOTE_FALLBACK_NAME", "X-Token")
		t.Setenv("REMOTE_FALLBACK_VALUE", "fallback-value")

		config := runtimeConfig{
			Path: filepath.Join(dir, defaultConfigFileName),
			Dir:  dir,
			Config: configtypes.VdlConfig{
				Version: 1,
				Remotes: &[]configtypes.VdlConfigRemote{
					{
						Host: "example.com",
						Auth: &configtypes.VdlConfigRemoteAuth{Header: &configtypes.VdlConfigRemoteAuthHeader{
							NameEnv:  "REMOTE_FALLBACK_NAME",
							ValueEnv: "REMOTE_FALLBACK_VALUE",
						}},
					},
					{
						Host: "example.com/plugins",
						Auth: &configtypes.VdlConfigRemoteAuth{Header: &configtypes.VdlConfigRemoteAuthHeader{
							NameEnv:  "REMOTE_HEADER_NAME",
							ValueEnv: "REMOTE_HEADER_VALUE",
						}},
					},
				},
				Plugins: &[]configtypes.VdlConfigPlugin{{
					Src:    "https://example.com/plugins/demo.js",
					Schema: "./schema.vdl",
					OutDir: "./gen",
				}},
			},
		}

		plugins, err := resolveRuntimePlugins(config)
		require.NoError(t, err)
		require.Len(t, plugins, 1)
		require.Equal(t, "specific-value", plugins[0].Source.Headers.Get("X-Token"))
	})
}

package plugin

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/varavelio/vdl/toolchain/internal/codegen/config"
	"github.com/varavelio/vdl/toolchain/internal/core/ir"
)

func TestGenerator(t *testing.T) {
	// Create a temporary directory for the plugin script
	tmpDir := t.TempDir()
	pluginScript := filepath.Join(tmpDir, "plugin.py")

	// Write the plugin script
	scriptContent := `
import sys
import json

# Read input from stdin
input_data = json.load(sys.stdin)

# Check if options are passed
prefix = input_data.get("options", {}).get("prefix", "DEFAULT")

# Generate output
output = {
    "files": [
        {
            "path": "test.txt",
            "content": f"{prefix}: Hello World"
        }
    ]
}

# Write output to stdout
json.dump(output, sys.stdout)
`
	err := os.WriteFile(pluginScript, []byte(scriptContent), 0755)
	require.NoError(t, err)

	// Create a dummy schema
	schema := &ir.Schema{
		RPCs: []ir.RPC{},
	}

	// Create configuration
	cfg := &config.PluginConfig{
		Command: []string{"python3", pluginScript},
		Options: map[string]any{
			"prefix": "TEST",
		},
		CommonConfig: config.CommonConfig{
			Output: tmpDir, // Not used by Generator directly but passed in context
		},
	}

	// Run generator
	gen := New(cfg)
	files, err := gen.Generate(context.Background(), schema)
	require.NoError(t, err)

	// Verify output
	require.Len(t, files, 1)
	assert.Equal(t, "test.txt", files[0].Path)
	assert.Equal(t, "TEST: Hello World", string(files[0].Content))
}

package codegen

import (
	"fmt"

	"github.com/varavelio/vdl/toolchain/internal/codegen/configtypes"
)

// Run runs the code generator and returns the total number of files generated and an error if one occurred.
func Run(configPath string) (int, error) {
	runtimeConfig, err := loadRuntimeConfig(configPath)
	if err != nil {
		return 0, fmt.Errorf("failed to load config: %w", err)
	}

	return runWithConfig(runtimeConfig)
}

type runtimeConfig struct {
	Path   string
	Config configtypes.VdlConfig
}

func runWithConfig(_ runtimeConfig) (int, error) {
	return 0, nil
}

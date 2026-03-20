package codegen

import (
	"fmt"

	"github.com/varavelio/vdl/toolchain/internal/codegen/configtypes"
)

// Run executes the full code generation pipeline and returns the number of
// written files.
func Run(configPath string) (int, error) {
	runtimeConfig, err := loadRuntimeConfig(configPath)
	if err != nil {
		return 0, fmt.Errorf("failed to load config: %w", err)
	}

	return runWithConfig(runtimeConfig)
}

type runtimeConfig struct {
	Path     string
	Dir      string
	LockPath string
	Config   configtypes.VdlConfig
}

// runWithConfig orchestrates the generation pipeline after the config file has
// already been loaded and normalized.
func runWithConfig(config runtimeConfig) (int, error) {
	if err := runPreGenerateHooks(config); err != nil {
		return 0, err
	}

	plugins, err := resolveRuntimePlugins(config)
	if err != nil {
		return 0, err
	}

	lockFile, err := loadLockFile(config.LockPath)
	if err != nil {
		return 0, err
	}

	if err := materializeRemotePlugins(plugins, &lockFile); err != nil {
		return 0, err
	}

	preparedPlugins, err := preparePlugins(plugins)
	if err != nil {
		return 0, err
	}

	results, err := executePlugins(preparedPlugins)
	if err != nil {
		return 0, err
	}
	applyGeneratedFileHeaders(results)

	plan, err := planOutputWrites(results)
	if err != nil {
		return 0, err
	}

	if err := writeLockFile(config.LockPath, lockFile); err != nil {
		return 0, err
	}

	if err := applyOutputWrites(config, plan); err != nil {
		return 0, err
	}

	runPostGenerateHooks(config)

	return len(plan.Writes), nil
}

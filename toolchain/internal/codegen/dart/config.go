package dart

import (
	"fmt"
)

// Config is the configuration for the Dart code generator.
type Config struct {
	// OutputDir is the directory to output the generated Dart package to.
	OutputDir string `toml:"output_dir"`
	// PackageName is the name of the Dart package.
	PackageName string `toml:"package_name"`
}

func (c Config) Validate() error {
	if c.OutputDir == "" {
		return fmt.Errorf("\"output_dir\" is required")
	}
	if c.PackageName == "" {
		return fmt.Errorf("\"package_name\" is required")
	}
	return nil
}

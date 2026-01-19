package playground

import "fmt"

// Header represents a key-value pair for HTTP headers.
type Header struct {
	Key   string `toml:"key"`
	Value string `toml:"value"`
}

// Config is the configuration for the playground generator.
type Config struct {
	// OutputDir is the directory to output the generated playground to.
	// This becomes the base path for all generated files.
	OutputDir string `toml:"output_dir"`

	// DefaultBaseURL is the default VDL base URL to use.
	DefaultBaseURL string `toml:"default_base_url"`

	// DefaultHeaders is the default headers to use.
	DefaultHeaders []Header `toml:"default_headers"`

	// FormattedSchema is the VDL schema source code, pre-formatted.
	// This is passed by the caller since the IR doesn't contain the original source.
	// If empty, the playground will not include a schema.vdl file.
	FormattedSchema string `toml:"-"`
}

// Validate validates the configuration.
func (c Config) Validate() error {
	if c.OutputDir == "" {
		return fmt.Errorf(`"output_dir" is required`)
	}
	return nil
}

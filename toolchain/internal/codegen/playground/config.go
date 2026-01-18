package playground

import "fmt"

type Header struct {
	Key   string `toml:"key"`
	Value string `toml:"value"`
}

// Config is the configuration for the playground generator.
type Config struct {
	// OutputDir is the directory to output the generated playground to.
	OutputDir string `toml:"output_dir"`

	// DefaultBaseURL is the default urpc base URL to use.
	DefaultBaseURL string `toml:"default_base_url"`

	// DefaultHeaders is the default headers to use.
	DefaultHeaders []Header `toml:"default_headers"`
}

func (c Config) Validate() error {
	if c.OutputDir == "" {
		return fmt.Errorf(`"output_dir" is required`)
	}
	return nil
}

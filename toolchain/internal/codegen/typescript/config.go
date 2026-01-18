package typescript

import (
	"fmt"
	"strings"
)

// Config is the configuration for the TypeScript code generator.
type Config struct {
	// OutputFile is the file to output the generated code to.
	OutputFile string `toml:"output_file"`
	// IncludeServer enables server code generation.
	IncludeServer bool `toml:"include_server"`
	// IncludeClient enables client code generation.
	IncludeClient bool `toml:"include_client"`
}

func (c Config) Validate() error {
	if c.OutputFile == "" {
		return fmt.Errorf("output_file is required")
	}
	if !strings.HasSuffix(c.OutputFile, ".ts") {
		return fmt.Errorf(`"output_file" must end with ".ts"`)
	}
	return nil
}

package golang

import (
	"fmt"
	"strings"
)

// Config is the configuration for the Go code generator.
type Config struct {
	// OutputFile is the file to output the generated code to.
	OutputFile string `toml:"output_file"`
	// PackageName is the name of the package to generate the code in.
	PackageName string `toml:"package_name"`
	// IncludeServer enables server code generation.
	IncludeServer bool `toml:"include_server"`
	// IncludeClient enables client code generation.
	IncludeClient bool `toml:"include_client"`
}

func (c Config) Validate() error {
	if c.OutputFile == "" {
		return fmt.Errorf(`"output_file" is required`)
	}
	if !strings.HasSuffix(c.OutputFile, ".go") {
		return fmt.Errorf(`"output_file" must end with ".go"`)
	}
	if c.PackageName == "" {
		return fmt.Errorf(`"package_name" is required`)
	}
	return nil
}

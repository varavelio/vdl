package openapi

import (
	"fmt"
	"strings"
)

// Config is the configuration for the OpenAPI generator.
type Config struct {
	// OutputFile is the file to output the generated code to.
	OutputFile string `toml:"output_file"`
	// Title is the title of the OpenAPI spec.
	Title string `toml:"title"`
	// Description is the description of the OpenAPI spec.
	Description string `toml:"description"`
	// Version is the version of the OpenAPI spec.
	Version string `toml:"version"`
	// BaseURL is the base URL to use for the OpenAPI spec.
	BaseURL string `toml:"base_url"`
	// ContactName is the name of the contact person for the OpenAPI spec.
	ContactName string `toml:"contact_name"`
	// ContactEmail is the email of the contact person for the OpenAPI spec.
	ContactEmail string `toml:"contact_email"`
	// LicenseName is the name of the license for the OpenAPI spec.
	LicenseName string `toml:"license_name"`
}

func (c Config) Validate() error {
	if c.OutputFile != "" &&
		!strings.HasSuffix(c.OutputFile, ".json") &&
		!strings.HasSuffix(c.OutputFile, ".yaml") &&
		!strings.HasSuffix(c.OutputFile, ".yml") {
		return fmt.Errorf(`"output_file" must end with ".json", ".yaml" or ".yml"`)
	}
	return nil
}

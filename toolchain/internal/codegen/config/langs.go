package config

// GoOptions contains configuration for the Go target.
type GoOptions struct {
	BaseCodeOptions `yaml:",inline" json:",inline"`
	Package         string `yaml:"package" json:"package" jsonschema:"required,description=The Go package name to use in generated files."`
}

// TypeScriptOptions contains configuration for the TypeScript target.
type TypeScriptOptions struct {
	BaseCodeOptions `yaml:",inline" json:",inline"`
}

// DartOptions contains configuration for the Dart target.
type DartOptions struct {
	BaseCodeOptions `yaml:",inline" json:",inline"`
	Package         string `yaml:"package" json:"package" jsonschema:"required,description=The name of the Dart package."`
}

// OpenAPIOptions contains configuration for the OpenAPI target.
type OpenAPIOptions struct {
	Filename     string `yaml:"filename" json:"filename,omitempty" jsonschema:"default=openapi.json,description=The name of the output file."`
	Title        string `yaml:"title" json:"title" jsonschema:"required"`
	Version      string `yaml:"version" json:"version" jsonschema:"required"`
	Description  string `yaml:"description" json:"description,omitempty"`
	BaseURL      string `yaml:"base_url" json:"base_url,omitempty"`
	ContactName  string `yaml:"contact_name" json:"contact_name,omitempty"`
	ContactEmail string `yaml:"contact_email" json:"contact_email,omitempty"`
	LicenseName  string `yaml:"license_name" json:"license_name,omitempty"`
}

// PlaygroundOptions contains configuration for the Playground target.
type PlaygroundOptions struct {
	DefaultBaseURL string `yaml:"default_base_url" json:"default_base_url,omitempty"`
	DefaultHeaders []struct {
		Key   string `yaml:"key" json:"key"`
		Value string `yaml:"value" json:"value"`
	} `yaml:"default_headers" json:"default_headers,omitempty"`
}

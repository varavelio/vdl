package config

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/kaptinlin/jsonschema"
	"gopkg.in/yaml.v3"
)

//go:embed config.schema.json
var schemaJSON []byte

type VDLConfig struct {
	Version int            `yaml:"version" json:"version" jsonschema:"required"`
	Schema  string         `yaml:"schema" json:"schema,omitempty" jsonschema:"description=Path to the default global VDL schema file."`
	Targets []TargetConfig `yaml:"targets" json:"targets" jsonschema:"required,minItems=1"`
}

// TargetConfig represents a configuration for a specific generation target.
// Only one of the fields must be set.
type TargetConfig struct {
	Go         *GoConfig         `yaml:"go,omitempty" json:"go,omitempty"`
	TypeScript *TypeScriptConfig `yaml:"typescript,omitempty" json:"typescript,omitempty"`
	Dart       *DartConfig       `yaml:"dart,omitempty" json:"dart,omitempty"`
	OpenAPI    *OpenAPIConfig    `yaml:"openapi,omitempty" json:"openapi,omitempty"`
	Playground *PlaygroundConfig `yaml:"playground,omitempty" json:"playground,omitempty"`
}

// CommonConfig defines the shared configuration options available to all generation targets.
type CommonConfig struct {
	Output string `yaml:"output" json:"output" jsonschema:"required,minLength=1,description=The output directory where the generated files will be placed."`
	Clean  bool   `yaml:"clean,omitempty" json:"clean,omitempty" jsonschema:"default=false,description=If true empties the output directory before generation."`
	Schema string `yaml:"schema,omitempty" json:"schema,omitempty" jsonschema:"description=Optional override for the VDL schema file specific to this target."`
}

// PatternsConfig defines configuration for generating patterns.
type PatternsConfig struct {
	GenPatterns *bool `yaml:"gen_patterns" json:"gen_patterns,omitempty" jsonschema:"default=true,description=Generate helper functions for patterns."`
}

// ShouldGenPatterns returns true if patterns should be generated (default: true).
func (b PatternsConfig) ShouldGenPatterns() bool {
	if b.GenPatterns == nil {
		return true
	}
	return *b.GenPatterns
}

// ConstsConfig defines configuration for generating constants.
type ConstsConfig struct {
	GenConsts *bool `yaml:"gen_consts" json:"gen_consts,omitempty" jsonschema:"default=true,description=Generate constant definitions."`
}

// ShouldGenConsts returns true if constants should be generated (default: true).
func (b ConstsConfig) ShouldGenConsts() bool {
	if b.GenConsts == nil {
		return true
	}
	return *b.GenConsts
}

// ClientConfig defines configuration for generating RPCs clients.
type ClientConfig struct {
	GenClient bool `yaml:"gen_client" json:"gen_client,omitempty" jsonschema:"default=false,description=Generate RPC client code."`
}

// ServerConfig defines configuration for generating RPCs servers.
type ServerConfig struct {
	GenServer bool `yaml:"gen_server" json:"gen_server,omitempty" jsonschema:"default=false,description=Generate RPC server interfaces and handlers."`
}

// GoConfig contains configuration for the Go target.
type GoConfig struct {
	CommonConfig   `yaml:",inline" json:",inline"`
	PatternsConfig `yaml:",inline" json:",inline"`
	ConstsConfig   `yaml:",inline" json:",inline"`
	ClientConfig   `yaml:",inline" json:",inline"`
	ServerConfig   `yaml:",inline" json:",inline"`
	Package        string `yaml:"package" json:"package" jsonschema:"required,description=The Go package name to use in generated files."`
}

// TypeScriptConfig contains configuration for the TypeScript target.
type TypeScriptConfig struct {
	CommonConfig   `yaml:",inline" json:",inline"`
	PatternsConfig `yaml:",inline" json:",inline"`
	ConstsConfig   `yaml:",inline" json:",inline"`
	ClientConfig   `yaml:",inline" json:",inline"`
	ServerConfig   `yaml:",inline" json:",inline"`
}

// DartConfig contains configuration for the Dart target.
type DartConfig struct {
	CommonConfig   `yaml:",inline" json:",inline"`
	PatternsConfig `yaml:",inline" json:",inline"`
	ConstsConfig   `yaml:",inline" json:",inline"`
	ClientConfig   `yaml:",inline" json:",inline"`
	Package        string `yaml:"package" json:"package" jsonschema:"required,description=The name of the Dart package."`
}

// OpenAPIConfig contains configuration for the OpenAPI target.
type OpenAPIConfig struct {
	CommonConfig `yaml:",inline" json:",inline"`
	Filename     string `yaml:"filename" json:"filename,omitempty" jsonschema:"default=openapi.yaml,description=The name of the output file (can be .yml\\, .yaml or .json)."`
	Title        string `yaml:"title" json:"title" jsonschema:"required"`
	Version      string `yaml:"version" json:"version" jsonschema:"required"`
	Description  string `yaml:"description" json:"description,omitempty"`
	BaseURL      string `yaml:"base_url" json:"base_url,omitempty"`
	ContactName  string `yaml:"contact_name" json:"contact_name,omitempty"`
	ContactEmail string `yaml:"contact_email" json:"contact_email,omitempty"`
	LicenseName  string `yaml:"license_name" json:"license_name,omitempty"`
}

// PlaygroundConfig contains configuration for the Playground target.
type PlaygroundConfig struct {
	CommonConfig   `yaml:",inline" json:",inline"`
	DefaultBaseURL string `yaml:"default_base_url" json:"default_base_url,omitempty"`
	DefaultHeaders []struct {
		Key   string `yaml:"key" json:"key"`
		Value string `yaml:"value" json:"value"`
	} `yaml:"default_headers" json:"default_headers,omitempty"`
}

func LoadConfig(path string) (*VDLConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
	return ParseConfig(data)
}

// ParseConfig parses and validates VDL configuration from bytes.
func ParseConfig(data []byte) (*VDLConfig, error) {
	return Validate(data)
}

// Validate parses and validates VDL configuration from bytes.
// It combines schema validation, unmarshaling, and logical validation.
func Validate(data []byte) (*VDLConfig, error) {
	// 1. Parse YAML into Node (Intermediate representation)
	// This avoids parsing the YAML text twice.
	var node yaml.Node
	if err := yaml.Unmarshal(data, &node); err != nil {
		return nil, fmt.Errorf("failed to parse yaml: %w", err)
	}

	// 2. Validate against JSON Schema
	// We first decode the node into a generic map/interface to convert to JSON,
	// because jsonschema requires JSON data (or equivalent interface).
	var raw any
	if err := node.Decode(&raw); err != nil {
		return nil, fmt.Errorf("failed to decode yaml for validation: %w", err)
	}

	jsonData, err := json.Marshal(raw)
	if err != nil {
		return nil, fmt.Errorf("failed to convert yaml to json: %w", err)
	}

	compiler := jsonschema.NewCompiler()
	schema, err := compiler.Compile(schemaJSON)
	if err != nil {
		return nil, fmt.Errorf("internal error: invalid embedded schema: %w", err)
	}

	if result := schema.Validate(jsonData); !result.IsValid() {
		var parts []string
		for path, err := range result.Errors {
			parts = append(parts, fmt.Sprintf("%s: %s", path, err.Message))
		}
		return nil, fmt.Errorf("%s", strings.Join(parts, "; "))
	}

	// 3. Decode into VDLConfig Struct
	var cfg VDLConfig
	if err := node.Decode(&cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// 4. Logical validation
	if cfg.Version != 1 {
		return nil, fmt.Errorf("unsupported version: %d", cfg.Version)
	}

	for i := range cfg.Targets {
		t := &cfg.Targets[i]
		if err := t.validateAndSetDefaults(cfg.Schema); err != nil {
			return nil, fmt.Errorf("target #%d: %w", i, err)
		}
	}

	return &cfg, nil
}

func (t *TargetConfig) validateAndSetDefaults(globalSchema string) error {
	count := 0
	var schema *string

	if t.Go != nil {
		count++
		schema = &t.Go.Schema
	}
	if t.TypeScript != nil {
		count++
		schema = &t.TypeScript.Schema
	}
	if t.Dart != nil {
		count++
		schema = &t.Dart.Schema
	}
	if t.OpenAPI != nil {
		count++
		schema = &t.OpenAPI.Schema
		if t.OpenAPI.Filename == "" {
			t.OpenAPI.Filename = "openapi.yaml"
		}
	}
	if t.Playground != nil {
		count++
		schema = &t.Playground.Schema
	}

	if count == 0 {
		return fmt.Errorf("no language configuration found for the target")
	}
	if count > 1 {
		return fmt.Errorf("multiple language configurations found in the same target block")
	}

	// Apply global schema if local one is missing
	if schema != nil {
		if *schema == "" {
			*schema = globalSchema
		}
		// Check again if it's still empty
		if *schema == "" {
			return fmt.Errorf("no schema defined for the target (must be defined globally or locally)")
		}
	}

	return nil
}

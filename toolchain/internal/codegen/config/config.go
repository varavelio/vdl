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

const (
	TargetGo         = "go"
	TargetTypeScript = "typescript"
	TargetDart       = "dart"
	TargetOpenAPI    = "openapi"
	TargetPlayground = "playground"
)

type VDLConfig struct {
	Version int            `yaml:"version" json:"version" jsonschema:"required,const=1"`
	Schema  string         `yaml:"schema" json:"schema,omitempty" jsonschema:"description=Path to the default global VDL schema file."`
	Targets []TargetConfig `yaml:"targets" json:"targets" jsonschema:"required,minItems=1"`
}

type TargetConfig struct {
	Target        string    `yaml:"target" json:"target" jsonschema:"required,enum=go,enum=typescript,enum=dart,enum=openapi,enum=playground"`
	Output        string    `yaml:"output" json:"output" jsonschema:"required,description=The output directory where the generated files will be placed."`
	Clean         bool      `yaml:"clean" json:"clean,omitempty" jsonschema:"default=false,description=If true empties the output directory before generation."`
	Schema        string    `yaml:"schema" json:"schema,omitempty" jsonschema:"description=Optional override for the VDL schema file specific to this target."`
	Options       yaml.Node `yaml:"options" json:"-"`
	ParsedOptions any       `yaml:"-" json:"-"`
}

// BaseCodeOptions defines standard options shared across code generators.
type BaseCodeOptions struct {
	GenClient   bool  `yaml:"gen_client" json:"gen_client,omitempty" jsonschema:"default=false,description=Generate RPC client code."`
	GenServer   bool  `yaml:"gen_server" json:"gen_server,omitempty" jsonschema:"default=false,description=Generate RPC server interfaces and handlers."`
	GenPatterns *bool `yaml:"gen_patterns" json:"gen_patterns,omitempty" jsonschema:"default=true,description=Generate helper functions for patterns."`
	GenConsts   *bool `yaml:"gen_consts" json:"gen_consts,omitempty" jsonschema:"default=true,description=Generate constant definitions."`
}

// ShouldGenPatterns returns true if patterns should be generated (default: true).
func (b BaseCodeOptions) ShouldGenPatterns() bool {
	if b.GenPatterns == nil {
		return true
	}
	return *b.GenPatterns
}

// ShouldGenConsts returns true if constants should be generated (default: true).
func (b BaseCodeOptions) ShouldGenConsts() bool {
	if b.GenConsts == nil {
		return true
	}
	return *b.GenConsts
}

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
		if t.Schema == "" {
			t.Schema = cfg.Schema
		}
		if t.Schema == "" {
			return nil, fmt.Errorf("target #%d (%s): no schema defined (must be defined globally or locally)", i, t.Target)
		}
		if err := t.decodeOptions(); err != nil {
			return nil, fmt.Errorf("target #%d (%s): %w", i, t.Target, err)
		}
	}

	return &cfg, nil
}

func (t *TargetConfig) decodeOptions() error {
	var opts any

	switch t.Target {
	case TargetGo:
		opts = &GoOptions{}
	case TargetTypeScript:
		opts = &TypeScriptOptions{}
	case TargetDart:
		opts = &DartOptions{}
	case TargetOpenAPI:
		o := &OpenAPIOptions{}
		if err := t.Options.Decode(o); err != nil {
			return err
		}
		if o.Filename == "" {
			o.Filename = "openapi.yaml"
		}
		t.ParsedOptions = o
		return nil
	case TargetPlayground:
		opts = &PlaygroundOptions{}
	default:
		return fmt.Errorf("unknown target: %q", t.Target)
	}

	if err := t.Options.Decode(opts); err != nil {
		return err
	}
	t.ParsedOptions = opts
	return nil
}

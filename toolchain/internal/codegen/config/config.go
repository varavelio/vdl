package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

const (
	TargetGo         = "go"
	TargetTypeScript = "typescript"
	TargetDart       = "dart"
	TargetOpenAPI    = "openapi"
	TargetPlayground = "playground"
)

type VDLConfig struct {
	Version int            `yaml:"version" json:"version" jsonschema:"required,const=1"`
	Schema  string         `yaml:"schema" json:"schema" jsonschema:"required,description=Path to the default global VDL schema file."`
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

func LoadConfig(path string) (*VDLConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg VDLConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (c *VDLConfig) validate() error {
	if c.Version != 1 {
		return fmt.Errorf("unsupported version: %d", c.Version)
	}

	for i := range c.Targets {
		t := &c.Targets[i]
		if t.Schema == "" {
			t.Schema = c.Schema
		}
		if t.Schema == "" {
			return fmt.Errorf("target #%d (%s): no schema defined", i, t.Target)
		}
		if err := t.decodeOptions(); err != nil {
			return fmt.Errorf("target #%d (%s): %w", i, t.Target, err)
		}
	}

	return nil
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

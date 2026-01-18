package codegen

import (
	"fmt"

	"github.com/BurntSushi/toml"
	"github.com/uforg/uforpc/urpc/internal/codegen/dart"
	"github.com/uforg/uforpc/urpc/internal/codegen/golang"
	"github.com/uforg/uforpc/urpc/internal/codegen/openapi"
	"github.com/uforg/uforpc/urpc/internal/codegen/playground"
	"github.com/uforg/uforpc/urpc/internal/codegen/typescript"
)

// Config is the configuration for the code generator.
type Config struct {
	Version int    `toml:"version"`
	Schema  string `toml:"schema"`

	OpenAPI    openapi.Config     `toml:"openapi"`
	Playground *playground.Config `toml:"playground"`

	// New split generators
	GolangServer     *golang.Config     `toml:"golang-server"`
	GolangClient     *golang.Config     `toml:"golang-client"`
	TypescriptClient *typescript.Config `toml:"typescript-client"`
	DartClient       *dart.Config       `toml:"dart-client"`
}

func (c *Config) HasOpenAPI() bool {
	return c.OpenAPI.OutputFile != ""
}

func (c *Config) HasPlayground() bool {
	return c.Playground != nil
}

func (c *Config) HasGolangServer() bool {
	return c.GolangServer != nil
}

func (c *Config) HasGolangClient() bool {
	return c.GolangClient != nil
}

func (c *Config) HasTypescriptClient() bool {
	return c.TypescriptClient != nil
}

func (c *Config) HasDartClient() bool {
	return c.DartClient != nil
}

func (c *Config) Unmarshal(data []byte) error {
	if err := toml.Unmarshal(data, c); err != nil {
		return fmt.Errorf("failed to unmarshal TOML config: %w", err)
	}
	return nil
}

func (c *Config) Validate() error {
	if c.Version == 0 {
		return fmt.Errorf(`"version" is required`)
	}

	if c.Version != 1 {
		return fmt.Errorf("unsupported version: %d", c.Version)
	}

	if c.Schema == "" {
		return fmt.Errorf(`"schema" is required`)
	}

	if err := c.OpenAPI.Validate(); err != nil {
		return fmt.Errorf("openapi config is invalid: %w", err)
	}

	if c.Playground != nil {
		if err := c.Playground.Validate(); err != nil {
			return fmt.Errorf("playground config is invalid: %w", err)
		}
	}

	if c.GolangServer != nil {
		if err := c.GolangServer.Validate(); err != nil {
			return fmt.Errorf("golang-server config is invalid: %w", err)
		}
	}

	if c.GolangClient != nil {
		if err := c.GolangClient.Validate(); err != nil {
			return fmt.Errorf("golang-client config is invalid: %w", err)
		}
	}

	if c.TypescriptClient != nil {
		if err := c.TypescriptClient.Validate(); err != nil {
			return fmt.Errorf("typescript-client config is invalid: %w", err)
		}
	}

	if c.DartClient != nil {
		if err := c.DartClient.Validate(); err != nil {
			return fmt.Errorf("dart-client config is invalid: %w", err)
		}
	}

	return nil
}

// UnmarshalAndValidate unmarshals and validates a TOML config
func (c *Config) UnmarshalAndValidate(configBytes []byte) error {
	if err := c.Unmarshal(configBytes); err != nil {
		return err
	}

	if err := c.Validate(); err != nil {
		return err
	}

	return nil
}

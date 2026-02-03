package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"

	"github.com/varavelio/vdl/toolchain/internal/codegen/config/configtypes"
	"gopkg.in/yaml.v3"
)

// LoadConfig loads and validates VDL configuration from a file path.
func LoadConfig(path string) (*configtypes.VdlConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
	return ParseConfig(data)
}

// ParseConfig parses and validates VDL configuration from bytes.
func ParseConfig(data []byte) (*configtypes.VdlConfig, error) {
	// Parse YAML into intermediate representation
	var raw any
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse yaml: %w", err)
	}

	// Convert to JSON for strict unmarshaling
	jsonData, err := json.Marshal(raw)
	if err != nil {
		return nil, fmt.Errorf("failed to convert yaml to json: %w", err)
	}

	// Unmarshal with strict mode to catch unknown fields
	var cfg configtypes.VdlConfig
	decoder := json.NewDecoder(bytes.NewReader(jsonData))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Validate
	if err := validate(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// Validate is an alias for ParseConfig for backward compatibility.
func Validate(data []byte) (*configtypes.VdlConfig, error) {
	return ParseConfig(data)
}

// validate performs logical validation on the parsed config.
func validate(cfg *configtypes.VdlConfig) error {
	if cfg.Version != 1 {
		return fmt.Errorf("unsupported version: %d", cfg.Version)
	}

	if len(cfg.Targets) == 0 {
		return fmt.Errorf("targets array must not be empty")
	}

	globalSchema := ""
	if cfg.Schema != nil {
		globalSchema = *cfg.Schema
	}

	for i := range cfg.Targets {
		if err := validateTarget(&cfg.Targets[i], globalSchema); err != nil {
			return fmt.Errorf("target #%d: %w", i, err)
		}
	}

	return nil
}

// validateTarget validates a single target and sets defaults.
func validateTarget(t *configtypes.TargetConfig, globalSchema string) error {
	count := 0
	var schemaPtr **string

	if t.Go != nil {
		count++
		schemaPtr = &t.Go.Schema
		if err := validateGo(t.Go); err != nil {
			return err
		}
	}
	if t.Typescript != nil {
		count++
		schemaPtr = &t.Typescript.Schema
		if err := validateCommon(t.Typescript.Output, "typescript"); err != nil {
			return err
		}
	}
	if t.Dart != nil {
		count++
		schemaPtr = &t.Dart.Schema
		if err := validateCommon(t.Dart.Output, "dart"); err != nil {
			return err
		}
	}
	if t.Python != nil {
		count++
		schemaPtr = &t.Python.Schema
		if err := validateCommon(t.Python.Output, "python"); err != nil {
			return err
		}
	}
	if t.Jsonschema != nil {
		count++
		schemaPtr = &t.Jsonschema.Schema
		if err := validateCommon(t.Jsonschema.Output, "jsonschema"); err != nil {
			return err
		}
		// Set default filename
		if t.Jsonschema.Filename == nil {
			t.Jsonschema.Filename = ptr("schema.json")
		}
	}
	if t.Openapi != nil {
		count++
		schemaPtr = &t.Openapi.Schema
		if err := validateOpenAPI(t.Openapi); err != nil {
			return err
		}
		// Set default filename
		if t.Openapi.Filename == nil {
			t.Openapi.Filename = ptr("openapi.yaml")
		}
	}
	if t.Playground != nil {
		count++
		schemaPtr = &t.Playground.Schema
		if err := validatePlayground(t.Playground); err != nil {
			return err
		}
	}
	if t.Plugin != nil {
		count++
		schemaPtr = &t.Plugin.Schema
		if err := validatePlugin(t.Plugin); err != nil {
			return err
		}
	}
	if t.Ir != nil {
		count++
		schemaPtr = &t.Ir.Schema
		if err := validateIr(t.Ir); err != nil {
			return err
		}
		// Set default filename
		if t.Ir.Filename == nil {
			t.Ir.Filename = ptr("ir.json")
		}
	}
	if t.Vdl != nil {
		count++
		schemaPtr = &t.Vdl.Schema
		if err := validateVdl(t.Vdl); err != nil {
			return err
		}
		// Set default filename
		if t.Vdl.Filename == nil {
			t.Vdl.Filename = ptr("schema.vdl")
		}
	}

	if count == 0 {
		return fmt.Errorf("no target configuration found")
	}
	if count > 1 {
		return fmt.Errorf("multiple target configurations found in the same target block (only one allowed per entry)")
	}

	// Apply global schema if local one is missing
	if schemaPtr != nil && (*schemaPtr == nil || **schemaPtr == "") {
		if globalSchema == "" {
			return fmt.Errorf("no schema defined (must be defined globally or per-target)")
		}
		*schemaPtr = &globalSchema
	}

	return nil
}

func validateCommon(output, targetName string) error {
	if output == "" {
		return fmt.Errorf("field 'output' is required for %s target", targetName)
	}
	return nil
}

func validateGo(cfg *configtypes.GoTargetConfig) error {
	if err := validateCommon(cfg.Output, "go"); err != nil {
		return err
	}
	if cfg.Package == "" {
		return fmt.Errorf("field 'package' is required for go target")
	}
	return nil
}

func validateOpenAPI(cfg *configtypes.OpenApiTargetConfig) error {
	if err := validateCommon(cfg.Output, "openapi"); err != nil {
		return err
	}
	if cfg.Title == "" {
		return fmt.Errorf("field 'title' is required for openapi target")
	}
	if cfg.Version == "" {
		return fmt.Errorf("field 'version' is required for openapi target")
	}
	return nil
}

func validatePlayground(cfg *configtypes.PlaygroundTargetConfig) error {
	if err := validateCommon(cfg.Output, "playground"); err != nil {
		return err
	}
	if cfg.DefaultHeaders != nil {
		for i, h := range *cfg.DefaultHeaders {
			if h.Key == "" {
				return fmt.Errorf("defaultHeaders[%d]: field 'key' is required", i)
			}
			if h.Value == "" {
				return fmt.Errorf("defaultHeaders[%d]: field 'value' is required", i)
			}
		}
	}
	return nil
}

func validatePlugin(cfg *configtypes.PluginTargetConfig) error {
	if err := validateCommon(cfg.Output, "plugin"); err != nil {
		return err
	}
	if len(cfg.Command) == 0 {
		return fmt.Errorf("field 'command' is required for plugin target")
	}
	return nil
}

func validateIr(cfg *configtypes.IrTargetConfig) error {
	if err := validateCommon(cfg.Output, "ir"); err != nil {
		return err
	}
	return nil
}

func validateVdl(cfg *configtypes.VdlTargetConfig) error {
	if err := validateCommon(cfg.Output, "vdl"); err != nil {
		return err
	}
	return nil
}

func ptr(s string) *string {
	return &s
}

// Helper functions to work with optional pointer fields.

// ShouldGenPatterns returns true if patterns should be generated (default: true).
func ShouldGenPatterns(genPatterns *bool) bool {
	return genPatterns == nil || *genPatterns
}

// ShouldGenConsts returns true if constants should be generated (default: true).
func ShouldGenConsts(genConsts *bool) bool {
	return genConsts == nil || *genConsts
}

// ShouldGenClient returns true if client code should be generated (default: false).
func ShouldGenClient(genClient *bool) bool {
	return genClient != nil && *genClient
}

// ShouldGenServer returns true if server code should be generated (default: false).
func ShouldGenServer(genServer *bool) bool {
	return genServer != nil && *genServer
}

// ShouldClean returns true if the output directory should be cleaned (default: false).
func ShouldClean(clean *bool) bool {
	return clean != nil && *clean
}

// GetSchema returns the schema path or empty string if not set.
func GetSchema(schema *string) string {
	if schema == nil {
		return ""
	}
	return *schema
}

// GetFilename returns the filename or empty string if not set.
func GetFilename(filename *string) string {
	if filename == nil {
		return ""
	}
	return *filename
}

// GetImportExtension returns the import extension or the default (none) if not set.
func GetImportExtension(ext *configtypes.TypescriptImportExtension) configtypes.TypescriptImportExtension {
	if ext == nil {
		return configtypes.TypescriptImportExtensionNone
	}
	return *ext
}

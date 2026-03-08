package codegen

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/varavelio/vdl/toolchain/internal/codegen/configtypes"
	"github.com/varavelio/vdl/toolchain/internal/core/analysis"
	"github.com/varavelio/vdl/toolchain/internal/core/ir"
	"github.com/varavelio/vdl/toolchain/internal/core/ir/irtypes"
	"github.com/varavelio/vdl/toolchain/internal/core/vfs"
)

const (
	defaultConfigFileName = "vdl.config.vdl"
	configConstName       = "config"
)

func loadRuntimeConfig(inputPath string) (runtimeConfig, error) {
	configPath, err := resolveConfigFilePath(inputPath)
	if err != nil {
		return runtimeConfig{}, err
	}

	fs := vfs.New()
	program, diagnostics := analysis.Analyze(fs, configPath)
	if len(diagnostics) > 0 {
		return runtimeConfig{}, diagnosticsToError(diagnostics)
	}

	schema := ir.FromProgram(program)
	constant, ok := findConstByName(schema.Constants, configConstName)
	if !ok {
		return runtimeConfig{}, fmt.Errorf("config file %q must declare const %q", configPath, configConstName)
	}

	payload, err := literalValueToJSON(constant.Value)
	if err != nil {
		return runtimeConfig{}, fmt.Errorf("failed to decode const %q: %w", configConstName, err)
	}

	var config configtypes.VdlConfig
	if err := json.Unmarshal(payload, &config); err != nil {
		return runtimeConfig{}, fmt.Errorf("failed to unmarshal const %q into VdlConfig: %w", configConstName, err)
	}

	return runtimeConfig{
		Path:   configPath,
		Config: config,
	}, nil
}

func resolveConfigFilePath(inputPath string) (string, error) {
	path := strings.TrimSpace(inputPath)
	if path == "" {
		path = "."
	}

	info, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", fmt.Errorf("path %q does not exist", path)
		}
		return "", fmt.Errorf("failed to inspect path %q: %w", path, err)
	}

	configPath := path
	if info.IsDir() {
		configPath = filepath.Join(path, defaultConfigFileName)
	}

	if filepath.Base(configPath) != defaultConfigFileName {
		return "", fmt.Errorf("invalid config file %q: expected %q", configPath, defaultConfigFileName)
	}

	configInfo, err := os.Stat(configPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", fmt.Errorf("could not find %q in %q", defaultConfigFileName, path)
		}
		return "", fmt.Errorf("failed to access config file %q: %w", configPath, err)
	}
	if configInfo.IsDir() {
		return "", fmt.Errorf("config path %q must be a file", configPath)
	}

	absPath, err := filepath.Abs(configPath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve config file path %q: %w", configPath, err)
	}

	return absPath, nil
}

func findConstByName(constants []irtypes.ConstantDef, name string) (*irtypes.ConstantDef, bool) {
	for i := range constants {
		if constants[i].Name == name {
			return &constants[i], true
		}
	}
	return nil, false
}

func literalValueToJSON(value irtypes.LiteralValue) ([]byte, error) {
	decoded, err := literalValueToAny(value)
	if err != nil {
		return nil, err
	}

	payload, err := json.Marshal(decoded)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal decoded literal: %w", err)
	}

	return payload, nil
}

func literalValueToAny(value irtypes.LiteralValue) (any, error) {
	switch value.Kind {
	case irtypes.LiteralKindString:
		return value.GetStringValue(), nil
	case irtypes.LiteralKindInt:
		return value.GetIntValue(), nil
	case irtypes.LiteralKindFloat:
		return value.GetFloatValue(), nil
	case irtypes.LiteralKindBool:
		return value.GetBoolValue(), nil
	case irtypes.LiteralKindObject:
		entries := value.GetObjectEntries()
		obj := make(map[string]any, len(entries))
		for _, entry := range entries {
			decoded, err := literalValueToAny(entry.Value)
			if err != nil {
				return nil, err
			}
			obj[entry.Key] = decoded
		}
		return obj, nil
	case irtypes.LiteralKindArray:
		items := value.GetArrayItems()
		arr := make([]any, len(items))
		for i := range items {
			decoded, err := literalValueToAny(items[i])
			if err != nil {
				return nil, err
			}
			arr[i] = decoded
		}
		return arr, nil
	default:
		return nil, fmt.Errorf("unsupported literal kind %q", value.Kind)
	}
}

func diagnosticsToError(diagnostics []analysis.Diagnostic) error {
	parts := make([]string, len(diagnostics))
	for i, diag := range diagnostics {
		parts[i] = diag.Error()
	}
	return errors.New(strings.Join(parts, "\n"))
}

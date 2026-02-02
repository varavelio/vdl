package main

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/varavelio/vdl/toolchain/internal/version"
)

//go:embed cmd_init_schema.vdl
var initSchema []byte

//go:embed cmd_init_config.yaml
var initConfig []byte

type cmdInitArgs struct {
	Path string `arg:"positional" help:"The directory path where VDL schema and config files will be created. Defaults to the current directory."`
}

func cmdInit(args *cmdInitArgs) {
	if args.Path == "" {
		args.Path = "."
	}

	// Validate that path is a directory
	if info, err := os.Stat(args.Path); err == nil && !info.IsDir() {
		fmt.Fprintf(os.Stderr, "VDL error: path must be a directory, not a file: %s\n", args.Path)
		os.Exit(1)
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(args.Path, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "VDL failed to create directory: %v\n", err)
		os.Exit(1)
	}

	// Generate unique filenames
	schemaName, schemaPath, _, configPath := generateUniqueFilenames(args.Path)

	// Write both files
	if err := os.WriteFile(schemaPath, initSchema, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "VDL failed to write schema file: %v\n", err)
		os.Exit(1)
	}

	initConfigStr := strings.ReplaceAll(string(initConfig), "{{schema_path}}", "./"+schemaName)
	initConfigStr = strings.ReplaceAll(initConfigStr, "{{config_schema_id}}", version.SchemaConfigID)
	if err := os.WriteFile(configPath, []byte(initConfigStr), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "VDL failed to write config file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("VDL: files initialized:\n- %s\n- %s\n", schemaPath, configPath)
}

// generateUniqueFilenames generates unique filenames for the schema and config files
//
// Returns:
// - schemaName: The name of the schema file
// - schemaPath: The path to the schema file
// - configName: The name of the config file
// - configPath: The path to the config file
func generateUniqueFilenames(dir string) (string, string, string, string) {
	schemaName := "schema.vdl"
	configName := "vdl.yaml"

	schemaPath := filepath.Join(dir, schemaName)
	configPath := filepath.Join(dir, configName)

	// Check if files exist
	if _, err := os.Stat(schemaPath); os.IsNotExist(err) {
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			return schemaName, schemaPath, configName, configPath
		}
	}

	// Generate unique suffix using unix timestamp
	timestamp := time.Now().Unix()
	schemaName = fmt.Sprintf("schema-%d.vdl", timestamp)
	configName = fmt.Sprintf("vdl-%d.yaml", timestamp)

	return schemaName, filepath.Join(dir, schemaName), configName, filepath.Join(dir, configName)
}

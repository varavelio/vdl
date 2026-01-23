package codegen

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/varavelio/vdl/toolchain/internal/codegen/config"
	"github.com/varavelio/vdl/toolchain/internal/codegen/dart"
	"github.com/varavelio/vdl/toolchain/internal/codegen/golang"
	"github.com/varavelio/vdl/toolchain/internal/codegen/jsonschema"
	"github.com/varavelio/vdl/toolchain/internal/codegen/openapi"
	"github.com/varavelio/vdl/toolchain/internal/codegen/playground"
	"github.com/varavelio/vdl/toolchain/internal/codegen/plugin"
	"github.com/varavelio/vdl/toolchain/internal/codegen/typescript"
	"github.com/varavelio/vdl/toolchain/internal/core/analysis"
	"github.com/varavelio/vdl/toolchain/internal/core/ir"
	"github.com/varavelio/vdl/toolchain/internal/core/vfs"
	"github.com/varavelio/vdl/toolchain/internal/formatter"
	"github.com/varavelio/vdl/toolchain/internal/util/filepathutil"
)

// Run runs the code generator and returns an error if one occurred.
func Run(configPath string) error {
	// Normalize config path first to ensure we resolve relative paths correctly
	absConfigPath, err := filepathutil.NormalizeFromWD(configPath)
	if err != nil {
		return fmt.Errorf("failed to normalize config path: %w", err)
	}
	absConfigDir := filepath.Dir(absConfigPath)

	cfg, err := config.LoadConfig(absConfigPath)
	if err != nil {
		return err
	}

	// Cache for parsed schemas to avoid reparsing the same file multiple times
	schemaCache := make(map[string]*ir.Schema)
	fs := vfs.New()

	// Helper to get or parse schema
	getSchema := func(schemaPath string) (*ir.Schema, *vfs.FileSystem, error) {
		// Schema path is relative to the config file
		absSchemaPath := filepath.Join(absConfigDir, schemaPath)
		if cached, ok := schemaCache[absSchemaPath]; ok {
			return cached, fs, nil
		}

		program, diagnostics := analysis.Analyze(fs, absSchemaPath)
		if len(diagnostics) > 0 {
			var errMsgs []string
			for _, d := range diagnostics {
				errMsgs = append(errMsgs, d.String())
			}
			return nil, nil, fmt.Errorf("schema validation failed for %s:\n%s", absSchemaPath, joinErrors(errMsgs))
		}

		schema := ir.FromProgram(program)
		schemaCache[absSchemaPath] = schema
		return schema, fs, nil
	}

	ctx := context.Background()

	for i, target := range cfg.Targets {
		// Note: validateAndSetDefaults ensures exactly one is set and Schema is populated.
		// We pass the pointer to the config struct directly.

		if target.Go != nil {
			schema, _, err := getSchema(target.Go.Schema)
			if err != nil {
				return err
			}
			if err := runGolang(ctx, absConfigDir, target.Go, schema); err != nil {
				return fmt.Errorf("target #%d (go): %w", i, err)
			}
		} else if target.TypeScript != nil {
			schema, _, err := getSchema(target.TypeScript.Schema)
			if err != nil {
				return err
			}
			if err := runTypeScript(ctx, absConfigDir, target.TypeScript, schema); err != nil {
				return fmt.Errorf("target #%d (typescript): %w", i, err)
			}
		} else if target.Dart != nil {
			schema, _, err := getSchema(target.Dart.Schema)
			if err != nil {
				return err
			}
			if err := runDart(ctx, absConfigDir, target.Dart, schema); err != nil {
				return fmt.Errorf("target #%d (dart): %w", i, err)
			}
		} else if target.JSONSchema != nil {
			schema, _, err := getSchema(target.JSONSchema.Schema)
			if err != nil {
				return err
			}
			if err := runJSONSchema(ctx, absConfigDir, target.JSONSchema, schema); err != nil {
				return fmt.Errorf("target #%d (jsonschema): %w", i, err)
			}
		} else if target.OpenAPI != nil {
			schema, _, err := getSchema(target.OpenAPI.Schema)
			if err != nil {
				return err
			}
			if err := runOpenAPI(ctx, absConfigDir, target.OpenAPI, schema); err != nil {
				return fmt.Errorf("target #%d (openapi): %w", i, err)
			}
		} else if target.Playground != nil {
			schema, fsRef, err := getSchema(target.Playground.Schema)
			if err != nil {
				return err
			}
			// Playground needs formatted schema
			absSchemaPath := filepath.Join(absConfigDir, target.Playground.Schema)
			formatted := getFormattedSchema(fsRef, absSchemaPath)

			if err := runPlayground(ctx, absConfigDir, target.Playground, schema, formatted); err != nil {
				return fmt.Errorf("target #%d (playground): %w", i, err)
			}
		} else if target.Plugin != nil {
			schema, _, err := getSchema(target.Plugin.Schema)
			if err != nil {
				return err
			}
			if err := runPlugin(ctx, absConfigDir, target.Plugin, schema); err != nil {
				return fmt.Errorf("target #%d (plugin): %w", i, err)
			}
		}
	}

	return nil
}

func runPlugin(ctx context.Context, absConfigDir string, config *config.PluginConfig, schema *ir.Schema) error {
	outputDir := filepath.Join(absConfigDir, config.Output)

	// Clean output directory if requested
	if config.Clean {
		if err := os.RemoveAll(outputDir); err != nil {
			return fmt.Errorf("failed to clean output directory: %w", err)
		}
	}

	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Generate the code using new Generator interface
	gen := plugin.New(config)
	files, err := gen.Generate(ctx, schema)
	if err != nil {
		return fmt.Errorf("failed to generate code: %w", err)
	}

	// Write the generated files
	for _, file := range files {
		outPath := filepath.Join(outputDir, file.Path)
		outDir := filepath.Dir(outPath)
		if err := os.MkdirAll(outDir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}
		if err := os.WriteFile(outPath, file.Content, 0644); err != nil {
			return fmt.Errorf("failed to write generated code to file: %w", err)
		}
	}

	return nil
}

// joinErrors joins multiple error messages with newlines.
func joinErrors(errs []string) string {
	result := ""
	for i, e := range errs {
		if i > 0 {
			result += "\n"
		}
		result += e
	}
	return result
}

// getFormattedSchema reads and formats the schema file.
func getFormattedSchema(fs *vfs.FileSystem, absSchemaPath string) string {
	content, err := fs.ReadFile(absSchemaPath)
	if err != nil {
		return ""
	}
	formatted, err := formatter.Format(absSchemaPath, string(content))
	if err != nil {
		return string(content) // Return original if formatting fails
	}
	return formatted
}

func runOpenAPI(ctx context.Context, absConfigDir string, config *config.OpenAPIConfig, schema *ir.Schema) error {
	outputDir := filepath.Join(absConfigDir, config.Output)

	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Generate the code using new Generator interface
	gen := openapi.New(config)
	files, err := gen.Generate(ctx, schema)
	if err != nil {
		return fmt.Errorf("failed to generate code: %w", err)
	}

	// Write the generated files
	for _, file := range files {
		outPath := filepath.Join(outputDir, file.RelativePath)
		if err := os.WriteFile(outPath, file.Content, 0644); err != nil {
			return fmt.Errorf("failed to write generated code to file: %w", err)
		}
	}

	return nil
}

func runPlayground(ctx context.Context, absConfigDir string, playgroundConfig *config.PlaygroundConfig, schema *ir.Schema, formattedSchema string) error {
	outputDir := filepath.Join(absConfigDir, playgroundConfig.Output)

	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Generate the playground using new Generator interface
	gen := playground.New(playgroundConfig, formattedSchema)
	files, err := gen.Generate(ctx, schema)
	if err != nil {
		return fmt.Errorf("failed to generate playground: %w", err)
	}

	// Write all generated files
	for _, file := range files {
		outPath := filepath.Join(outputDir, file.RelativePath)
		outDir := filepath.Dir(outPath)
		if err := os.MkdirAll(outDir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}
		if err := os.WriteFile(outPath, file.Content, 0644); err != nil {
			return fmt.Errorf("failed to write generated file %s: %w", outPath, err)
		}
	}

	// Generate the openapi.yaml file for the playground
	// Synthesize an OpenAPI config
	openAPIConfig := &config.OpenAPIConfig{
		CommonConfig: config.CommonConfig{
			Output: playgroundConfig.Output,
			Schema: playgroundConfig.Schema,
		},
		Filename: "openapi.yaml",
		Title:    "VDL API",
		Version:  "1.0.0",
		BaseURL:  playgroundConfig.DefaultBaseURL,
	}

	// Re-use runOpenAPI logic? No, runOpenAPI calculates path relative to absConfigDir.
	// Here playgroundConfig.Output is already relative to absConfigDir.
	// So we can call runOpenAPI directly!
	if err := runOpenAPI(ctx, absConfigDir, openAPIConfig, schema); err != nil {
		return fmt.Errorf("failed to generate openapi.yaml for playground: %w", err)
	}

	return nil
}

func runGolang(ctx context.Context, absConfigDir string, config *config.GoConfig, schema *ir.Schema) error {
	outputDir := filepath.Join(absConfigDir, config.Output)

	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Generate the code using new Generator interface
	gen := golang.New(config)
	files, err := gen.Generate(ctx, schema)
	if err != nil {
		return fmt.Errorf("failed to generate code: %w", err)
	}

	// Write the generated files
	for _, file := range files {
		outPath := filepath.Join(outputDir, file.RelativePath)
		outDir := filepath.Dir(outPath)
		if err := os.MkdirAll(outDir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}
		if err := os.WriteFile(outPath, file.Content, 0644); err != nil {
			return fmt.Errorf("failed to write generated code to file: %w", err)
		}
	}

	return nil
}

func runTypeScript(ctx context.Context, absConfigDir string, config *config.TypeScriptConfig, schema *ir.Schema) error {
	outputDir := filepath.Join(absConfigDir, config.Output)

	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Generate the code using new Generator interface
	gen := typescript.New(config)
	files, err := gen.Generate(ctx, schema)
	if err != nil {
		return fmt.Errorf("failed to generate code: %w", err)
	}

	// Write the generated files
	for _, file := range files {
		outPath := filepath.Join(outputDir, file.RelativePath)
		outDir := filepath.Dir(outPath)
		if err := os.MkdirAll(outDir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}
		if err := os.WriteFile(outPath, file.Content, 0644); err != nil {
			return fmt.Errorf("failed to write generated code to file: %w", err)
		}
	}

	return nil
}

func runDart(ctx context.Context, absConfigDir string, config *config.DartConfig, schema *ir.Schema) error {
	outputDir := filepath.Join(absConfigDir, config.Output)

	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Generate the code using new Generator interface
	gen := dart.New(config)
	files, err := gen.Generate(ctx, schema)
	if err != nil {
		return fmt.Errorf("failed to generate code: %w", err)
	}

	// Write the generated files
	for _, file := range files {
		outputFile := filepath.Join(outputDir, file.RelativePath)
		outputFileDir := filepath.Dir(outputFile)
		if err := os.MkdirAll(outputFileDir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}

		if err := os.WriteFile(outputFile, file.Content, 0644); err != nil {
			return fmt.Errorf("failed to write generated code to file %s: %w", outputFile, err)
		}
	}

	return nil
}

func runJSONSchema(ctx context.Context, absConfigDir string, config *config.JSONSchemaConfig, schema *ir.Schema) error {
	outputDir := filepath.Join(absConfigDir, config.Output)

	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Generate the code using new Generator interface
	gen := jsonschema.New(config)
	files, err := gen.Generate(ctx, schema)
	if err != nil {
		return fmt.Errorf("failed to generate code: %w", err)
	}

	// Write the generated files
	for _, file := range files {
		outputFile := filepath.Join(outputDir, file.RelativePath)
		outputFileDir := filepath.Dir(outputFile)
		if err := os.MkdirAll(outputFileDir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}

		if err := os.WriteFile(outputFile, file.Content, 0644); err != nil {
			return fmt.Errorf("failed to write generated code to file %s: %w", outputFile, err)
		}
	}

	return nil
}

package codegen

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/varavelio/vdl/toolchain/internal/codegen/config"
	"github.com/varavelio/vdl/toolchain/internal/codegen/dart"
	"github.com/varavelio/vdl/toolchain/internal/codegen/golang"
	"github.com/varavelio/vdl/toolchain/internal/codegen/jsonschema"
	"github.com/varavelio/vdl/toolchain/internal/codegen/openapi"
	"github.com/varavelio/vdl/toolchain/internal/codegen/playground"
	"github.com/varavelio/vdl/toolchain/internal/codegen/plugin"
	"github.com/varavelio/vdl/toolchain/internal/codegen/python"
	"github.com/varavelio/vdl/toolchain/internal/codegen/typescript"
	"github.com/varavelio/vdl/toolchain/internal/core/analysis"
	"github.com/varavelio/vdl/toolchain/internal/core/ir"
	"github.com/varavelio/vdl/toolchain/internal/core/vfs"
	"github.com/varavelio/vdl/toolchain/internal/transform"
	"github.com/varavelio/vdl/toolchain/internal/util/filepathutil"
)

// GeneratedFile represents a file produced by a generator.
type GeneratedFile struct {
	Path    string
	Content []byte
}

// prepareOutputDir cleans (if requested) and creates the output directory.
func prepareOutputDir(outputDir string, clean bool) error {
	if clean {
		if err := os.RemoveAll(outputDir); err != nil {
			return fmt.Errorf("failed to clean output directory: %w", err)
		}
	}
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}
	return nil
}

// writeGeneratedFiles writes a slice of generated files to the output directory.
func writeGeneratedFiles(outputDir string, files []GeneratedFile) error {
	for _, file := range files {
		outPath := filepath.Join(outputDir, file.Path)
		outDir := filepath.Dir(outPath)
		if err := os.MkdirAll(outDir, 0755); err != nil {
			return fmt.Errorf("failed to create directory for %s: %w", file.Path, err)
		}
		if err := os.WriteFile(outPath, file.Content, 0644); err != nil {
			return fmt.Errorf("failed to write file %s: %w", file.Path, err)
		}
	}
	return nil
}

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

	// Cache for parsed schemas and programs to avoid reparsing the same file multiple times
	schemaCache := make(map[string]*ir.Schema)
	programCache := make(map[string]*analysis.Program)
	fs := vfs.New()

	// Helper to get or parse schema (returns IR schema and program for advanced uses)
	getSchema := func(schemaPath string) (*ir.Schema, *analysis.Program, error) {
		// Schema path is relative to the config file
		absSchemaPath := filepath.Join(absConfigDir, schemaPath)
		if cached, ok := schemaCache[absSchemaPath]; ok {
			return cached, programCache[absSchemaPath], nil
		}

		program, diagnostics := analysis.Analyze(fs, absSchemaPath)
		if len(diagnostics) > 0 {
			var errMsgs []string
			for _, d := range diagnostics {
				errMsgs = append(errMsgs, d.String())
			}
			return nil, nil, fmt.Errorf("schema validation failed for %s:\n%s", absSchemaPath, strings.Join(errMsgs, "\n"))
		}

		schema := ir.FromProgram(program)
		schemaCache[absSchemaPath] = schema
		programCache[absSchemaPath] = program
		return schema, program, nil
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
		} else if target.Python != nil {
			schema, _, err := getSchema(target.Python.Schema)
			if err != nil {
				return err
			}
			if err := runPython(ctx, absConfigDir, target.Python, schema); err != nil {
				return fmt.Errorf("target #%d (python): %w", i, err)
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
			schema, program, err := getSchema(target.Playground.Schema)
			if err != nil {
				return err
			}
			// Playground needs merged and formatted schema (all includes resolved into one file)
			formatted := transform.MergeAndFormat(program)

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

func runPlugin(ctx context.Context, absConfigDir string, cfg *config.PluginConfig, schema *ir.Schema) error {
	outputDir := filepath.Join(absConfigDir, cfg.Output)
	if err := prepareOutputDir(outputDir, cfg.Clean); err != nil {
		return err
	}

	gen := plugin.New(cfg)
	files, err := gen.Generate(ctx, schema)
	if err != nil {
		return fmt.Errorf("failed to generate code: %w", err)
	}

	generatedFiles := make([]GeneratedFile, len(files))
	for i, f := range files {
		generatedFiles[i] = GeneratedFile{Path: f.Path, Content: f.Content}
	}
	return writeGeneratedFiles(outputDir, generatedFiles)
}

func runOpenAPI(ctx context.Context, absConfigDir string, cfg *config.OpenAPIConfig, schema *ir.Schema) error {
	outputDir := filepath.Join(absConfigDir, cfg.Output)
	if err := prepareOutputDir(outputDir, cfg.Clean); err != nil {
		return err
	}

	gen := openapi.New(cfg)
	files, err := gen.Generate(ctx, schema)
	if err != nil {
		return fmt.Errorf("failed to generate code: %w", err)
	}

	generatedFiles := make([]GeneratedFile, len(files))
	for i, f := range files {
		generatedFiles[i] = GeneratedFile{Path: f.RelativePath, Content: f.Content}
	}
	return writeGeneratedFiles(outputDir, generatedFiles)
}

func runPlayground(ctx context.Context, absConfigDir string, cfg *config.PlaygroundConfig, schema *ir.Schema, formattedSchema string) error {
	outputDir := filepath.Join(absConfigDir, cfg.Output)
	if err := prepareOutputDir(outputDir, cfg.Clean); err != nil {
		return err
	}

	gen := playground.New(cfg, formattedSchema)
	files, err := gen.Generate(ctx, schema)
	if err != nil {
		return fmt.Errorf("failed to generate playground: %w", err)
	}

	generatedFiles := make([]GeneratedFile, len(files))
	for i, f := range files {
		generatedFiles[i] = GeneratedFile{Path: f.RelativePath, Content: f.Content}
	}
	if err := writeGeneratedFiles(outputDir, generatedFiles); err != nil {
		return err
	}

	// Generate the openapi.yaml file for the playground
	openAPIConfig := &config.OpenAPIConfig{
		CommonConfig: config.CommonConfig{
			Output: cfg.Output,
			Schema: cfg.Schema,
			Clean:  false, // Don't clean again, we already cleaned above
		},
		Filename: "openapi.yaml",
		Title:    "VDL API",
		Version:  "1.0.0",
		BaseURL:  cfg.DefaultBaseURL,
	}

	if err := runOpenAPI(ctx, absConfigDir, openAPIConfig, schema); err != nil {
		return fmt.Errorf("failed to generate openapi.yaml for playground: %w", err)
	}

	return nil
}

func runGolang(ctx context.Context, absConfigDir string, cfg *config.GoConfig, schema *ir.Schema) error {
	outputDir := filepath.Join(absConfigDir, cfg.Output)
	if err := prepareOutputDir(outputDir, cfg.Clean); err != nil {
		return err
	}

	gen := golang.New(cfg)
	files, err := gen.Generate(ctx, schema)
	if err != nil {
		return fmt.Errorf("failed to generate code: %w", err)
	}

	generatedFiles := make([]GeneratedFile, len(files))
	for i, f := range files {
		generatedFiles[i] = GeneratedFile{Path: f.RelativePath, Content: f.Content}
	}
	return writeGeneratedFiles(outputDir, generatedFiles)
}

func runTypeScript(ctx context.Context, absConfigDir string, cfg *config.TypeScriptConfig, schema *ir.Schema) error {
	outputDir := filepath.Join(absConfigDir, cfg.Output)
	if err := prepareOutputDir(outputDir, cfg.Clean); err != nil {
		return err
	}

	gen := typescript.New(cfg)
	files, err := gen.Generate(ctx, schema)
	if err != nil {
		return fmt.Errorf("failed to generate code: %w", err)
	}

	generatedFiles := make([]GeneratedFile, len(files))
	for i, f := range files {
		generatedFiles[i] = GeneratedFile{Path: f.RelativePath, Content: f.Content}
	}
	return writeGeneratedFiles(outputDir, generatedFiles)
}

func runDart(ctx context.Context, absConfigDir string, cfg *config.DartConfig, schema *ir.Schema) error {
	outputDir := filepath.Join(absConfigDir, cfg.Output)
	if err := prepareOutputDir(outputDir, cfg.Clean); err != nil {
		return err
	}

	gen := dart.New(cfg)
	files, err := gen.Generate(ctx, schema)
	if err != nil {
		return fmt.Errorf("failed to generate code: %w", err)
	}

	generatedFiles := make([]GeneratedFile, len(files))
	for i, f := range files {
		generatedFiles[i] = GeneratedFile{Path: f.RelativePath, Content: f.Content}
	}
	return writeGeneratedFiles(outputDir, generatedFiles)
}

func runPython(ctx context.Context, absConfigDir string, cfg *config.PythonConfig, schema *ir.Schema) error {
	outputDir := filepath.Join(absConfigDir, cfg.Output)
	if err := prepareOutputDir(outputDir, cfg.Clean); err != nil {
		return err
	}

	gen := python.New(cfg)
	files, err := gen.Generate(ctx, schema)
	if err != nil {
		return fmt.Errorf("failed to generate code: %w", err)
	}

	generatedFiles := make([]GeneratedFile, len(files))
	for i, f := range files {
		generatedFiles[i] = GeneratedFile{Path: f.RelativePath, Content: f.Content}
	}
	return writeGeneratedFiles(outputDir, generatedFiles)
}

func runJSONSchema(ctx context.Context, absConfigDir string, cfg *config.JSONSchemaConfig, schema *ir.Schema) error {
	outputDir := filepath.Join(absConfigDir, cfg.Output)
	if err := prepareOutputDir(outputDir, cfg.Clean); err != nil {
		return err
	}

	gen := jsonschema.New(cfg)
	files, err := gen.Generate(ctx, schema)
	if err != nil {
		return fmt.Errorf("failed to generate code: %w", err)
	}

	generatedFiles := make([]GeneratedFile, len(files))
	for i, f := range files {
		generatedFiles[i] = GeneratedFile{Path: f.RelativePath, Content: f.Content}
	}
	return writeGeneratedFiles(outputDir, generatedFiles)
}

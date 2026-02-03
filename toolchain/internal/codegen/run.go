package codegen

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/varavelio/vdl/toolchain/internal/codegen/config"
	"github.com/varavelio/vdl/toolchain/internal/codegen/config/configtypes"
	"github.com/varavelio/vdl/toolchain/internal/codegen/dart"
	"github.com/varavelio/vdl/toolchain/internal/codegen/golang"
	"github.com/varavelio/vdl/toolchain/internal/codegen/irjson"
	"github.com/varavelio/vdl/toolchain/internal/codegen/jsonschema"
	"github.com/varavelio/vdl/toolchain/internal/codegen/openapi"
	"github.com/varavelio/vdl/toolchain/internal/codegen/playground"
	"github.com/varavelio/vdl/toolchain/internal/codegen/plugin"
	"github.com/varavelio/vdl/toolchain/internal/codegen/python"
	"github.com/varavelio/vdl/toolchain/internal/codegen/typescript"
	"github.com/varavelio/vdl/toolchain/internal/codegen/vdlschema"
	"github.com/varavelio/vdl/toolchain/internal/core/analysis"
	"github.com/varavelio/vdl/toolchain/internal/core/ir"
	"github.com/varavelio/vdl/toolchain/internal/core/ir/irtypes"
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

// Run runs the code generator and returns the total number of files generated and an error if one occurred.
func Run(configPath string) (int, error) {
	// Normalize config path first to ensure we resolve relative paths correctly
	absConfigPath, err := filepathutil.NormalizeFromWD(configPath)
	if err != nil {
		return 0, fmt.Errorf("failed to normalize config path: %w", err)
	}
	absConfigDir := filepath.Dir(absConfigPath)

	cfg, err := config.LoadConfig(absConfigPath)
	if err != nil {
		return 0, fmt.Errorf("failed to load config: %w", err)
	}

	// Cache for parsed schemas and programs to avoid reparsing the same file multiple times
	schemaCache := make(map[string]*irtypes.IrSchema)
	programCache := make(map[string]*analysis.Program)
	fs := vfs.New()

	// Helper to get or parse schema (returns IR schema and program for advanced uses)
	getSchema := func(schemaPath string) (*irtypes.IrSchema, *analysis.Program, error) {
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
	totalFiles := 0

	for i, target := range cfg.Targets {
		// Note: validateAndSetDefaults ensures exactly one is set and Schema is populated.
		// We pass the pointer to the config struct directly.

		if target.Go != nil {
			schema, _, err := getSchema(*target.Go.Schema)
			if err != nil {
				return 0, err
			}
			count, err := runGolang(ctx, absConfigDir, target.Go, schema)
			if err != nil {
				return 0, fmt.Errorf("target #%d (go): %w", i, err)
			}
			totalFiles += count
		} else if target.Typescript != nil {
			schema, _, err := getSchema(*target.Typescript.Schema)
			if err != nil {
				return 0, err
			}
			count, err := runTypeScript(ctx, absConfigDir, target.Typescript, schema)
			if err != nil {
				return 0, fmt.Errorf("target #%d (typescript): %w", i, err)
			}
			totalFiles += count
		} else if target.Dart != nil {
			schema, _, err := getSchema(*target.Dart.Schema)
			if err != nil {
				return 0, err
			}
			count, err := runDart(ctx, absConfigDir, target.Dart, schema)
			if err != nil {
				return 0, fmt.Errorf("target #%d (dart): %w", i, err)
			}
			totalFiles += count
		} else if target.Python != nil {
			schema, _, err := getSchema(*target.Python.Schema)
			if err != nil {
				return 0, err
			}
			count, err := runPython(ctx, absConfigDir, target.Python, schema)
			if err != nil {
				return 0, fmt.Errorf("target #%d (python): %w", i, err)
			}
			totalFiles += count
		} else if target.Jsonschema != nil {
			schema, _, err := getSchema(*target.Jsonschema.Schema)
			if err != nil {
				return 0, err
			}
			count, err := runJSONSchema(ctx, absConfigDir, target.Jsonschema, schema)
			if err != nil {
				return 0, fmt.Errorf("target #%d (jsonschema): %w", i, err)
			}
			totalFiles += count
		} else if target.Openapi != nil {
			schema, _, err := getSchema(*target.Openapi.Schema)
			if err != nil {
				return 0, err
			}
			count, err := runOpenAPI(ctx, absConfigDir, target.Openapi, schema)
			if err != nil {
				return 0, fmt.Errorf("target #%d (openapi): %w", i, err)
			}
			totalFiles += count
		} else if target.Playground != nil {
			schema, program, err := getSchema(*target.Playground.Schema)
			if err != nil {
				return 0, err
			}
			// Playground needs merged and formatted schema (all includes resolved into one file)
			formatted := transform.MergeAndFormat(program)

			count, err := runPlayground(ctx, absConfigDir, target.Playground, schema, formatted)
			if err != nil {
				return 0, fmt.Errorf("target #%d (playground): %w", i, err)
			}
			totalFiles += count
		} else if target.Plugin != nil {
			schema, _, err := getSchema(*target.Plugin.Schema)
			if err != nil {
				return 0, err
			}
			count, err := runPlugin(ctx, absConfigDir, target.Plugin, schema)
			if err != nil {
				return 0, fmt.Errorf("target #%d (plugin): %w", i, err)
			}
			totalFiles += count
		} else if target.Ir != nil {
			schema, _, err := getSchema(*target.Ir.Schema)
			if err != nil {
				return 0, err
			}
			count, err := runIR(ctx, absConfigDir, target.Ir, schema)
			if err != nil {
				return 0, fmt.Errorf("target #%d (ir): %w", i, err)
			}
			totalFiles += count
		} else if target.Vdl != nil {
			schema, program, err := getSchema(*target.Vdl.Schema)
			if err != nil {
				return 0, err
			}
			// VDL Schema needs merged and formatted schema (all includes resolved into one file)
			formatted := transform.MergeAndFormat(program)

			count, err := runVdl(ctx, absConfigDir, target.Vdl, schema, formatted)
			if err != nil {
				return 0, fmt.Errorf("target #%d (vdl): %w", i, err)
			}
			totalFiles += count
		}
	}

	return totalFiles, nil
}

func runPlugin(ctx context.Context, absConfigDir string, cfg *configtypes.PluginConfig, schema *irtypes.IrSchema) (int, error) {
	outputDir := filepath.Join(absConfigDir, cfg.Output)
	if err := prepareOutputDir(outputDir, config.ShouldClean(cfg.Clean)); err != nil {
		return 0, err
	}

	gen := plugin.New(cfg)
	files, err := gen.Generate(ctx, schema)
	if err != nil {
		return 0, fmt.Errorf("failed to generate code: %w", err)
	}

	generatedFiles := make([]GeneratedFile, len(files))
	for i, f := range files {
		generatedFiles[i] = GeneratedFile{Path: f.Path, Content: f.Content}
	}
	if err := writeGeneratedFiles(outputDir, generatedFiles); err != nil {
		return 0, err
	}

	return len(generatedFiles), nil
}

func runOpenAPI(ctx context.Context, absConfigDir string, cfg *configtypes.OpenApiConfig, schema *irtypes.IrSchema) (int, error) {
	outputDir := filepath.Join(absConfigDir, cfg.Output)
	if err := prepareOutputDir(outputDir, config.ShouldClean(cfg.Clean)); err != nil {
		return 0, err
	}

	gen := openapi.New(cfg)
	files, err := gen.Generate(ctx, schema)
	if err != nil {
		return 0, fmt.Errorf("failed to generate code: %w", err)
	}

	generatedFiles := make([]GeneratedFile, len(files))
	for i, f := range files {
		generatedFiles[i] = GeneratedFile{Path: f.RelativePath, Content: f.Content}
	}
	if err := writeGeneratedFiles(outputDir, generatedFiles); err != nil {
		return 0, err
	}

	return len(generatedFiles), nil
}

func runPlayground(ctx context.Context, absConfigDir string, cfg *configtypes.PlaygroundConfig, schema *irtypes.IrSchema, formattedSchema string) (int, error) {
	outputDir := filepath.Join(absConfigDir, cfg.Output)
	if err := prepareOutputDir(outputDir, config.ShouldClean(cfg.Clean)); err != nil {
		return 0, err
	}

	gen := playground.New(cfg, formattedSchema)
	files, err := gen.Generate(ctx, schema)
	if err != nil {
		return 0, fmt.Errorf("failed to generate playground: %w", err)
	}

	generatedFiles := make([]GeneratedFile, len(files))
	for i, f := range files {
		generatedFiles[i] = GeneratedFile{Path: f.RelativePath, Content: f.Content}
	}
	if err := writeGeneratedFiles(outputDir, generatedFiles); err != nil {
		return 0, err
	}

	// Generate the openapi.yaml file for the playground
	filename := "openapi.yaml"
	openAPIConfig := &configtypes.OpenApiConfig{
		Output:   cfg.Output,
		Schema:   cfg.Schema,
		Clean:    nil, // Don't clean again, we already cleaned above
		Filename: &filename,
		Title:    "VDL API",
		Version:  "1.0.0",
		BaseUrl:  cfg.DefaultBaseUrl,
	}

	openAPICount, err := runOpenAPI(ctx, absConfigDir, openAPIConfig, schema)
	if err != nil {
		return 0, fmt.Errorf("failed to generate openapi.yaml for playground: %w", err)
	}

	return len(generatedFiles) + openAPICount, nil
}

func runGolang(ctx context.Context, absConfigDir string, cfg *configtypes.GoConfig, schema *irtypes.IrSchema) (int, error) {
	outputDir := filepath.Join(absConfigDir, cfg.Output)
	if err := prepareOutputDir(outputDir, config.ShouldClean(cfg.Clean)); err != nil {
		return 0, err
	}

	gen := golang.New(cfg)
	files, err := gen.Generate(ctx, schema)
	if err != nil {
		return 0, fmt.Errorf("failed to generate code: %w", err)
	}

	generatedFiles := make([]GeneratedFile, len(files))
	for i, f := range files {
		generatedFiles[i] = GeneratedFile{Path: f.RelativePath, Content: f.Content}
	}
	if err := writeGeneratedFiles(outputDir, generatedFiles); err != nil {
		return 0, err
	}

	return len(generatedFiles), nil
}

func runTypeScript(ctx context.Context, absConfigDir string, cfg *configtypes.TypeScriptConfig, schema *irtypes.IrSchema) (int, error) {
	outputDir := filepath.Join(absConfigDir, cfg.Output)
	if err := prepareOutputDir(outputDir, config.ShouldClean(cfg.Clean)); err != nil {
		return 0, err
	}

	gen := typescript.New(cfg)
	files, err := gen.Generate(ctx, schema)
	if err != nil {
		return 0, fmt.Errorf("failed to generate code: %w", err)
	}

	generatedFiles := make([]GeneratedFile, len(files))
	for i, f := range files {
		generatedFiles[i] = GeneratedFile{Path: f.RelativePath, Content: f.Content}
	}
	if err := writeGeneratedFiles(outputDir, generatedFiles); err != nil {
		return 0, err
	}

	return len(generatedFiles), nil
}

func runDart(ctx context.Context, absConfigDir string, cfg *configtypes.DartConfig, schema *irtypes.IrSchema) (int, error) {
	outputDir := filepath.Join(absConfigDir, cfg.Output)
	if err := prepareOutputDir(outputDir, config.ShouldClean(cfg.Clean)); err != nil {
		return 0, err
	}

	gen := dart.New(cfg)
	files, err := gen.Generate(ctx, schema)
	if err != nil {
		return 0, fmt.Errorf("failed to generate code: %w", err)
	}

	generatedFiles := make([]GeneratedFile, len(files))
	for i, f := range files {
		generatedFiles[i] = GeneratedFile{Path: f.RelativePath, Content: f.Content}
	}
	if err := writeGeneratedFiles(outputDir, generatedFiles); err != nil {
		return 0, err
	}

	return len(generatedFiles), nil
}

func runPython(ctx context.Context, absConfigDir string, cfg *configtypes.PythonConfig, schema *irtypes.IrSchema) (int, error) {
	outputDir := filepath.Join(absConfigDir, cfg.Output)
	if err := prepareOutputDir(outputDir, config.ShouldClean(cfg.Clean)); err != nil {
		return 0, err
	}

	gen := python.New(cfg)
	files, err := gen.Generate(ctx, schema)
	if err != nil {
		return 0, fmt.Errorf("failed to generate code: %w", err)
	}

	generatedFiles := make([]GeneratedFile, len(files))
	for i, f := range files {
		generatedFiles[i] = GeneratedFile{Path: f.RelativePath, Content: f.Content}
	}
	if err := writeGeneratedFiles(outputDir, generatedFiles); err != nil {
		return 0, err
	}

	return len(generatedFiles), nil
}

func runJSONSchema(ctx context.Context, absConfigDir string, cfg *configtypes.JsonSchemaConfig, schema *irtypes.IrSchema) (int, error) {
	outputDir := filepath.Join(absConfigDir, cfg.Output)
	if err := prepareOutputDir(outputDir, config.ShouldClean(cfg.Clean)); err != nil {
		return 0, err
	}

	gen := jsonschema.New(cfg)
	files, err := gen.Generate(ctx, schema)
	if err != nil {
		return 0, fmt.Errorf("failed to generate code: %w", err)
	}

	generatedFiles := make([]GeneratedFile, len(files))
	for i, f := range files {
		generatedFiles[i] = GeneratedFile{Path: f.RelativePath, Content: f.Content}
	}
	if err := writeGeneratedFiles(outputDir, generatedFiles); err != nil {
		return 0, err
	}

	return len(generatedFiles), nil
}

func runIR(ctx context.Context, absConfigDir string, cfg *configtypes.IrConfig, schema *irtypes.IrSchema) (int, error) {
	outputDir := filepath.Join(absConfigDir, cfg.Output)
	if err := prepareOutputDir(outputDir, config.ShouldClean(cfg.Clean)); err != nil {
		return 0, err
	}

	gen := irjson.New(cfg)
	files, err := gen.Generate(ctx, schema)
	if err != nil {
		return 0, fmt.Errorf("failed to generate IR JSON: %w", err)
	}

	generatedFiles := make([]GeneratedFile, len(files))
	for i, f := range files {
		generatedFiles[i] = GeneratedFile{Path: f.RelativePath, Content: f.Content}
	}
	if err := writeGeneratedFiles(outputDir, generatedFiles); err != nil {
		return 0, err
	}

	return len(generatedFiles), nil
}

func runVdl(ctx context.Context, absConfigDir string, cfg *configtypes.VdlTargetConfig, schema *irtypes.IrSchema, formattedSchema string) (int, error) {
	outputDir := filepath.Join(absConfigDir, cfg.Output)
	if err := prepareOutputDir(outputDir, config.ShouldClean(cfg.Clean)); err != nil {
		return 0, err
	}

	gen := vdlschema.New(cfg, formattedSchema)
	files, err := gen.Generate(ctx, schema)
	if err != nil {
		return 0, fmt.Errorf("failed to generate unified VDL schema: %w", err)
	}

	generatedFiles := make([]GeneratedFile, len(files))
	for i, f := range files {
		generatedFiles[i] = GeneratedFile{Path: f.RelativePath, Content: f.Content}
	}
	if err := writeGeneratedFiles(outputDir, generatedFiles); err != nil {
		return 0, err
	}

	return len(generatedFiles), nil
}

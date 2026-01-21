package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/invopop/jsonschema"
	"github.com/varavelio/vdl/toolchain/internal/codegen/config"
	"github.com/varavelio/vdl/toolchain/internal/version"
)

// Shadow types for JSON Schema generation with jsonschema tags.

type baseTarget struct {
	Output string `json:"output" jsonschema:"minLength=1,description=The output directory where the generated files will be placed."`
	Clean  bool   `json:"clean,omitempty" jsonschema:"default=false,description=If true empties the output directory before generation."`
	Schema string `json:"schema,omitempty" jsonschema:"description=Optional override for the VDL schema file specific to this target."`
}

type (
	goTarget struct {
		baseTarget
		Target  string            `json:"target" jsonschema:"const=go"`
		Options *config.GoOptions `json:"options,omitempty"`
	}
	tsTarget struct {
		baseTarget
		Target  string                    `json:"target" jsonschema:"const=typescript"`
		Options *config.TypeScriptOptions `json:"options,omitempty"`
	}
	dartTarget struct {
		baseTarget
		Target  string              `json:"target" jsonschema:"const=dart"`
		Options *config.DartOptions `json:"options,omitempty"`
	}
	openAPITarget struct {
		baseTarget
		Target  string                 `json:"target" jsonschema:"const=openapi"`
		Options *config.OpenAPIOptions `json:"options,omitempty"`
	}
	playgroundTarget struct {
		baseTarget
		Target  string                    `json:"target" jsonschema:"const=playground"`
		Options *config.PlaygroundOptions `json:"options,omitempty"`
	}
)

// targets maps each target to its schema generation metadata.
var targets = []struct {
	defName, optName, constVal string
	model, options             any
}{
	{"GoTarget", "GoOptions", config.TargetGo, &goTarget{}, &config.GoOptions{}},
	{"TsTarget", "TypeScriptOptions", config.TargetTypeScript, &tsTarget{}, &config.TypeScriptOptions{}},
	{"DartTarget", "DartOptions", config.TargetDart, &dartTarget{}, &config.DartOptions{}},
	{"OpenAPITarget", "OpenAPIOptions", config.TargetOpenAPI, &openAPITarget{}, &config.OpenAPIOptions{}},
	{"PlaygroundTarget", "PlaygroundOptions", config.TargetPlayground, &playgroundTarget{}, &config.PlaygroundOptions{}},
}

type targetWrapper struct{}

func (targetWrapper) JSONSchema() *jsonschema.Schema {
	refs := make([]*jsonschema.Schema, len(targets))
	for i, t := range targets {
		refs[i] = &jsonschema.Schema{Ref: "#/$defs/" + t.defName}
	}
	return &jsonschema.Schema{OneOf: refs}
}

type configSchema struct {
	Version int             `json:"version" jsonschema:"const=1"`
	Schema  string          `json:"schema,omitempty"`
	Targets []targetWrapper `json:"targets" jsonschema:"minItems=1"`
}

func generateConfigSchema() {
	r := &jsonschema.Reflector{ExpandedStruct: true}
	schema := r.Reflect(&configSchema{})

	schema.ID = jsonschema.ID(fmt.Sprintf("https://vdl.varavel.com/schemas/v%s/config.schema.json", version.VersionMajor))
	schema.Title = "VDL Config Schema"
	schema.Description = "JSON Schema for the VDL Config"

	if v, ok := schema.Properties.Get("version"); ok {
		v.Const = 1
	}

	for _, t := range targets {
		schema.Definitions[t.optName] = r.Reflect(t.options)

		ts := r.Reflect(t.model)
		ts.Definitions = nil
		if ts.Properties != nil {
			if opt, ok := ts.Properties.Get("options"); ok {
				opt.Ref = "#/$defs/" + t.optName
			}
			if tgt, ok := ts.Properties.Get("target"); ok {
				tgt.Const = t.constVal
			}
		}
		schema.Definitions[t.defName] = ts
	}

	data, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		log.Fatalf("failed to marshal schema: %v", err)
	}
	data = append(data, '\n')

	outPath := filepath.Join("internal", "codegen", "config", "config.schema.json")
	if err := os.WriteFile(outPath, data, 0644); err != nil {
		log.Fatalf("failed to write schema: %v", err)
	}

	log.Printf("Generated %s", outPath)
}

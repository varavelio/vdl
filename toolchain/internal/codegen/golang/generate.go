package golang

import (
	"strings"

	"github.com/uforg/ufogenkit"
	"github.com/uforg/uforpc/urpc/internal/schema"
	"golang.org/x/tools/imports"
)

// Generate takes a schema and a config and generates the Go code for the schema.
func Generate(sch schema.Schema, config Config) (string, error) {
	subGenerators := []func(schema.Schema, Config) (string, error){
		generatePackage,
		generateCoreTypes,
		generateDomainTypes,
		generateProcedureTypes,
		generateStreamTypes,
		generateOptional,
		generateServer,
		generateClient,
	}

	g := ufogenkit.NewGenKit().WithTabs()
	for _, generator := range subGenerators {
		codeChunk, err := generator(sch, config)
		if err != nil {
			return "", err
		}

		codeChunk = strings.TrimSpace(codeChunk)
		g.Raw(codeChunk)
		g.Break()
		g.Break()
	}

	generatedCode := g.String()

	// Try to format the generated code (might not work when running in WebAssembly)
	formattedCode, err := imports.Process("", []byte(generatedCode), nil)
	if err == nil {
		generatedCode = string(formattedCode)
	}

	return generatedCode, nil
}

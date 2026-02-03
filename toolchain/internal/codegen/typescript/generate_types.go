package typescript

import (
	"strings"

	"github.com/varavelio/gen"
	"github.com/varavelio/vdl/toolchain/internal/codegen/config/configtypes"
	"github.com/varavelio/vdl/toolchain/internal/core/ir/irtypes"
)

func generateTypes(schema *irtypes.IrSchema, cfg *configtypes.TypeScriptConfig) (string, error) {
	g := gen.New().WithSpaces(2)

	if len(schema.Procedures) > 0 || len(schema.Streams) > 0 {
		generateImport(g, []string{"Response"}, "./core", true, cfg)
	}
	g.Break()

	// Helper to append content if not empty
	appendContent := func(g *gen.Generator, genFunc func(*irtypes.IrSchema, *configtypes.TypeScriptConfig) (string, error)) error {
		content, err := genFunc(schema, cfg)
		if err != nil {
			return err
		}
		if strings.TrimSpace(content) != "" {
			g.Raw(content)
			g.Break()
		}
		return nil
	}

	if err := appendContent(g, generateEnums); err != nil {
		return "", err
	}
	if err := appendContent(g, generateDomainTypes); err != nil {
		return "", err
	}
	if err := appendContent(g, generateProcedureTypes); err != nil {
		return "", err
	}
	if err := appendContent(g, generateStreamTypes); err != nil {
		return "", err
	}

	typesContent := g.String()
	if strings.TrimSpace(typesContent) == "" {
		g.Line("export {};")
	}

	return g.String(), nil
}

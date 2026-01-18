package openapi

import (
	"fmt"

	"github.com/uforg/uforpc/urpc/internal/schema"
)

func generatePaths(sch schema.Schema) (Paths, error) {
	paths := Paths{}

	for _, procNode := range sch.GetProcNodes() {
		name := procNode.Name
		inputName := fmt.Sprintf("%sInput", name)
		outputName := fmt.Sprintf("%sOutput", name)

		doc := ""
		if procNode.Doc != nil {
			doc = *procNode.Doc
		}

		paths["/"+name] = map[string]any{
			"post": map[string]any{
				"deprecated":  procNode.Deprecated != nil,
				"tags":        []string{"procedures"},
				"description": doc,
				"requestBody": map[string]any{
					"$ref": fmt.Sprintf("#/components/requestBodies/%s", inputName),
				},
				"responses": map[string]any{
					"200": map[string]any{
						"$ref": fmt.Sprintf("#/components/responses/%s", outputName),
					},
				},
			},
		}
	}

	for _, streamNode := range sch.GetStreamNodes() {
		name := streamNode.Name
		inputName := fmt.Sprintf("%sInput", name)
		outputName := fmt.Sprintf("%sOutput", name)

		doc := ""
		if streamNode.Doc != nil {
			doc = *streamNode.Doc
		}

		paths["/"+name] = map[string]any{
			"post": map[string]any{
				"deprecated":  streamNode.Deprecated != nil,
				"tags":        []string{"streams"},
				"description": doc,
				"requestBody": map[string]any{
					"$ref": fmt.Sprintf("#/components/requestBodies/%s", inputName),
				},
				"responses": map[string]any{
					"200": map[string]any{
						"$ref": fmt.Sprintf("#/components/responses/%s", outputName),
					},
				},
			},
		}
	}

	return paths, nil
}

package openapi

import (
	"fmt"

	"github.com/varavelio/vdl/toolchain/internal/core/ir/irtypes"
)

// generatePaths generates OpenAPI paths from the IR schema.
// Paths follow the VDL request lifecycle spec: /{RPCName}/{EndpointName}
func generatePaths(schema *irtypes.IrSchema) Paths {
	paths := Paths{}

	// Generate paths for procedures
	for _, proc := range schema.Procedures {
		path := "/" + proc.RpcName + "/" + proc.Name
		inputName := proc.RpcName + proc.Name + "Input"
		outputName := proc.RpcName + proc.Name + "Output"

		operation := map[string]any{
			"tags": []string{proc.RpcName + "Procedures"},
			"requestBody": map[string]any{
				"$ref": fmt.Sprintf("#/components/requestBodies/%s", inputName),
			},
			"responses": map[string]any{
				"200": map[string]any{
					"$ref": fmt.Sprintf("#/components/responses/%s", outputName),
				},
			},
		}

		doc := proc.GetDoc()
		if doc != "" {
			operation["description"] = doc
		}

		if proc.Deprecated != nil {
			operation["deprecated"] = true
		}

		paths[path] = map[string]any{
			"post": operation,
		}
	}

	// Generate paths for streams
	for _, stream := range schema.Streams {
		path := "/" + stream.RpcName + "/" + stream.Name
		inputName := stream.RpcName + stream.Name + "Input"
		outputName := stream.RpcName + stream.Name + "Output"

		operation := map[string]any{
			"tags": []string{stream.RpcName + "Streams"},
			"requestBody": map[string]any{
				"$ref": fmt.Sprintf("#/components/requestBodies/%s", inputName),
			},
			"responses": map[string]any{
				"200": map[string]any{
					"$ref": fmt.Sprintf("#/components/responses/%s", outputName),
				},
			},
		}

		doc := stream.GetDoc()
		if doc != "" {
			operation["description"] = doc
		}

		if stream.Deprecated != nil {
			operation["deprecated"] = true
		}

		paths[path] = map[string]any{
			"post": operation,
		}
	}

	return paths
}

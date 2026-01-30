package openapi

import (
	"fmt"

	"github.com/varavelio/vdl/toolchain/internal/core/ir"
)

// generatePaths generates OpenAPI paths from the IR schema.
// Paths follow the VDL request lifecycle spec: /{RPCName}/{EndpointName}
func generatePaths(schema *ir.Schema) Paths {
	paths := Paths{}

	for _, rpc := range schema.RPCs {
		// Generate paths for procedures
		for _, proc := range rpc.Procs {
			path := "/" + rpc.Name + "/" + proc.Name
			inputName := rpc.Name + proc.Name + "Input"
			outputName := rpc.Name + proc.Name + "Output"

			operation := map[string]any{
				"tags": []string{rpc.Name + "Procedures"},
				"requestBody": map[string]any{
					"$ref": fmt.Sprintf("#/components/requestBodies/%s", inputName),
				},
				"responses": map[string]any{
					"200": map[string]any{
						"$ref": fmt.Sprintf("#/components/responses/%s", outputName),
					},
				},
			}

			if proc.Doc != "" {
				operation["description"] = proc.Doc
			}

			if proc.Deprecated != nil {
				operation["deprecated"] = true
			}

			paths[path] = map[string]any{
				"post": operation,
			}
		}

		// Generate paths for streams
		for _, stream := range rpc.Streams {
			path := "/" + rpc.Name + "/" + stream.Name
			inputName := rpc.Name + stream.Name + "Input"
			outputName := rpc.Name + stream.Name + "Output"

			operation := map[string]any{
				"tags": []string{rpc.Name + "Streams"},
				"requestBody": map[string]any{
					"$ref": fmt.Sprintf("#/components/requestBodies/%s", inputName),
				},
				"responses": map[string]any{
					"200": map[string]any{
						"$ref": fmt.Sprintf("#/components/responses/%s", outputName),
					},
				},
			}

			if stream.Doc != "" {
				operation["description"] = stream.Doc
			}

			if stream.Deprecated != nil {
				operation["deprecated"] = true
			}

			paths[path] = map[string]any{
				"post": operation,
			}
		}
	}

	return paths
}

package openapi

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/varavelio/vdl/toolchain/internal/core/ir/irtypes"
)

const authTokenDescription = `
Enter the full value for the Authorization header. The specific format (Bearer, Basic, API Key, etc.) is determined by the server's implementation.

---
**Examples:**
- **Bearer Token:** ''Bearer eyJhbGciOiJIUzI1Ni...'' (a JWT token)
- **Basic Auth:** ''Basic dXNlcm5hbWU6cGFzc3dvcmQ='' (a base64 encoding of ''username:password'')
- **API Key:** ''sk_live_123abc456def'' (a raw token)
`

// generateComponents generates OpenAPI components from the IR schema.
func generateComponents(schema *irtypes.IrSchema) Components {
	components := Components{
		SecuritySchemes: map[string]any{
			"AuthToken": map[string]any{
				"type":        "apiKey",
				"in":          "header",
				"name":        "Authorization",
				"description": strings.TrimSpace(strings.ReplaceAll(authTokenDescription, "''", "`")),
			},
		},
		Schemas:       map[string]any{},
		RequestBodies: map[string]any{},
		Responses:     map[string]any{},
	}

	// Generate schemas for custom types
	for _, t := range schema.Types {
		components.Schemas[t.Name] = generateTypeSchema(t)
	}

	// Generate schemas for enums
	for _, e := range schema.Enums {
		components.Schemas[e.Name] = generateEnumSchema(e)
	}

	// Generate request/response bodies for each procedure
	for _, proc := range schema.Procedures {
		inputName := proc.RpcName + proc.Name + "Input"
		outputName := proc.RpcName + proc.Name + "Output"

		components.RequestBodies[inputName] = generateRequestBody(
			proc.Input,
			fmt.Sprintf("Request body for %s/%s procedure", proc.RpcName, proc.Name),
		)

		components.Responses[outputName] = generateProcedureResponse(
			proc.Output,
			fmt.Sprintf("Response for %s/%s procedure", proc.RpcName, proc.Name),
		)
	}

	// Generate request/response bodies for each stream
	for _, stream := range schema.Streams {
		inputName := stream.RpcName + stream.Name + "Input"
		outputName := stream.RpcName + stream.Name + "Output"

		components.RequestBodies[inputName] = generateRequestBody(
			stream.Input,
			fmt.Sprintf("Request body for %s/%s stream subscription", stream.RpcName, stream.Name),
		)

		components.Responses[outputName] = generateStreamResponse(
			stream.Output,
			fmt.Sprintf("Server-Sent Events for %s/%s stream", stream.RpcName, stream.Name),
		)
	}

	return components
}

// generateTypeSchema generates an OpenAPI schema for an IR type.
func generateTypeSchema(t irtypes.TypeDef) map[string]any {
	properties, required := generatePropertiesFromFields(t.Fields)

	schema := map[string]any{
		"type":       "object",
		"properties": properties,
	}

	doc := t.GetDoc()
	if doc != "" {
		schema["description"] = doc
	}

	if t.Deprecated != nil {
		schema["deprecated"] = true
		deprecated := t.GetDeprecated()
		if deprecated != "" {
			desc := schema["description"]
			if desc == nil {
				desc = ""
			}
			schema["description"] = fmt.Sprintf("%s\n\nDeprecated: %s", desc, deprecated)
		}
	}

	if len(required) > 0 {
		schema["required"] = required
	}

	return schema
}

// generateEnumSchema generates an OpenAPI schema for an IR enum.
func generateEnumSchema(e irtypes.EnumDef) map[string]any {
	schema := map[string]any{}

	if e.EnumType == irtypes.EnumTypeString {
		values := []string{}
		for _, m := range e.Members {
			values = append(values, m.Value)
		}
		schema["type"] = "string"
		schema["enum"] = values
	} else {
		values := []int{}
		for _, m := range e.Members {
			v, _ := strconv.Atoi(m.Value)
			values = append(values, v)
		}
		schema["type"] = "integer"
		schema["enum"] = values
	}

	doc := e.GetDoc()
	if doc != "" {
		schema["description"] = doc
	}

	if e.Deprecated != nil {
		schema["deprecated"] = true
	}

	return schema
}

// generateRequestBody generates an OpenAPI request body from IR fields.
func generateRequestBody(fields []irtypes.Field, description string) map[string]any {
	properties, required := generatePropertiesFromFields(fields)

	schema := map[string]any{
		"type":       "object",
		"properties": properties,
	}
	if len(required) > 0 {
		schema["required"] = required
	}

	return map[string]any{
		"description": description,
		"content": map[string]any{
			"application/json": map[string]any{
				"schema": schema,
			},
		},
	}
}

// generateProcedureResponse generates an OpenAPI response for a procedure.
func generateProcedureResponse(fields []irtypes.Field, description string) map[string]any {
	properties, required := generateOutputProperties(fields)

	schema := map[string]any{
		"type":       "object",
		"properties": properties,
	}
	if len(required) > 0 {
		schema["required"] = required
	}

	return map[string]any{
		"description": description,
		"content": map[string]any{
			"application/json": map[string]any{
				"schema": schema,
			},
		},
	}
}

// generateStreamResponse generates an OpenAPI response for a stream (SSE).
func generateStreamResponse(fields []irtypes.Field, description string) map[string]any {
	properties, required := generateOutputProperties(fields)

	schema := map[string]any{
		"type":       "object",
		"properties": properties,
	}
	if len(required) > 0 {
		schema["required"] = required
	}

	return map[string]any{
		"description": description,
		"content": map[string]any{
			"text/event-stream": map[string]any{
				"schema": schema,
			},
		},
	}
}

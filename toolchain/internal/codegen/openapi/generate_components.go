package openapi

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/varavelio/vdl/toolchain/internal/core/ir"
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
func generateComponents(schema *ir.Schema) Components {
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

	// Generate request/response bodies for each RPC endpoint
	for _, rpc := range schema.RPCs {
		// Procedures
		for _, proc := range rpc.Procs {
			inputName := rpc.Name + "_" + proc.Name + "Input"
			outputName := rpc.Name + "_" + proc.Name + "Output"

			components.RequestBodies[inputName] = generateRequestBody(
				proc.Input,
				fmt.Sprintf("Request body for %s/%s procedure", rpc.Name, proc.Name),
			)

			components.Responses[outputName] = generateProcedureResponse(
				proc.Output,
				fmt.Sprintf("Response for %s/%s procedure", rpc.Name, proc.Name),
			)
		}

		// Streams
		for _, stream := range rpc.Streams {
			inputName := rpc.Name + "_" + stream.Name + "Input"
			outputName := rpc.Name + "_" + stream.Name + "Output"

			components.RequestBodies[inputName] = generateRequestBody(
				stream.Input,
				fmt.Sprintf("Request body for %s/%s stream subscription", rpc.Name, stream.Name),
			)

			components.Responses[outputName] = generateStreamResponse(
				stream.Output,
				fmt.Sprintf("Server-Sent Events for %s/%s stream", rpc.Name, stream.Name),
			)
		}
	}

	return components
}

// generateTypeSchema generates an OpenAPI schema for an IR type.
func generateTypeSchema(t ir.Type) map[string]any {
	properties, required := generatePropertiesFromFields(t.Fields)

	schema := map[string]any{
		"type":       "object",
		"properties": properties,
	}

	if t.Doc != "" {
		schema["description"] = t.Doc
	}

	if t.Deprecated != nil {
		schema["deprecated"] = true
		if t.Deprecated.Message != "" {
			desc := schema["description"]
			if desc == nil {
				desc = ""
			}
			schema["description"] = fmt.Sprintf("%s\n\nDeprecated: %s", desc, t.Deprecated.Message)
		}
	}

	if len(required) > 0 {
		schema["required"] = required
	}

	return schema
}

// generateEnumSchema generates an OpenAPI schema for an IR enum.
func generateEnumSchema(e ir.Enum) map[string]any {
	schema := map[string]any{}

	if e.ValueType == ir.EnumValueTypeString {
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

	if e.Doc != "" {
		schema["description"] = e.Doc
	}

	if e.Deprecated != nil {
		schema["deprecated"] = true
	}

	return schema
}

// generateRequestBody generates an OpenAPI request body from IR fields.
func generateRequestBody(fields []ir.Field, description string) map[string]any {
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
func generateProcedureResponse(fields []ir.Field, description string) map[string]any {
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
func generateStreamResponse(fields []ir.Field, description string) map[string]any {
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

package schema

import (
	_ "embed"
	"fmt"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

//go:embed schema.json
var jsonSchemaRaw string

// compileJSONSchema compiles the JSON schema to be used for schema input validation
func compileJSONSchema() (*jsonschema.Schema, error) {
	dummySchemaURL := "https://raw.githubusercontent.com/uforg/uforpc/refs/heads/main/internal/schema/schema.json"

	unmarshaled, err := jsonschema.UnmarshalJSON(strings.NewReader(jsonSchemaRaw))
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal json schema: %w", err)
	}

	c := jsonschema.NewCompiler()

	if err := c.AddResource(dummySchemaURL, unmarshaled); err != nil {
		return nil, err
	}

	return c.Compile(dummySchemaURL)
}

// validateSchema validates the input schema against the defined JSON schema
func validateSchema(inputSchema string) error {
	unmarshaled, err := jsonschema.UnmarshalJSON(strings.NewReader(inputSchema))
	if err != nil {
		return fmt.Errorf("failed to unmarshal input schema: %w", err)
	}

	jsonSchema, err := compileJSONSchema()
	if err != nil {
		return fmt.Errorf("failed to compile json schema: %w", err)
	}

	if err := jsonSchema.Validate(unmarshaled); err != nil {
		return fmt.Errorf("invalid input: %w", err)
	}

	return nil
}

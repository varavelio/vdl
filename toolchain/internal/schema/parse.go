package schema

import (
	"encoding/json"
	"fmt"
)

// ParseSchema parses and validates a JSON schema string into a Schema struct
func ParseSchema(schemaStr string) (Schema, error) {
	if err := validateSchema(schemaStr); err != nil {
		return Schema{}, fmt.Errorf("error validating against JSON schema: %w", err)
	}

	var schema Schema
	if err := json.Unmarshal([]byte(schemaStr), &schema); err != nil {
		return Schema{}, fmt.Errorf("error decoding schema: %w", err)
	}

	// TODO: Resolve the circular dependency and run the semantic analyzer

	// astSchema, err := transpile.ToURPC(schema)
	// if err != nil {
	// 	return Schema{}, fmt.Errorf("error transpiling to URPC: %w", err)
	// }

	// analyzer, err := analyzer.NewAnalyzer(docstore.NewDocstore())
	// if err != nil {
	// 	return Schema{}, fmt.Errorf("error creating analyzer: %w", err)
	// }

	// _, err = analyzer.AnalyzeAstSchema(&astSchema)
	// if err != nil {
	// 	return Schema{}, fmt.Errorf("error analyzing URPC schema: %w", err)
	// }

	return schema, nil
}

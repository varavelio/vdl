package typescript

import (
	"github.com/varavelio/vdl/toolchain/internal/core/ir"
)

// collectAllTypeNames returns a list of all type names that should be imported
// from the types file.
func collectAllTypeNames(schema *ir.Schema) []string {
	names := make(map[string]bool)

	// Domain Types
	for _, t := range schema.Types {
		names[t.Name] = true
	}

	// Enums
	for _, e := range schema.Enums {
		names[e.Name] = true
	}

	// Procedure Types (Input, Output, Response, Hydrate)
	for _, proc := range schema.Procedures {
		fullName := proc.FullName()
		names[fullName+"Input"] = true
		names[fullName+"Output"] = true
		names[fullName+"Response"] = true
		names["hydrate"+fullName+"Output"] = true
	}

	// Stream Types (Input, Output, Response, Hydrate)
	for _, stream := range schema.Streams {
		fullName := stream.FullName()
		names[fullName+"Input"] = true
		names[fullName+"Output"] = true
		names[fullName+"Response"] = true
		names["hydrate"+fullName+"Output"] = true
	}

	// Convert map to slice
	var result []string
	for name := range names {
		result = append(result, name)
	}
	return result
}

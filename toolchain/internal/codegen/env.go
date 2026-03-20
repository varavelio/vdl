package codegen

import (
	"os"
	"strings"
)

// isTruthyEnv reports whether an environment variable exists and is set to a
// common truthy value.
func isTruthyEnv(name string) bool {
	value, ok := os.LookupEnv(name)
	if !ok {
		return false
	}

	value = strings.TrimSpace(strings.ToLower(value))
	return value == "1" || value == "true" || value == "yes" || value == "y" || value == "on" || value == "enable" || value == "enabled"
}

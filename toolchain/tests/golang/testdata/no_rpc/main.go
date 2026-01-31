package main

import (
	"e2e/gen"
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	s := gen.Something{Field: "value"}
	if s.Field != "value" {
		panic("field mismatch")
	}

	// Verify catalog.go does not exist
	catalogPath := filepath.Join("gen", "catalog.go")
	if _, err := os.Stat(catalogPath); !os.IsNotExist(err) {
		panic("catalog.go should not exist")
	}

	fmt.Println("Success")
}

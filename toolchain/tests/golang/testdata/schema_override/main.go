package main

import (
	"fmt"
	"os"

	gen "schema_override/gen"
)

func main() {
	if !gen.OVERRIDE_WORKS {
		fmt.Println("OverrideWorks constant is false or missing")
		os.Exit(1)
	}
}

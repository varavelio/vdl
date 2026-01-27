package main

import (
	"fmt"
	"math"
	"os"

	"test/gen"
)

func main() {
	// Verify String constant
	if gen.VERSION != "1.2.3" {
		fail("VERSION", "1.2.3", gen.VERSION)
	}
	if gen.GREETING != "Hello, VDL!" {
		fail("GREETING", "Hello, VDL!", gen.GREETING)
	}

	// Verify Int constant
	if gen.MAX_RETRIES != 5 {
		fail("MAX_RETRIES", "5", fmt.Sprintf("%d", gen.MAX_RETRIES))
	}
	if gen.TIMEOUT_SECONDS != 30 {
		fail("TIMEOUT_SECONDS", "30", fmt.Sprintf("%d", gen.TIMEOUT_SECONDS))
	}

	// Verify Float constant
	if math.Abs(gen.PI-3.14159) > 1e-9 {
		fail("PI", "3.14159", fmt.Sprintf("%f", gen.PI))
	}

	// Verify Bool constant
	if gen.IS_ENABLED != true {
		fail("IS_ENABLED", "true", fmt.Sprintf("%v", gen.IS_ENABLED))
	}

	fmt.Println("Constants verification successful")
}

func fail(name, expected, actual string) {
	fmt.Fprintf(os.Stderr, "Constant %s mismatch: expected %q, got %q\n", name, expected, actual)
	os.Exit(1)
}

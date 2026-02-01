// Verifies constants are generated correctly in a separate constants.ts file.
// Import directly from constants.ts to verify the file exists
import {
  VERSION,
  MAX_RETRIES,
  TIMEOUT_SECONDS,
  PI,
  IS_ENABLED,
  GREETING,
} from "./gen/constants.ts";
// Also verify constants are exported via index.ts
import * as gen from "./gen/index.ts";

function fail(name: string, expected: unknown, actual: unknown): never {
  console.error(
    `Constant ${name} mismatch: expected ${expected}, got ${actual}`,
  );
  process.exit(1);
}

function main() {
  // Verify String constant
  if (VERSION !== "1.2.3") {
    fail("VERSION", "1.2.3", VERSION);
  }
  if (GREETING !== "Hello, VDL!") {
    fail("GREETING", "Hello, VDL!", GREETING);
  }

  // Verify Int constant
  if (MAX_RETRIES !== 5) {
    fail("MAX_RETRIES", 5, MAX_RETRIES);
  }
  if (TIMEOUT_SECONDS !== 30) {
    fail("TIMEOUT_SECONDS", 30, TIMEOUT_SECONDS);
  }

  // Verify Float constant
  if (Math.abs(PI - 3.14159) > 1e-9) {
    fail("PI", 3.14159, PI);
  }

  // Verify Bool constant
  if (IS_ENABLED !== true) {
    fail("IS_ENABLED", true, IS_ENABLED);
  }

  // Verify constants are exported via index.ts
  if (gen.VERSION !== "1.2.3") {
    fail("gen.VERSION", "1.2.3", gen.VERSION);
  }
  if (gen.MAX_RETRIES !== 5) {
    fail("gen.MAX_RETRIES", 5, gen.MAX_RETRIES);
  }

  console.log("Constants verification successful");
  process.exit(0);
}

main();

// Verifies constants are generated correctly.
import {
  VERSION,
  MAX_RETRIES,
  TIMEOUT_SECONDS,
  PI,
  IS_ENABLED,
  GREETING,
} from "./gen/index.ts";

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

  console.log("Constants verification successful");
  process.exit(0);
}

main();

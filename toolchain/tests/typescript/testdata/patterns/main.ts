// Verifies patterns are generated correctly in a separate patterns.ts file.
// Import directly from patterns.ts to verify the file exists
// Also verify patterns are exported via index.ts
import * as gen from "./gen/index.ts";
import {
  UserTopic,
  CacheKey,
  SimpleRoute,
  PrefixedPath,
  StaticPath,
  DuplicatedPlaceholder,
} from "./gen/patterns.ts";

function fail(name: string, expected: unknown, actual: unknown): never {
  console.error(
    `Pattern ${name} mismatch: expected ${expected}, got ${actual}`,
  );
  process.exit(1);
}

function main() {
  // Verify UserTopic pattern (multiple placeholders)
  const userTopic = UserTopic("123", "login");
  if (userTopic !== "events.users.123.login") {
    fail("UserTopic", "events.users.123.login", userTopic);
  }

  // Verify CacheKey pattern (multiple placeholders)
  const cacheKey = CacheKey("user", "profile-456");
  if (cacheKey !== "cache:user:profile-456") {
    fail("CacheKey", "cache:user:profile-456", cacheKey);
  }

  // Verify SimpleRoute pattern (single placeholder)
  const simpleRoute = SimpleRoute("users");
  if (simpleRoute !== "/api/v1/users") {
    fail("SimpleRoute", "/api/v1/users", simpleRoute);
  }

  // Verify PrefixedPath pattern (placeholder at start)
  const prefixedPath = PrefixedPath("myprefix");
  if (prefixedPath !== "myprefix/suffix") {
    fail("PrefixedPath", "myprefix/suffix", prefixedPath);
  }

  // Verify StaticPath pattern (no placeholders)
  const staticPath = StaticPath();
  if (staticPath !== "static/path/to/resource") {
    fail("StaticPath", "static/path/to/resource", staticPath);
  }

  // Verify patterns are exported via index.ts
  const genUserTopic = gen.UserTopic("abc", "logout");
  if (genUserTopic !== "events.users.abc.logout") {
    fail("gen.UserTopic", "events.users.abc.logout", genUserTopic);
  }

  // Verify DuplicatedPlaceholder pattern (duplicated placeholders)
  const duplicatedPlaceholder = DuplicatedPlaceholder("123", "login");
  if (duplicatedPlaceholder !== "users.123.login.123") {
    fail("UserTopic", "users.123.login.123", duplicatedPlaceholder);
  }

  console.log("Patterns verification successful");
  process.exit(0);
}

main();

// Verifies pattern generation: pattern templates with placeholders should
// generate functions that construct the correct strings.
import {
  DuplicatedSegment,
  SessionCacheKey,
  SimpleKey,
  TopicChannel,
  UserEventSubject,
} from "./gen/patterns.ts";

function fail(name: string, expected: string, actual: string): never {
  console.error(`${name}: expected ${JSON.stringify(expected)}, got ${JSON.stringify(actual)}`);
  process.exit(1);
}

function main() {
  // Test pattern with two placeholders
  let result = UserEventSubject("user123", "created");
  let expected = "events.users.user123.created";
  if (result !== expected) {
    fail("UserEventSubject", expected, result);
  }

  // Test pattern with one placeholder
  result = SessionCacheKey("sess-abc");
  expected = "cache:session:sess-abc";
  if (result !== expected) {
    fail("SessionCacheKey", expected, result);
  }

  // Test pattern with two placeholders and different separators
  result = TopicChannel("chan-1", "msg-42");
  expected = "channels.chan-1.messages.msg-42";
  if (result !== expected) {
    fail("TopicChannel", expected, result);
  }

  // Test static pattern (no placeholders)
  result = SimpleKey();
  expected = "static-key";
  if (result !== expected) {
    fail("SimpleKey", expected, result);
  }

  // Test with empty strings
  result = UserEventSubject("", "");
  expected = "events.users..";
  if (result !== expected) {
    fail("UserEventSubject with empty", expected, result);
  }

  // Test with special characters
  result = UserEventSubject("user/123", "event:type");
  expected = "events.users.user/123.event:type";
  if (result !== expected) {
    fail("UserEventSubject with special chars", expected, result);
  }

  // Test duplicated segments
  // Pattern: "events.users.{userId}.{eventType}.{userId}"
  // Expected function signature: DuplicatedSegment(userId: string, eventType: string): string
  // Notice that userId appears twice in the pattern but only once in the arguments.
  result = DuplicatedSegment("user123", "login");
  expected = "events.users.user123.login.user123";
  if (result !== expected) {
    fail("DuplicatedSegment", expected, result);
  }

  console.log("Success");
  process.exit(0);
}

main();

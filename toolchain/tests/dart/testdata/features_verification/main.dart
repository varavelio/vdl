// Dart E2E test for features verification
// Verifies that constants and patterns are generated correctly

import 'gen/index.dart';

void main() {
  testIntConstants();
  testStringConstants();
  testFloatConstants();
  testBoolConstants();
  testPatternFunctions();
  testPatternInterpolation();

  print('Success');
}

void testIntConstants() {
  assert(MAX_PAGE_SIZE == 100, 'MAX_PAGE_SIZE should be 100');
  assert(MIN_PAGE_SIZE == 10, 'MIN_PAGE_SIZE should be 10');

  // Verify types
  assert(MAX_PAGE_SIZE is int, 'MAX_PAGE_SIZE should be int');
  assert(MIN_PAGE_SIZE is int, 'MIN_PAGE_SIZE should be int');
}

void testStringConstants() {
  assert(API_VERSION == '2.0.0', 'API_VERSION should be "2.0.0"');

  // Verify type
  assert(API_VERSION is String, 'API_VERSION should be String');
}

void testFloatConstants() {
  assert(PI_APPROX == 3.14159, 'PI_APPROX should be 3.14159');

  // Verify type
  assert(PI_APPROX is double, 'PI_APPROX should be double');
}

void testBoolConstants() {
  assert(DEBUG_MODE == false, 'DEBUG_MODE should be false');
  assert(ENABLE_CACHE == true, 'ENABLE_CACHE should be true');

  // Verify types
  assert(DEBUG_MODE is bool, 'DEBUG_MODE should be bool');
  assert(ENABLE_CACHE is bool, 'ENABLE_CACHE should be bool');
}

void testPatternFunctions() {
  // Test UserEventSubject pattern
  final userEvent = UserEventSubject('user123', 'login');
  assert(
    userEvent == 'events.users.user123.login',
    'UserEventSubject pattern failed: $userEvent',
  );

  // Test CacheKey pattern
  final cacheKey = CacheKey('sessions', 'abc123');
  assert(
    cacheKey == 'cache:sessions:abc123',
    'CacheKey pattern failed: $cacheKey',
  );

  // Test ApiEndpoint pattern
  final endpoint = ApiEndpoint('2', 'users', '42');
  assert(
    endpoint == '/api/v2/users/42',
    'ApiEndpoint pattern failed: $endpoint',
  );
}

void testPatternInterpolation() {
  // Test with special characters that need escaping
  final keyWithSpecial = CacheKey('user:data', 'key/with/slashes');
  assert(
    keyWithSpecial == 'cache:user:data:key/with/slashes',
    'Pattern with special chars failed',
  );

  // Test with empty strings
  final emptyParts = CacheKey('', '');
  assert(emptyParts == 'cache::', 'Pattern with empty strings failed');

  // Test with numbers as strings
  final numeric = UserEventSubject('123', '456');
  assert(
    numeric == 'events.users.123.456',
    'Pattern with numeric strings failed',
  );

  // Test duplicated segments
  final ds = DuplicatedSegment('123', 'abc');
  assert(ds == 'users.123.abc.123', 'Pattern with dup-segment failed');
}

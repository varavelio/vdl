// Dart E2E test for numeric stability
// Verifies that int/double coercion works correctly when JSON has
// "dirty" numeric types (integers as doubles and vice versa).

import 'gen/index.dart';

void main() {
  testIntAsDouble();
  testDoubleAsInt();
  testOptionalNumerics();
  testNumericArrays();
  testNumericMaps();
  testRoundTrip();

  print('Success');
}

void testIntAsDouble() {
  // Simulate JSON from JavaScript where integers come as doubles
  final dirtyJson = <String, dynamic>{
    'intField': 42.0, // int field receiving a double
    'floatField': 3.14,
  };

  final data = NumericData.fromJson(dirtyJson);

  assert(data.intField == 42, 'intField should be 42, got ${data.intField}');
  assert(data.intField is int, 'intField should be int type');
  assert(data.floatField == 3.14, 'floatField should be 3.14');
  assert(data.floatField is double, 'floatField should be double type');
}

void testDoubleAsInt() {
  // Simulate JSON where floats come as integers
  final dirtyJson = <String, dynamic>{
    'intField': 100,
    'floatField': 10, // float field receiving an int
  };

  final data = NumericData.fromJson(dirtyJson);

  assert(data.intField == 100, 'intField should be 100');
  assert(data.floatField == 10.0, 'floatField should be 10.0');
  assert(data.floatField is double, 'floatField should be double type');
}

void testOptionalNumerics() {
  // Test with optional fields present as wrong types
  final dirtyJson = <String, dynamic>{
    'intField': 1.0, // int as double
    'floatField': 2, // float as int
    'optionalInt': 3.0, // optional int as double
    'optionalFloat': 4, // optional float as int
  };

  final data = NumericData.fromJson(dirtyJson);

  assert(data.intField == 1, 'intField should be 1');
  assert(data.floatField == 2.0, 'floatField should be 2.0');
  assert(data.optionalInt == 3, 'optionalInt should be 3');
  assert(data.optionalFloat == 4.0, 'optionalFloat should be 4.0');

  // Test with optional fields missing
  final minimalJson = <String, dynamic>{'intField': 5, 'floatField': 6.0};

  final minimal = NumericData.fromJson(minimalJson);
  assert(minimal.optionalInt == null, 'optionalInt should be null');
  assert(minimal.optionalFloat == null, 'optionalFloat should be null');
}

void testNumericArrays() {
  // Test arrays with mixed numeric types
  final dirtyJson = <String, dynamic>{
    'intArray': [1.0, 2.0, 3.0], // int array with doubles
    'floatArray': [1, 2, 3], // float array with ints
  };

  final data = NumericArrays.fromJson(dirtyJson);

  assert(data.intArray.length == 3, 'intArray should have 3 elements');
  assert(data.intArray[0] == 1, 'intArray[0] should be 1');
  assert(data.intArray[1] == 2, 'intArray[1] should be 2');
  assert(data.intArray[2] == 3, 'intArray[2] should be 3');
  assert(data.intArray[0] is int, 'intArray elements should be int type');

  assert(data.floatArray.length == 3, 'floatArray should have 3 elements');
  assert(data.floatArray[0] == 1.0, 'floatArray[0] should be 1.0');
  assert(data.floatArray[1] == 2.0, 'floatArray[1] should be 2.0');
  assert(data.floatArray[2] == 3.0, 'floatArray[2] should be 3.0');
  assert(
    data.floatArray[0] is double,
    'floatArray elements should be double type',
  );
}

void testNumericMaps() {
  // Test maps with mixed numeric types
  final dirtyJson = <String, dynamic>{
    'intMap': {'a': 1.0, 'b': 2.0}, // int map with doubles
    'floatMap': {'x': 1, 'y': 2}, // float map with ints
  };

  final data = NumericMaps.fromJson(dirtyJson);

  assert(data.intMap['a'] == 1, 'intMap["a"] should be 1');
  assert(data.intMap['b'] == 2, 'intMap["b"] should be 2');
  assert(data.intMap['a'] is int, 'intMap values should be int type');

  assert(data.floatMap['x'] == 1.0, 'floatMap["x"] should be 1.0');
  assert(data.floatMap['y'] == 2.0, 'floatMap["y"] should be 2.0');
  assert(data.floatMap['x'] is double, 'floatMap values should be double type');
}

void testRoundTrip() {
  // Create instance, serialize, deserialize, verify
  final original = NumericData(
    intField: 42,
    floatField: 3.14159,
    optionalInt: 100,
    optionalFloat: 2.718,
  );

  final json = original.toJson();
  final restored = NumericData.fromJson(json);

  assert(restored.intField == original.intField, 'intField round-trip failed');
  assert(
    restored.floatField == original.floatField,
    'floatField round-trip failed',
  );
  assert(
    restored.optionalInt == original.optionalInt,
    'optionalInt round-trip failed',
  );
  assert(
    restored.optionalFloat == original.optionalFloat,
    'optionalFloat round-trip failed',
  );

  assert(restored == original, 'Equality check failed after round-trip');
}

// Dart E2E test for primitives
// This test verifies that all primitive types are generated correctly
// and can be serialized/deserialized properly.

import 'gen/index.dart';

void main() {
  // Test AllPrimitives type
  testAllPrimitives();

  // Test OptionalPrimitives type
  testOptionalPrimitives();

  // Test procedure types
  testProcedureTypes();

  print('Success');
}

void testAllPrimitives() {
  final now = DateTime.now().toUtc();

  // Create an instance
  final original = AllPrimitives(
    stringField: 'hello',
    intField: 42,
    floatField: 3.14,
    boolField: true,
    datetimeField: now,
  );

  // Serialize to JSON
  final json = original.toJson();

  // Verify JSON structure
  assert(json['stringField'] == 'hello', 'stringField mismatch');
  assert(json['intField'] == 42, 'intField mismatch');
  assert(json['floatField'] == 3.14, 'floatField mismatch');
  assert(json['boolField'] == true, 'boolField mismatch');
  assert(
    json['datetimeField'] is String,
    'datetimeField should be ISO8601 string',
  );

  // Deserialize from JSON
  final deserialized = AllPrimitives.fromJson(json);

  // Verify fields match
  assert(
    deserialized.stringField == original.stringField,
    'stringField round-trip failed',
  );
  assert(
    deserialized.intField == original.intField,
    'intField round-trip failed',
  );
  assert(
    deserialized.floatField == original.floatField,
    'floatField round-trip failed',
  );
  assert(
    deserialized.boolField == original.boolField,
    'boolField round-trip failed',
  );
  // DateTime comparison (within 1 second tolerance for milliseconds)
  assert(
    deserialized.datetimeField
            .difference(original.datetimeField)
            .inSeconds
            .abs() <
        1,
    'datetimeField round-trip failed',
  );

  // Test equality
  assert(
    original == original,
    '== operator should return true for same instance',
  );

  // Test toString
  assert(
    original.toString().contains('AllPrimitives'),
    'toString should contain class name',
  );

  // Test copyWith
  final modified = original.copyWith(stringField: 'world');
  assert(modified.stringField == 'world', 'copyWith stringField failed');
  assert(
    modified.intField == original.intField,
    'copyWith should preserve other fields',
  );

  // Test hashCode
}

void testOptionalPrimitives() {
  // Test with all fields null
  final empty = OptionalPrimitives();
  final emptyJson = empty.toJson();
  assert(emptyJson.isEmpty, 'Empty optional type should produce empty JSON');

  // Test with some fields set
  final partial = OptionalPrimitives(stringField: 'test', intField: 123);
  final partialJson = partial.toJson();
  assert(partialJson['stringField'] == 'test', 'stringField should be present');
  assert(partialJson['intField'] == 123, 'intField should be present');
  assert(
    !partialJson.containsKey('floatField'),
    'floatField should not be present',
  );
  assert(
    !partialJson.containsKey('boolField'),
    'boolField should not be present',
  );
  assert(
    !partialJson.containsKey('datetimeField'),
    'datetimeField should not be present',
  );

  // Deserialize partial JSON
  final deserializedPartial = OptionalPrimitives.fromJson(partialJson);
  assert(
    deserializedPartial.stringField == 'test',
    'stringField deserialization failed',
  );
  assert(
    deserializedPartial.intField == 123,
    'intField deserialization failed',
  );
  assert(deserializedPartial.floatField == null, 'floatField should be null');
  assert(deserializedPartial.boolField == null, 'boolField should be null');
  assert(
    deserializedPartial.datetimeField == null,
    'datetimeField should be null',
  );
}

void testProcedureTypes() {
  // Test ServiceEchoInput
  final now = DateTime.now().toUtc();
  final input = ServiceEchoInput(
    data: AllPrimitives(
      stringField: 'test',
      intField: 1,
      floatField: 1.5,
      boolField: false,
      datetimeField: now,
    ),
  );

  final inputJson = input.toJson();
  assert(inputJson['data'] is Map, 'data should be a Map in JSON');

  final inputDeserialized = ServiceEchoInput.fromJson(inputJson);
  assert(
    inputDeserialized.data.stringField == 'test',
    'Nested type deserialization failed',
  );

  // Test ServiceEchoOutput
  final output = ServiceEchoOutput(
    data: AllPrimitives(
      stringField: 'response',
      intField: 2,
      floatField: 2.5,
      boolField: true,
      datetimeField: now,
    ),
  );

  final outputJson = output.toJson();
  assert(outputJson['data'] is Map, 'data should be a Map in JSON');

  // Test Response type
  final response = Response<ServiceEchoOutput>.ok(output);
  assert(response.ok == true, 'Response.ok should be true');
  assert(
    response.output?.data.stringField == 'response',
    'Response output should be accessible',
  );

  final errorResponse = Response<ServiceEchoOutput>.error(
    VdlError(message: 'Something went wrong', code: 'ERR001'),
  );
  assert(errorResponse.ok == false, 'Error response.ok should be false');
  assert(
    errorResponse.error?.message == 'Something went wrong',
    'Error message should be accessible',
  );
  assert(
    errorResponse.error?.code == 'ERR001',
    'Error code should be accessible',
  );
}

// Dart E2E test for malformed JSON handling
// Verifies that the generated code fails gracefully with bad data

import 'gen/index.dart';

void main() {
  testMissingRequiredField();
  testWrongTypeField();
  testNestedMalformed();
  testListWrongType();
  testValidJsonWorks();

  print('Success');
}

void testMissingRequiredField() {
  // Missing 'age' field
  final malformedJson = <String, dynamic>{
    'name': 'John',
    'active': true,
    // 'age' is missing
  };

  bool threwError = false;
  try {
    RequiredFields.fromJson(malformedJson);
  } catch (e) {
    threwError = true;
    // Verify it's a type-related error
    assert(
      e is TypeError || e is NoSuchMethodError || e is Error,
      'Should throw a type-related error, got: ${e.runtimeType}',
    );
  }

  assert(threwError, 'Should throw error when required field is missing');
}

void testWrongTypeField() {
  // 'age' is a string instead of int
  final malformedJson = <String, dynamic>{
    'name': 'John',
    'age': 'not a number', // Should be int
    'active': true,
  };

  bool threwError = false;
  try {
    RequiredFields.fromJson(malformedJson);
  } catch (e) {
    threwError = true;
    // Should be a type error or cast error
    assert(
      e is TypeError || e is Error,
      'Should throw type error, got: ${e.runtimeType}',
    );
  }

  assert(threwError, 'Should throw error when field has wrong type');
}

void testNestedMalformed() {
  // Nested child is missing required fields
  final malformedJson = <String, dynamic>{
    'child': {
      'name': 'Jane',
      // 'age' and 'active' missing
    },
  };

  bool threwError = false;
  try {
    NestedRequired.fromJson(malformedJson);
  } catch (e) {
    threwError = true;
  }

  assert(threwError, 'Should throw error when nested object is malformed');
}

void testListWrongType() {
  // 'items' should be a list but we pass a string
  final malformedJson = <String, dynamic>{'items': 'not a list'};

  bool threwError = false;
  try {
    WithList.fromJson(malformedJson);
  } catch (e) {
    threwError = true;
    assert(
      e is TypeError || e is Error,
      'Should throw type error for list mismatch',
    );
  }

  assert(threwError, 'Should throw error when list field is not a list');
}

void testValidJsonWorks() {
  // Ensure valid JSON still works after all these failure tests
  final validJson = <String, dynamic>{
    'name': 'Alice',
    'age': 30,
    'active': true,
  };

  final data = RequiredFields.fromJson(validJson);
  assert(data.name == 'Alice', 'Valid JSON should work');
  assert(data.age == 30, 'Valid JSON should work');
  assert(data.active == true, 'Valid JSON should work');
}

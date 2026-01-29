// Dart E2E test for null safety boundaries
// Verifies proper handling of optional fields, null values, and empty JSON

import 'gen/index.dart';

void main() {
  testOptionalListNull();
  testOptionalMapNull();
  testOptionalNestedNull();
  testAllOptionalEmpty();
  testDeepOptionalNull();
  testToJsonOmitsNull();
  testFromEmptyJson();

  print('Success');
}

void testOptionalListNull() {
  // Deserialize with missing optional list
  final json = <String, dynamic>{};
  final data = WithOptionalList.fromJson(json);

  assert(data.items == null, 'Optional list should be null when missing');

  // Deserialize with null value
  final jsonWithNull = <String, dynamic>{'items': null};
  final dataWithNull = WithOptionalList.fromJson(jsonWithNull);
  assert(dataWithNull.items == null, 'Optional list should be null when null');

  // Deserialize with empty list
  final jsonWithEmpty = <String, dynamic>{'items': []};
  final dataWithEmpty = WithOptionalList.fromJson(jsonWithEmpty);
  assert(dataWithEmpty.items != null, 'Empty list should not be null');
  assert(dataWithEmpty.items!.isEmpty, 'Empty list should be empty');
}

void testOptionalMapNull() {
  // Deserialize with missing optional map
  final json = <String, dynamic>{};
  final data = WithOptionalMap.fromJson(json);

  assert(data.metadata == null, 'Optional map should be null when missing');

  // Deserialize with null value
  final jsonWithNull = <String, dynamic>{'metadata': null};
  final dataWithNull = WithOptionalMap.fromJson(jsonWithNull);
  assert(dataWithNull.metadata == null, 'Optional map should be null');

  // Deserialize with empty map
  final jsonWithEmpty = <String, dynamic>{'metadata': {}};
  final dataWithEmpty = WithOptionalMap.fromJson(jsonWithEmpty);
  assert(dataWithEmpty.metadata != null, 'Empty map should not be null');
  assert(dataWithEmpty.metadata!.isEmpty, 'Empty map should be empty');
}

void testOptionalNestedNull() {
  // Deserialize with missing optional nested type
  final json = <String, dynamic>{};
  final data = WithOptionalNested.fromJson(json);

  assert(data.child == null, 'Optional nested should be null when missing');

  // Deserialize with null value
  final jsonWithNull = <String, dynamic>{'child': null};
  final dataWithNull = WithOptionalNested.fromJson(jsonWithNull);
  assert(dataWithNull.child == null, 'Optional nested should be null');

  // Deserialize with value
  final jsonWithValue = <String, dynamic>{
    'child': {'name': 'test', 'value': 42},
  };
  final dataWithValue = WithOptionalNested.fromJson(jsonWithValue);
  assert(dataWithValue.child != null, 'Nested should not be null');
  assert(dataWithValue.child!.name == 'test', 'Nested name mismatch');
  assert(dataWithValue.child!.value == 42, 'Nested value mismatch');
}

void testAllOptionalEmpty() {
  // Create with all nulls
  final empty = AllOptional();

  assert(empty.stringField == null, 'stringField should be null');
  assert(empty.intField == null, 'intField should be null');
  assert(empty.listField == null, 'listField should be null');
  assert(empty.mapField == null, 'mapField should be null');
  assert(empty.nestedField == null, 'nestedField should be null');

  // Serialize and verify empty JSON
  final json = empty.toJson();
  assert(json.isEmpty, 'All-null object should produce empty JSON');
}

void testDeepOptionalNull() {
  // All levels null
  final deep = DeepOptional();
  assert(deep.level1 == null, 'level1 should be null');

  // Level 1 present, level 2 null
  final partial = DeepOptional(level1: Level1());
  assert(partial.level1 != null, 'level1 should not be null');
  assert(partial.level1!.level2 == null, 'level2 should be null');

  // All levels present
  final full = DeepOptional(
    level1: Level1(level2: Level2(value: 'deep')),
  );
  assert(full.level1!.level2!.value == 'deep', 'Deep value mismatch');

  // Round-trip
  final json = full.toJson();
  final restored = DeepOptional.fromJson(json);
  assert(restored.level1!.level2!.value == 'deep', 'Deep round-trip failed');
}

void testToJsonOmitsNull() {
  // Verify that null fields are omitted from JSON
  final partial = AllOptional(stringField: 'test', intField: 42);
  final json = partial.toJson();

  assert(json.containsKey('stringField'), 'stringField should be present');
  assert(json.containsKey('intField'), 'intField should be present');
  assert(!json.containsKey('listField'), 'null listField should be omitted');
  assert(!json.containsKey('mapField'), 'null mapField should be omitted');
  assert(
    !json.containsKey('nestedField'),
    'null nestedField should be omitted',
  );

  assert(json.length == 2, 'JSON should only have 2 keys');
}

void testFromEmptyJson() {
  // Test that fromJson handles completely empty JSON for all-optional types
  final json = <String, dynamic>{};
  final data = AllOptional.fromJson(json);

  assert(data.stringField == null, 'stringField should be null from empty');
  assert(data.intField == null, 'intField should be null from empty');
  assert(data.listField == null, 'listField should be null from empty');
  assert(data.mapField == null, 'mapField should be null from empty');
  assert(data.nestedField == null, 'nestedField should be null from empty');
}

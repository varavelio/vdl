// Dart E2E test for field name handling
// Verifies that camelCase field names work correctly with JSON serialization

import 'gen/index.dart';

void main() {
  testFieldsCompile();
  testSerialization();
  testOptionalFields();

  print('Success');
}

void testFieldsCompile() {
  // If this compiles, field name handling works
  final data = SafeFields(
    finalValue: 'final_value',
    varName: 'var_value',
    constData: 'const_value',
    className: 'class_value',
    isActive: true,
    inRange: false,
    withPrefix: 'with_value',
    asType: 'as_value',
    newItem: 'new_value',
    defaultVal: 'default_value',
  );

  // Access each field to verify they work
  assert(data.finalValue == 'final_value', 'finalValue access failed');
  assert(data.varName == 'var_value', 'varName access failed');
  assert(data.constData == 'const_value', 'constData access failed');
  assert(data.className == 'class_value', 'className access failed');
  assert(data.isActive == true, 'isActive access failed');
  assert(data.inRange == false, 'inRange access failed');
  assert(data.withPrefix == 'with_value', 'withPrefix access failed');
  assert(data.asType == 'as_value', 'asType access failed');
  assert(data.newItem == 'new_value', 'newItem access failed');
  assert(data.defaultVal == 'default_value', 'defaultVal access failed');
}

void testSerialization() {
  final data = SafeFields(
    finalValue: 'a',
    varName: 'b',
    constData: 'c',
    className: 'd',
    isActive: true,
    inRange: false,
    withPrefix: 'g',
    asType: 'h',
    newItem: 'i',
    defaultVal: 'j',
  );

  final json = data.toJson();

  // JSON keys should be camelCase
  assert(json['finalValue'] == 'a', 'JSON key should be "finalValue"');
  assert(json['varName'] == 'b', 'JSON key should be "varName"');
  assert(json['constData'] == 'c', 'JSON key should be "constData"');
  assert(json['className'] == 'd', 'JSON key should be "className"');
  assert(json['isActive'] == true, 'JSON key should be "isActive"');
  assert(json['inRange'] == false, 'JSON key should be "inRange"');
  assert(json['withPrefix'] == 'g', 'JSON key should be "withPrefix"');
  assert(json['asType'] == 'h', 'JSON key should be "asType"');
  assert(json['newItem'] == 'i', 'JSON key should be "newItem"');
  assert(json['defaultVal'] == 'j', 'JSON key should be "defaultVal"');

  // Round-trip test
  final restored = SafeFields.fromJson(json);
  assert(
    restored.finalValue == data.finalValue,
    'Round-trip failed for finalValue',
  );
  assert(restored.varName == data.varName, 'Round-trip failed for varName');
  assert(
    restored.className == data.className,
    'Round-trip failed for className',
  );
}

void testOptionalFields() {
  // Test with all null
  final empty = OptionalSafe();
  final emptyJson = empty.toJson();
  assert(emptyJson.isEmpty, 'Empty should produce empty JSON');

  // Test with values
  final data = OptionalSafe(finalValue: 'test', varName: 42, className: true);
  final json = data.toJson();

  assert(json['finalValue'] == 'test', 'Optional finalValue serialization');
  assert(json['varName'] == 42, 'Optional varName serialization');
  assert(json['className'] == true, 'Optional className serialization');

  // Deserialize
  final restored = OptionalSafe.fromJson(json);
  assert(restored.finalValue == 'test', 'Optional finalValue deserialization');
  assert(restored.varName == 42, 'Optional varName deserialization');
  assert(restored.className == true, 'Optional className deserialization');
}

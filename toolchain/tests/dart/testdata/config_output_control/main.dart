// Dart E2E test for config output control
// Verifies that gen_consts: false and gen_patterns: false work correctly
//
// This test verifies:
// 1. The code compiles (index.dart doesn't try to export non-existent files)
// 2. Types still work
// 3. Constants and patterns are NOT generated

import 'dart:io';
import 'gen/index.dart';

void main() {
  testTypesStillWork();
  testConstantsFileNotGenerated();
  testPatternsFileNotGenerated();
  testIndexDoesNotExportMissing();

  print('Success');
}

void testTypesStillWork() {
  // Types should still be generated
  final data = SimpleType(id: '123', name: 'Test');

  assert(data.id == '123', 'Type should work');
  assert(data.name == 'Test', 'Type should work');

  // Round trip
  final json = data.toJson();
  final restored = SimpleType.fromJson(json);
  assert(restored.id == data.id, 'Round trip should work');
}

void testConstantsFileNotGenerated() {
  // constants.dart should NOT exist
  final constantsFile = File('gen/constants.dart');
  assert(
    !constantsFile.existsSync(),
    'constants.dart should NOT be generated when gen_consts: false',
  );
}

void testPatternsFileNotGenerated() {
  // patterns.dart should NOT exist
  final patternsFile = File('gen/patterns.dart');
  assert(
    !patternsFile.existsSync(),
    'patterns.dart should NOT be generated when gen_patterns: false',
  );
}

void testIndexDoesNotExportMissing() {
  // If index.dart tried to export constants.dart or patterns.dart,
  // this file wouldn't compile at all (which is the test).
  // If we reach this point, it means index.dart is correct.

  // Read index.dart and verify it doesn't contain exports for missing files
  final indexFile = File('gen/index.dart');
  if (indexFile.existsSync()) {
    final content = indexFile.readAsStringSync();
    assert(
      !content.contains("export 'constants.dart'"),
      'index.dart should not export constants.dart',
    );
    assert(
      !content.contains("export 'patterns.dart'"),
      'index.dart should not export patterns.dart',
    );
  }
}

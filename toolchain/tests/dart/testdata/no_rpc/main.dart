import 'dart:io';
import 'gen/index.dart';

void main() {
  final s = Something(field: 'value');
  if (s.field != 'value') throw 'field mismatch';

  // Verify catalog.dart does not exist
  if (File('gen/catalog.dart').existsSync()) {
    throw 'catalog.dart should not exist';
  }

  print('Success');
}

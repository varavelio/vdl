import 'gen/index.dart';

void main() {
  // 1. Array of inline objects
  final files = [ComplexInlineTypesFiles(path: 'a', content: 'b')];

  // 2. Map of inline objects
  final meta = {'v1': ComplexInlineTypesMeta(createdAt: '2023', author: 'me')};

  // 3. Map of arrays of inline objects
  final groupedFiles = {
    'group1': [ComplexInlineTypesGroupedFiles(name: 'f1', size: 10)],
  };

  // 4. Array of maps of inline objects
  final configs = [
    {'conf1': ComplexInlineTypesConfigs(key: 'k', value: 'v')},
  ];

  // 5. Nested arrays of inline objects
  final grid = [
    [ComplexInlineTypesGrid(x: 1, y: 2)],
  ];

  // 6. Simple inline object
  final simple = ComplexInlineTypesSimple(name: 'test', enabled: true);

  // 7. Deeply nested inline objects
  final deepNest = ComplexInlineTypesDeepNest(
    level1: 'l1',
    child: ComplexInlineTypesDeepNestChild(
      level2: 2,
      grandChild: ComplexInlineTypesDeepNestChildGrandChild(
        level3: true,
        greatGrandChild:
            ComplexInlineTypesDeepNestChildGrandChildGreatGrandChild(
              level4: 4.5,
              data: 'end',
            ),
      ),
    ),
  );

  final output = ComplexInlineTypes(
    files: files,
    meta: meta,
    groupedFiles: groupedFiles,
    configs: configs,
    grid: grid,
    simple: simple,
    deepNest: deepNest,
  );

  // Verify integrity
  if (output.files[0].path != 'a') throw 'files mismatch';
  if (output.meta['v1']!.author != 'me') throw 'meta mismatch';
  if (output.groupedFiles['group1']![0].size != 10)
    throw 'groupedFiles mismatch';
  if (output.configs[0]['conf1']!.value != 'v') throw 'configs mismatch';
  if (output.grid[0][0].y != 2) throw 'grid mismatch';
  if (output.simple.name != 'test') throw 'simple mismatch';
  if (output.deepNest.child.grandChild.greatGrandChild.data != 'end')
    throw 'deepNest mismatch';

  print('Success');
}

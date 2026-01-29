import 'dart:convert';
import 'gen/index.dart';

void main() {
  testAllPrimitives();
  testMediumComplex();
  testNightmareConstruction();
  testNightmareRoundTrip();
  testDeepAbyss();
  testOptionalMadness();
  print('All extreme_nesting tests passed!');
}

void testAllPrimitives() {
  final now = DateTime.now().toUtc();
  final ap = AllPrimitives(s: 'test', i: 42, f: 3.14, b: true, d: now);

  final json = ap.toJson();
  assert(json['s'] == 'test', 's failed');
  assert(json['i'] == 42, 'i failed');
  assert(json['f'] == 3.14, 'f failed');
  assert(json['b'] == true, 'b failed');

  final parsed = AllPrimitives.fromJson(json);
  assert(parsed.s == 'test', 'parsed s failed');
  assert(parsed.i == 42, 'parsed i failed');
}

void testMediumComplex() {
  final now = DateTime.now().toUtc();

  final medium = MediumComplex(
    grid: [
      [
        MediumComplexGrid(x: 0, y: 0, data: {'key': 'val00'}),
        MediumComplexGrid(x: 0, y: 1, data: {'key': 'val01'}),
      ],
      [
        MediumComplexGrid(x: 1, y: 0, data: {'key': 'val10'}),
      ],
    ],
    catalog: {
      'items': [
        MediumComplexCatalog(
          name: 'Item1',
          count: 10,
          price: 99.99,
          available: true,
          lastUpdate: now,
        ),
      ],
    },
    config: MediumComplexConfig(
      inner: MediumComplexConfigInner(value: 'nested'),
    ),
  );

  final json = medium.toJson();
  final grid = json['grid'] as List;
  assert(grid.length == 2, 'grid rows');
  assert((grid[0] as List).length == 2, 'grid[0] cols');

  final catalog = json['catalog'] as Map;
  assert((catalog['items'] as List).length == 1, 'catalog items');

  // Round trip
  final jsonStr = jsonEncode(json);
  final parsed = MediumComplex.fromJson(jsonDecode(jsonStr));
  assert(parsed.grid.length == 2, 'parsed grid');
  assert(parsed.config?.inner?.value == 'nested', 'parsed config');
}

void testNightmareConstruction() {
  final now = DateTime.now().toUtc();

  // Build the hyperCube - 4D array of inline objects
  final hyperCubeCell = NightmareHyperCube(
    coords: [1, 2, 3],
    data: {
      'key1': NightmareHyperCubeData(values: [1.1, 2.2], flags: [true, false]),
    },
  );

  // Build tripleNestedMap
  final tripleMap = {
    'l1': {
      'l2': {
        'l3': NightmareTripleNestedMap(timestamps: [now], active: true),
      },
    },
  };

  // Build chaosArray
  final chaosItem = NightmareChaosArray(
    str: 'chaos',
    number: 666,
    dec: 6.66,
    flag: true,
    time: now,
    nested: NightmareChaosArrayNested(
      deep: 'deep',
      deeper: NightmareChaosArrayNestedDeeper(deepest: 999),
    ),
  );

  final nightmare = Nightmare(
    hyperCube: [
      [
        [
          [hyperCubeCell],
        ],
      ],
    ],
    tripleNestedMap: tripleMap,
    chaosArray: [
      {
        'chaos': [chaosItem],
      },
    ],
    optionalMadness: null,
    abyss: NightmareAbyss(
      level1: NightmareAbyssLevel1(
        level2: NightmareAbyssLevel1Level2(
          level3: NightmareAbyssLevel1Level2Level3(
            level4: NightmareAbyssLevel1Level2Level3Level4(
              level5: NightmareAbyssLevel1Level2Level3Level4Level5(
                bottom: 'BOTTOM',
                allTypes: AllPrimitives(s: 's', i: 1, f: 1.0, b: true, d: now),
              ),
            ),
          ),
        ),
      ),
    ),
    matrixOfMapsOfArrays: [
      [
        {
          'row': NightmareMatrixOfMapsOfArrays(
            items: [
              NightmareMatrixOfMapsOfArraysItems(id: 1, tags: ['a', 'b']),
            ],
          ),
        },
      ],
    ],
    primitiveMap: {
      'partial': NightmarePrimitiveMap(
        s: 'only string',
        i: null,
        f: null,
        b: null,
        d: null,
      ),
      'full': NightmarePrimitiveMap(s: 's', i: 1, f: 1.1, b: true, d: now),
    },
  );

  final json = nightmare.toJson();
  assert(json['hyperCube'] != null, 'hyperCube exists');
  assert(json['tripleNestedMap'] != null, 'tripleNestedMap exists');
  assert(json['chaosArray'] != null, 'chaosArray exists');
  assert(!json.containsKey('optionalMadness'), 'optionalMadness omitted');
  assert(json['abyss'] != null, 'abyss exists');
}

void testNightmareRoundTrip() {
  final now = DateTime.utc(2025, 1, 1, 12, 0, 0);

  final nightmare = Nightmare(
    hyperCube: [
      [
        [
          [
            NightmareHyperCube(
              coords: [1],
              data: {
                'k': NightmareHyperCubeData(values: [1.0], flags: [true]),
              },
            ),
          ],
        ],
      ],
    ],
    tripleNestedMap: {
      'a': {
        'b': {
          'c': NightmareTripleNestedMap(timestamps: [now], active: false),
        },
      },
    },
    chaosArray: [
      {
        'x': [
          NightmareChaosArray(
            str: 'x',
            number: 1,
            dec: 1.0,
            flag: false,
            time: now,
            nested: null,
          ),
        ],
      },
    ],
    optionalMadness: {
      'opt': [
        NightmareOptionalMadness(
          required: 'req',
          opt1: 1,
          opt2: null,
          opt3: true,
          opt4: null,
          optNested: NightmareOptionalMadnessOptNested(also: 'also'),
        ),
      ],
    },
    abyss: NightmareAbyss(
      level1: NightmareAbyssLevel1(
        level2: NightmareAbyssLevel1Level2(
          level3: NightmareAbyssLevel1Level2Level3(
            level4: NightmareAbyssLevel1Level2Level3Level4(
              level5: NightmareAbyssLevel1Level2Level3Level4Level5(
                bottom: 'end',
                allTypes: AllPrimitives(s: '', i: 0, f: 0.0, b: false, d: now),
              ),
            ),
          ),
        ),
      ),
    ),
    matrixOfMapsOfArrays: [],
    primitiveMap: {},
  );

  // Full round trip
  final jsonStr = jsonEncode(nightmare.toJson());
  final parsed = Nightmare.fromJson(jsonDecode(jsonStr));

  // Verify deep structures survived
  assert(parsed.hyperCube[0][0][0][0].coords[0] == 1, 'hyperCube roundtrip');
  assert(
    parsed.tripleNestedMap['a']!['b']!['c']!.active == false,
    'tripleMap roundtrip',
  );
  assert(parsed.chaosArray[0]['x']![0].str == 'x', 'chaosArray roundtrip');
  assert(
    parsed.optionalMadness!['opt']![0].opt1 == 1,
    'optionalMadness roundtrip',
  );
  assert(
    parsed.optionalMadness!['opt']![0].opt2 == null,
    'opt2 null preserved',
  );
  assert(
    parsed.abyss.level1.level2.level3.level4.level5.bottom == 'end',
    'abyss roundtrip',
  );
}

void testDeepAbyss() {
  final now = DateTime.now().toUtc();

  // Test the 5-level deep inline object
  final abyss = NightmareAbyss(
    level1: NightmareAbyssLevel1(
      level2: NightmareAbyssLevel1Level2(
        level3: NightmareAbyssLevel1Level2Level3(
          level4: NightmareAbyssLevel1Level2Level3Level4(
            level5: NightmareAbyssLevel1Level2Level3Level4Level5(
              bottom: 'THE_BOTTOM',
              allTypes: AllPrimitives(
                s: 'string_at_bottom',
                i: 12345,
                f: 123.456,
                b: true,
                d: now,
              ),
            ),
          ),
        ),
      ),
    ),
  );

  final json = abyss.toJson();

  // Navigate deep
  final l1 = json['level1'] as Map<String, dynamic>;
  final l2 = l1['level2'] as Map<String, dynamic>;
  final l3 = l2['level3'] as Map<String, dynamic>;
  final l4 = l3['level4'] as Map<String, dynamic>;
  final l5 = l4['level5'] as Map<String, dynamic>;

  assert(l5['bottom'] == 'THE_BOTTOM', 'bottom value');
  assert((l5['allTypes'] as Map)['i'] == 12345, 'allTypes.i');

  // Round trip
  final parsed = NightmareAbyss.fromJson(json);
  assert(
    parsed.level1.level2.level3.level4.level5.allTypes.s == 'string_at_bottom',
    'parsed allTypes.s',
  );
}

void testOptionalMadness() {
  // Test with all optionals null
  final allNull = NightmareOptionalMadness(
    required: 'only_required',
    opt1: null,
    opt2: null,
    opt3: null,
    opt4: null,
    optNested: null,
  );

  var json = allNull.toJson();
  assert(json['required'] == 'only_required', 'required present');
  assert(!json.containsKey('opt1'), 'opt1 omitted');
  assert(!json.containsKey('opt2'), 'opt2 omitted');
  assert(!json.containsKey('opt3'), 'opt3 omitted');
  assert(!json.containsKey('opt4'), 'opt4 omitted');
  assert(!json.containsKey('optNested'), 'optNested omitted');

  // Test with all optionals present
  final now = DateTime.now().toUtc();
  final allPresent = NightmareOptionalMadness(
    required: 'all_present',
    opt1: 42,
    opt2: 3.14,
    opt3: true,
    opt4: now,
    optNested: NightmareOptionalMadnessOptNested(also: 'nested_value'),
  );

  json = allPresent.toJson();
  assert(json['opt1'] == 42, 'opt1 present');
  assert(json['opt2'] == 3.14, 'opt2 present');
  assert(json['opt3'] == true, 'opt3 present');
  assert(json['optNested'] != null, 'optNested present');

  // Round trip
  final parsed = NightmareOptionalMadness.fromJson(json);
  assert(parsed.opt1 == 42, 'parsed opt1');
  assert(parsed.optNested?.also == 'nested_value', 'parsed optNested.also');
}

import 'dart:convert';
import 'gen/index.dart';

void main() {
  testMatrix3D();
  testNestedMaps();
  testComplexNested();
  testMixedDeep();
  testJsonRoundTrip();
  print('All deep_nested_collections tests passed!');
}

void testMatrix3D() {
  // Test 3D array construction and serialization
  final matrix = Matrix3D(
    data: [
      [
        [1, 2, 3],
        [4, 5, 6],
      ],
      [
        [7, 8, 9],
        [10, 11, 12],
      ],
    ],
  );

  final json = matrix.toJson();
  assert(json['data'] != null, 'data should be present');
  assert((json['data'] as List).length == 2, 'Should have 2 outer elements');

  // Verify deep access
  final data = json['data'] as List;
  final firstPlane = data[0] as List;
  final firstRow = firstPlane[0] as List;
  assert(firstRow[0] == 1, 'First element should be 1');
  assert(firstRow[2] == 3, 'Third element should be 3');

  // Test deserialization from JSON
  final jsonStr = '{"data":[[[1,2],[3,4]],[[5,6],[7,8]]]}';
  final parsed = Matrix3D.fromJson(jsonDecode(jsonStr));
  assert(parsed.data.length == 2, 'Parsed should have 2 planes');
  assert(parsed.data[0][0][0] == 1, 'First value should be 1');
  assert(parsed.data[1][1][1] == 8, 'Last value should be 8');
}

void testNestedMaps() {
  // Test nested map construction (map<map<int>> means Map<String, Map<String, int>>)
  final nested = NestedMaps(
    lookup: {
      'users': {'alice': 1, 'bob': 2},
      'admins': {'carol': 3},
    },
  );

  final json = nested.toJson();
  assert(json['lookup'] != null, 'lookup should be present');

  final lookup = json['lookup'] as Map<String, dynamic>;
  assert(lookup['users'] != null, 'users should be present');
  assert((lookup['users'] as Map)['alice'] == 1, 'alice should be 1');

  // Test deserialization
  final jsonStr = '{"lookup":{"x":{"a":10,"b":20},"y":{"c":30}}}';
  final parsed = NestedMaps.fromJson(jsonDecode(jsonStr));
  assert(parsed.lookup['x']!['a'] == 10, 'x.a should be 10');
  assert(parsed.lookup['y']!['c'] == 30, 'y.c should be 30');
}

void testComplexNested() {
  // Test map containing arrays (map<int[]> means Map<String, List<int>>)
  final complex = ComplexNested(
    arrayMap: {
      'primes': [2, 3, 5, 7],
      'evens': [2, 4, 6, 8],
    },
    mapArray: [
      {'name': 'first'},
      {'name': 'second', 'extra': 'value'},
    ],
    deepMap: {
      'category1': {
        'subcatA': ['item1', 'item2'],
        'subcatB': ['item3'],
      },
      'category2': {
        'subcatC': ['item4', 'item5', 'item6'],
      },
    },
  );

  final json = complex.toJson();

  // Verify arrayMap
  final arrayMap = json['arrayMap'] as Map<String, dynamic>;
  assert(
    (arrayMap['primes'] as List).length == 4,
    'primes should have 4 elements',
  );

  // Verify mapArray
  final mapArray = json['mapArray'] as List;
  assert(mapArray.length == 2, 'mapArray should have 2 elements');
  assert(
    (mapArray[0] as Map)['name'] == 'first',
    'First map name should be first',
  );

  // Verify deepMap
  final deepMap = json['deepMap'] as Map<String, dynamic>;
  final cat1 = deepMap['category1'] as Map<String, dynamic>;
  final subcatA = cat1['subcatA'] as List;
  assert(subcatA[0] == 'item1', 'First item should be item1');

  // Test round-trip
  final jsonStr = jsonEncode(json);
  final parsed = ComplexNested.fromJson(jsonDecode(jsonStr));
  assert(
    parsed.arrayMap['primes']![0] == 2,
    'Round-trip primes[0] should be 2',
  );
  assert(
    parsed.deepMap['category1']!['subcatA']![0] == 'item1',
    'Round-trip deepMap access',
  );
}

void testMixedDeep() {
  // Test with values present
  final withValues = MixedDeep(
    optionalMatrix: [
      [1, 2],
      [3, 4],
    ],
    optionalNestedMap: {
      'outer': {'inner': 42},
    },
  );

  var json = withValues.toJson();
  assert(json['optionalMatrix'] != null, 'optionalMatrix should be present');
  assert(
    json['optionalNestedMap'] != null,
    'optionalNestedMap should be present',
  );

  // Test with null values
  final withNulls = MixedDeep(optionalMatrix: null, optionalNestedMap: null);

  json = withNulls.toJson();
  assert(
    !json.containsKey('optionalMatrix'),
    'optionalMatrix should be omitted when null',
  );
  assert(
    !json.containsKey('optionalNestedMap'),
    'optionalNestedMap should be omitted when null',
  );

  // Test deserialization with missing fields
  final parsed = MixedDeep.fromJson({});
  assert(parsed.optionalMatrix == null, 'Missing field should be null');
  assert(parsed.optionalNestedMap == null, 'Missing field should be null');
}

void testJsonRoundTrip() {
  // Test numeric coercion in nested structures
  // JSON numbers may come as doubles even for int fields
  final jsonWithDoubles = {
    'data': [
      [
        [1.0, 2.0],
        [3.0, 4.0],
      ],
    ],
  };

  final matrix = Matrix3D.fromJson(jsonWithDoubles);
  assert(matrix.data[0][0][0] == 1, 'Should coerce 1.0 to 1');
  assert(matrix.data[0][0][1] == 2, 'Should coerce 2.0 to 2');

  // Verify the round-trip produces clean ints
  final backToJson = matrix.toJson();
  final reData = backToJson['data'] as List;
  final firstVal = (reData[0] as List)[0] as List;
  assert(firstVal[0] is int, 'Should be int after round-trip');
}

import 'dart:convert';
import 'gen/index.dart';

void main() {
  testSingleSpread();
  testMultipleSpreads();
  testChainedSpread();
  testSpreadWithNested();
  testDeepChain();
  testSpreadWithOptionals();
  testSpreadInAnonymous();
  testSpreadRoundTrip();
  print('All spreads tests passed!');
}

void testSingleSpread() {
  // Entity has ...Identifiable which adds 'id'
  final entity = Entity(id: 'entity-123', name: 'Test Entity');

  final json = entity.toJson();
  assert(json['id'] == 'entity-123', 'id from spread');
  assert(json['name'] == 'Test Entity', 'own field');

  final parsed = Entity.fromJson(json);
  assert(parsed.id == 'entity-123', 'parsed id');
  assert(parsed.name == 'Test Entity', 'parsed name');
}

void testMultipleSpreads() {
  final now = DateTime.now().toUtc();

  // TrackedEntity has ...Identifiable, ...Timestamps, ...Metadata
  final tracked = TrackedEntity(
    id: 'tracked-1',
    createdAt: now,
    updatedAt: now,
    tags: ['tag1', 'tag2'],
    labels: {'env': 'prod', 'version': '1.0'},
    name: 'Tracked Thing',
    active: true,
  );

  final json = tracked.toJson();

  // From Identifiable
  assert(json['id'] == 'tracked-1', 'id from Identifiable');

  // From Timestamps
  assert(json['createdAt'] != null, 'createdAt from Timestamps');
  assert(json['updatedAt'] != null, 'updatedAt from Timestamps');

  // From Metadata
  assert((json['tags'] as List).length == 2, 'tags from Metadata');
  assert((json['labels'] as Map)['env'] == 'prod', 'labels from Metadata');

  // Own fields
  assert(json['name'] == 'Tracked Thing', 'own name');
  assert(json['active'] == true, 'own active');
}

void testChainedSpread() {
  final now = DateTime.now().toUtc();

  // AuditedRecord has ...Auditable, which has ...Timestamps
  // So AuditedRecord gets: createdAt, updatedAt, createdBy, updatedBy
  final record = AuditedRecord(
    createdAt: now,
    updatedAt: now,
    createdBy: 'user-1',
    updatedBy: 'user-2',
    recordType: 'invoice',
    data: '{"amount": 100}',
  );

  final json = record.toJson();

  // From Timestamps (via Auditable)
  assert(json['createdAt'] != null, 'createdAt from chain');
  assert(json['updatedAt'] != null, 'updatedAt from chain');

  // From Auditable
  assert(json['createdBy'] == 'user-1', 'createdBy from Auditable');
  assert(json['updatedBy'] == 'user-2', 'updatedBy from Auditable');

  // Own fields
  assert(json['recordType'] == 'invoice', 'own recordType');
}

void testSpreadWithNested() {
  final now = DateTime.now().toUtc();

  // Person has ...PersonBase (which has ...Identifiable, ...Timestamps)
  // Plus its own email and address fields
  final person = Person(
    id: 'person-1',
    createdAt: now,
    updatedAt: now,
    name: 'John Doe',
    email: 'john@example.com',
    address: Address(street: '123 Main St', city: 'NYC'),
  );

  final json = person.toJson();
  assert(json['id'] == 'person-1', 'id from PersonBase->Identifiable');
  assert(json['name'] == 'John Doe', 'name from PersonBase');
  assert(json['email'] == 'john@example.com', 'own email');
  assert((json['address'] as Map)['city'] == 'NYC', 'nested address');

  final parsed = Person.fromJson(json);
  assert(parsed.address.street == '123 Main St', 'parsed nested');
}

void testDeepChain() {
  // Level3 -> Level2 -> Level1
  // Level3 gets: field1 (from Level1), field2 (from Level2), field3 (own)
  final level = Level3(field1: 'from_level1', field2: 42, field3: true);

  final json = level.toJson();
  assert(json['field1'] == 'from_level1', 'field1 from Level1');
  assert(json['field2'] == 42, 'field2 from Level2');
  assert(json['field3'] == true, 'field3 own');

  final parsed = Level3.fromJson(json);
  assert(parsed.field1 == 'from_level1', 'parsed field1');
  assert(parsed.field2 == 42, 'parsed field2');
  assert(parsed.field3 == true, 'parsed field3');
}

void testSpreadWithOptionals() {
  // ExtendedOptional has ...OptionalBase (required, optional?)
  // Plus anotherRequired, anotherOptional?

  // With all optionals null
  final minimal = ExtendedOptional(
    required: 'req',
    optional: null,
    anotherRequired: 'also_req',
    anotherOptional: null,
  );

  var json = minimal.toJson();
  assert(json['required'] == 'req', 'required present');
  assert(!json.containsKey('optional'), 'optional omitted');
  assert(json['anotherRequired'] == 'also_req', 'anotherRequired present');
  assert(!json.containsKey('anotherOptional'), 'anotherOptional omitted');

  // With all optionals present
  final full = ExtendedOptional(
    required: 'req',
    optional: 42,
    anotherRequired: 'also_req',
    anotherOptional: true,
  );

  json = full.toJson();
  assert(json['optional'] == 42, 'optional present');
  assert(json['anotherOptional'] == true, 'anotherOptional present');
}

void testSpreadInAnonymous() {
  final now = DateTime.now().toUtc();

  // ServiceSpreadInAnonymousInput.wrapper has ...Identifiable
  // ServiceSpreadInAnonymousInput.wrapper.inner has ...Timestamps, ...Metadata
  final input = ServiceSpreadInAnonymousInput(
    wrapper: ServiceSpreadInAnonymousInputWrapper(
      id: 'wrapper-id',
      inner: ServiceSpreadInAnonymousInputWrapperInner(
        createdAt: now,
        updatedAt: now,
        tags: ['nested', 'tags'],
        labels: {'key': 'value'},
        value: 'inner_value',
      ),
    ),
  );

  final json = input.toJson();
  final wrapper = json['wrapper'] as Map<String, dynamic>;
  assert(wrapper['id'] == 'wrapper-id', 'wrapper.id from spread');

  final inner = wrapper['inner'] as Map<String, dynamic>;
  assert(inner['createdAt'] != null, 'inner.createdAt from spread');
  assert((inner['tags'] as List).contains('nested'), 'inner.tags from spread');
  assert(inner['value'] == 'inner_value', 'inner.value own');
}

void testSpreadRoundTrip() {
  final now = DateTime.utc(2025, 6, 15, 10, 30, 54);

  final tracked = TrackedEntity(
    id: 'roundtrip-1',
    createdAt: now,
    updatedAt: now,
    tags: ['a', 'b', 'c'],
    labels: {'x': 'y'},
    name: 'Roundtrip Test',
    active: false,
  );

  // Full round trip through JSON string
  final jsonStr = jsonEncode(tracked.toJson());
  final parsed = TrackedEntity.fromJson(jsonDecode(jsonStr));

  assert(parsed.id == 'roundtrip-1', 'roundtrip id');
  assert(parsed.tags.length == 3, 'roundtrip tags');
  assert(parsed.labels['x'] == 'y', 'roundtrip labels');
  assert(parsed.name == 'Roundtrip Test', 'roundtrip name');
  assert(parsed.active == false, 'roundtrip active');

  // Verify datetime survived
  assert(parsed.createdAt.year == 2025, 'roundtrip year');
  assert(parsed.createdAt.month == 6, 'roundtrip month');
  assert(parsed.createdAt.day == 15, 'roundtrip day');
  assert(parsed.createdAt.hour == 10, 'roundtrip hour');
  assert(parsed.createdAt.minute == 30, 'roundtrip minute');
  assert(parsed.createdAt.second == 54, 'roundtrip second');
}

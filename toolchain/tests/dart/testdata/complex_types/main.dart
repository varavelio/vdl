// Dart E2E test for complex types
// This test verifies that nested types, arrays, maps, and inline objects
// are generated correctly and can be serialized/deserialized properly.

import 'gen/index.dart';

void main() {
  // Test nested custom types
  testAddress();

  // Test arrays
  testArrayContainer();

  // Test maps
  testMapContainer();

  // Test inline objects
  testUserProfile();

  // Test deeply nested types
  testOrder();

  // Test copyWith, equality, hashCode, toString
  testClassMethods();

  print('Success');
}

void testAddress() {
  final address = Address(
    street: '123 Main St',
    city: 'Springfield',
    country: 'USA',
  );

  // Serialize to JSON
  final json = address.toJson();
  assert(json['street'] == '123 Main St', 'street mismatch');
  assert(json['city'] == 'Springfield', 'city mismatch');
  assert(json['country'] == 'USA', 'country mismatch');

  // Deserialize from JSON
  final deserialized = Address.fromJson(json);
  assert(deserialized.street == address.street, 'street round-trip failed');
  assert(deserialized.city == address.city, 'city round-trip failed');
  assert(deserialized.country == address.country, 'country round-trip failed');
}

void testArrayContainer() {
  final container = ArrayContainer(
    tags: ['dart', 'flutter', 'vdl'],
    numbers: [1, 2, 3, 4, 5],
    matrix: [
      [1, 2, 3],
      [4, 5, 6],
    ],
    addresses: [
      Address(street: 'Street 1', city: 'City 1', country: 'Country 1'),
      Address(street: 'Street 2', city: 'City 2', country: 'Country 2'),
    ],
  );

  // Serialize to JSON
  final json = container.toJson();
  assert(json['tags'] is List, 'tags should be a List');
  assert((json['tags'] as List).length == 3, 'tags should have 3 items');
  assert(json['numbers'] is List, 'numbers should be a List');
  assert((json['numbers'] as List).length == 5, 'numbers should have 5 items');
  assert(json['matrix'] is List, 'matrix should be a List');
  assert((json['matrix'] as List).length == 2, 'matrix should have 2 rows');
  assert(
    ((json['matrix'] as List)[0] as List).length == 3,
    'matrix row should have 3 columns',
  );
  assert(json['addresses'] is List, 'addresses should be a List');
  assert(
    (json['addresses'] as List).length == 2,
    'addresses should have 2 items',
  );

  // Deserialize from JSON
  final deserialized = ArrayContainer.fromJson(json);
  assert(deserialized.tags.length == 3, 'tags deserialization failed');
  assert(deserialized.tags[0] == 'dart', 'tags[0] should be "dart"');
  assert(deserialized.numbers.length == 5, 'numbers deserialization failed');
  assert(deserialized.numbers[4] == 5, 'numbers[4] should be 5');
  assert(deserialized.matrix.length == 2, 'matrix deserialization failed');
  assert(deserialized.matrix[1][2] == 6, 'matrix[1][2] should be 6');
  assert(
    deserialized.addresses.length == 2,
    'addresses deserialization failed',
  );
  assert(
    deserialized.addresses[0].street == 'Street 1',
    'addresses[0].street mismatch',
  );
}

void testMapContainer() {
  final container = MapContainer(
    stringMap: {'key1': 'value1', 'key2': 'value2'},
    intMap: {'a': 1, 'b': 2, 'c': 3},
    addressMap: {
      'home': Address(street: 'Home St', city: 'Home City', country: 'HC'),
      'work': Address(street: 'Work St', city: 'Work City', country: 'WC'),
    },
  );

  // Serialize to JSON
  final json = container.toJson();
  assert(json['stringMap'] is Map, 'stringMap should be a Map');
  assert(
    (json['stringMap'] as Map)['key1'] == 'value1',
    'stringMap["key1"] should be "value1"',
  );
  assert(json['intMap'] is Map, 'intMap should be a Map');
  assert((json['intMap'] as Map)['b'] == 2, 'intMap["b"] should be 2');
  assert(json['addressMap'] is Map, 'addressMap should be a Map');
  assert(
    ((json['addressMap'] as Map)['home'] as Map)['street'] == 'Home St',
    'addressMap["home"]["street"] should be "Home St"',
  );

  // Deserialize from JSON
  final deserialized = MapContainer.fromJson(json);
  assert(
    deserialized.stringMap['key1'] == 'value1',
    'stringMap deserialization failed',
  );
  assert(deserialized.intMap['c'] == 3, 'intMap deserialization failed');
  assert(
    deserialized.addressMap['work']?.city == 'Work City',
    'addressMap deserialization failed',
  );
}

void testUserProfile() {
  final now = DateTime.now().toUtc();

  // Test with inline objects
  final profile = UserProfile(
    name: 'John Doe',
    primaryAddress: UserProfilePrimaryAddress(
      street: '456 Oak Ave',
      city: 'Metropolis',
      zip: '12345',
    ),
    metadata: UserProfileMetadata(
      createdAt: now,
      tags: ['premium', 'verified'],
    ),
  );

  // Serialize to JSON
  final json = profile.toJson();
  assert(json['name'] == 'John Doe', 'name mismatch');
  assert(json['primaryAddress'] is Map, 'primaryAddress should be a Map');
  assert(
    (json['primaryAddress'] as Map)['zip'] == '12345',
    'primaryAddress.zip mismatch',
  );
  assert(json['metadata'] is Map, 'metadata should be a Map');
  assert(
    ((json['metadata'] as Map)['tags'] as List).length == 2,
    'metadata.tags should have 2 items',
  );

  // Deserialize from JSON
  final deserialized = UserProfile.fromJson(json);
  assert(deserialized.name == 'John Doe', 'name deserialization failed');
  assert(
    deserialized.primaryAddress.city == 'Metropolis',
    'primaryAddress.city deserialization failed',
  );
  assert(
    deserialized.metadata?.tags.length == 2,
    'metadata.tags deserialization failed',
  );
  assert(
    deserialized.metadata?.tags[0] == 'premium',
    'metadata.tags[0] should be "premium"',
  );

  // Test with optional inline object null
  final profileNoMeta = UserProfile(
    name: 'Jane Doe',
    primaryAddress: UserProfilePrimaryAddress(
      street: '789 Pine Rd',
      city: 'Gotham',
      zip: '54321',
    ),
  );

  final jsonNoMeta = profileNoMeta.toJson();
  assert(
    !jsonNoMeta.containsKey('metadata'),
    'null optional inline object should not be in JSON',
  );

  final deserializedNoMeta = UserProfile.fromJson(jsonNoMeta);
  assert(
    deserializedNoMeta.metadata == null,
    'optional inline object should be null',
  );
}

void testOrder() {
  final order = Order(
    id: 'ORDER-001',
    shippingAddress: Address(
      street: 'Shipping St',
      city: 'Ship City',
      country: 'SC',
    ),
    billingAddress: Address(
      street: 'Billing St',
      city: 'Bill City',
      country: 'BC',
    ),
    items: [
      OrderItem(productId: 'PROD-1', quantity: 2, price: 19.99),
      OrderItem(productId: 'PROD-2', quantity: 1, price: 49.99),
    ],
  );

  // Serialize to JSON
  final json = order.toJson();
  assert(json['id'] == 'ORDER-001', 'id mismatch');
  assert(json['shippingAddress'] is Map, 'shippingAddress should be a Map');
  assert(json['billingAddress'] is Map, 'billingAddress should be a Map');
  assert(json['items'] is List, 'items should be a List');
  assert((json['items'] as List).length == 2, 'items should have 2 items');

  // Deserialize from JSON
  final deserialized = Order.fromJson(json);
  assert(deserialized.id == 'ORDER-001', 'id deserialization failed');
  assert(
    deserialized.shippingAddress.city == 'Ship City',
    'shippingAddress deserialization failed',
  );
  assert(
    deserialized.billingAddress?.country == 'BC',
    'billingAddress deserialization failed',
  );
  assert(deserialized.items.length == 2, 'items deserialization failed');
  assert(
    deserialized.items[0].productId == 'PROD-1',
    'items[0].productId mismatch',
  );
  assert(deserialized.items[1].price == 49.99, 'items[1].price mismatch');

  // Test with optional billingAddress null
  final orderNoBilling = Order(
    id: 'ORDER-002',
    shippingAddress: Address(
      street: 'Only Shipping',
      city: 'Ship Only',
      country: 'SO',
    ),
    items: [],
  );

  final jsonNoBilling = orderNoBilling.toJson();
  assert(
    !jsonNoBilling.containsKey('billingAddress'),
    'null optional nested type should not be in JSON',
  );

  final deserializedNoBilling = Order.fromJson(jsonNoBilling);
  assert(
    deserializedNoBilling.billingAddress == null,
    'optional nested type should be null',
  );
}

void testClassMethods() {
  final address1 = Address(
    street: '123 Test St',
    city: 'Test City',
    country: 'TC',
  );

  final address2 = Address(
    street: '123 Test St',
    city: 'Test City',
    country: 'TC',
  );

  final address3 = Address(
    street: 'Different St',
    city: 'Test City',
    country: 'TC',
  );

  // Test equality
  assert(address1 == address2, 'Equal objects should be equal');
  assert(address1 != address3, 'Different objects should not be equal');

  // Test hashCode
  assert(
    address1.hashCode == address2.hashCode,
    'Equal objects should have equal hashCodes',
  );

  // Test toString
  assert(
    address1.toString().contains('Address'),
    'toString should contain class name',
  );
  assert(
    address1.toString().contains('123 Test St'),
    'toString should contain field values',
  );

  // Test copyWith
  final modified = address1.copyWith(street: 'Modified St');
  assert(modified.street == 'Modified St', 'copyWith should update street');
  assert(modified.city == address1.city, 'copyWith should preserve city');
  assert(
    modified.country == address1.country,
    'copyWith should preserve country',
  );

  // Test copyWith on type with arrays
  final container = ArrayContainer(
    tags: ['a', 'b'],
    numbers: [1, 2],
    matrix: [
      [1],
    ],
    addresses: [address1],
  );

  final modifiedContainer = container.copyWith(tags: ['x', 'y', 'z']);
  assert(modifiedContainer.tags.length == 3, 'copyWith should update tags');
  assert(
    modifiedContainer.numbers.length == 2,
    'copyWith should preserve numbers',
  );
}

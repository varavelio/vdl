import 'dart:convert';
import 'gen/index.dart';

void main() {
  testSharedTypes();
  testUserTypes();
  testOrderTypes();
  testCrossDomainTypes();
  testEnumsFromShared();
  testConstantsGenerated();
  testPatternsGenerated();
  testCompleteRoundTrip();
  print('All multi_file tests passed!');
}

void testSharedTypes() {
  // Test Timestamps
  final now = DateTime.now().toUtc();
  final timestamps = Timestamps(createdAt: now, updatedAt: now);
  final json = timestamps.toJson();
  assert(json['createdAt'] != null, 'createdAt exists');

  // Test Identifiable
  final ident = Identifiable(id: 'test-id');
  assert(ident.toJson()['id'] == 'test-id', 'id works');

  // Test Address
  final addr = Address(
    street: '123 Main',
    city: 'NYC',
    country: 'USA',
    postalCode: '10001',
  );
  final addrJson = addr.toJson();
  assert(addrJson['postalCode'] == '10001', 'optional postalCode');

  // Without optional
  final addrNoPostal = Address(
    street: '456 Oak',
    city: 'LA',
    country: 'USA',
    postalCode: null,
  );
  final noPostalJson = addrNoPostal.toJson();
  assert(!noPostalJson.containsKey('postalCode'), 'postalCode omitted');
}

void testUserTypes() {
  final now = DateTime.now().toUtc();

  final user = User(
    id: 'user-123',
    createdAt: now,
    updatedAt: now,
    username: 'johndoe',
    email: 'john@example.com',
    status: Status.Active,
    address: Address(
      street: '123 Main',
      city: 'NYC',
      country: 'USA',
      postalCode: null,
    ),
    roles: ['admin', 'user'],
    metadata: {'theme': 'dark', 'lang': 'en'},
  );

  final json = user.toJson();

  // Verify spread fields
  assert(json['id'] == 'user-123', 'id from Identifiable');
  assert(json['createdAt'] != null, 'createdAt from Timestamps');

  // Verify own fields
  assert(json['username'] == 'johndoe', 'username');
  assert(json['status'] == 'Active', 'status enum');
  assert((json['roles'] as List).length == 2, 'roles array');
  assert((json['metadata'] as Map)['theme'] == 'dark', 'metadata map');

  // Round trip
  final parsed = User.fromJson(json);
  assert(parsed.status == Status.Active, 'parsed status');
  assert(parsed.roles.contains('admin'), 'parsed roles');
}

void testOrderTypes() {
  final now = DateTime.now().toUtc();

  final order = Order(
    id: 'order-456',
    createdAt: now,
    updatedAt: now,
    userId: 'user-123',
    items: [
      OrderItem(
        productId: 'prod-1',
        name: 'Widget',
        quantity: 2,
        unitPrice: 9.99,
        totalPrice: 19.98,
      ),
      OrderItem(
        productId: 'prod-2',
        name: 'Gadget',
        quantity: 1,
        unitPrice: 49.99,
        totalPrice: 49.99,
      ),
    ],
    status: Status.Pending,
    priority: Priority.High,
    shippingAddress: Address(
      street: '789 Ship',
      city: 'LA',
      country: 'USA',
      postalCode: null,
    ),
    billingAddress: null,
    total: 69.97,
    notes: 'Rush delivery',
  );

  final json = order.toJson();

  // Verify items array
  final items = json['items'] as List;
  assert(items.length == 2, 'two items');
  assert((items[0] as Map)['name'] == 'Widget', 'first item name');

  // Verify enums
  assert(json['status'] == 'Pending', 'status enum value');
  assert(json['priority'] == 'High', 'priority enum value');

  // Verify optional present
  assert(json['notes'] == 'Rush delivery', 'notes present');

  // Verify optional absent
  assert(!json.containsKey('billingAddress'), 'billingAddress omitted');
}

void testCrossDomainTypes() {
  final now = DateTime.now().toUtc();

  final stats = UserOrderStats(
    user: User(
      id: 'user-1',
      createdAt: now,
      updatedAt: now,
      username: 'testuser',
      email: 'test@test.com',
      status: Status.Active,
      address: null,
      roles: [],
      metadata: {},
    ),
    totalOrders: 42,
    totalSpent: 1234.56,
    recentOrders: [
      OrderSummary(
        id: 'order-1',
        userId: 'user-1',
        itemCount: 3,
        total: 99.99,
        status: Status.Active,
      ),
    ],
    favoriteAddress: Address(
      street: 'Favorite St',
      city: 'Home',
      country: 'Here',
      postalCode: '00000',
    ),
  );

  final json = stats.toJson();
  assert((json['user'] as Map)['username'] == 'testuser', 'nested user');
  assert(json['totalOrders'] == 42, 'totalOrders');
  assert((json['recentOrders'] as List).length == 1, 'recentOrders');

  // Round trip
  final parsed = UserOrderStats.fromJson(json);
  assert(parsed.user.email == 'test@test.com', 'parsed user.email');
  assert(parsed.recentOrders[0].itemCount == 3, 'parsed recentOrders');
}

void testEnumsFromShared() {
  // Verify enums defined in shared.vdl work
  assert(Status.Active.toJson() == 'Active', 'Status.Active');
  assert(Status.Inactive.toJson() == 'Inactive', 'Status.Inactive');
  assert(Status.Pending.toJson() == 'Pending', 'Status.Pending');

  assert(Priority.Low.toJson() == 'Low', 'Priority.Low');
  assert(Priority.Medium.toJson() == 'Medium', 'Priority.Medium');
  assert(Priority.High.toJson() == 'High', 'Priority.High');

  // Verify fromJson
  assert(StatusJson.fromJson('Active') == Status.Active, 'Status fromJson');
  assert(PriorityJson.fromJson('High') == Priority.High, 'Priority fromJson');

  // Verify lists
  assert(statusList.length == 3, 'statusList length');
  assert(priorityList.length == 3, 'priorityList length');
}

void testConstantsGenerated() {
  // Verify constants from shared.vdl (uses UPPER_CASE naming)
  assert(API_VERSION == 'v1', 'API_VERSION constant');
  assert(MAX_RESULTS == 100, 'MAX_RESULTS constant');
  assert(DEFAULT_TIMEOUT == 30.5, 'DEFAULT_TIMEOUT constant');
}

void testPatternsGenerated() {
  // Verify pattern functions from shared.vdl (uses PascalCase naming)
  final userTopic = UserTopic('u123', 'login');
  assert(userTopic == 'events.users.u123.login', 'userTopic pattern');

  final orderTopic = OrderTopic('us-east', 'ord-456');
  assert(orderTopic == 'orders.us-east.ord-456', 'orderTopic pattern');
}

void testCompleteRoundTrip() {
  final now = DateTime.utc(2025, 7, 4, 12, 0, 0);

  // Build complex nested structure across domains
  final stats = UserOrderStats(
    user: User(
      id: 'complete-user',
      createdAt: now,
      updatedAt: now,
      username: 'completetest',
      email: 'complete@test.com',
      status: Status.Active,
      address: Address(
        street: '100 Complete St',
        city: 'Testville',
        country: 'Testland',
        postalCode: '12345',
      ),
      roles: ['role1', 'role2', 'role3'],
      metadata: {'key1': 'val1', 'key2': 'val2'},
    ),
    totalOrders: 100,
    totalSpent: 9999.99,
    recentOrders: [
      OrderSummary(
        id: 'o1',
        userId: 'complete-user',
        itemCount: 5,
        total: 500.0,
        status: Status.Active,
      ),
      OrderSummary(
        id: 'o2',
        userId: 'complete-user',
        itemCount: 3,
        total: 300.0,
        status: Status.Pending,
      ),
      OrderSummary(
        id: 'o3',
        userId: 'complete-user',
        itemCount: 1,
        total: 100.0,
        status: Status.Inactive,
      ),
    ],
    favoriteAddress: Address(
      street: 'Favorite Lane',
      city: 'Happy Town',
      country: 'Joy',
      postalCode: null,
    ),
  );

  // Serialize to JSON string
  final jsonStr = jsonEncode(stats.toJson());

  // Deserialize back
  final parsed = UserOrderStats.fromJson(jsonDecode(jsonStr));

  // Verify everything survived
  assert(parsed.user.id == 'complete-user', 'user.id');
  assert(parsed.user.username == 'completetest', 'user.username');
  assert(parsed.user.status == Status.Active, 'user.status');
  assert(parsed.user.address?.city == 'Testville', 'user.address.city');
  assert(parsed.user.roles.length == 3, 'user.roles');
  assert(parsed.user.metadata['key1'] == 'val1', 'user.metadata');

  assert(parsed.totalOrders == 100, 'totalOrders');
  assert(parsed.totalSpent == 9999.99, 'totalSpent');

  assert(parsed.recentOrders.length == 3, 'recentOrders.length');
  assert(
    parsed.recentOrders[0].status == Status.Active,
    'recentOrders[0].status',
  );
  assert(
    parsed.recentOrders[1].status == Status.Pending,
    'recentOrders[1].status',
  );
  assert(
    parsed.recentOrders[2].status == Status.Inactive,
    'recentOrders[2].status',
  );

  assert(
    parsed.favoriteAddress?.street == 'Favorite Lane',
    'favoriteAddress.street',
  );
  assert(
    parsed.favoriteAddress?.postalCode == null,
    'favoriteAddress.postalCode null',
  );

  // Verify datetime
  assert(parsed.user.createdAt.year == 2025, 'datetime year');
  assert(parsed.user.createdAt.month == 7, 'datetime month');
  assert(parsed.user.createdAt.day == 4, 'datetime day');
}

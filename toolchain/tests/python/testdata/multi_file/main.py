import datetime
import json
import sys
import os

sys.path.append(os.getcwd())

from gen import *


def test_shared_types():
    now = datetime.datetime.now(datetime.UTC)
    timestamps = Timestamps(created_at=now, updated_at=now)
    data = timestamps.to_dict()
    assert data["createdAt"] is not None

    ident = Identifiable(id_="test-id")
    assert ident.to_dict()["id"] == "test-id"

    addr = Address(street="123 Main", city="NYC", country="USA", postal_code="10001")
    addr_json = addr.to_dict()
    assert addr_json["postalCode"] == "10001"

    addr_no_postal = Address(street="456 Oak", city="LA", country="USA", postal_code=None)
    no_postal_json = addr_no_postal.to_dict()
    assert "postalCode" not in no_postal_json


def test_user_types():
    now = datetime.datetime.now(datetime.UTC)
    user = User(
        id_="user-123",
        created_at=now,
        updated_at=now,
        username="johndoe",
        email="john@example.com",
        status=Status.ACTIVE,
        address=Address(
            street="123 Main", city="NYC", country="USA", postal_code=None
        ),
        roles=["admin", "user"],
        metadata={"theme": "dark", "lang": "en"},
    )
    data = user.to_dict()
    assert data["id"] == "user-123"
    assert data["createdAt"] is not None
    assert data["username"] == "johndoe"
    assert data["status"] == "Active"
    assert len(data["roles"]) == 2
    assert data["metadata"]["theme"] == "dark"

    parsed = User.from_dict(data)
    assert parsed.status == Status.ACTIVE
    assert "admin" in parsed.roles


def test_order_types():
    now = datetime.datetime.now(datetime.UTC)
    order = Order(
        id_="order-456",
        created_at=now,
        updated_at=now,
        user_id="user-123",
        items=[
            OrderItem(
                product_id="prod-1",
                name="Widget",
                quantity=2,
                unit_price=9.99,
                total_price=19.98,
            ),
            OrderItem(
                product_id="prod-2",
                name="Gadget",
                quantity=1,
                unit_price=49.99,
                total_price=49.99,
            ),
        ],
        status=Status.PENDING,
        priority=Priority.HIGH,
        shipping_address=Address(
            street="789 Ship", city="LA", country="USA", postal_code=None
        ),
        billing_address=None,
        total=69.97,
        notes="Rush delivery",
    )
    data = order.to_dict()
    items = data["items"]
    assert len(items) == 2
    assert items[0]["name"] == "Widget"
    assert data["status"] == "Pending"
    assert data["priority"] == "High"
    assert data["notes"] == "Rush delivery"
    assert "billingAddress" not in data


def test_cross_domain_types():
    now = datetime.datetime.now(datetime.UTC)
    stats = UserOrderStats(
        user=User(
            id_="user-1",
            created_at=now,
            updated_at=now,
            username="testuser",
            email="test@test.com",
            status=Status.ACTIVE,
            address=None,
            roles=[],
            metadata={},
        ),
        total_orders=42,
        total_spent=1234.56,
        recent_orders=[
            OrderSummary(
                id_="order-1",
                user_id="user-1",
                item_count=3,
                total=99.99,
                status=Status.ACTIVE,
            )
        ],
        favorite_address=Address(
            street="Favorite St",
            city="Home",
            country="Here",
            postal_code="00000",
        ),
    )
    data = stats.to_dict()
    assert data["user"]["username"] == "testuser"
    assert data["totalOrders"] == 42
    assert len(data["recentOrders"]) == 1

    parsed = UserOrderStats.from_dict(data)
    assert parsed.user.email == "test@test.com"
    assert parsed.recent_orders[0].item_count == 3


def test_enums_from_shared():
    assert Status.ACTIVE.value == "Active"
    assert Status.INACTIVE.value == "Inactive"
    assert Status.PENDING.value == "Pending"

    assert Priority.LOW.value == "Low"
    assert Priority.MEDIUM.value == "Medium"
    assert Priority.HIGH.value == "High"

    assert Status.from_value("Active") == Status.ACTIVE
    assert Priority.from_value("High") == Priority.HIGH


def test_constants_generated():
    assert API_VERSION == "v1"
    assert MAX_RESULTS == 100
    assert DEFAULT_TIMEOUT == 30.5


def test_patterns_generated():
    user_topic_value = user_topic("u123", "login")
    assert user_topic_value == "events.users.u123.login"

    order_topic_value = order_topic("us-east", "ord-456")
    assert order_topic_value == "orders.us-east.ord-456"


def test_complete_round_trip():
    now = datetime.datetime(2025, 7, 4, 12, 0, 0, tzinfo=datetime.UTC)
    stats = UserOrderStats(
        user=User(
            id_="complete-user",
            created_at=now,
            updated_at=now,
            username="completetest",
            email="complete@test.com",
            status=Status.ACTIVE,
            address=Address(
                street="100 Complete St",
                city="Testville",
                country="Testland",
                postal_code="12345",
            ),
            roles=["role1", "role2", "role3"],
            metadata={"key1": "val1", "key2": "val2"},
        ),
        total_orders=100,
        total_spent=9999.99,
        recent_orders=[
            OrderSummary(
                id_="o1",
                user_id="complete-user",
                item_count=5,
                total=500.0,
                status=Status.ACTIVE,
            ),
            OrderSummary(
                id_="o2",
                user_id="complete-user",
                item_count=3,
                total=300.0,
                status=Status.PENDING,
            ),
            OrderSummary(
                id_="o3",
                user_id="complete-user",
                item_count=1,
                total=100.0,
                status=Status.INACTIVE,
            ),
        ],
        favorite_address=Address(
            street="Favorite Lane",
            city="Happy Town",
            country="Joy",
            postal_code=None,
        ),
    )

    parsed = UserOrderStats.from_dict(json.loads(json.dumps(stats.to_dict())))
    assert parsed.user.id_ == "complete-user"
    assert parsed.user.username == "completetest"
    assert parsed.user.status == Status.ACTIVE
    assert parsed.user.address.city == "Testville"
    assert len(parsed.user.roles) == 3
    assert parsed.user.metadata["key1"] == "val1"
    assert parsed.total_orders == 100
    assert parsed.total_spent == 9999.99
    assert len(parsed.recent_orders) == 3
    assert parsed.recent_orders[0].status == Status.ACTIVE
    assert parsed.recent_orders[1].status == Status.PENDING
    assert parsed.recent_orders[2].status == Status.INACTIVE
    assert parsed.favorite_address.street == "Favorite Lane"
    assert parsed.favorite_address.postal_code is None
    assert parsed.user.created_at.year == 2025
    assert parsed.user.created_at.month == 7
    assert parsed.user.created_at.day == 4


if __name__ == "__main__":
    test_shared_types()
    test_user_types()
    test_order_types()
    test_cross_domain_types()
    test_enums_from_shared()
    test_constants_generated()
    test_patterns_generated()
    test_complete_round_trip()
    print("All multi_file tests passed!")

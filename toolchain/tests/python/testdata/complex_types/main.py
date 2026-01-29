import datetime
import sys
import os

sys.path.append(os.getcwd())

from gen import *


def test_address():
    address = Address(street="123 Main St", city="Springfield", country="USA")
    data = address.to_dict()
    assert data["street"] == "123 Main St"
    assert data["city"] == "Springfield"
    assert data["country"] == "USA"

    restored = Address.from_dict(data)
    assert restored.street == address.street
    assert restored.city == address.city
    assert restored.country == address.country


def test_array_container():
    container = ArrayContainer(
        tags=["dart", "flutter", "vdl"],
        numbers=[1, 2, 3, 4, 5],
        matrix=[[1, 2, 3], [4, 5, 6]],
        addresses=[
            Address(street="Street 1", city="City 1", country="Country 1"),
            Address(street="Street 2", city="City 2", country="Country 2"),
        ],
    )

    data = container.to_dict()
    assert isinstance(data["tags"], list)
    assert len(data["tags"]) == 3
    assert isinstance(data["numbers"], list)
    assert len(data["numbers"]) == 5
    assert isinstance(data["matrix"], list)
    assert len(data["matrix"]) == 2
    assert len(data["matrix"][0]) == 3
    assert isinstance(data["addresses"], list)
    assert len(data["addresses"]) == 2

    restored = ArrayContainer.from_dict(data)
    assert len(restored.tags) == 3
    assert restored.tags[0] == "dart"
    assert restored.numbers[4] == 5
    assert restored.matrix[1][2] == 6
    assert restored.addresses[0].street == "Street 1"


def test_map_container():
    container = MapContainer(
        string_map={"key1": "value1", "key2": "value2"},
        int_map={"a": 1, "b": 2, "c": 3},
        address_map={
            "home": Address(street="Home St", city="Home City", country="HC"),
            "work": Address(street="Work St", city="Work City", country="WC"),
        },
    )

    data = container.to_dict()
    assert isinstance(data["stringMap"], dict)
    assert data["stringMap"]["key1"] == "value1"
    assert isinstance(data["intMap"], dict)
    assert data["intMap"]["b"] == 2
    assert isinstance(data["addressMap"], dict)
    assert data["addressMap"]["home"]["street"] == "Home St"

    restored = MapContainer.from_dict(data)
    assert restored.string_map["key1"] == "value1"
    assert restored.int_map["c"] == 3
    assert restored.address_map["work"].city == "Work City"


def test_user_profile():
    now = datetime.datetime.now(datetime.UTC)
    profile = UserProfile(
        name="John Doe",
        primary_address=UserProfilePrimaryAddress(
            street="456 Oak Ave", city="Metropolis", zip="12345"
        ),
        metadata=UserProfileMetadata(created_at=now, tags=["premium", "verified"]),
    )

    data = profile.to_dict()
    assert data["name"] == "John Doe"
    assert isinstance(data["primaryAddress"], dict)
    assert data["primaryAddress"]["zip"] == "12345"
    assert isinstance(data["metadata"], dict)
    assert len(data["metadata"]["tags"]) == 2

    restored = UserProfile.from_dict(data)
    assert restored.name == "John Doe"
    assert restored.primary_address.city == "Metropolis"
    assert restored.metadata.tags[0] == "premium"

    no_meta = UserProfile(
        name="Jane Doe",
        primary_address=UserProfilePrimaryAddress(
            street="789 Pine Rd", city="Gotham", zip="54321"
        ),
    )
    data_no_meta = no_meta.to_dict()
    assert "metadata" not in data_no_meta
    restored_no_meta = UserProfile.from_dict(data_no_meta)
    assert restored_no_meta.metadata is None


def test_order():
    order = Order(
        id_="ORDER-001",
        shipping_address=Address(
            street="Shipping St", city="Ship City", country="SC"
        ),
        billing_address=Address(
            street="Billing St", city="Bill City", country="BC"
        ),
        items=[
            OrderItem(product_id="PROD-1", quantity=2, price=19.99),
            OrderItem(product_id="PROD-2", quantity=1, price=49.99),
        ],
    )

    data = order.to_dict()
    assert data["id"] == "ORDER-001"
    assert isinstance(data["shippingAddress"], dict)
    assert isinstance(data["billingAddress"], dict)
    assert isinstance(data["items"], list)
    assert len(data["items"]) == 2

    restored = Order.from_dict(data)
    assert restored.id_ == "ORDER-001"
    assert restored.shipping_address.city == "Ship City"
    assert restored.billing_address.country == "BC"
    assert len(restored.items) == 2
    assert restored.items[0].product_id == "PROD-1"
    assert restored.items[1].price == 49.99

    order_no_billing = Order(
        id_="ORDER-002",
        shipping_address=Address(
            street="Only Shipping", city="Ship Only", country="SO"
        ),
        items=[],
    )
    data_no_billing = order_no_billing.to_dict()
    assert "billingAddress" not in data_no_billing
    restored_no_billing = Order.from_dict(data_no_billing)
    assert restored_no_billing.billing_address is None


if __name__ == "__main__":
    test_address()
    test_array_container()
    test_map_container()
    test_user_profile()
    test_order()
    print("Success")

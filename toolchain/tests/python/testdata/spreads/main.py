import datetime
import json
import sys
import os

sys.path.append(os.getcwd())

from gen import *


def test_single_spread():
    entity = Entity(id_="entity-123", name="Test Entity")
    data = entity.to_dict()
    assert data["id"] == "entity-123"
    assert data["name"] == "Test Entity"

    parsed = Entity.from_dict(data)
    assert parsed.id_ == "entity-123"
    assert parsed.name == "Test Entity"


def test_multiple_spreads():
    now = datetime.datetime.now(datetime.UTC)
    tracked = TrackedEntity(
        id_="tracked-1",
        created_at=now,
        updated_at=now,
        tags=["tag1", "tag2"],
        labels={"env": "prod", "version": "1.0"},
        name="Tracked Thing",
        active=True,
    )
    data = tracked.to_dict()
    assert data["id"] == "tracked-1"
    assert data["createdAt"] is not None
    assert data["updatedAt"] is not None
    assert len(data["tags"]) == 2
    assert data["labels"]["env"] == "prod"
    assert data["name"] == "Tracked Thing"
    assert data["active"] is True


def test_chained_spread():
    now = datetime.datetime.now(datetime.UTC)
    record = AuditedRecord(
        created_at=now,
        updated_at=now,
        created_by="user-1",
        updated_by="user-2",
        record_type="invoice",
        data="{\"amount\": 100}",
    )
    data = record.to_dict()
    assert data["createdAt"] is not None
    assert data["updatedAt"] is not None
    assert data["createdBy"] == "user-1"
    assert data["updatedBy"] == "user-2"
    assert data["recordType"] == "invoice"


def test_spread_with_nested():
    now = datetime.datetime.now(datetime.UTC)
    person = Person(
        id_="person-1",
        created_at=now,
        updated_at=now,
        name="John Doe",
        email="john@example.com",
        address=Address(street="123 Main St", city="NYC"),
    )
    data = person.to_dict()
    assert data["id"] == "person-1"
    assert data["name"] == "John Doe"
    assert data["email"] == "john@example.com"
    assert data["address"]["city"] == "NYC"

    parsed = Person.from_dict(data)
    assert parsed.address.street == "123 Main St"


def test_deep_chain():
    level = Level3(field1="from_level1", field2=42, field3=True)
    data = level.to_dict()
    assert data["field1"] == "from_level1"
    assert data["field2"] == 42
    assert data["field3"] is True

    parsed = Level3.from_dict(data)
    assert parsed.field1 == "from_level1"
    assert parsed.field2 == 42
    assert parsed.field3 is True


def test_spread_with_optionals():
    minimal = ExtendedOptional(
        required="req",
        optional=None,
        another_required="also_req",
        another_optional=None,
    )
    data = minimal.to_dict()
    assert data["required"] == "req"
    assert "optional" not in data
    assert data["anotherRequired"] == "also_req"
    assert "anotherOptional" not in data

    full = ExtendedOptional(
        required="req",
        optional=42,
        another_required="also_req",
        another_optional=True,
    )
    data = full.to_dict()
    assert data["optional"] == 42
    assert data["anotherOptional"] is True


def test_spread_in_anonymous():
    now = datetime.datetime.now(datetime.UTC)
    input_data = ServiceSpreadInAnonymousInput(
        wrapper=ServiceSpreadInAnonymousInputWrapper(
            id_="wrapper-id",
            inner=ServiceSpreadInAnonymousInputWrapperInner(
                created_at=now,
                updated_at=now,
                tags=["nested", "tags"],
                labels={"key": "value"},
                value="inner_value",
            ),
        )
    )
    data = input_data.to_dict()
    wrapper = data["wrapper"]
    assert wrapper["id"] == "wrapper-id"
    inner = wrapper["inner"]
    assert inner["createdAt"] is not None
    assert "nested" in inner["tags"]
    assert inner["value"] == "inner_value"


def test_spread_round_trip():
    now = datetime.datetime(2025, 6, 15, 10, 30, 54, tzinfo=datetime.UTC)
    tracked = TrackedEntity(
        id_="roundtrip-1",
        created_at=now,
        updated_at=now,
        tags=["a", "b", "c"],
        labels={"x": "y"},
        name="Roundtrip Test",
        active=False,
    )
    parsed = TrackedEntity.from_dict(json.loads(json.dumps(tracked.to_dict())))
    assert parsed.id_ == "roundtrip-1"
    assert len(parsed.tags) == 3
    assert parsed.labels["x"] == "y"
    assert parsed.name == "Roundtrip Test"
    assert parsed.active is False
    assert parsed.created_at.year == 2025
    assert parsed.created_at.month == 6
    assert parsed.created_at.day == 15
    assert parsed.created_at.hour == 10
    assert parsed.created_at.minute == 30
    assert parsed.created_at.second == 54


if __name__ == "__main__":
    test_single_spread()
    test_multiple_spreads()
    test_chained_spread()
    test_spread_with_nested()
    test_deep_chain()
    test_spread_with_optionals()
    test_spread_in_anonymous()
    test_spread_round_trip()
    print("All spreads tests passed!")

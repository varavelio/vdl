import sys
import os

sys.path.append(os.getcwd())

from gen import *


def test_missing_required_field():
    malformed = {"name": "John", "active": True}
    threw = False
    try:
        RequiredFields.from_dict(malformed)
    except Exception as exc:
        threw = True
        assert isinstance(exc, (TypeError, AttributeError, Exception))
    assert threw


def test_wrong_type_field():
    malformed = {"name": "John", "age": "not a number", "active": True}
    threw = False
    try:
        RequiredFields.from_dict(malformed)
    except Exception as exc:
        threw = True
        assert isinstance(exc, (TypeError, AttributeError, Exception))
    assert threw


def test_nested_malformed():
    malformed = {"child": {"name": "Jane"}}
    threw = False
    try:
        NestedRequired.from_dict(malformed)
    except Exception:
        threw = True
    assert threw


def test_list_wrong_type():
    malformed = {"items": "not a list"}
    threw = False
    try:
        WithList.from_dict(malformed)
    except Exception as exc:
        threw = True
        assert isinstance(exc, (TypeError, AttributeError, Exception))
    assert threw


def test_valid_json_works():
    valid = {"name": "Alice", "age": 30, "active": True}
    data = RequiredFields.from_dict(valid)
    assert data.name == "Alice"
    assert data.age == 30
    assert data.active is True


if __name__ == "__main__":
    test_missing_required_field()
    test_wrong_type_field()
    test_nested_malformed()
    test_list_wrong_type()
    test_valid_json_works()
    print("Success")

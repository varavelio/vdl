import sys
import os

sys.path.append(os.getcwd())

from gen import *


def test_optional_list_null():
    data = WithOptionalList.from_dict({})
    assert data.items is None

    data_null = WithOptionalList.from_dict({"items": None})
    assert data_null.items is None

    data_empty = WithOptionalList.from_dict({"items": []})
    assert data_empty.items is not None
    assert len(data_empty.items) == 0


def test_optional_map_null():
    data = WithOptionalMap.from_dict({})
    assert data.metadata is None

    data_null = WithOptionalMap.from_dict({"metadata": None})
    assert data_null.metadata is None

    data_empty = WithOptionalMap.from_dict({"metadata": {}})
    assert data_empty.metadata is not None
    assert len(data_empty.metadata) == 0


def test_optional_nested_null():
    data = WithOptionalNested.from_dict({})
    assert data.child is None

    data_null = WithOptionalNested.from_dict({"child": None})
    assert data_null.child is None

    data_value = WithOptionalNested.from_dict({"child": {"name": "test", "value": 42}})
    assert data_value.child is not None
    assert data_value.child.name == "test"
    assert data_value.child.value == 42


def test_all_optional_empty():
    empty = AllOptional()
    assert empty.string_field is None
    assert empty.int_field is None
    assert empty.list_field is None
    assert empty.map_field is None
    assert empty.nested_field is None

    data = empty.to_dict()
    assert data == {}


def test_deep_optional_null():
    deep = DeepOptional()
    assert deep.level1 is None

    partial = DeepOptional(level1=Level1())
    assert partial.level1 is not None
    assert partial.level1.level2 is None

    full = DeepOptional(level1=Level1(level2=Level2(value="deep")))
    assert full.level1.level2.value == "deep"

    data = full.to_dict()
    restored = DeepOptional.from_dict(data)
    assert restored.level1.level2.value == "deep"


def test_to_json_omits_null():
    partial = AllOptional(string_field="test", int_field=42)
    data = partial.to_dict()
    assert "stringField" in data
    assert "intField" in data
    assert "listField" not in data
    assert "mapField" not in data
    assert "nestedField" not in data
    assert len(data) == 2


def test_from_empty_json():
    data = AllOptional.from_dict({})
    assert data.string_field is None
    assert data.int_field is None
    assert data.list_field is None
    assert data.map_field is None
    assert data.nested_field is None


if __name__ == "__main__":
    test_optional_list_null()
    test_optional_map_null()
    test_optional_nested_null()
    test_all_optional_empty()
    test_deep_optional_null()
    test_to_json_omits_null()
    test_from_empty_json()
    print("Success")

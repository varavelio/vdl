import json
import sys
import os

sys.path.append(os.getcwd())

from gen import *


def test_int_as_double():
    dirty = {"intField": 42.0, "floatField": 3.14}
    data = NumericData.from_dict(dirty)
    assert data.int_field == 42
    assert isinstance(data.int_field, int)
    assert data.float_field == 3.14
    assert isinstance(data.float_field, float)


def test_double_as_int():
    dirty = {"intField": 100, "floatField": 10}
    data = NumericData.from_dict(dirty)
    assert data.int_field == 100
    assert data.float_field == 10.0
    assert isinstance(data.float_field, float)


def test_optional_numerics():
    dirty = {
        "intField": 1.0,
        "floatField": 2,
        "optionalInt": 3.0,
        "optionalFloat": 4,
    }
    data = NumericData.from_dict(dirty)
    assert data.int_field == 1
    assert data.float_field == 2.0
    assert data.optional_int == 3
    assert data.optional_float == 4.0

    minimal = NumericData.from_dict({"intField": 5, "floatField": 6.0})
    assert minimal.optional_int is None
    assert minimal.optional_float is None


def test_numeric_arrays():
    dirty = {"intArray": [1.0, 2.0, 3.0], "floatArray": [1, 2, 3]}
    data = NumericArrays.from_dict(dirty)
    assert len(data.int_array) == 3
    assert data.int_array[0] == 1
    assert isinstance(data.int_array[0], int)

    assert len(data.float_array) == 3
    assert data.float_array[0] == 1.0
    assert isinstance(data.float_array[0], float)


def test_numeric_maps():
    dirty = {"intMap": {"a": 1.0, "b": 2.0}, "floatMap": {"x": 1, "y": 2}}
    data = NumericMaps.from_dict(dirty)
    assert data.int_map["a"] == 1
    assert isinstance(data.int_map["a"], int)

    assert data.float_map["x"] == 1.0
    assert isinstance(data.float_map["x"], float)


def test_round_trip():
    original = NumericData(int_field=42, float_field=3.14159, optional_int=100, optional_float=2.718)
    parsed = NumericData.from_dict(json.loads(json.dumps(original.to_dict())))
    assert parsed.int_field == original.int_field
    assert parsed.float_field == original.float_field
    assert parsed.optional_int == original.optional_int
    assert parsed.optional_float == original.optional_float


if __name__ == "__main__":
    test_int_as_double()
    test_double_as_int()
    test_optional_numerics()
    test_numeric_arrays()
    test_numeric_maps()
    test_round_trip()
    print("Success")

import json
import sys
import os

sys.path.append(os.getcwd())

from gen import *


def test_matrix_3d():
    matrix = Matrix3D(
        data=[
            [[1, 2, 3], [4, 5, 6]],
            [[7, 8, 9], [10, 11, 12]],
        ]
    )
    data = matrix.to_dict()
    assert data["data"] is not None
    assert len(data["data"]) == 2

    first_plane = data["data"][0]
    first_row = first_plane[0]
    assert first_row[0] == 1
    assert first_row[2] == 3

    parsed = Matrix3D.from_dict(json.loads('{"data":[[[1,2],[3,4]],[[5,6],[7,8]]]}'))
    assert len(parsed.data) == 2
    assert parsed.data[0][0][0] == 1
    assert parsed.data[1][1][1] == 8


def test_nested_maps():
    nested = NestedMaps(
        lookup={"users": {"alice": 1, "bob": 2}, "admins": {"carol": 3}}
    )
    data = nested.to_dict()
    assert data["lookup"] is not None
    lookup = data["lookup"]
    assert lookup["users"]["alice"] == 1

    parsed = NestedMaps.from_dict(
        json.loads('{"lookup":{"x":{"a":10,"b":20},"y":{"c":30}}}')
    )
    assert parsed.lookup["x"]["a"] == 10
    assert parsed.lookup["y"]["c"] == 30


def test_complex_nested():
    complex_data = ComplexNested(
        array_map={"primes": [2, 3, 5, 7], "evens": [2, 4, 6, 8]},
        map_array=[{"name": "first"}, {"name": "second", "extra": "value"}],
        deep_map={
            "category1": {"subcatA": ["item1", "item2"], "subcatB": ["item3"]},
            "category2": {"subcatC": ["item4", "item5", "item6"]},
        },
    )

    data = complex_data.to_dict()
    array_map = data["arrayMap"]
    assert len(array_map["primes"]) == 4

    map_array = data["mapArray"]
    assert len(map_array) == 2
    assert map_array[0]["name"] == "first"

    deep_map = data["deepMap"]
    subcat_a = deep_map["category1"]["subcatA"]
    assert subcat_a[0] == "item1"

    parsed = ComplexNested.from_dict(json.loads(json.dumps(data)))
    assert parsed.array_map["primes"][0] == 2
    assert parsed.deep_map["category1"]["subcatA"][0] == "item1"


def test_mixed_deep():
    with_values = MixedDeep(
        optional_matrix=[[1, 2], [3, 4]],
        optional_nested_map={"outer": {"inner": 42}},
    )
    data = with_values.to_dict()
    assert data["optionalMatrix"] is not None
    assert data["optionalNestedMap"] is not None

    with_nulls = MixedDeep(optional_matrix=None, optional_nested_map=None)
    data = with_nulls.to_dict()
    assert "optionalMatrix" not in data
    assert "optionalNestedMap" not in data

    parsed = MixedDeep.from_dict({})
    assert parsed.optional_matrix is None
    assert parsed.optional_nested_map is None


def test_json_round_trip():
    json_with_doubles = {"data": [[[1.0, 2.0], [3.0, 4.0]]]} 
    matrix = Matrix3D.from_dict(json_with_doubles)
    assert matrix.data[0][0][0] == 1
    assert matrix.data[0][0][1] == 2

    back = matrix.to_dict()
    first_val = back["data"][0][0]
    assert isinstance(first_val[0], int)


if __name__ == "__main__":
    test_matrix_3d()
    test_nested_maps()
    test_complex_nested()
    test_mixed_deep()
    test_json_round_trip()
    print("All deep_nested_collections tests passed!")

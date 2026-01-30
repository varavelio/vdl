import datetime
import json
import sys
import os

sys.path.append(os.getcwd())

from gen import *


def test_all_primitives():
    now = datetime.datetime.now(datetime.UTC)
    ap = AllPrimitives(s="test", i=42, f=3.14, b=True, d=now)
    data = ap.to_dict()
    assert data["s"] == "test"
    assert data["i"] == 42
    assert data["f"] == 3.14
    assert data["b"] is True

    parsed = AllPrimitives.from_dict(data)
    assert parsed.s == "test"
    assert parsed.i == 42


def test_medium_complex():
    now = datetime.datetime.now(datetime.UTC)
    medium = MediumComplex(
        grid=[
            [
                MediumComplexGrid(x=0, y=0, data={"key": "val00"}),
                MediumComplexGrid(x=0, y=1, data={"key": "val01"}),
            ],
            [MediumComplexGrid(x=1, y=0, data={"key": "val10"})],
        ],
        catalog={
            "items": [
                MediumComplexCatalog(
                    name="Item1",
                    count=10,
                    price=99.99,
                    available=True,
                    last_update=now,
                )
            ]
        },
        config=MediumComplexConfig(inner=MediumComplexConfigInner(value="nested")),
    )

    data = medium.to_dict()
    grid = data["grid"]
    assert len(grid) == 2
    assert len(grid[0]) == 2
    catalog = data["catalog"]
    assert len(catalog["items"]) == 1

    parsed = MediumComplex.from_dict(json.loads(json.dumps(data)))
    assert len(parsed.grid) == 2
    assert parsed.config.inner.value == "nested"


def test_nightmare_construction():
    now = datetime.datetime.now(datetime.UTC)
    hyper_cell = NightmareHyperCube(
        coords=[1, 2, 3],
        data={"key1": NightmareHyperCubeData(values=[1.1, 2.2], flags=[True, False])},
    )
    triple_map = {"l1": {"l2": {"l3": NightmareTripleNestedMap(timestamps=[now], active=True)}}}
    chaos_item = NightmareChaosArray(
        str_="chaos",
        number=666,
        dec=6.66,
        flag=True,
        time=now,
        nested=NightmareChaosArrayNested(
            deep="deep",
            deeper=NightmareChaosArrayNestedDeeper(deepest=999),
        ),
    )
    nightmare = Nightmare(
        hyper_cube=[[[[hyper_cell]]]],
        triple_nested_map=triple_map,
        chaos_array=[{"chaos": [chaos_item]}],
        optional_madness=None,
        abyss=NightmareAbyss(
            level1=NightmareAbyssLevel1(
                level2=NightmareAbyssLevel1Level2(
                    level3=NightmareAbyssLevel1Level2Level3(
                        level4=NightmareAbyssLevel1Level2Level3Level4(
                            level5=NightmareAbyssLevel1Level2Level3Level4Level5(
                                bottom="BOTTOM",
                                all_types=AllPrimitives(
                                    s="s", i=1, f=1.0, b=True, d=now
                                ),
                            )
                        )
                    )
                )
            )
        ),
        matrix_of_maps_of_arrays=[
            [
                {
                    "row": NightmareMatrixOfMapsOfArrays(
                        items=[
                            NightmareMatrixOfMapsOfArraysItems(id_=1, tags=["a", "b"])
                        ]
                    )
                }
            ]
        ],
        primitive_map={
            "partial": NightmarePrimitiveMap(
                s="only string", i=None, f=None, b=None, d=None
            ),
            "full": NightmarePrimitiveMap(s="s", i=1, f=1.1, b=True, d=now),
        },
    )

    data = nightmare.to_dict()
    assert data["hyperCube"] is not None
    assert data["tripleNestedMap"] is not None
    assert data["chaosArray"] is not None
    assert "optionalMadness" not in data
    assert data["abyss"] is not None


def test_nightmare_round_trip():
    now = datetime.datetime(2025, 1, 1, 12, 0, 0, tzinfo=datetime.UTC)
    nightmare = Nightmare(
        hyper_cube=[[[[NightmareHyperCube(coords=[1], data={"k": NightmareHyperCubeData(values=[1.0], flags=[True])})]]]],
        triple_nested_map={"a": {"b": {"c": NightmareTripleNestedMap(timestamps=[now], active=False)}}},
        chaos_array=[{"x": [NightmareChaosArray(str_="x", number=1, dec=1.0, flag=False, time=now, nested=None)]}],
        optional_madness={
            "opt": [
                NightmareOptionalMadness(
                    required="req",
                    opt1=1,
                    opt2=None,
                    opt3=True,
                    opt4=None,
                    opt_nested=NightmareOptionalMadnessOptNested(also="also"),
                )
            ]
        },
        abyss=NightmareAbyss(
            level1=NightmareAbyssLevel1(
                level2=NightmareAbyssLevel1Level2(
                    level3=NightmareAbyssLevel1Level2Level3(
                        level4=NightmareAbyssLevel1Level2Level3Level4(
                            level5=NightmareAbyssLevel1Level2Level3Level4Level5(
                                bottom="end",
                                all_types=AllPrimitives(s="", i=0, f=0.0, b=False, d=now),
                            )
                        )
                    )
                )
            )
        ),
        matrix_of_maps_of_arrays=[],
        primitive_map={},
    )

    parsed = Nightmare.from_dict(json.loads(json.dumps(nightmare.to_dict())))
    assert parsed.hyper_cube[0][0][0][0].coords[0] == 1
    assert parsed.triple_nested_map["a"]["b"]["c"].active is False
    assert parsed.chaos_array[0]["x"][0].str_ == "x"
    assert parsed.optional_madness["opt"][0].opt1 == 1
    assert parsed.optional_madness["opt"][0].opt2 is None
    assert parsed.abyss.level1.level2.level3.level4.level5.bottom == "end"


def test_deep_abyss():
    now = datetime.datetime.now(datetime.UTC)
    abyss = NightmareAbyss(
        level1=NightmareAbyssLevel1(
            level2=NightmareAbyssLevel1Level2(
                level3=NightmareAbyssLevel1Level2Level3(
                    level4=NightmareAbyssLevel1Level2Level3Level4(
                        level5=NightmareAbyssLevel1Level2Level3Level4Level5(
                            bottom="THE_BOTTOM",
                            all_types=AllPrimitives(
                                s="string_at_bottom",
                                i=12345,
                                f=123.456,
                                b=True,
                                d=now,
                            ),
                        )
                    )
                )
            )
        )
    )

    data = abyss.to_dict()
    l5 = data["level1"]["level2"]["level3"]["level4"]["level5"]
    assert l5["bottom"] == "THE_BOTTOM"
    assert l5["allTypes"]["i"] == 12345

    parsed = NightmareAbyss.from_dict(data)
    assert (
        parsed.level1.level2.level3.level4.level5.all_types.s
        == "string_at_bottom"
    )


def test_optional_madness():
    all_null = NightmareOptionalMadness(
        required="only_required",
        opt1=None,
        opt2=None,
        opt3=None,
        opt4=None,
        opt_nested=None,
    )
    data = all_null.to_dict()
    assert data["required"] == "only_required"
    assert "opt1" not in data
    assert "opt2" not in data
    assert "opt3" not in data
    assert "opt4" not in data
    assert "optNested" not in data

    now = datetime.datetime.now(datetime.UTC)
    all_present = NightmareOptionalMadness(
        required="all_present",
        opt1=42,
        opt2=3.14,
        opt3=True,
        opt4=now,
        opt_nested=NightmareOptionalMadnessOptNested(also="nested_value"),
    )
    data = all_present.to_dict()
    assert data["opt1"] == 42
    assert data["opt2"] == 3.14
    assert data["opt3"] is True
    assert data["optNested"] is not None

    parsed = NightmareOptionalMadness.from_dict(data)
    assert parsed.opt1 == 42
    assert parsed.opt_nested.also == "nested_value"


if __name__ == "__main__":
    test_all_primitives()
    test_medium_complex()
    test_nightmare_construction()
    test_nightmare_round_trip()
    test_deep_abyss()
    test_optional_madness()
    print("All extreme_nesting tests passed!")

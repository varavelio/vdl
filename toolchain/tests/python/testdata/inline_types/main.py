import sys
import os

sys.path.append(os.getcwd())

from gen import *

def main():
    # 1. Array of inline objects
    files = [
        ComplexInlineTypesFiles(path="a", content="b")
    ]

    # 2. Map of inline objects
    meta = {
        "v1": ComplexInlineTypesMeta(created_at="2023", author="me")
    }

    # 3. Map of arrays of inline objects
    grouped_files = {
        "group1": [ComplexInlineTypesGroupedFiles(name="f1", size=10)]
    }

    # 4. Array of maps of inline objects
    configs = [
        {"conf1": ComplexInlineTypesConfigs(key="k", value="v")}
    ]

    # 5. Nested arrays of inline objects
    grid = [
        [ComplexInlineTypesGrid(x=1, y=2)]
    ]

    # 6. Simple inline object
    simple = ComplexInlineTypesSimple(
        name="test",
        enabled=True
    )

    # 7. Deeply nested inline objects
    deep_nest = ComplexInlineTypesDeepNest(
        level1="l1",
        child=ComplexInlineTypesDeepNestChild(
            level2=2,
            grand_child=ComplexInlineTypesDeepNestChildGrandChild(
                level3=True,
                great_grand_child=ComplexInlineTypesDeepNestChildGrandChildGreatGrandChild(
                    level4=4.5,
                    data="end"
                )
            )
        )
    )

    output = ComplexInlineTypes(
        files=files,
        meta=meta,
        grouped_files=grouped_files,
        configs=configs,
        grid=grid,
        simple=simple,
        deep_nest=deep_nest
    )

    # Verify integrity
    if output.files[0].path != "a": raise Exception("files mismatch")
    if output.meta["v1"].author != "me": raise Exception("meta mismatch")
    if output.grouped_files["group1"][0].size != 10: raise Exception("groupedFiles mismatch")
    if output.configs[0]["conf1"].value != "v": raise Exception("configs mismatch")
    if output.grid[0][0].y != 2: raise Exception("grid mismatch")
    if output.simple.name != "test": raise Exception("simple mismatch")
    if output.deep_nest.child.grand_child.great_grand_child.data != "end": raise Exception("deepNest mismatch")

    print("Success")

if __name__ == "__main__":
    main()

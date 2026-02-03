import sys
import os

sys.path.append(os.getcwd())

from gen import *


def test_self_referencing():
    self_ref = SelfReferencing(
        id_="root",
        name="Root Node",
        parent=SelfReferencing(
            id_="parent",
            name="Parent Node",
            parent=SelfReferencing(
                id_="grandparent",
                name="Grandparent Node",
                parent=None,
            ),
        ),
    )

    data = self_ref.to_dict()
    restored = SelfReferencing.from_dict(data)

    assert restored.id_ == "root"
    assert restored.name == "Root Node"
    assert restored.parent is not None
    assert restored.parent.id_ == "parent"
    assert restored.parent.parent is not None
    assert restored.parent.parent.id_ == "grandparent"
    assert restored.parent.parent.parent is None


def test_chain_with_optional():
    chain = NodeA(
        value="A",
        node_b=NodeB(
            value="B",
            node_c=NodeC(
                value="C",
                node_d=NodeD(
                    value="D",
                    node_e=NodeE(
                        value="E",
                        back_to_a=None,
                    ),
                ),
            ),
        ),
    )

    data = chain.to_dict()
    restored = NodeA.from_dict(data)

    assert restored.value == "A"
    assert restored.node_b.value == "B"
    assert restored.node_b.node_c.value == "C"
    assert restored.node_b.node_c.node_d.value == "D"
    assert restored.node_b.node_c.node_d.node_e.value == "E"
    assert restored.node_b.node_c.node_d.node_e.back_to_a is None


def test_fully_optional_chain():
    optional = FullyOptionalA(
        id_="A",
        b=FullyOptionalB(
            id_="B",
            c=FullyOptionalC(
                id_="C",
                d=FullyOptionalD(
                    id_="D",
                    a=None,
                ),
            ),
        ),
    )

    data = optional.to_dict()
    restored = FullyOptionalA.from_dict(data)

    assert restored.id_ == "A"
    assert restored.b is not None
    assert restored.b.id_ == "B"
    assert restored.b.c is not None
    assert restored.b.c.id_ == "C"
    assert restored.b.c.d is not None
    assert restored.b.c.d.id_ == "D"
    assert restored.b.c.d.a is None


if __name__ == "__main__":
    test_self_referencing()
    test_chain_with_optional()
    test_fully_optional_chain()
    print("Success")

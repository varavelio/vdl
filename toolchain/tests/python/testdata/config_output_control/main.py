import os
import sys

sys.path.append(os.getcwd())

from gen import *


def test_types_still_work():
    data = SimpleType(id_="123", name="Test")
    assert data.id_ == "123"
    assert data.name == "Test"
    restored = SimpleType.from_dict(data.to_dict())
    assert restored.id_ == data.id_


def test_constants_file_not_generated():
    assert not os.path.exists("gen/constants.py")


def test_patterns_file_not_generated():
    assert not os.path.exists("gen/patterns.py")


def test_init_does_not_export_missing():
    init_file = "gen/__init__.py"
    if os.path.exists(init_file):
        content = open(init_file, "r", encoding="utf-8").read()
        assert "constants" not in content
        assert "patterns" not in content


if __name__ == "__main__":
    test_types_still_work()
    test_constants_file_not_generated()
    test_patterns_file_not_generated()
    test_init_does_not_export_missing()
    print("Success")

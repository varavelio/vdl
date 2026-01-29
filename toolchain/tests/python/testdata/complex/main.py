import sys
import os

sys.path.append(os.getcwd())

from gen.types import Inner, Outer

def test_complex():
    o = Outer(
        list_=[Inner(val=1), Inner(val=2)],
        my_map={"k1": Inner(val=3)},
        inner=Inner(val=4)
    )
    
    d = o.to_dict()
    assert d["list"] == [{"val": 1}, {"val": 2}]
    assert d["myMap"] == {"k1": {"val": 3}}
    assert d["inner"] == {"val": 4}
    
    o2 = Outer.from_dict(d)
    assert len(o2.list_) == 2
    assert o2.list_[0].val == 1
    assert o2.list_[1].val == 2
    assert o2.my_map["k1"].val == 3
    assert o2.inner.val == 4

if __name__ == "__main__":
    try:
        test_complex()
        print("PASS")
    except Exception as e:
        print(f"FAIL: {e}")
        import traceback
        traceback.print_exc()
        sys.exit(1)

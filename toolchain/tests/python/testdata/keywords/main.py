import sys
import os

sys.path.append(os.getcwd())

from gen.types import Reserved

def test_keywords():
    r = Reserved(
        class_="cls",
        def_=1,
        from_=True
    )
    
    assert r.class_ == "cls"
    assert r.def_ == 1
    assert r.from_ is True
    
    d = r.to_dict()
    assert d["class"] == "cls"
    assert d["def"] == 1
    assert d["from"] is True
    
    r2 = Reserved.from_dict(d)
    assert r2.class_ == "cls"
    assert r2.def_ == 1
    assert r2.from_ is True

if __name__ == "__main__":
    try:
        test_keywords()
        print("PASS")
    except Exception as e:
        print(f"FAIL: {e}")
        import traceback
        traceback.print_exc()
        sys.exit(1)

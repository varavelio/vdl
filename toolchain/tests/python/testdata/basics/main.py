import sys
import os

# Add current directory to path so we can import gen
sys.path.append(os.getcwd())

from gen.types import User, Status

def test_basics():
    u = User(
        id_="u1",
        age=30,
        score=99.5,
        active=True,
        status=Status.ACTIVE,
        tags=["a", "b"],
        meta={"key": "val"}
    )
    
    assert u.id_ == "u1"
    assert u.age == 30
    assert u.score == 99.5
    assert u.active is True
    assert u.status == Status.ACTIVE
    assert u.tags == ["a", "b"]
    assert u.meta == {"key": "val"}
    
    # Serialization
    d = u.to_dict()
    assert d["id"] == "u1"
    assert d["age"] == 30
    assert d["score"] == 99.5
    assert d["active"] is True
    assert d["status"] == "Active" # Enum value
    assert d["tags"] == ["a", "b"]
    assert d["meta"] == {"key": "val"}
    
    # Deserialization
    u2 = User.from_dict(d)
    assert u2.id_ == "u1"
    assert u2.age == 30
    assert u2.score == 99.5
    assert u2.active is True
    assert u2.status == Status.ACTIVE
    assert u2.tags == ["a", "b"]
    assert u2.meta == {"key": "val"}

if __name__ == "__main__":
    try:
        test_basics()
        print("PASS")
    except Exception as e:
        print(f"FAIL: {e}")
        # print stack trace
        import traceback
        traceback.print_exc()
        sys.exit(1)

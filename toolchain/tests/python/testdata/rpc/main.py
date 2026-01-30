import sys
import os

sys.path.append(os.getcwd())

from gen.rpc_catalog import VDL_PROCEDURES, VDLPaths, OperationDefinition
from gen.types import GreeterGreetInput, GreeterGreetOutput, GreeterGreetResponse
from gen.core_types import Response

def test_rpc():
    assert len(VDL_PROCEDURES) == 1
    op = VDL_PROCEDURES[0]
    
    assert op.name == "Greet"
    assert op.path == "/Greeter/Greet"
    assert op.type.name == "PROC"
    
    assert VDLPaths.greeter.greet == "/Greeter/Greet"
    
    # Check types
    inp = GreeterGreetInput(name="World")
    assert inp.name == "World"
    
    out = GreeterGreetOutput(message="Hello World")
    assert out.message == "Hello World"
    
    res = GreeterGreetResponse(result=out)
    assert res.result.message == "Hello World"

if __name__ == "__main__":
    try:
        test_rpc()
        print("PASS")
    except Exception as e:
        print(f"FAIL: {e}")
        import traceback
        traceback.print_exc()
        sys.exit(1)

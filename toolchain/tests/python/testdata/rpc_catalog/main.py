import sys
import os

sys.path.append(os.getcwd())

from gen import *


def test_operation_type():
    assert len(OperationType) == 2
    assert OperationType.PROC in list(OperationType)
    assert OperationType.STREAM in list(OperationType)


def test_vdl_procedures():
    assert len(VDL_PROCEDURES) == 6
    proc_names = [f"{p.rpc_name}/{p.name}" for p in VDL_PROCEDURES]
    assert "Chat/Send" in proc_names
    assert "Users/Create" in proc_names
    assert "Users/Delete" in proc_names
    assert "Users/Get" in proc_names
    assert "Users/List" in proc_names
    assert "Users/Remove" in proc_names

    for proc in VDL_PROCEDURES:
        assert proc.type == OperationType.PROC

    chat_send = next(p for p in VDL_PROCEDURES if p.name == "Send")
    assert chat_send.rpc_name == "Chat"
    assert chat_send.path == "/Chat/Send"


def test_vdl_streams():
    assert len(VDL_STREAMS) == 2
    stream_names = [f"{s.rpc_name}/{s.name}" for s in VDL_STREAMS]
    assert "Chat/Messages" in stream_names
    assert "Chat/Typing" in stream_names

    for stream in VDL_STREAMS:
        assert stream.type == OperationType.STREAM

    messages = next(s for s in VDL_STREAMS if s.name == "Messages")
    assert messages.rpc_name == "Chat"
    assert messages.path == "/Chat/Messages"


def test_vdl_paths():
    assert VDLPaths.chat.send == "/Chat/Send"
    assert VDLPaths.chat.messages == "/Chat/Messages"
    assert VDLPaths.chat.typing == "/Chat/Typing"

    assert VDLPaths.users.create == "/Users/Create"
    assert VDLPaths.users.delete == "/Users/Delete"
    assert VDLPaths.users.get == "/Users/Get"
    assert VDLPaths.users.list == "/Users/List"
    assert VDLPaths.users.remove == "/Users/Remove"


def test_operation_definition():
    op = OperationDefinition(rpc_name="TestRpc", name="TestOp", type=OperationType.PROC)
    assert op.rpc_name == "TestRpc"
    assert op.name == "TestOp"
    assert op.type == OperationType.PROC
    assert op.path == "/TestRpc/TestOp"


if __name__ == "__main__":
    test_operation_type()
    test_vdl_procedures()
    test_vdl_streams()
    test_vdl_paths()
    test_operation_definition()
    print("Success")

import datetime
import sys
import os

sys.path.append(os.getcwd())

from gen import *


def test_all_primitives():
    now = datetime.datetime.now(datetime.UTC)

    original = AllPrimitives(
        string_field="hello",
        int_field=42,
        float_field=3.14,
        bool_field=True,
        datetime_field=now,
    )

    data = original.to_dict()
    assert data["stringField"] == "hello"
    assert data["intField"] == 42
    assert data["floatField"] == 3.14
    assert data["boolField"] is True
    assert isinstance(data["datetimeField"], str)

    restored = AllPrimitives.from_dict(data)
    assert restored.string_field == original.string_field
    assert restored.int_field == original.int_field
    assert restored.float_field == original.float_field
    assert restored.bool_field == original.bool_field
    assert abs((restored.datetime_field - original.datetime_field).total_seconds()) < 1


def test_optional_primitives():
    empty = OptionalPrimitives()
    empty_json = empty.to_dict()
    assert empty_json == {}

    partial = OptionalPrimitives(string_field="test", int_field=123)
    partial_json = partial.to_dict()
    assert partial_json["stringField"] == "test"
    assert partial_json["intField"] == 123
    assert "floatField" not in partial_json
    assert "boolField" not in partial_json
    assert "datetimeField" not in partial_json

    restored = OptionalPrimitives.from_dict(partial_json)
    assert restored.string_field == "test"
    assert restored.int_field == 123
    assert restored.float_field is None
    assert restored.bool_field is None
    assert restored.datetime_field is None


def test_procedure_types():
    now = datetime.datetime.now(datetime.UTC)
    input_data = ServiceEchoInput(
        data=AllPrimitives(
            string_field="test",
            int_field=1,
            float_field=1.5,
            bool_field=False,
            datetime_field=now,
        )
    )
    input_json = input_data.to_dict()
    assert isinstance(input_json["data"], dict)
    input_restored = ServiceEchoInput.from_dict(input_json)
    assert input_restored.data.string_field == "test"

    output_data = ServiceEchoOutput(
        data=AllPrimitives(
            string_field="response",
            int_field=2,
            float_field=2.5,
            bool_field=True,
            datetime_field=now,
        )
    )
    output_json = output_data.to_dict()
    assert isinstance(output_json["data"], dict)

    response = Response(result=output_data)
    assert response.result.data.string_field == "response"

    err = VdlError(message="Something went wrong", code="ERR001")
    error_response = Response(result=None, error=err)
    assert error_response.error.message == "Something went wrong"
    assert error_response.error.code == "ERR001"


if __name__ == "__main__":
    test_all_primitives()
    test_optional_primitives()
    test_procedure_types()
    print("Success")

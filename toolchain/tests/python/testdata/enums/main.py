import sys
import os

sys.path.append(os.getcwd())

from gen import *


def test_color_enum():
    assert Color.RED.value == "red"
    assert Color.GREEN.value == "green"
    assert Color.BLUE.value == "blue"

    assert Color.from_value("red") == Color.RED
    assert Color.from_value("green") == Color.GREEN
    assert Color.from_value("blue") == Color.BLUE
    assert Color.from_value("invalid") is None

    assert Color.RED.value == "red"


def test_status_enum():
    assert Status.PENDING.value == "Pending"
    assert Status.ACTIVE.value == "Active"
    assert Status.COMPLETED.value == "Completed"

    assert Status.from_value("Pending") == Status.PENDING
    assert Status.from_value("Active") == Status.ACTIVE
    assert Status.from_value("Completed") == Status.COMPLETED


def test_priority_enum():
    assert Priority.LOW.value == 1
    assert Priority.MEDIUM.value == 2
    assert Priority.HIGH.value == 3
    assert Priority.CRITICAL.value == 10

    assert Priority.from_value(1) == Priority.LOW
    assert Priority.from_value(10) == Priority.CRITICAL
    assert Priority.from_value(999) is None


def test_task_with_enums():
    task = Task(
        name="Test Task",
        status=Status.ACTIVE,
        priority=Priority.HIGH,
        color=Color.BLUE,
    )

    data = task.to_dict()
    assert data["name"] == "Test Task"
    assert data["status"] == "Active"
    assert data["priority"] == 3
    assert data["color"] == "blue"

    restored = Task.from_dict(data)
    assert restored.name == "Test Task"
    assert restored.status == Status.ACTIVE
    assert restored.priority == Priority.HIGH
    assert restored.color == Color.BLUE

    no_color = Task(name="No Color", status=Status.PENDING, priority=Priority.LOW)
    json_no_color = no_color.to_dict()
    assert "color" not in json_no_color
    restored_no_color = Task.from_dict(json_no_color)
    assert restored_no_color.color is None


def test_procedure_with_enums():
    input_data = ServiceCreateTaskInput(
        name="New Task",
        status=Status.PENDING,
        priority=Priority.MEDIUM,
    )
    input_json = input_data.to_dict()
    assert input_json["status"] == "Pending"
    assert input_json["priority"] == 2

    input_restored = ServiceCreateTaskInput.from_dict(input_json)
    assert input_restored.status == Status.PENDING
    assert input_restored.priority == Priority.MEDIUM

    output_data = ServiceCreateTaskOutput(
        task=Task(name="Created", status=Status.ACTIVE, priority=Priority.HIGH)
    )
    output_json = output_data.to_dict()
    assert output_json["task"]["status"] == "Active"

    output_restored = ServiceCreateTaskOutput.from_dict(output_json)
    assert output_restored.task.status == Status.ACTIVE


if __name__ == "__main__":
    test_color_enum()
    test_status_enum()
    test_priority_enum()
    test_task_with_enums()
    test_procedure_with_enums()
    print("Success")

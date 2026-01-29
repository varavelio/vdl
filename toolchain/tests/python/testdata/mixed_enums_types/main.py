import json
import sys
import os

sys.path.append(os.getcwd())

from gen import *


def test_enum_list_serialization():
    task_list = TaskList(
        statuses=[Status.PENDING, Status.ACTIVE, Status.COMPLETED],
        priorities=[Priority.HIGH, Priority.LOW],
        categories=[Category.ELECTRONICS, Category.FOOD],
    )
    data = task_list.to_dict()
    statuses = data["statuses"]
    assert len(statuses) == 3
    assert statuses[0] == "Pending"
    assert statuses[1] == "Active"
    assert statuses[2] == "Completed"

    priorities = data["priorities"]
    assert priorities[0] == "High"
    assert priorities[1] == "Low"

    categories = data["categories"]
    assert categories[0] == "ELECTRONICS"
    assert categories[1] == "FOOD"

    with_null = TaskList(
        statuses=[Status.PENDING], priorities=[Priority.MEDIUM], categories=None
    )
    data_null = with_null.to_dict()
    assert "categories" not in data_null


def test_enum_map_serialization():
    status_map = StatusMap(
        tasks_by_status={
            "task1": Status.PENDING,
            "task2": Status.ACTIVE,
            "task3": Status.COMPLETED,
        },
        priority_by_name={
            "urgent": Priority.CRITICAL,
            "normal": Priority.MEDIUM,
            "backlog": Priority.LOW,
        },
    )
    data = status_map.to_dict()
    tasks_by_status = data["tasksByStatus"]
    assert tasks_by_status["task1"] == "Pending"
    assert tasks_by_status["task2"] == "Active"
    assert tasks_by_status["task3"] == "Completed"

    priority_by_name = data["priorityByName"]
    assert priority_by_name["urgent"] == "Critical"
    assert priority_by_name["normal"] == "Medium"


def test_complex_enum_containers():
    complex_data = ComplexEnumContainers(
        tasks=[
            Task(id_=1, title="First", status=Status.PENDING, priority=Priority.HIGH),
            Task(id_=2, title="Second", status=Status.ACTIVE, priority=Priority.LOW),
        ],
        status_groups={
            "todo": [Status.PENDING],
            "in_progress": [Status.ACTIVE],
            "done": [Status.COMPLETED, Status.CANCELLED],
        },
        priority_matrix=[[Priority.HIGH, Priority.CRITICAL], [Priority.LOW, Priority.MEDIUM]],
    )
    data = complex_data.to_dict()
    tasks = data["tasks"]
    assert len(tasks) == 2
    assert tasks[0]["status"] == "Pending"
    assert tasks[0]["priority"] == "High"

    status_groups = data["statusGroups"]
    done_group = status_groups["done"]
    assert len(done_group) == 2
    assert done_group[0] == "Completed"
    assert done_group[1] == "Cancelled"

    matrix = data["priorityMatrix"]
    first_row = matrix[0]
    assert first_row[0] == "High"
    assert first_row[1] == "Critical"


def test_enum_deserialization():
    json_str = '{"statuses":["Pending","Active","Completed"],"priorities":["Low","High"],"categories":["ELECTRONICS","OTHER"]}'
    task_list = TaskList.from_dict(json.loads(json_str))
    assert len(task_list.statuses) == 3
    assert task_list.statuses[0] == Status.PENDING
    assert task_list.statuses[1] == Status.ACTIVE
    assert task_list.statuses[2] == Status.COMPLETED

    assert len(task_list.priorities) == 2
    assert task_list.priorities[0] == Priority.LOW
    assert task_list.priorities[1] == Priority.HIGH

    assert task_list.categories is not None
    assert len(task_list.categories) == 2
    assert task_list.categories[0] == Category.ELECTRONICS
    assert task_list.categories[1] == Category.OTHER


def test_enum_round_trip():
    original = ComplexEnumContainers(
        tasks=[Task(id_=1, title="Test", status=Status.ACTIVE, priority=Priority.CRITICAL)],
        status_groups={
            "all": [Status.PENDING, Status.ACTIVE, Status.COMPLETED, Status.CANCELLED]
        },
        priority_matrix=[
            [Priority.LOW],
            [Priority.MEDIUM],
            [Priority.HIGH],
            [Priority.CRITICAL],
        ],
    )
    parsed = ComplexEnumContainers.from_dict(json.loads(json.dumps(original.to_dict())))
    assert len(parsed.tasks) == 1
    assert parsed.tasks[0].status == Status.ACTIVE
    assert parsed.tasks[0].priority == Priority.CRITICAL
    assert len(parsed.status_groups["all"]) == 4
    assert parsed.status_groups["all"][0] == Status.PENDING
    assert parsed.priority_matrix is not None
    assert len(parsed.priority_matrix) == 4
    assert parsed.priority_matrix[3][0] == Priority.CRITICAL


def test_custom_enum_values():
    task_list = TaskList(
        statuses=[Status.PENDING],
        priorities=[Priority.HIGH],
        categories=[
            Category.ELECTRONICS,
            Category.CLOTHING,
            Category.FOOD,
            Category.OTHER,
        ],
    )
    data = task_list.to_dict()
    categories = data["categories"]
    assert categories[0] == "ELECTRONICS"
    assert categories[1] == "CLOTHING"
    assert categories[2] == "FOOD"
    assert categories[3] == "OTHER"

    parsed = TaskList.from_dict(
        json.loads(
            '{"statuses":["Pending"],"priorities":["High"],"categories":["ELECTRONICS","OTHER"]}'
        )
    )
    assert parsed.categories[0] == Category.ELECTRONICS
    assert parsed.categories[1] == Category.OTHER


if __name__ == "__main__":
    test_enum_list_serialization()
    test_enum_map_serialization()
    test_complex_enum_containers()
    test_enum_deserialization()
    test_enum_round_trip()
    test_custom_enum_values()
    print("All mixed_enums_types tests passed!")

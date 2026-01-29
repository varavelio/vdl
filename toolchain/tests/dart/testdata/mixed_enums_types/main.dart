import 'dart:convert';
import 'gen/index.dart';

void main() {
  testEnumListSerialization();
  testEnumMapSerialization();
  testComplexEnumContainers();
  testEnumDeserialization();
  testEnumRoundTrip();
  testCustomEnumValues();
  print('All mixed_enums_types tests passed!');
}

void testEnumListSerialization() {
  final taskList = TaskList(
    statuses: [Status.Pending, Status.Active, Status.Completed],
    priorities: [Priority.High, Priority.Low],
    categories: [Category.Electronics, Category.Food],
  );

  final json = taskList.toJson();

  // Verify enum lists serialize to primitive values
  final statuses = json['statuses'] as List;
  assert(statuses.length == 3, 'Should have 3 statuses');
  assert(statuses[0] == 'Pending', 'First status should be "Pending"');
  assert(statuses[1] == 'Active', 'Second status should be "Active"');
  assert(statuses[2] == 'Completed', 'Third status should be "Completed"');

  final priorities = json['priorities'] as List;
  assert(priorities[0] == 'High', 'First priority should be "High"');
  assert(priorities[1] == 'Low', 'Second priority should be "Low"');

  // Custom enum values
  final categories = json['categories'] as List;
  assert(
    categories[0] == 'ELECTRONICS',
    'First category should be "ELECTRONICS"',
  );
  assert(categories[1] == 'FOOD', 'Second category should be "FOOD"');

  // Test with null optional
  final withNull = TaskList(
    statuses: [Status.Pending],
    priorities: [Priority.Medium],
    categories: null,
  );

  final jsonNull = withNull.toJson();
  assert(
    !jsonNull.containsKey('categories'),
    'Null categories should be omitted',
  );
}

void testEnumMapSerialization() {
  // map<Status> means Map<String, Status>
  final statusMap = StatusMap(
    tasksByStatus: {
      'task1': Status.Pending,
      'task2': Status.Active,
      'task3': Status.Completed,
    },
    priorityByName: {
      'urgent': Priority.Critical,
      'normal': Priority.Medium,
      'backlog': Priority.Low,
    },
  );

  final json = statusMap.toJson();

  // Verify enum map values serialize to primitives
  final tasksByStatus = json['tasksByStatus'] as Map<String, dynamic>;
  assert(tasksByStatus['task1'] == 'Pending', 'task1 should be "Pending"');
  assert(tasksByStatus['task2'] == 'Active', 'task2 should be "Active"');
  assert(tasksByStatus['task3'] == 'Completed', 'task3 should be "Completed"');

  final priorityByName = json['priorityByName'] as Map<String, dynamic>;
  assert(priorityByName['urgent'] == 'Critical', 'urgent should be "Critical"');
  assert(priorityByName['normal'] == 'Medium', 'normal should be "Medium"');
}

void testComplexEnumContainers() {
  final complex = ComplexEnumContainers(
    tasks: [
      Task(
        id: 1,
        title: 'First',
        status: Status.Pending,
        priority: Priority.High,
      ),
      Task(
        id: 2,
        title: 'Second',
        status: Status.Active,
        priority: Priority.Low,
      ),
    ],
    statusGroups: {
      'todo': [Status.Pending],
      'in_progress': [Status.Active],
      'done': [Status.Completed, Status.Cancelled],
    },
    priorityMatrix: [
      [Priority.High, Priority.Critical],
      [Priority.Low, Priority.Medium],
    ],
  );

  final json = complex.toJson();

  // Verify nested tasks with enums
  final tasks = json['tasks'] as List;
  assert(tasks.length == 2, 'Should have 2 tasks');
  final firstTask = tasks[0] as Map<String, dynamic>;
  assert(
    firstTask['status'] == 'Pending',
    'First task status should be "Pending"',
  );
  assert(
    firstTask['priority'] == 'High',
    'First task priority should be "High"',
  );

  // Verify map of enum arrays
  final statusGroups = json['statusGroups'] as Map<String, dynamic>;
  final doneGroup = statusGroups['done'] as List;
  assert(doneGroup.length == 2, 'done group should have 2 items');
  assert(
    doneGroup[0] == 'Completed',
    'First done status should be "Completed"',
  );
  assert(
    doneGroup[1] == 'Cancelled',
    'Second done status should be "Cancelled"',
  );

  // Verify 2D enum array
  final matrix = json['priorityMatrix'] as List;
  final firstRow = matrix[0] as List;
  assert(firstRow[0] == 'High', 'matrix[0][0] should be "High"');
  assert(firstRow[1] == 'Critical', 'matrix[0][1] should be "Critical"');
}

void testEnumDeserialization() {
  // Test deserializing enums from JSON strings
  final jsonStr = '''
  {
    "statuses": ["Pending", "Active", "Completed"],
    "priorities": ["Low", "High"],
    "categories": ["ELECTRONICS", "OTHER"]
  }
  ''';

  final taskList = TaskList.fromJson(jsonDecode(jsonStr));

  assert(taskList.statuses.length == 3, 'Should have 3 statuses');
  assert(taskList.statuses[0] == Status.Pending, 'First should be Pending');
  assert(taskList.statuses[1] == Status.Active, 'Second should be Active');
  assert(taskList.statuses[2] == Status.Completed, 'Third should be Completed');

  assert(taskList.priorities.length == 2, 'Should have 2 priorities');
  assert(
    taskList.priorities[0] == Priority.Low,
    'First priority should be Low',
  );
  assert(
    taskList.priorities[1] == Priority.High,
    'Second priority should be High',
  );

  // Custom enum values
  assert(taskList.categories != null, 'categories should be present');
  assert(taskList.categories!.length == 2, 'Should have 2 categories');
  assert(
    taskList.categories![0] == Category.Electronics,
    'First should be Electronics',
  );
  assert(taskList.categories![1] == Category.Other, 'Second should be Other');
}

void testEnumRoundTrip() {
  // Full round-trip test
  final original = ComplexEnumContainers(
    tasks: [
      Task(
        id: 1,
        title: 'Test',
        status: Status.Active,
        priority: Priority.Critical,
      ),
    ],
    statusGroups: {
      'all': [
        Status.Pending,
        Status.Active,
        Status.Completed,
        Status.Cancelled,
      ],
    },
    priorityMatrix: [
      [Priority.Low],
      [Priority.Medium],
      [Priority.High],
      [Priority.Critical],
    ],
  );

  // Serialize to JSON string
  final jsonStr = jsonEncode(original.toJson());

  // Deserialize back
  final parsed = ComplexEnumContainers.fromJson(jsonDecode(jsonStr));

  // Verify
  assert(parsed.tasks.length == 1, 'Should have 1 task');
  assert(
    parsed.tasks[0].status == Status.Active,
    'Task status should be Active',
  );
  assert(
    parsed.tasks[0].priority == Priority.Critical,
    'Task priority should be Critical',
  );

  assert(
    parsed.statusGroups['all']!.length == 4,
    'all group should have 4 statuses',
  );
  assert(
    parsed.statusGroups['all']![0] == Status.Pending,
    'First all status should be Pending',
  );

  assert(parsed.priorityMatrix != null, 'priorityMatrix should be present');
  assert(parsed.priorityMatrix!.length == 4, 'Should have 4 rows');
  assert(
    parsed.priorityMatrix![3][0] == Priority.Critical,
    'Last row should be Critical',
  );
}

void testCustomEnumValues() {
  // Test that enums with custom string values work correctly
  final taskList = TaskList(
    statuses: [Status.Pending],
    priorities: [Priority.High],
    categories: [
      Category.Electronics,
      Category.Clothing,
      Category.Food,
      Category.Other,
    ],
  );

  final json = taskList.toJson();
  final categories = json['categories'] as List;

  // Verify custom values are used
  assert(categories[0] == 'ELECTRONICS', 'Should use custom value ELECTRONICS');
  assert(categories[1] == 'CLOTHING', 'Should use custom value CLOTHING');
  assert(categories[2] == 'FOOD', 'Should use custom value FOOD');
  assert(categories[3] == 'OTHER', 'Should use custom value OTHER');

  // Verify deserialization with custom values
  final jsonStr =
      '{"statuses":["Pending"],"priorities":["High"],"categories":["ELECTRONICS","OTHER"]}';
  final parsed = TaskList.fromJson(jsonDecode(jsonStr));
  assert(
    parsed.categories![0] == Category.Electronics,
    'Should parse ELECTRONICS correctly',
  );
  assert(
    parsed.categories![1] == Category.Other,
    'Should parse OTHER correctly',
  );
}

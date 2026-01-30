// Dart E2E test for enums
// This test verifies that both string and integer enums are generated correctly
// and can be serialized/deserialized properly.

import 'gen/index.dart';

void main() {
  // Test string enum with explicit values
  testColorEnum();

  // Test string enum with implicit values
  testStatusEnum();

  // Test integer enum
  testPriorityEnum();

  // Test enums in types
  testTaskWithEnums();

  // Test enums in procedure types
  testProcedureWithEnums();

  print('Success');
}

void testColorEnum() {
  // Test all Color values
  assert(Color.Red.value == 'red', 'Color.Red value should be "red"');
  assert(Color.Green.value == 'green', 'Color.Green value should be "green"');
  assert(Color.Blue.value == 'blue', 'Color.Blue value should be "blue"');

  // Test fromValue
  assert(
    Color.fromValue('red') == Color.Red,
    'fromValue("red") should return Color.Red',
  );
  assert(
    Color.fromValue('green') == Color.Green,
    'fromValue("green") should return Color.Green',
  );
  assert(
    Color.fromValue('blue') == Color.Blue,
    'fromValue("blue") should return Color.Blue',
  );
  assert(
    Color.fromValue('invalid') == null,
    'fromValue("invalid") should return null',
  );

  // Test JSON extension
  assert(Color.Red.toJson() == 'red', 'Color.Red.toJson() should return "red"');
  assert(
    ColorJson.fromJson('green') == Color.Green,
    'ColorJson.fromJson should work',
  );

  // Test list
  assert(colorList.length == 3, 'colorList should have 3 values');
  assert(colorList.contains(Color.Red), 'colorList should contain Color.Red');
}

void testStatusEnum() {
  // Test implicit values (member name used)
  assert(
    Status.Pending.value == 'Pending',
    'Status.Pending value should be "Pending"',
  );
  assert(
    Status.Active.value == 'Active',
    'Status.Active value should be "Active"',
  );
  assert(
    Status.Completed.value == 'Completed',
    'Status.Completed value should be "Completed"',
  );

  // Test fromValue
  assert(
    Status.fromValue('Pending') == Status.Pending,
    'fromValue should work for Status',
  );
  assert(
    Status.fromValue('Active') == Status.Active,
    'fromValue should work for Status',
  );

  // Test JSON
  assert(
    Status.Active.toJson() == 'Active',
    'Status.Active.toJson() should return "Active"',
  );
  assert(
    StatusJson.fromJson('Completed') == Status.Completed,
    'StatusJson.fromJson should work',
  );
}

void testPriorityEnum() {
  // Test integer values
  assert(Priority.Low.value == 1, 'Priority.Low value should be 1');
  assert(Priority.Medium.value == 2, 'Priority.Medium value should be 2');
  assert(Priority.High.value == 3, 'Priority.High value should be 3');
  assert(Priority.Critical.value == 10, 'Priority.Critical value should be 10');

  // Test fromValue
  assert(
    Priority.fromValue(1) == Priority.Low,
    'fromValue(1) should return Priority.Low',
  );
  assert(
    Priority.fromValue(10) == Priority.Critical,
    'fromValue(10) should return Priority.Critical',
  );
  assert(Priority.fromValue(999) == null, 'fromValue(999) should return null');

  // Test JSON
  assert(Priority.High.toJson() == 3, 'Priority.High.toJson() should return 3');
  assert(
    PriorityJson.fromJson(2) == Priority.Medium,
    'PriorityJson.fromJson should work',
  );
}

void testTaskWithEnums() {
  // Create a Task with enums
  final task = Task(
    name: 'Test Task',
    status: Status.Active,
    priority: Priority.High,
    color: Color.Blue,
  );

  // Serialize to JSON
  final json = task.toJson();
  assert(json['name'] == 'Test Task', 'name should be serialized');
  assert(json['status'] == 'Active', 'status should be serialized as string');
  assert(json['priority'] == 3, 'priority should be serialized as int');
  assert(json['color'] == 'blue', 'color should be serialized as string');

  // Deserialize from JSON
  final deserialized = Task.fromJson(json);
  assert(deserialized.name == 'Test Task', 'name should be deserialized');
  assert(deserialized.status == Status.Active, 'status should be deserialized');
  assert(
    deserialized.priority == Priority.High,
    'priority should be deserialized',
  );
  assert(deserialized.color == Color.Blue, 'color should be deserialized');

  // Test with optional enum null
  final taskNoColor = Task(
    name: 'No Color Task',
    status: Status.Pending,
    priority: Priority.Low,
  );
  final jsonNoColor = taskNoColor.toJson();
  assert(
    !jsonNoColor.containsKey('color'),
    'null optional enum should not be in JSON',
  );

  final deserializedNoColor = Task.fromJson(jsonNoColor);
  assert(deserializedNoColor.color == null, 'optional enum should be null');
}

void testProcedureWithEnums() {
  // Test input type with enums
  final input = ServiceCreateTaskInput(
    name: 'New Task',
    status: Status.Pending,
    priority: Priority.Medium,
  );

  final inputJson = input.toJson();
  assert(
    inputJson['status'] == 'Pending',
    'status in input should be serialized',
  );
  assert(inputJson['priority'] == 2, 'priority in input should be serialized');

  // Deserialize input
  final inputDeserialized = ServiceCreateTaskInput.fromJson(inputJson);
  assert(
    inputDeserialized.status == Status.Pending,
    'status should be deserialized',
  );
  assert(
    inputDeserialized.priority == Priority.Medium,
    'priority should be deserialized',
  );

  // Test output type with nested Task containing enums
  final output = ServiceCreateTaskOutput(
    task: Task(
      name: 'Created Task',
      status: Status.Active,
      priority: Priority.High,
    ),
  );

  final outputJson = output.toJson();
  assert(
    (outputJson['task'] as Map)['status'] == 'Active',
    'nested enum should be serialized',
  );

  final outputDeserialized = ServiceCreateTaskOutput.fromJson(outputJson);
  assert(
    outputDeserialized.task.status == Status.Active,
    'nested enum should be deserialized',
  );
}

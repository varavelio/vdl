// Dart E2E test for RPC catalog
// This test verifies that the RPC catalog (vdlProcedures, vdlStreams, VDLPaths)
// is generated correctly and provides accurate introspection.

import 'gen/index.dart';

void main() {
  // Test OperationType enum
  testOperationType();

  // Test procedure catalog
  testVdlProcedures();

  // Test stream catalog
  testVdlStreams();

  // Test VDLPaths
  testVdlPaths();

  // Test OperationDefinition
  testOperationDefinition();

  print('Success');
}

void testOperationType() {
  // Verify OperationType enum values
  assert(
    OperationType.values.length == 2,
    'OperationType should have 2 values',
  );
  assert(
    OperationType.values.contains(OperationType.proc),
    'OperationType should have proc',
  );
  assert(
    OperationType.values.contains(OperationType.stream),
    'OperationType should have stream',
  );
}

void testVdlProcedures() {
  // Verify all procedures are listed
  assert(vdlProcedures.length == 6, 'Should have 6 procedures');

  // Check procedure names
  final procNames = vdlProcedures.map((p) => '${p.rpcName}/${p.name}').toList();

  assert(procNames.contains('Chat/Send'), 'Should contain Chat/Send');
  assert(procNames.contains('Users/Create'), 'Should contain Users/Create');
  assert(procNames.contains('Users/Delete'), 'Should contain Users/Delete');
  assert(procNames.contains('Users/Get'), 'Should contain Users/Get');
  assert(procNames.contains('Users/List'), 'Should contain Users/List');
  assert(procNames.contains('Users/Remove'), 'Should contain Users/Remove');

  // Verify all are of type proc
  for (final proc in vdlProcedures) {
    assert(proc.type == OperationType.proc, '${proc.name} should be type proc');
  }

  // Check specific procedure details
  final chatSend = vdlProcedures.firstWhere((p) => p.name == 'Send');
  assert(chatSend.rpcName == 'Chat', 'Send should be in Chat RPC');
  assert(chatSend.path == '/Chat/Send', 'Send path should be /Chat/Send');
}

void testVdlStreams() {
  // Verify all streams are listed
  assert(vdlStreams.length == 2, 'Should have 2 streams');

  // Check stream names
  final streamNames = vdlStreams.map((s) => '${s.rpcName}/${s.name}').toList();

  assert(streamNames.contains('Chat/Messages'), 'Should contain Chat/Messages');
  assert(streamNames.contains('Chat/Typing'), 'Should contain Chat/Typing');

  // Verify all are of type stream
  for (final stream in vdlStreams) {
    assert(
      stream.type == OperationType.stream,
      '${stream.name} should be type stream',
    );
  }

  // Check specific stream details
  final messages = vdlStreams.firstWhere((s) => s.name == 'Messages');
  assert(messages.rpcName == 'Chat', 'Messages should be in Chat RPC');
  assert(
    messages.path == '/Chat/Messages',
    'Messages path should be /Chat/Messages',
  );
}

void testVdlPaths() {
  // Test Chat paths
  assert(VDLPaths.chat.send == '/Chat/Send', 'chat.send path mismatch');
  assert(
    VDLPaths.chat.messages == '/Chat/Messages',
    'chat.messages path mismatch',
  );
  assert(VDLPaths.chat.typing == '/Chat/Typing', 'chat.typing path mismatch');

  // Test Users paths
  assert(
    VDLPaths.users.create == '/Users/Create',
    'users.create path mismatch',
  );
  assert(
    VDLPaths.users.delete == '/Users/Delete',
    'users.delete path mismatch',
  );
  assert(VDLPaths.users.get == '/Users/Get', 'users.get path mismatch');
  assert(VDLPaths.users.list == '/Users/List', 'users.list path mismatch');
  assert(
    VDLPaths.users.remove == '/Users/Remove',
    'users.remove path mismatch',
  );
}

void testOperationDefinition() {
  // Create an OperationDefinition manually and verify
  const op = OperationDefinition(
    rpcName: 'TestRpc',
    name: 'TestOp',
    type: OperationType.proc,
  );

  assert(op.rpcName == 'TestRpc', 'rpcName should be TestRpc');
  assert(op.name == 'TestOp', 'name should be TestOp');
  assert(op.type == OperationType.proc, 'type should be proc');
  assert(op.path == '/TestRpc/TestOp', 'path should be /TestRpc/TestOp');
}

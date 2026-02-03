import 'gen/index.dart';

void main() {
  testSelfReferencing();
  testChainWithOptional();
  testFullyOptionalChain();

  print('Success');
}

void testSelfReferencing() {
  final selfRef = SelfReferencing(
    id: 'root',
    name: 'Root Node',
    parent: SelfReferencing(
      id: 'parent',
      name: 'Parent Node',
      parent: SelfReferencing(
        id: 'grandparent',
        name: 'Grandparent Node',
        parent: null,
      ),
    ),
  );

  final json = selfRef.toJson();
  final deserialized = SelfReferencing.fromJson(json);

  assert(deserialized.id == selfRef.id, 'Self-referencing id mismatch');
  assert(deserialized.name == selfRef.name, 'Self-referencing name mismatch');
  assert(deserialized.parent != null, 'Parent should not be null');
  assert(deserialized.parent!.id == 'parent', 'Parent id mismatch');
  assert(deserialized.parent!.parent != null, 'Grandparent should not be null');
  assert(
    deserialized.parent!.parent!.id == 'grandparent',
    'Grandparent id mismatch',
  );
  assert(
    deserialized.parent!.parent!.parent == null,
    'Grandparent parent should be null',
  );
}

void testChainWithOptional() {
  final chain = NodeA(
    value: 'A',
    nodeB: NodeB(
      value: 'B',
      nodeC: NodeC(
        value: 'C',
        nodeD: NodeD(
          value: 'D',
          nodeE: NodeE(value: 'E', backToA: null),
        ),
      ),
    ),
  );

  final json = chain.toJson();
  final deserialized = NodeA.fromJson(json);

  assert(deserialized.value == 'A', 'NodeA value mismatch');
  assert(deserialized.nodeB.value == 'B', 'NodeB value mismatch');
  assert(deserialized.nodeB.nodeC.value == 'C', 'NodeC value mismatch');
  assert(deserialized.nodeB.nodeC.nodeD.value == 'D', 'NodeD value mismatch');
  assert(
    deserialized.nodeB.nodeC.nodeD.nodeE.value == 'E',
    'NodeE value mismatch',
  );
  assert(
    deserialized.nodeB.nodeC.nodeD.nodeE.backToA == null,
    'BackToA should be null',
  );
}

void testFullyOptionalChain() {
  final optional = FullyOptionalA(
    id: 'A',
    b: FullyOptionalB(
      id: 'B',
      c: FullyOptionalC(
        id: 'C',
        d: FullyOptionalD(id: 'D', a: null),
      ),
    ),
  );

  final json = optional.toJson();
  final deserialized = FullyOptionalA.fromJson(json);

  assert(deserialized.id == 'A', 'FullyOptionalA id mismatch');
  assert(deserialized.b != null, 'FullyOptionalB should not be null');
  assert(deserialized.b!.id == 'B', 'FullyOptionalB id mismatch');
  assert(deserialized.b!.c != null, 'FullyOptionalC should not be null');
  assert(deserialized.b!.c!.id == 'C', 'FullyOptionalC id mismatch');
  assert(deserialized.b!.c!.d != null, 'FullyOptionalD should not be null');
  assert(deserialized.b!.c!.d!.id == 'D', 'FullyOptionalD id mismatch');
  assert(deserialized.b!.c!.d!.a == null, 'FullyOptionalA should be null');
}

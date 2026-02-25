// Verifies spread (type composition with ...) works correctly.
// Tests simple spreads, chained spreads, multiple spreads, spreads with complex nested types,
// and spreads directly in input/output blocks and nested anonymous objects.

import { createServer } from "http";
import { createNodeHandler } from "./gen/adapters/node.ts";
import type {
  Address,
  AuditedEntity,
  ContactInfo,
  Document,
  Order,
  OrderItem,
  Person,
  ServiceConfig,
  User,
  UserWithTimestamps,
} from "./gen/index.ts";
import { type Client, NewClient, Server } from "./gen/index.ts";

function deepEqual(a: unknown, b: unknown): boolean {
  return JSON.stringify(a) === JSON.stringify(b);
}

async function main() {
  const server = new Server();

  // Register all handlers as echo handlers
  server.rpcs
    .service()
    .procs.EchoUser()
    .handle(async ({ input }) => {
      return { user: input.user };
    });

  server.rpcs
    .service()
    .procs.EchoUserWithTimestamps()
    .handle(async ({ input }) => {
      return { user: input.user };
    });

  server.rpcs
    .service()
    .procs.EchoAuditedEntity()
    .handle(async ({ input }) => {
      return { entity: input.entity };
    });

  server.rpcs
    .service()
    .procs.EchoPerson()
    .handle(async ({ input }) => {
      return { person: input.person };
    });

  server.rpcs
    .service()
    .procs.EchoDocument()
    .handle(async ({ input }) => {
      return { doc: input.doc };
    });

  server.rpcs
    .service()
    .procs.EchoServiceConfig()
    .handle(async ({ input }) => {
      return { config: input.config };
    });

  server.rpcs
    .service()
    .procs.EchoOrder()
    .handle(async ({ input }) => {
      return { order: input.order };
    });

  // Handlers for spread in input/output/anonymous
  server.rpcs
    .service()
    .procs.SpreadInInput()
    .handle(async ({ input }) => {
      return {
        id: input.id,
        createdAt: input.createdAt,
        updatedAt: input.updatedAt,
        customField: input.customField,
      };
    });

  server.rpcs
    .service()
    .procs.SpreadInOutput()
    .handle(async ({ input }) => {
      const now = new Date();
      return {
        id: input.id,
        createdAt: now,
        updatedAt: new Date(now.getTime() + 3600000),
        tags: ["generated", "processed"],
        metadata: { source: "server" },
        name: input.name,
        processed: true,
      };
    });

  server.rpcs
    .service()
    .procs.SpreadInNestedAnonymous()
    .handle(async ({ input }) => {
      return {
        wrapper: {
          id: input.wrapper.id,
          data: {
            createdAt: input.wrapper.data.createdAt,
            updatedAt: input.wrapper.data.updatedAt,
            tags: input.wrapper.data.tags,
            metadata: input.wrapper.data.metadata,
            value: input.wrapper.data.value,
          },
        },
      };
    });

  server.rpcs
    .service()
    .procs.DeepNestedSpreads()
    .handle(async ({ input }) => {
      return {
        level1: {
          id: input.level1.id,
          level2: {
            createdAt: input.level1.level2.createdAt,
            updatedAt: input.level1.level2.updatedAt,
            level3: {
              tags: input.level1.level2.level3.tags,
              metadata: input.level1.level2.level3.metadata,
              deepValue: input.level1.level2.level3.deepValue,
            },
          },
        },
      };
    });

  const handler = createNodeHandler(server, undefined, { prefix: "/rpc" });

  const httpServer = createServer(async (req, res) => {
    if (req.method !== "POST") {
      res.writeHead(405);
      res.end();
      return;
    }

    await handler(req, res);
  });

  await new Promise<void>((resolve) => {
    httpServer.listen(0, resolve);
  });

  const addr = httpServer.address() as any;
  const port = addr.port;
  const baseUrl = `http://localhost:${port}/rpc`;

  const client = NewClient(baseUrl).build();

  const now = new Date();
  now.setMilliseconds(0); // Truncate milliseconds for comparison

  try {
    // Test 1: Simple spread (User has ...Identifiable)
    await testSimpleSpread(client);

    // Test 2: Multiple spreads (UserWithTimestamps has ...Identifiable and ...Timestamps)
    await testMultipleSpreads(client, now);

    // Test 3: Chained spreads (AuditedEntity has ...Auditable which has ...Timestamps)
    await testChainedSpreads(client, now);

    // Test 4: Spread with nested objects (Person has ...Identifiable, ...Timestamps, and contact: ContactInfo)
    await testSpreadWithNestedObjects(client, now);

    // Test 5: Spread with arrays and maps (Document has ...Identifiable, ...Timestamps, ...Taggable)
    await testSpreadWithArraysAndMaps(client, now);

    // Test 6: Deep chain of spreads (ServiceConfig has ...AdvancedConfig which has ...BaseConfig)
    await testDeepChainOfSpreads(client);

    // Test 7: Spread with array of objects (Order has ...Identifiable, ...Timestamps, items: OrderItem[])
    await testSpreadWithArrayOfObjects(client, now);

    // Test 8: Spread directly in input block
    await testSpreadInInput(client, now);

    // Test 9: Spread directly in output block
    await testSpreadInOutput(client);

    // Test 10: Spread in nested anonymous objects
    await testSpreadInNestedAnonymous(client, now);

    // Test 11: Deep nested spreads in anonymous objects
    await testDeepNestedSpreads(client, now);

    console.log("Success");
  } catch (e) {
    console.error("Error:", e);
    process.exit(1);
  }

  httpServer.close();
  process.exit(0);
}

async function testSimpleSpread(client: Client) {
  const user: User = {
    id: "user-123",
    name: "Alice",
    email: "alice@example.com",
  };

  const res = await client.procs.serviceEchoUser().execute({ user });

  if (!deepEqual(res.user, user)) {
    throw new Error(`User mismatch: got ${JSON.stringify(res.user)}, want ${JSON.stringify(user)}`);
  }
}

async function testMultipleSpreads(client: Client, now: Date) {
  const user: UserWithTimestamps = {
    id: "user-456",
    createdAt: now,
    updatedAt: new Date(now.getTime() + 3600000),
    name: "Bob",
    email: "bob@example.com",
    deletedAt: new Date(now.getTime() + 24 * 3600000),
  };

  const res = await client.procs.serviceEchoUserWithTimestamps().execute({ user });

  if (!deepEqual(res.user, user)) {
    throw new Error(
      `UserWithTimestamps mismatch: got ${JSON.stringify(res.user)}, want ${JSON.stringify(user)}`,
    );
  }
}

async function testChainedSpreads(client: Client, now: Date) {
  const entity: AuditedEntity = {
    createdAt: now,
    updatedAt: new Date(now.getTime() + 3600000),
    createdBy: "system",
    updatedBy: "admin",
    entityType: "document",
    entityId: "doc-789",
  };

  const res = await client.procs.serviceEchoAuditedEntity().execute({ entity });

  if (!deepEqual(res.entity, entity)) {
    throw new Error(
      `AuditedEntity mismatch: got ${JSON.stringify(res.entity)}, want ${JSON.stringify(entity)}`,
    );
  }
}

async function testSpreadWithNestedObjects(client: Client, now: Date) {
  const address: Address = {
    street: "123 Main St",
    city: "Springfield",
    country: "USA",
  };

  const contact: ContactInfo = {
    email: "charlie@example.com",
    phone: "+1-555-1234",
    address: address,
  };

  const person: Person = {
    id: "person-111",
    createdAt: now,
    updatedAt: new Date(now.getTime() + 60000),
    name: "Charlie",
    contact: contact,
  };

  const res = await client.procs.serviceEchoPerson().execute({ person });

  if (!deepEqual(res.person, person)) {
    throw new Error(
      `Person mismatch: got ${JSON.stringify(res.person)}, want ${JSON.stringify(person)}`,
    );
  }
}

async function testSpreadWithArraysAndMaps(client: Client, now: Date) {
  const doc: Document = {
    id: "doc-222",
    createdAt: now,
    updatedAt: new Date(now.getTime() + 1000),
    tags: ["important", "urgent", "draft"],
    metadata: {
      author: "Diana",
      version: "1.0",
      category: "technical",
    },
    title: "Technical Specification",
    content: "This document describes the system architecture...",
  };

  const res = await client.procs.serviceEchoDocument().execute({ doc });

  if (!deepEqual(res.doc, doc)) {
    throw new Error(
      `Document mismatch: got ${JSON.stringify(res.doc)}, want ${JSON.stringify(doc)}`,
    );
  }
}

async function testDeepChainOfSpreads(client: Client) {
  const config: ServiceConfig = {
    enabled: true,
    priority: 10,
    retryCount: 3,
    timeout: 5000,
    serviceName: "api-gateway",
    endpoints: ["https://api.example.com", "https://api-backup.example.com"],
  };

  const res = await client.procs.serviceEchoServiceConfig().execute({ config });

  if (!deepEqual(res.config, config)) {
    throw new Error(
      `ServiceConfig mismatch: got ${JSON.stringify(res.config)}, want ${JSON.stringify(config)}`,
    );
  }
}

async function testSpreadWithArrayOfObjects(client: Client, now: Date) {
  const items: OrderItem[] = [
    { productId: "prod-1", quantity: 2, price: 29.99 },
    { productId: "prod-2", quantity: 1, price: 149.5 },
    { productId: "prod-3", quantity: 5, price: 9.99 },
  ];

  const order: Order = {
    id: "order-333",
    createdAt: now,
    updatedAt: new Date(now.getTime() + 500),
    items: items,
    total: 259.43,
    status: "pending",
  };

  const res = await client.procs.serviceEchoOrder().execute({ order });

  if (!deepEqual(res.order, order)) {
    throw new Error(
      `Order mismatch: got ${JSON.stringify(res.order)}, want ${JSON.stringify(order)}`,
    );
  }
}

async function testSpreadInInput(client: Client, now: Date) {
  const input = {
    id: "input-spread-1",
    createdAt: now,
    updatedAt: new Date(now.getTime() + 60000),
    customField: "my-custom-value",
  };

  const res = await client.procs.serviceSpreadInInput().execute(input);

  if (res.id !== input.id) {
    throw new Error(`Id mismatch: got ${res.id}, want ${input.id}`);
  }
  if (!deepEqual(res.createdAt, input.createdAt)) {
    throw new Error(`CreatedAt mismatch`);
  }
  if (!deepEqual(res.updatedAt, input.updatedAt)) {
    throw new Error(`UpdatedAt mismatch`);
  }
  if (res.customField !== input.customField) {
    throw new Error(`CustomField mismatch: got ${res.customField}, want ${input.customField}`);
  }
}

async function testSpreadInOutput(client: Client) {
  const input = {
    id: "output-spread-1",
    name: "test-name",
  };

  const res = await client.procs.serviceSpreadInOutput().execute(input);

  // Verify spread fields from Identifiable
  if (res.id !== input.id) {
    throw new Error(`Id mismatch: got ${res.id}, want ${input.id}`);
  }

  // Verify spread fields from Timestamps (server generates these)
  if (!res.createdAt) {
    throw new Error("CreatedAt should not be empty");
  }
  if (!res.updatedAt) {
    throw new Error("UpdatedAt should not be empty");
  }

  // Verify spread fields from Taggable
  if (!res.tags || res.tags.length === 0) {
    throw new Error("Tags should not be empty");
  }
  if (!res.metadata || Object.keys(res.metadata).length === 0) {
    throw new Error("Metadata should not be empty");
  }

  // Verify own fields
  if (res.name !== input.name) {
    throw new Error(`Name mismatch: got ${res.name}, want ${input.name}`);
  }
  if (!res.processed) {
    throw new Error("Processed should be true");
  }
}

async function testSpreadInNestedAnonymous(client: Client, now: Date) {
  const input = {
    wrapper: {
      id: "nested-anon-1",
      data: {
        createdAt: now,
        updatedAt: new Date(now.getTime() + 3600000),
        tags: ["nested", "anonymous"],
        metadata: { key: "value" },
        value: "inner-value",
      },
    },
  };

  const res = await client.procs.serviceSpreadInNestedAnonymous().execute(input);

  // Verify wrapper.id from Identifiable spread
  if (res.wrapper.id !== input.wrapper.id) {
    throw new Error(`Wrapper.Id mismatch: got ${res.wrapper.id}, want ${input.wrapper.id}`);
  }

  // Verify wrapper.data fields
  if (!deepEqual(res.wrapper.data.createdAt, input.wrapper.data.createdAt)) {
    throw new Error("Wrapper.Data.CreatedAt mismatch");
  }
  if (!deepEqual(res.wrapper.data.updatedAt, input.wrapper.data.updatedAt)) {
    throw new Error("Wrapper.Data.UpdatedAt mismatch");
  }
  if (!deepEqual(res.wrapper.data.tags, input.wrapper.data.tags)) {
    throw new Error("Wrapper.Data.Tags mismatch");
  }
  if (!deepEqual(res.wrapper.data.metadata, input.wrapper.data.metadata)) {
    throw new Error("Wrapper.Data.Metadata mismatch");
  }
  if (res.wrapper.data.value !== input.wrapper.data.value) {
    throw new Error("Wrapper.Data.Value mismatch");
  }
}

async function testDeepNestedSpreads(client: Client, now: Date) {
  const input = {
    level1: {
      id: "deep-1",
      level2: {
        createdAt: now,
        updatedAt: new Date(now.getTime() + 30 * 60000),
        level3: {
          tags: ["level3", "deep"],
          metadata: { depth: "3" },
          deepValue: "innermost-value",
        },
      },
    },
  };

  const res = await client.procs.serviceDeepNestedSpreads().execute(input);

  // Verify level1 (Identifiable spread)
  if (res.level1.id !== input.level1.id) {
    throw new Error("Level1.Id mismatch");
  }

  // Verify level2 (Timestamps spread)
  if (!deepEqual(res.level1.level2.createdAt, input.level1.level2.createdAt)) {
    throw new Error("Level1.Level2.CreatedAt mismatch");
  }
  if (!deepEqual(res.level1.level2.updatedAt, input.level1.level2.updatedAt)) {
    throw new Error("Level1.Level2.UpdatedAt mismatch");
  }

  // Verify level3 (Taggable spread)
  if (!deepEqual(res.level1.level2.level3.tags, input.level1.level2.level3.tags)) {
    throw new Error("Level1.Level2.Level3.Tags mismatch");
  }
  if (!deepEqual(res.level1.level2.level3.metadata, input.level1.level2.level3.metadata)) {
    throw new Error("Level1.Level2.Level3.Metadata mismatch");
  }
  if (res.level1.level2.level3.deepValue !== input.level1.level2.level3.deepValue) {
    throw new Error("Level1.Level2.Level3.DeepValue mismatch");
  }
}

main().catch((err) => {
  console.error(err);
  process.exit(1);
});

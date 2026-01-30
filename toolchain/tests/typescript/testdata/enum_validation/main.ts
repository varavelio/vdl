// Verifies enum serialization and round-trip: both string enums and int enums
// are echoed correctly through client/server communication.
import { Server, NewClient, Client } from "./gen/index.ts";
import { createNodeHandler } from "./gen/adapters/node.ts";
import type { Color, Priority, LogLevel } from "./gen/index.ts";
import { createServer } from "http";

async function main() {
  const server = new Server();

  server.rpcs
    .service()
    .procs.Echo()
    .handle(async ({ input }) => {
      return {
        color: input.color,
        priority: input.priority,
      };
    });

  server.rpcs
    .service()
    .procs.EchoOptional()
    .handle(async ({ input }) => {
      return {
        container: input.container,
      };
    });

  server.rpcs
    .service()
    .procs.EchoLogLevel()
    .handle(async ({ input }) => {
      return {
        level: input.level,
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

  try {
    // Test valid enum round-trip
    await testValidEnumRoundTrip(client);

    // Test optional enum fields
    await testOptionalEnums(client);

    // Test explicit-value enum round-trip
    await testExplicitValueRoundTrip(client);

    console.log("Success");
  } catch (e) {
    console.error("Error:", e);
    process.exit(1);
  }

  httpServer.close();
  process.exit(0);
}

async function testValidEnumRoundTrip(client: Client) {
  const testCases: { color: Color; priority: Priority }[] = [
    { color: "Red", priority: 1 },
    { color: "Green", priority: 2 },
    { color: "Blue", priority: 3 },
  ];

  for (const tc of testCases) {
    const res = await client.procs.serviceEcho().execute({
      color: tc.color,
      priority: tc.priority,
    });

    if (res.color !== tc.color) {
      throw new Error(`expected color ${tc.color}, got ${res.color}`);
    }
    if (res.priority !== tc.priority) {
      throw new Error(`expected priority ${tc.priority}, got ${res.priority}`);
    }
  }
}

async function testOptionalEnums(client: Client) {
  // Test with absent optional enums
  const res = await client.procs.serviceEchoOptional().execute({
    container: {},
  });

  if (res.container.color !== undefined) {
    throw new Error("color should be absent");
  }
  if (res.container.priority !== undefined) {
    throw new Error("priority should be absent");
  }

  // Test with present valid optional enums
  const res2 = await client.procs.serviceEchoOptional().execute({
    container: {
      color: "Blue",
      priority: 3,
    },
  });

  if (res2.container.color !== "Blue") {
    throw new Error(`expected "Blue", got ${res2.container.color}`);
  }
  if (res2.container.priority !== 3) {
    throw new Error(`expected 3, got ${res2.container.priority}`);
  }

  // Test optional explicit-value enum
  const res3 = await client.procs.serviceEchoOptional().execute({
    container: {
      logLevel: "WARN",
    },
  });

  if (res3.container.logLevel !== "WARN") {
    throw new Error(`expected "WARN", got ${res3.container.logLevel}`);
  }
}

async function testExplicitValueRoundTrip(client: Client) {
  const testCases: { level: LogLevel; expectedValue: string }[] = [
    { level: "DEBUG", expectedValue: "DEBUG" },
    { level: "INFO", expectedValue: "INFO" },
    { level: "WARN", expectedValue: "WARN" },
    { level: "ERROR", expectedValue: "ERROR" },
    { level: "CRITICAL", expectedValue: "CRITICAL" },
  ];

  for (const tc of testCases) {
    const res = await client.procs.serviceEchoLogLevel().execute({
      level: tc.level,
    });

    if (res.level !== tc.level) {
      throw new Error(
        `round-trip failed: expected ${tc.level}, got ${res.level}`,
      );
    }
    // Verify constant value matches expected wire format
    if (tc.level !== tc.expectedValue) {
      throw new Error(
        `value mismatch: expected ${tc.expectedValue}, got ${tc.level}`,
      );
    }
  }
}

main().catch((err) => {
  console.error(err);
  process.exit(1);
});

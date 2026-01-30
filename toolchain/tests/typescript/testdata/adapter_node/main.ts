/**
 * E2E Test: Node.js Adapter
 *
 * This test verifies that the generated Node.js adapter (gen/adapters/node.ts)
 * works correctly with both native Node.js http and with pre-parsed body.
 */

import { createServer } from "http";
import { Server } from "./gen/server.ts";
import { NewClient } from "./gen/client.ts";
import { createNodeHandler, NodeAdapter } from "./gen/adapters/node.ts";

async function main() {
  // =========================================================================
  // Test 1: Using createNodeHandler helper
  // =========================================================================

  const server = new Server();

  // Register handlers
  server.rpcs
    .greeter()
    .procs.Hello()
    .handle(async ({ input }) => {
      return { greeting: `Hello, ${input.name}!` };
    });

  server.rpcs
    .greeter()
    .procs.Echo()
    .handle(async ({ input }) => {
      return { message: input.message };
    });

  server.rpcs
    .calculator()
    .procs.Add()
    .handle(async ({ input }) => {
      return { result: input.a + input.b };
    });

  // Create the handler with prefix
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

  console.log("Server started on port", port);

  const client = NewClient(baseUrl).build();

  try {
    // Test 1: Simple greeting
    console.log("Test 1: Greeter.Hello");
    const greetResult = await client.procs
      .greeterHello()
      .execute({ name: "World" });
    if (greetResult.greeting !== "Hello, World!") {
      console.error(
        "Test 1 FAILED: Expected 'Hello, World!' but got:",
        greetResult.greeting,
      );
      process.exit(1);
    }
    console.log("Test 1 PASSED");

    // Test 2: Echo
    console.log("Test 2: Greeter.Echo");
    const echoResult = await client.procs
      .greeterEcho()
      .execute({ message: "Test message" });
    if (echoResult.message !== "Test message") {
      console.error(
        "Test 2 FAILED: Expected 'Test message' but got:",
        echoResult.message,
      );
      process.exit(1);
    }
    console.log("Test 2 PASSED");

    // Test 3: Calculator.Add
    console.log("Test 3: Calculator.Add");
    const addResult = await client.procs
      .calculatorAdd()
      .execute({ a: 5, b: 3 });
    if (addResult.result !== 8) {
      console.error("Test 3 FAILED: Expected 8 but got:", addResult.result);
      process.exit(1);
    }
    console.log("Test 3 PASSED");

    // Test 4: 404 for invalid path
    console.log("Test 4: Invalid path returns 404");
    const invalidResponse = await fetch(`${baseUrl}`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({}),
    });
    if (invalidResponse.status !== 404) {
      console.error(
        "Test 4 FAILED: Expected 404 but got:",
        invalidResponse.status,
      );
      process.exit(1);
    }
    console.log("Test 4 PASSED");
  } catch (e) {
    console.error("Test failed with error:", e);
    process.exit(1);
  }

  // =========================================================================
  // Test 5: Using NodeAdapter directly with pre-parsed body
  // =========================================================================

  console.log("Test 5: NodeAdapter with pre-parsed body");

  // Create a second server to test pre-parsed body scenario
  const httpServer2 = createServer(async (req, res) => {
    if (req.method !== "POST") {
      res.writeHead(405);
      res.end();
      return;
    }

    // Simulate middleware that pre-parses body (like Express body-parser)
    const buffers: Buffer[] = [];
    for await (const chunk of req) buffers.push(chunk);
    const bodyStr = Buffer.concat(buffers).toString();
    const parsedBody = bodyStr ? JSON.parse(bodyStr) : {};

    // Extract RPC and operation from URL
    const parts = req.url?.split("/") || [];
    const rpcName = parts[2];
    const operationName = parts[3];

    // Use NodeAdapter directly with pre-parsed body
    const adapter = new NodeAdapter(req, res, parsedBody);
    await server.handleRequest({}, rpcName, operationName, adapter);
  });

  await new Promise<void>((resolve) => {
    httpServer2.listen(0, resolve);
  });

  const addr2 = httpServer2.address() as any;
  const port2 = addr2.port;
  const baseUrl2 = `http://localhost:${port2}/rpc`;

  const client2 = NewClient(baseUrl2).build();

  try {
    const result = await client2.procs
      .greeterHello()
      .execute({ name: "PreParsed" });
    if (result.greeting !== "Hello, PreParsed!") {
      console.error(
        "Test 5 FAILED: Expected 'Hello, PreParsed!' but got:",
        result.greeting,
      );
      process.exit(1);
    }
    console.log("Test 5 PASSED");
  } catch (e) {
    console.error("Test 5 failed with error:", e);
    process.exit(1);
  }

  httpServer.close();
  httpServer2.close();

  console.log("\n=== All Node Adapter tests PASSED ===");
  process.exit(0);
}

main().catch((err) => {
  console.error("Fatal error:", err);
  process.exit(1);
});

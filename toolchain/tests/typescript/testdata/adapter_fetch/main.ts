/**
 * E2E Test: Fetch Adapter
 *
 * This test verifies that the generated Fetch adapter (gen/adapters/fetch.ts)
 * works correctly with Web Standards Request/Response objects.
 *
 * Since we're running in Node.js and not a real edge runtime,
 * we simulate the Fetch API flow by directly invoking the handler.
 */
import {
  createFetchHandler,
  FetchAdapter,
  parseRpcPath,
} from "./gen/adapters/fetch.ts";
import { Server } from "./gen/server.ts";

async function main() {
  // =========================================================================
  // Setup: Create server with handlers
  // =========================================================================

  const server = new Server<{ requestId: string }>();

  // Register handlers
  server.rpcs
    .greeter()
    .procs.Hello()
    .handle(async ({ input, props }) => {
      // Verify props is passed through
      console.log("Handler received requestId:", props.requestId);
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

  // =========================================================================
  // Test 1: parseRpcPath helper
  // =========================================================================

  console.log("Test 1: parseRpcPath helper");

  const parsed1 = parseRpcPath("http://localhost/rpc/Greeter/Hello", "/rpc");
  if (
    !parsed1 ||
    parsed1.rpcName !== "Greeter" ||
    parsed1.operationName !== "Hello"
  ) {
    console.error("Test 1a FAILED: Expected Greeter/Hello but got:", parsed1);
    process.exit(1);
  }

  const parsed2 = parseRpcPath("http://localhost/Greeter/Hello");
  if (
    !parsed2 ||
    parsed2.rpcName !== "Greeter" ||
    parsed2.operationName !== "Hello"
  ) {
    console.error("Test 1b FAILED: Expected Greeter/Hello but got:", parsed2);
    process.exit(1);
  }

  const parsed3 = parseRpcPath(
    "http://localhost/api/v1/rpc/Calculator/Add",
    "/api/v1/rpc",
  );
  if (
    !parsed3 ||
    parsed3.rpcName !== "Calculator" ||
    parsed3.operationName !== "Add"
  ) {
    console.error("Test 1c FAILED: Expected Calculator/Add but got:", parsed3);
    process.exit(1);
  }

  const parsed4 = parseRpcPath("http://localhost/invalid");
  if (parsed4 !== null) {
    console.error(
      "Test 1d FAILED: Expected null for invalid path but got:",
      parsed4,
    );
    process.exit(1);
  }

  console.log("Test 1 PASSED");

  // =========================================================================
  // Test 2: createFetchHandler with context
  // =========================================================================

  console.log("Test 2: createFetchHandler with context");

  const handler = createFetchHandler(
    server,
    async (req) => {
      // Create context from request
      return { requestId: req.headers.get("X-Request-ID") || "unknown" };
    },
    { prefix: "/rpc" },
  );

  const request = new Request("http://localhost/rpc/Greeter/Hello", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      "X-Request-ID": "test-123",
    },
    body: JSON.stringify({ name: "World" }),
  });

  const response = await handler(request);

  if (response.status !== 200) {
    console.error(
      "Test 2 FAILED: Expected status 200 but got:",
      response.status,
    );
    process.exit(1);
  }

  const result = (await response.json()) as any;
  if (!result.ok || result.output.greeting !== "Hello, World!") {
    console.error("Test 2 FAILED: Unexpected response:", result);
    process.exit(1);
  }

  console.log("Test 2 PASSED");

  // =========================================================================
  // Test 3: Multiple RPC calls
  // =========================================================================

  console.log("Test 3: Multiple RPC calls");

  const echoRequest = new Request("http://localhost/rpc/Greeter/Echo", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ message: "Test message" }),
  });

  const echoResponse = await handler(echoRequest);
  const echoResult = (await echoResponse.json()) as any;

  if (!echoResult.ok || echoResult.output.message !== "Test message") {
    console.error("Test 3a FAILED: Unexpected echo response:", echoResult);
    process.exit(1);
  }

  const addRequest = new Request("http://localhost/rpc/Calculator/Add", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ a: 10, b: 20 }),
  });

  const addResponse = await handler(addRequest);
  const addResult = (await addResponse.json()) as any;

  if (!addResult.ok || addResult.output.result !== 30) {
    console.error("Test 3b FAILED: Unexpected add response:", addResult);
    process.exit(1);
  }

  console.log("Test 3 PASSED");

  // =========================================================================
  // Test 4: Invalid path returns 404
  // =========================================================================

  console.log("Test 4: Invalid path returns 404");

  const invalidRequest = new Request("http://localhost/rpc/invalid", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({}),
  });

  const invalidResponse = await handler(invalidRequest);

  if (invalidResponse.status !== 404) {
    console.error(
      "Test 4 FAILED: Expected 404 but got:",
      invalidResponse.status,
    );
    process.exit(1);
  }

  const invalidResult = (await invalidResponse.json()) as any;
  if (invalidResult.ok !== false) {
    console.error(
      "Test 4 FAILED: Expected error response but got:",
      invalidResult,
    );
    process.exit(1);
  }

  console.log("Test 4 PASSED");

  // =========================================================================
  // Test 5: FetchAdapter directly
  // =========================================================================

  console.log("Test 5: FetchAdapter directly");

  const directRequest = new Request("http://localhost/Greeter/Echo", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ message: "Direct adapter test" }),
  });

  const adapter = new FetchAdapter(directRequest);

  // Process request
  await server.handleRequest(
    { requestId: "direct-test" },
    "Greeter",
    "Echo",
    adapter,
  );

  const directResponse = adapter.toResponse();
  const directResult = (await directResponse.json()) as any;

  if (
    !directResult.ok ||
    directResult.output.message !== "Direct adapter test"
  ) {
    console.error("Test 5 FAILED: Unexpected direct response:", directResult);
    process.exit(1);
  }

  console.log("Test 5 PASSED");

  // =========================================================================
  // Test 6: Invalid operation returns error
  // =========================================================================

  console.log("Test 6: Invalid operation returns error");

  const invalidOpRequest = new Request(
    "http://localhost/rpc/Greeter/NonExistent",
    {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({}),
    },
  );

  const invalidOpResponse = await handler(invalidOpRequest);
  const invalidOpResult = (await invalidOpResponse.json()) as any;

  if (invalidOpResult.ok !== false) {
    console.error(
      "Test 6 FAILED: Expected error for non-existent operation:",
      invalidOpResult,
    );
    process.exit(1);
  }

  console.log("Test 6 PASSED");

  console.log("\n=== All Fetch Adapter tests PASSED ===");
  process.exit(0);
}

main().catch((err) => {
  console.error("Fatal error:", err);
  process.exit(1);
});

// Verifies that the server properly validates enum values and rejects invalid ones.

import { createServer } from "http";
import { createNodeHandler } from "./gen/adapters/node.ts";
import { NewClient, Server, VdlError } from "./gen/index.ts";

async function main() {
  const server = new Server();

  // Register all handlers that simply echo the input
  server.rpcs
    .service()
    .procs.EchoColor()
    .handle(async ({ input }) => ({ color: input.color }));

  server.rpcs
    .service()
    .procs.EchoPriority()
    .handle(async ({ input }) => ({ priority: input.priority }));

  server.rpcs
    .service()
    .procs.EchoLogLevel()
    .handle(async ({ input }) => ({ level: input.level }));

  server.rpcs
    .service()
    .procs.EchoSettings()
    .handle(async ({ input }) => ({ settings: input.settings }));

  server.rpcs
    .service()
    .procs.EchoOptional()
    .handle(async ({ input }) => ({ settings: input.settings }));

  server.rpcs
    .service()
    .procs.EchoPalette()
    .handle(async ({ input }) => ({ palette: input.palette }));

  server.rpcs
    .service()
    .procs.EchoColorMap()
    .handle(async ({ input }) => ({ colorMap: input.colorMap }));

  server.rpcs
    .service()
    .procs.EchoNested()
    .handle(async ({ input }) => ({ nested: input.nested }));

  server.rpcs
    .service()
    .streams.ColorStream()
    .handle(async (c, emit) => {
      await emit(c, { color: c.input.color });
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
    // ========== Test 1: Valid enum values should work ==========
    console.log("Test 1: Valid enum values...");
    const colorRes = await client.procs.serviceEchoColor().execute({ color: "Red" });
    if (colorRes.color !== "Red") throw new Error(`Expected "Red", got ${colorRes.color}`);

    const priorityRes = await client.procs.serviceEchoPriority().execute({ priority: 2 });
    if (priorityRes.priority !== 2) throw new Error(`Expected 2, got ${priorityRes.priority}`);

    const logLevelRes = await client.procs.serviceEchoLogLevel().execute({ level: "DEBUG" });
    if (logLevelRes.level !== "DEBUG")
      throw new Error(`Expected "DEBUG", got ${logLevelRes.level}`);
    console.log("  PASS: Valid enum values work");

    // ========== Test 2: Invalid string enum value should be rejected ==========
    console.log("Test 2: Invalid string enum value...");
    await testRawInvalidRequest(baseUrl, "/Service/EchoColor", { color: "Purple" }, "Color");
    console.log("  PASS: Invalid string enum rejected");

    // ========== Test 3: Invalid integer enum value should be rejected ==========
    console.log("Test 3: Invalid integer enum value...");
    await testRawInvalidRequest(baseUrl, "/Service/EchoPriority", { priority: 99 }, "Priority");
    console.log("  PASS: Invalid integer enum rejected");

    // ========== Test 4: Invalid explicit-value enum should be rejected ==========
    console.log("Test 4: Invalid explicit-value enum...");
    await testRawInvalidRequest(baseUrl, "/Service/EchoLogLevel", { level: "TRACE" }, "LogLevel");
    console.log("  PASS: Invalid explicit-value enum rejected");

    // ========== Test 5: Invalid enum in nested type should be rejected ==========
    console.log("Test 5: Invalid enum in nested type...");
    await testRawInvalidRequest(
      baseUrl,
      "/Service/EchoSettings",
      {
        settings: {
          color: "Yellow", // Invalid!
          priority: 1,
          logLevel: "DEBUG",
        },
      },
      "Color",
    );
    console.log("  PASS: Invalid enum in nested type rejected");

    // ========== Test 6: Valid optional enums (missing) should work ==========
    console.log("Test 6: Valid optional enums (missing)...");
    const optRes = await client.procs.serviceEchoOptional().execute({ settings: {} });
    if (optRes.settings.color !== undefined) throw new Error("Expected undefined color");
    console.log("  PASS: Optional enums work when missing");

    // ========== Test 7: Invalid optional enum should be rejected ==========
    console.log("Test 7: Invalid optional enum value...");
    await testRawInvalidRequest(
      baseUrl,
      "/Service/EchoOptional",
      { settings: { color: "Orange" } }, // Invalid!
      "Color",
    );
    console.log("  PASS: Invalid optional enum rejected");

    // ========== Test 8: Valid enum arrays should work ==========
    console.log("Test 8: Valid enum arrays...");
    const paletteRes = await client.procs.serviceEchoPalette().execute({
      palette: {
        colors: ["Red", "Green", "Blue"],
        priorities: [1, 2, 3],
      },
    });
    if (paletteRes.palette.colors.length !== 3) throw new Error("Expected 3 colors");
    console.log("  PASS: Valid enum arrays work");

    // ========== Test 9: Invalid enum in array should be rejected ==========
    console.log("Test 9: Invalid enum in array...");
    await testRawInvalidRequest(
      baseUrl,
      "/Service/EchoPalette",
      {
        palette: {
          colors: ["Red", "Magenta", "Blue"], // Invalid!
          priorities: [1, 2, 3],
        },
      },
      "Color",
    );
    console.log("  PASS: Invalid enum in array rejected");

    // ========== Test 10: Valid enum maps should work ==========
    console.log("Test 10: Valid enum maps...");
    const mapRes = await client.procs.serviceEchoColorMap().execute({
      colorMap: {
        colorByName: { primary: "Red", secondary: "Blue" },
        priorityByTask: { task1: 1, task2: 2 },
      },
    });
    if (mapRes.colorMap.colorByName.primary !== "Red") throw new Error("Expected Red");
    console.log("  PASS: Valid enum maps work");

    // ========== Test 11: Invalid enum in map should be rejected ==========
    console.log("Test 11: Invalid enum in map...");
    await testRawInvalidRequest(
      baseUrl,
      "/Service/EchoColorMap",
      {
        colorMap: {
          colorByName: { primary: "Cyan" }, // Invalid!
          priorityByTask: { task1: 1 },
        },
      },
      "Color",
    );
    console.log("  PASS: Invalid enum in map rejected");

    // ========== Test 12: Deeply nested invalid enum should be rejected ==========
    console.log("Test 12: Deeply nested invalid enum...");
    await testRawInvalidRequest(
      baseUrl,
      "/Service/EchoNested",
      {
        nested: {
          settings: {
            color: "Red",
            priority: 999, // Invalid!
            logLevel: "DEBUG",
          },
        },
      },
      "Priority",
    );
    console.log("  PASS: Deeply nested invalid enum rejected");

    // ========== Test 13: Invalid enum in stream should be rejected ==========
    console.log("Test 13: Invalid enum in stream...");
    await testStreamInvalidRequest(baseUrl, "/Service/ColorStream", { color: "Pink" }, "Color");
    console.log("  PASS: Invalid enum in stream rejected");

    console.log("\nSuccess: All enum validation tests passed!");
  } catch (e) {
    console.error("Test failed:", e);
    process.exit(1);
  }

  httpServer.close();
  process.exit(0);
}

async function testStreamInvalidRequest(
  baseUrl: string,
  path: string,
  payload: unknown,
  expectedEnumName: string,
) {
  const resp = await fetch(`${baseUrl}${path}`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
  });

  // Streams return SSE format, so we need to read the text and parse the first event
  const text = await resp.text();

  // SSE format is "data: {...}\n\n"
  const firstLine = text.split("\n").find((l) => l.startsWith("data: "));
  if (!firstLine) {
    throw new Error(`Expected SSE data line, got: ${text}`);
  }

  const jsonStr = firstLine.slice(6); // Remove "data: " prefix
  const result = JSON.parse(jsonStr) as {
    ok: boolean;
    error?: { code?: string; message?: string };
  };

  if (result.ok) {
    throw new Error(
      `Expected validation error for invalid ${expectedEnumName}, but request succeeded`,
    );
  }

  if (result.error?.code !== "INVALID_INPUT") {
    throw new Error(
      `Expected INVALID_INPUT error code, got: ${result.error?.code} (message: ${result.error?.message})`,
    );
  }

  if (!result.error.message?.includes(expectedEnumName)) {
    throw new Error(
      `Expected error message to mention "${expectedEnumName}", got: ${result.error?.message}`,
    );
  }
}

async function testRawInvalidRequest(
  baseUrl: string,
  path: string,
  payload: unknown,
  expectedEnumName: string,
) {
  const resp = await fetch(`${baseUrl}${path}`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
  });

  const result = (await resp.json()) as {
    ok: boolean;
    error?: { code?: string; message?: string };
  };

  if (result.ok) {
    throw new Error(
      `Expected validation error for invalid ${expectedEnumName}, but request succeeded`,
    );
  }

  if (result.error?.code !== "INVALID_INPUT") {
    throw new Error(
      `Expected INVALID_INPUT error code, got: ${result.error?.code} (message: ${result.error?.message})`,
    );
  }

  if (!result.error.message?.includes(expectedEnumName)) {
    throw new Error(
      `Expected error message to mention "${expectedEnumName}", got: ${result.error?.message}`,
    );
  }
}

main().catch((err) => {
  console.error(err);
  process.exit(1);
});

// Verifies scalar edge cases: zero values (0, 0.0, false, "", zero datetime)
// should be correctly transmitted and distinguishable from absent values.

import { createServer } from "http";
import { createNodeHandler } from "./gen/adapters/node.ts";
import { type Client, NewClient, Server } from "./gen/index.ts";

async function main() {
  const server = new Server();

  server.rpcs
    .service()
    .procs.Echo()
    .handle(async ({ input }) => {
      return {
        intVal: input.intVal,
        floatVal: input.floatVal,
        boolVal: input.boolVal,
        stringVal: input.stringVal,
        datetimeVal: input.datetimeVal,
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
    // Test 1: Zero values via raw HTTP
    await testZeroValuesRaw(baseUrl);

    // Test 2: Zero values via generated client
    await testZeroValuesClient(client);

    // Test 3: Non-zero values
    await testNonZeroValues(client);

    console.log("Success");
  } catch (e) {
    console.error("Error:", e);
    process.exit(1);
  }

  httpServer.close();
  process.exit(0);
}

async function testZeroValuesRaw(baseUrl: string) {
  // Zero datetime - epoch time
  const zeroTime = new Date(0);

  const payload = {
    intVal: 0,
    floatVal: 0.0,
    boolVal: false,
    stringVal: "",
    datetimeVal: zeroTime.toISOString(),
  };

  const resp = await fetch(`${baseUrl}/Service/Echo`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
  });

  const result = (await resp.json()) as any;

  if (result.ok !== true) {
    throw new Error(`expected ok=true, got: ${JSON.stringify(result)}`);
  }

  const output = result.output;

  // Verify int zero
  if (output.intVal !== 0) {
    throw new Error(`expected intVal=0, got: ${output.intVal}`);
  }

  // Verify float zero
  if (output.floatVal !== 0) {
    throw new Error(`expected floatVal=0, got: ${output.floatVal}`);
  }

  // Verify bool false
  if (output.boolVal !== false) {
    throw new Error(`expected boolVal=false, got: ${output.boolVal}`);
  }

  // Verify empty string
  if (output.stringVal !== "") {
    throw new Error(`expected stringVal='', got: ${output.stringVal}`);
  }

  // Verify datetime is present
  if (output.datetimeVal === null || output.datetimeVal === undefined) {
    throw new Error("expected datetimeVal to be present");
  }
}

async function testZeroValuesClient(client: Client) {
  // Zero datetime - epoch time
  const zeroTime = new Date(0);

  const result = await client.procs.serviceEcho().execute({
    intVal: 0,
    floatVal: 0.0,
    boolVal: false,
    stringVal: "",
    datetimeVal: zeroTime,
  });

  if (result.intVal !== 0) {
    throw new Error(`expected intVal=0, got: ${result.intVal}`);
  }
  if (result.floatVal !== 0.0) {
    throw new Error(`expected floatVal=0.0, got: ${result.floatVal}`);
  }
  if (result.boolVal !== false) {
    throw new Error(`expected boolVal=false, got: ${result.boolVal}`);
  }
  if (result.stringVal !== "") {
    throw new Error(`expected stringVal='', got: ${result.stringVal}`);
  }

  const resultDate = new Date(result.datetimeVal);
  if (resultDate.getTime() !== zeroTime.getTime()) {
    throw new Error(
      `expected datetimeVal=${zeroTime.toISOString()}, got: ${resultDate.toISOString()}`,
    );
  }
}

async function testNonZeroValues(client: Client) {
  const now = new Date();
  now.setMilliseconds(0); // Truncate to seconds like Go does

  const result = await client.procs.serviceEcho().execute({
    intVal: 42,
    floatVal: 3.14159,
    boolVal: true,
    stringVal: "hello world",
    datetimeVal: now,
  });

  if (result.intVal !== 42) {
    throw new Error(`expected intVal=42, got: ${result.intVal}`);
  }
  if (result.floatVal !== 3.14159) {
    throw new Error(`expected floatVal=3.14159, got: ${result.floatVal}`);
  }
  if (result.boolVal !== true) {
    throw new Error(`expected boolVal=true, got: ${result.boolVal}`);
  }
  if (result.stringVal !== "hello world") {
    throw new Error(`expected stringVal='hello world', got: ${result.stringVal}`);
  }

  const resultDate = new Date(result.datetimeVal);
  if (resultDate.getTime() !== now.getTime()) {
    throw new Error(`expected datetimeVal=${now.toISOString()}, got: ${resultDate.toISOString()}`);
  }
}

main().catch((err) => {
  console.error(err);
  process.exit(1);
});

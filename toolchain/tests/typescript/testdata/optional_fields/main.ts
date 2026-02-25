// Verifies optional field behavior: fields can be omitted or provided with values.

import { createServer } from "http";
import { createNodeHandler } from "./gen/adapters/node.ts";
import { NewClient, Server } from "./gen/index.ts";

async function main() {
  const server = new Server();

  server.rpcs
    .service()
    .procs.Echo()
    .handle(async (ctx) => {
      // Echo back exactly what was received
      return {
        required: ctx.input.required,
        optional: ctx.input.optional,
        optionalInt: ctx.input.optionalInt,
        optionalBool: ctx.input.optionalBool,
        optionalArray: ctx.input.optionalArray,
        optionalObject: ctx.input.optionalObject,
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

  // Test 1: All optional fields absent (only required)
  const result1 = await client.procs.serviceEcho().execute({
    required: "hello",
  });
  if (result1.required !== "hello") {
    throw new Error(`Expected required='hello', got '${result1.required}'`);
  }
  if (result1.optional !== undefined) {
    throw new Error(`Expected optional=undefined, got '${result1.optional}'`);
  }
  if (result1.optionalInt !== undefined) {
    throw new Error(`Expected optionalInt=undefined, got '${result1.optionalInt}'`);
  }

  // Test 2: All optional fields present with values
  const result2 = await client.procs.serviceEcho().execute({
    required: "test",
    optional: "value",
    optionalInt: 42,
    optionalBool: true,
    optionalArray: ["a", "b", "c"],
    optionalObject: { street: "123 Main", city: "NYC" },
  });
  if (result2.optional !== "value") {
    throw new Error(`Expected optional='value', got '${result2.optional}'`);
  }
  if (result2.optionalInt !== 42) {
    throw new Error(`Expected optionalInt=42, got '${result2.optionalInt}'`);
  }
  if (result2.optionalBool !== true) {
    throw new Error(`Expected optionalBool=true, got '${result2.optionalBool}'`);
  }
  if (!Array.isArray(result2.optionalArray) || result2.optionalArray.length !== 3) {
    throw new Error(
      `Expected optionalArray=['a','b','c'], got '${JSON.stringify(result2.optionalArray)}'`,
    );
  }
  if (result2.optionalObject?.city !== "NYC") {
    throw new Error(`Expected optionalObject.city='NYC', got '${result2.optionalObject?.city}'`);
  }

  // Test 3: Optional fields with "zero" values (empty string, 0, false)
  const result3 = await client.procs.serviceEcho().execute({
    required: "test",
    optional: "",
    optionalInt: 0,
    optionalBool: false,
  });
  if (result3.optional !== "") {
    throw new Error(`Expected optional='', got '${result3.optional}'`);
  }
  if (result3.optionalInt !== 0) {
    throw new Error(`Expected optionalInt=0, got '${result3.optionalInt}'`);
  }
  if (result3.optionalBool !== false) {
    throw new Error(`Expected optionalBool=false, got '${result3.optionalBool}'`);
  }

  // Test 4: Mixed - some optional present, some absent
  const result4 = await client.procs.serviceEcho().execute({
    required: "mixed",
    optional: "present",
    // optionalInt is absent
    optionalBool: true,
  });
  if (result4.optional !== "present") {
    throw new Error(`Expected optional='present', got '${result4.optional}'`);
  }
  if (result4.optionalInt !== undefined) {
    throw new Error(`Expected optionalInt=undefined, got '${result4.optionalInt}'`);
  }
  if (result4.optionalBool !== true) {
    throw new Error(`Expected optionalBool=true, got '${result4.optionalBool}'`);
  }

  console.log("Success");
  httpServer.close();
  process.exit(0);
}

main().catch((err) => {
  console.error(err);
  process.exit(1);
});

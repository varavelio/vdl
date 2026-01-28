// Verifies enum serialization: enums should be transmitted as strings on the wire,
// and round-trip correctly through client->server->client.
// Tests both implicit-value enums (name=value) and explicit-value enums (name!=value).
import {
  Server,
  NewClient,
  Client,
  ColorValues,
  StatusValues,
  HttpStatusValues,
} from "./gen/index.ts";
import { createNodeHandler } from "./gen/adapters/node.ts";
import type { Color, Status, HttpStatus } from "./gen/index.ts";
import { createServer } from "http";

async function main() {
  const server = new Server();

  server.rpcs
    .service()
    .procs.Echo()
    .handle(async ({ input }) => {
      return {
        color: input.color,
        status: input.status,
      };
    });

  server.rpcs
    .service()
    .procs.GetDefaults()
    .handle(async () => {
      return {
        color: ColorValues.Red,
        status: StatusValues.Pending,
      };
    });

  server.rpcs
    .service()
    .procs.EchoHttpStatus()
    .handle(async ({ input }) => {
      return {
        status: input.status,
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
    // Implicit-value enum tests
    await testWireFormat(baseUrl);
    await testGeneratedClient(client);
    await testAllEnumValues(client);

    // Explicit-value enum tests
    await testExplicitValueWireFormat(baseUrl);
    await testExplicitValueClient(client);
    await testExplicitValueAllMembers(client);

    console.log("Success");
  } catch (e) {
    console.error("Error:", e);
    process.exit(1);
  }

  httpServer.close();
  process.exit(0);
}

async function testWireFormat(baseUrl: string) {
  // Send raw JSON to verify wire format
  const payload = JSON.stringify({ color: "Blue", status: "Active" });
  const resp = await fetch(`${baseUrl}/Service/Echo`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: payload,
  });

  const result = (await resp.json()) as {
    ok: boolean;
    output?: { color: string; status: string };
  };

  if (result.ok !== true) {
    throw new Error(`expected ok=true, got: ${JSON.stringify(result)}`);
  }

  if (result.output?.color !== "Blue") {
    throw new Error(`expected color='Blue', got: ${result.output?.color}`);
  }
  if (result.output?.status !== "Active") {
    throw new Error(`expected status='Active', got: ${result.output?.status}`);
  }
}

async function testGeneratedClient(client: Client) {
  const result = await client.procs.serviceEcho().execute({
    color: ColorValues.Green,
    status: StatusValues.Completed,
  });

  if (result.color !== ColorValues.Green) {
    throw new Error(`expected ColorValues.Green, got: ${result.color}`);
  }
  if (result.status !== StatusValues.Completed) {
    throw new Error(`expected StatusValues.Completed, got: ${result.status}`);
  }
}

async function testAllEnumValues(client: Client) {
  const colors: Color[] = [
    ColorValues.Red,
    ColorValues.Green,
    ColorValues.Blue,
  ];
  const statuses: Status[] = [
    StatusValues.Pending,
    StatusValues.Active,
    StatusValues.Completed,
    StatusValues.Cancelled,
  ];

  for (const color of colors) {
    for (const status of statuses) {
      const result = await client.procs.serviceEcho().execute({
        color: color,
        status: status,
      });

      if (result.color !== color) {
        throw new Error(
          `color mismatch: expected ${color}, got ${result.color}`,
        );
      }
      if (result.status !== status) {
        throw new Error(
          `status mismatch: expected ${status}, got ${result.status}`,
        );
      }
    }
  }
}

// testExplicitValueWireFormat verifies that explicit-value enums use the VALUE (not the name) on the wire.
async function testExplicitValueWireFormat(baseUrl: string) {
  // Send explicit value string - should work
  const payload = JSON.stringify({ status: "BAD_REQUEST" });
  const resp = await fetch(`${baseUrl}/Service/EchoHttpStatus`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: payload,
  });

  const result = (await resp.json()) as {
    ok: boolean;
    output?: { status: string };
  };

  if (result.ok !== true) {
    throw new Error(
      `expected ok=true for BAD_REQUEST, got: ${JSON.stringify(result)}`,
    );
  }

  // Wire format must use the VALUE, not the name
  if (result.output?.status !== "BAD_REQUEST") {
    throw new Error(
      `expected status='BAD_REQUEST' on wire, got: ${result.output?.status}`,
    );
  }
}

// testExplicitValueClient verifies the generated client uses correct values.
async function testExplicitValueClient(client: Client) {
  // Test each explicit-value enum member
  const testCases: { input: HttpStatus; expected: string }[] = [
    { input: HttpStatusValues.Ok, expected: "OK" },
    { input: HttpStatusValues.Created, expected: "CREATED" },
    { input: HttpStatusValues.BadRequest, expected: "BAD_REQUEST" },
    { input: HttpStatusValues.NotFound, expected: "NOT_FOUND" },
    {
      input: HttpStatusValues.InternalError,
      expected: "INTERNAL_SERVER_ERROR",
    },
  ];

  for (const tc of testCases) {
    const result = await client.procs.serviceEchoHttpStatus().execute({
      status: tc.input,
    });

    if (result.status !== tc.input) {
      throw new Error(
        `round-trip failed: expected ${tc.input}, got ${result.status}`,
      );
    }
    // Verify the constant value matches expected wire format
    if (tc.input !== tc.expected) {
      throw new Error(
        `constant value mismatch: expected ${tc.expected}, got ${tc.input}`,
      );
    }
  }
}

// testExplicitValueAllMembers verifies all explicit-value enum members round-trip correctly.
async function testExplicitValueAllMembers(client: Client) {
  const allStatuses: HttpStatus[] = [
    HttpStatusValues.Ok,
    HttpStatusValues.Created,
    HttpStatusValues.BadRequest,
    HttpStatusValues.NotFound,
    HttpStatusValues.InternalError,
  ];

  for (const status of allStatuses) {
    const result = await client.procs.serviceEchoHttpStatus().execute({
      status: status,
    });

    if (result.status !== status) {
      throw new Error(
        `status mismatch: expected ${status}, got ${result.status}`,
      );
    }
  }
}

main().catch((err) => {
  console.error(err);
  process.exit(1);
});

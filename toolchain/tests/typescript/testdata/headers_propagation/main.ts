// Verifies custom header propagation: X-Trace-ID and other headers
// should be correctly passed from client to server and accessible in handlers.
import { createNodeHandler } from "./gen/adapters/node.ts";
import { Server, NewClient } from "./gen/index.ts";
import { createServer, IncomingMessage, ServerResponse } from "http";

interface AppProps {
  traceId: string;
}

async function main() {
  const server = new Server<AppProps>();

  server.rpcs
    .service()
    .procs.Echo()
    .handle(async (c) => {
      return {
        data: c.input.data,
        receivedTraceId: c.props.traceId,
      };
    });

  const handler = createNodeHandler<AppProps>(
    server,
    (req) => {
      // Extract header from request and put into props
      // Node.js headers are lower-cased
      const traceId = (req.headers["x-trace-id"] as string) || "";
      return { traceId };
    },
    { prefix: "/rpc" },
  );

  const httpServer = createServer(
    async (req: IncomingMessage, res: ServerResponse) => {
      if (req.method !== "POST") {
        res.writeHead(405);
        res.end();
        return;
      }
      await handler(req, res);
    },
  );

  await new Promise<void>((resolve) => {
    httpServer.listen(0, resolve);
  });

  const addr = httpServer.address() as any;
  const port = addr.port;
  const baseUrl = `http://localhost:${port}/rpc`;

  try {
    // Test 1: With trace ID header at client level
    await testClientLevelHeader(baseUrl);

    // Test 2: Without trace ID header
    await testWithoutTraceID(baseUrl);

    // Test 3: With request-level header
    await testRequestLevelHeader(baseUrl);

    console.log("Success");
  } catch (e) {
    console.error("Error:", e);
    process.exit(1);
  }

  httpServer.close();
  process.exit(0);
}

async function testClientLevelHeader(baseUrl: string) {
  const client = NewClient(baseUrl)
    .withGlobalHeader("X-Trace-ID", "trace-123-abc")
    .build();

  const result = await client.procs.serviceEcho().execute({ data: "hello" });

  if (result.receivedTraceId !== "trace-123-abc") {
    throw new Error(
      `expected receivedTraceId='trace-123-abc', got: ${result.receivedTraceId}`,
    );
  }
}

async function testWithoutTraceID(baseUrl: string) {
  const client = NewClient(baseUrl).build();

  const result = await client.procs.serviceEcho().execute({ data: "hello" });

  if (result.receivedTraceId !== "") {
    throw new Error(
      `expected receivedTraceId='', got: ${result.receivedTraceId}`,
    );
  }
}

async function testRequestLevelHeader(baseUrl: string) {
  const client = NewClient(baseUrl).build();

  const result = await client.procs
    .serviceEcho()
    .withHeader("X-Trace-ID", "trace-456-def")
    .execute({ data: "test" });

  if (result.receivedTraceId !== "trace-456-def") {
    throw new Error(
      `expected receivedTraceId='trace-456-def', got: ${result.receivedTraceId}`,
    );
  }
}

main().catch((err) => {
  console.error(err);
  process.exit(1);
});

// Verifies SSE ping handling: the server sends pings during a long-running stream.
// Uses a raw HTTP client to verify pings are actually on the wire (the generated client filters them out).
import { createNodeHandler } from "./gen/adapters/node.ts";
import { Server, NewClient } from "./gen/index.ts";
import { createServer } from "http";

function sleep(ms: number) {
  return new Promise((r) => setTimeout(r, ms));
}

async function main() {
  // Test 1: Raw HTTP client to verify pings are actually sent
  await testRawPings();

  // Test 2: Generated client correctly filters pings and receives events
  await testGeneratedClient();

  console.log("Success");
  process.exit(0);
}

async function testRawPings() {
  const server = new Server();

  server.setStreamConfig({ pingIntervalMs: 50 });

  server.rpcs
    .clock()
    .streams.Ticks()
    .handle(async (c, emit) => {
      await sleep(200);
      await emit(c, { iso: "event1" });
      await sleep(200);
      await emit(c, { iso: "event2" });
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
  const baseUrl = `http://localhost:${addr.port}/rpc/Clock/Ticks`;

  // Make raw HTTP request to see pings on the wire
  const response = await fetch(baseUrl, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      Accept: "text/event-stream",
    },
    body: "{}",
  });

  if (!response.ok) {
    console.error(`unexpected status: ${response.status}`);
    process.exit(1);
  }

  const body = await response.text();
  const lines = body.split("\n");

  let pingCount = 0;
  let eventCount = 0;

  for (const line of lines) {
    if (line === ": ping") {
      pingCount++;
    }
    if (line.startsWith("data:")) {
      eventCount++;
    }
  }

  // With 50ms ping interval and ~400ms total stream time, expect at least 3 pings
  if (pingCount < 3) {
    console.error(`expected at least 3 pings, got ${pingCount}`);
    process.exit(1);
  }

  if (eventCount !== 2) {
    console.error(`expected 2 events, got ${eventCount}`);
    process.exit(1);
  }

  httpServer.close();
}

async function testGeneratedClient() {
  const server = new Server();

  server.setStreamConfig({ pingIntervalMs: 50 });

  server.rpcs
    .clock()
    .streams.Ticks()
    .handle(async (c, emit) => {
      await sleep(200);
      await emit(c, { iso: "event1" });
      await sleep(200);
      await emit(c, { iso: "event2" });
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
  const baseUrl = `http://localhost:${addr.port}/rpc`;
  const client = NewClient(baseUrl).build();

  const { stream } = client.streams
    .clockTicks()
    .withReconnect({ maxAttempts: 0 })
    .execute({});

  const received: string[] = [];
  for await (const evt of stream) {
    if (!evt.ok) {
      console.error(`stream error: ${JSON.stringify(evt.error)}`);
      process.exit(1);
    }
    received.push(evt.output.iso);
  }

  if (
    received.length !== 2 ||
    received[0] !== "event1" ||
    received[1] !== "event2"
  ) {
    console.error(`expected [event1, event2], got ${JSON.stringify(received)}`);
    process.exit(1);
  }

  httpServer.close();
}

main().catch((err) => {
  console.error(err);
  process.exit(1);
});

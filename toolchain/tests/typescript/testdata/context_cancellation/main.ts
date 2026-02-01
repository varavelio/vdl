// Verifies context cancellation behavior: when the client cancels the AbortSignal
// mid-stream, the server should stop sending events and clean up resources.
import { createNodeHandler } from "./gen/adapters/node.ts";
import { Server, NewClient } from "./gen/index.ts";
import { createServer } from "http";

function sleep(ms: number) {
  return new Promise((r) => setTimeout(r, ms));
}

async function main() {
  let emitCount = 0;
  let cleanedUp = false;

  const server = new Server();

  server.rpcs
    .service()
    .streams.Counter()
    .handle(async (c, emit) => {
      try {
        for (let i = 0; ; i++) {
          // Check if signal is aborted (similar to Go's ctx.Done())
          if (c.signal.aborted) {
            return;
          }
          await emit(c, { count: i });
          emitCount++;
          await sleep(50);
        }
      } finally {
        cleanedUp = true;
      }
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

  // Create an AbortController to cancel the stream
  const abortController = new AbortController();

  const { stream } = client.streams
    .serviceCounter()
    .withReconnect({ maxAttempts: 0 }) // Disable reconnection
    .withSignal(abortController.signal)
    .execute({});

  // Receive a few events
  let receivedCount = 0;
  for await (const evt of stream) {
    if (evt.ok) {
      receivedCount++;
      if (receivedCount >= 3) {
        // Cancel after receiving 3 events
        abortController.abort();
        break;
      }
    } else {
      // If error occurs due to cancellation, that's expected
      break;
    }
  }

  // Wait for cleanup
  await sleep(200);

  if (receivedCount < 3) {
    console.error(
      `expected at least 3 events before cancel, got ${receivedCount}`,
    );
    process.exit(1);
  }

  if (!cleanedUp) {
    console.error("expected server handler to clean up after cancellation");
    process.exit(1);
  }

  // Verify server stopped emitting (count should stabilize)
  const countBefore = emitCount;
  await sleep(200);
  const countAfter = emitCount;

  if (countAfter > countBefore + 2) {
    console.error(
      `server continued emitting after cancellation: before=${countBefore}, after=${countAfter}`,
    );
    process.exit(1);
  }

  console.log("Success");
  httpServer.close();
  process.exit(0);
}

main().catch((err) => {
  console.error(err);
  process.exit(1);
});

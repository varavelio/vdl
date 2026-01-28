// Verifies stream lifecycle hooks (onConnect, onDisconnect, onReconnect) work correctly.
import { Server, NewClient } from "./gen/index.ts";
import { createNodeHandler } from "./gen/adapters/node.ts";
import { createServer } from "http";

function sleep(ms: number) {
  return new Promise((r) => setTimeout(r, ms));
}

async function main() {
  const server = new Server();

  server.rpcs
    .events()
    .streams.Watch()
    .handle(async (c, emit) => {
      for (let i = 0; i < c.input.count; i++) {
        await emit(c, { value: i });
        await sleep(10);
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

  // Track lifecycle events
  let connectCalled = false;
  let disconnectCalled = false;
  let disconnectError: Error | null = null;

  try {
    // Test 1: onConnect and onDisconnect hooks
    const { stream } = client.streams
      .eventsWatch()
      .onConnect(() => {
        console.log("onConnect called");
        connectCalled = true;
      })
      .onDisconnect((error) => {
        console.log("onDisconnect called, error:", error);
        disconnectCalled = true;
        disconnectError = error;
      })
      .execute({ count: 3 });

    let received = 0;
    for await (const event of stream) {
      if (event.ok) {
        console.log("Received:", event.output.value);
        received++;
      } else {
        throw new Error("Stream error: " + JSON.stringify(event.error));
      }
    }

    // Verify lifecycle hooks were called
    if (!connectCalled) {
      throw new Error("onConnect was not called");
    }

    if (!disconnectCalled) {
      throw new Error("onDisconnect was not called");
    }

    // Normal completion should have null error
    if (disconnectError !== null) {
      throw new Error(
        "onDisconnect should have null error on normal completion",
      );
    }

    if (received !== 3) {
      throw new Error(`Expected 3 events, got ${received}`);
    }

    console.log("Test 1 passed: onConnect and onDisconnect work correctly");

    // Test 2: onDisconnect with cancel()
    let cancelDisconnectCalled = false;
    const { stream: stream2, cancel } = client.streams
      .eventsWatch()
      .onDisconnect(() => {
        cancelDisconnectCalled = true;
      })
      .execute({ count: 100 }); // Many events so we can cancel mid-stream

    let received2 = 0;
    for await (const event of stream2) {
      if (event.ok) {
        received2++;
        if (received2 >= 2) {
          // Cancel after receiving 2 events
          cancel();
          break;
        }
      }
    }

    // Give time for onDisconnect to be called
    await sleep(50);

    if (!cancelDisconnectCalled) {
      throw new Error("onDisconnect was not called after cancel()");
    }

    console.log("Test 2 passed: onDisconnect called after cancel()");

    // Test 3: onReconnect hook with a flaky server
    // Create a separate server that fails on first attempt, succeeds on second
    let flakyRequestCount = 0;
    const flakyServer = createServer(async (req, res) => {
      if (req.method !== "POST") {
        res.writeHead(405);
        res.end();
        return;
      }

      flakyRequestCount++;
      if (flakyRequestCount === 1) {
        // First request: return 503 to trigger reconnect
        res.writeHead(503);
        res.end("Service Unavailable");
        return;
      }

      // Subsequent requests: handle normally
      await handler(req, res);
    });

    await new Promise<void>((resolve) => {
      flakyServer.listen(0, resolve);
    });

    const flakyAddr = flakyServer.address() as any;
    const flakyBaseUrl = `http://localhost:${flakyAddr.port}/rpc`;
    const flakyClient = NewClient(flakyBaseUrl).build();

    let reconnectCalled = false;
    let reconnectAttempt = 0;
    let reconnectDelayMs = 0;

    const { stream: stream3 } = flakyClient.streams
      .eventsWatch()
      .withReconnect({
        maxAttempts: 5,
        initialDelayMs: 50,
        maxDelayMs: 200,
        delayMultiplier: 1.5,
      })
      .onReconnect((attempt, delayMs) => {
        console.log(
          `onReconnect called: attempt=${attempt}, delayMs=${delayMs}`,
        );
        reconnectCalled = true;
        reconnectAttempt = attempt;
        reconnectDelayMs = delayMs;
      })
      .execute({ count: 2 });

    let received3 = 0;
    for await (const event of stream3) {
      if (event.ok) {
        received3++;
      }
    }

    flakyServer.close();

    if (!reconnectCalled) {
      throw new Error("onReconnect was not called when server returned 503");
    }

    if (reconnectAttempt !== 1) {
      throw new Error(`Expected reconnect attempt 1, got ${reconnectAttempt}`);
    }

    if (reconnectDelayMs <= 0) {
      throw new Error(`Expected positive delay, got ${reconnectDelayMs}`);
    }

    if (received3 !== 2) {
      throw new Error(`Expected 2 events after reconnect, got ${received3}`);
    }

    console.log("Test 3 passed: onReconnect called on server failure");

    console.log("All tests passed!");
  } catch (e) {
    console.error("Error:", e);
    process.exit(1);
  }

  httpServer.close();
  process.exit(0);
}

main().catch((err) => {
  console.error(err);
  process.exit(1);
});

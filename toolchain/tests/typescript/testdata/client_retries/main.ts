// Verifies client retry logic: the server fails with 500 on first 2 attempts,
// then succeeds on the 3rd. The client should automatically retry and eventually succeed.
import { NewClient } from "./gen/index.ts";
import { createServer, IncomingMessage, ServerResponse } from "http";

async function main() {
  let attempts = 0;

  // Create a custom server that fails the first 2 requests with 500
  const httpServer = createServer(
    async (req: IncomingMessage, res: ServerResponse) => {
      if (req.method !== "POST") {
        res.writeHead(405);
        res.end();
        return;
      }

      attempts++;
      if (attempts < 3) {
        // Fail with 500 (retryable error)
        res.writeHead(500);
        res.end("Internal Server Error");
        return;
      }

      // Success on 3rd attempt
      const response = { ok: true, output: {} };
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify(response));
    },
  );

  await new Promise<void>((resolve) => {
    httpServer.listen(0, resolve);
  });

  const addr = httpServer.address() as any;
  const port = addr.port;
  const baseUrl = `http://localhost:${port}/rpc`;

  const client = NewClient(baseUrl).build();

  try {
    // Configure retries: 3 attempts with minimal delays
    await client.procs
      .serviceFlaky()
      .withRetries({
        maxAttempts: 3,
        initialDelayMs: 10,
        maxDelayMs: 100,
        delayMultiplier: 1.0,
        jitter: 0,
      })
      .execute({});

    // Verify we made 3 attempts
    if (attempts !== 3) {
      console.error(`expected 3 attempts, got ${attempts}`);
      process.exit(1);
    }

    console.log("Success");
  } catch (e) {
    console.error("expected success after retries, got:", e);
    process.exit(1);
  }

  httpServer.close();
  process.exit(0);
}

main().catch((err) => {
  console.error(err);
  process.exit(1);
});

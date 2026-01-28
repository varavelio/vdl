// Verifies stream error events: when a handler returns an error, the client
// receives it as an error event with the expected message.
import { Server, NewClient } from "./gen/index.ts";
import { createNodeHandler } from "./gen/adapters/node.ts";
import { createServer } from "http";

function sleep(ms: number) {
  return new Promise((r) => setTimeout(r, ms));
}

async function main() {
  const server = new Server();

  server.rpcs
    .service()
    .streams.Data()
    .handle(async (c, emit) => {
      await sleep(50);
      throw new Error("something went wrong");
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

  try {
    const { stream } = client.streams
      .serviceData()
      .withReconnect({ maxAttempts: 0 })
      .execute({});

    let gotError = false;
    const timeout = setTimeout(() => {
      console.error("timeout waiting for error event");
      process.exit(1);
    }, 2000);

    for await (const event of stream) {
      clearTimeout(timeout);
      if (event.ok) {
        console.error("expected error event, got success");
        process.exit(1);
      }
      if (event.error.message !== "something went wrong") {
        console.error(
          `expected 'something went wrong', got '${event.error.message}'`,
        );
        process.exit(1);
      }
      gotError = true;
      break;
    }

    if (!gotError) {
      console.error("did not receive error event");
      process.exit(1);
    }

    console.log("Success");
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

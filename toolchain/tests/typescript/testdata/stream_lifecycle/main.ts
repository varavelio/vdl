import { createNodeHandler } from "./gen/adapters/node.ts";
import { Server, NewClient } from "./gen/index.ts";
import { createServer } from "http";

function sleep(ms: number) {
  return new Promise((r) => setTimeout(r, ms));
}

async function main() {
  const server = new Server();

  server.rpcs
    .streamer()
    .streams.Counter()
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

  try {
    const { stream } = client.streams.streamerCounter().execute({ count: 5 });
    let received = 0;
    for await (const event of stream) {
      if (event.ok) {
        console.log("Received:", event.output.value);
        if (event.output.value !== received) {
          throw new Error(`Expected ${received}, got ${event.output.value}`);
        }
        received++;
      } else {
        throw new Error("Stream error: " + JSON.stringify(event.error));
      }
    }
    if (received !== 5) throw new Error(`Expected 5 events, got ${received}`);
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

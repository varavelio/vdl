import { Server, NewClient, HTTPAdapter } from "./gen/index.ts";
import { createServer } from "http";

class NodeAdapter implements HTTPAdapter {
  constructor(
    private req: any,
    private res: any,
    private body: any,
  ) {}
  async json() {
    return this.body;
  }
  setHeader(k: string, v: string) {
    this.res.setHeader(k, v);
  }
  write(d: string) {
    console.log("Server write:", JSON.stringify(d));
    this.res.write(d);
  }
  flush() {
    if (this.res.flush) this.res.flush();
  }
  end() {
    this.res.end();
  }
  onClose(cb: () => void) {
    this.req.on("close", cb);
  }
}

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

  const httpServer = createServer(async (req, res) => {
    if (req.method !== "POST") {
      res.writeHead(405);
      res.end();
      return;
    }

    const buffers = [];
    for await (const chunk of req) buffers.push(chunk);
    const bodyStr = Buffer.concat(buffers).toString();
    const body = bodyStr ? JSON.parse(bodyStr) : {};

    const parts = req.url?.split("/") || [];
    const service = parts[2];
    const method = parts[3];

    const adapter = new NodeAdapter(req, res, body);
    await server.handleRequest({}, service, method, adapter);
  });

  httpServer.listen(0, async () => {
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
  });
}

main().catch((err) => {
  console.error(err);
  process.exit(1);
});

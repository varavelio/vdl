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
    this.res.write(d);
  }
  end() {
    this.res.end();
  }
  onClose(cb: () => void) {
    this.req.on("close", cb);
  }
}

async function main() {
  const server = new Server();

  server.rpcs
    .greeter()
    .procs.Hello()
    .handle(async ({ input }) => {
      return { result: `Hello ${input.name}!` };
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

    // Extract URL components /rpc/Greeter/Hello
    // req.url is e.g. /rpc/Greeter/Hello
    const parts = req.url?.split("/") || [];
    // ["", "rpc", "Greeter", "Hello"]
    const service = parts[2];
    const method = parts[3];

    const adapter = new NodeAdapter(req, res, body);
    await server.handleRequest({}, service, method, adapter);
  });

  httpServer.listen(0, async () => {
    const addr = httpServer.address() as any;
    const port = addr.port;
    const baseUrl = `http://localhost:${port}/rpc`;

    const client = NewClient(baseUrl).build();
    // Client uses fetch. In Node 20+ fetch is global.
    // Assuming test runner has global fetch (tsx does).

    try {
      const response = await client.procs
        .greeterHello()
        .execute({ name: "World" });
      console.log("Response:", response);

      if (response.result !== "Hello World!") {
        console.error("Unexpected response:", response);
        process.exit(1);
      }
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

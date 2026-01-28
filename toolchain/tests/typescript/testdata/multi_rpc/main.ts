// Verifies multiple RPC blocks with different procs work correctly.
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
    .a()
    .procs.X()
    .handle(async () => {
      return {};
    });

  server.rpcs
    .b()
    .procs.Y()
    .handle(async () => {
      return {};
    });

  const httpServer = createServer(async (req, res) => {
    if (req.method !== "POST") {
      res.writeHead(405);
      res.end();
      return;
    }

    const buffers: Buffer[] = [];
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
    const port = addr.port;
    const baseUrl = `http://localhost:${port}/rpc`;

    const client = NewClient(baseUrl).build();

    // Test A.X
    await client.procs.aX().execute({});

    // Test B.Y
    await client.procs.bY().execute({});

    console.log("Success");
    httpServer.close();
    process.exit(0);
  });
}

main().catch((err) => {
  console.error(err);
  process.exit(1);
});

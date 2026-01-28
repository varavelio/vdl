// Verifies that nested structures and arrays serialize/deserialize correctly.
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
    .service()
    .procs.ValidatePerson()
    .handle(async () => {
      return {};
    });

  server.rpcs
    .service()
    .procs.ValidateArray()
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

    // Test 1: Valid person - should succeed
    await client.procs.serviceValidatePerson().execute({
      person: {
        name: "John",
        address: { street: "123 Main", city: "NYC" },
      },
    });

    // Test 2: Valid array - should succeed
    await client.procs.serviceValidateArray().execute({
      people: [
        { name: "Alice", address: { street: "1st St", city: "LA" } },
        { name: "Bob", address: { street: "2nd St", city: "SF" } },
      ],
    });

    // Test 3: Empty array - should succeed
    await client.procs.serviceValidateArray().execute({
      people: [],
    });

    console.log("Success");
    httpServer.close();
    process.exit(0);
  });
}

main().catch((err) => {
  console.error(err);
  process.exit(1);
});

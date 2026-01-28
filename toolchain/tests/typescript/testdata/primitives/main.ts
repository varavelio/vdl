// Verifies primitive type serialization: int, float, bool, string, and datetime.
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
    .procs.Echo()
    .handle(async ({ input }) => {
      return {
        i: input.i,
        f: input.f,
        b: input.b,
        s: input.s,
        d: input.d,
      };
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

    try {
      const now = new Date();
      now.setMilliseconds(0); // Truncate to seconds like Go does

      const input = {
        i: 42,
        f: 3.14159,
        b: true,
        s: "Hello VDL",
        d: now,
      };

      const response = await client.procs.serviceEcho().execute(input);

      if (response.i !== input.i) {
        console.error("int mismatch:", response.i, "!==", input.i);
        process.exit(1);
      }
      if (response.f !== input.f) {
        console.error("float mismatch:", response.f, "!==", input.f);
        process.exit(1);
      }
      if (response.b !== input.b) {
        console.error("bool mismatch:", response.b, "!==", input.b);
        process.exit(1);
      }
      if (response.s !== input.s) {
        console.error("string mismatch:", response.s, "!==", input.s);
        process.exit(1);
      }
      // Compare dates
      const responseDate = new Date(response.d);
      if (responseDate.getTime() !== input.d.getTime()) {
        console.error("datetime mismatch:", responseDate, "!==", input.d);
        process.exit(1);
      }

      console.log("Success");
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

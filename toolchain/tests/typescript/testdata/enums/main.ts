// Verifies enum serialization: both string enums and int enums are echoed correctly.
import {
  Server,
  NewClient,
  HTTPAdapter,
  ColorValues,
  PriorityValues,
} from "./gen/index.ts";
import type { Color, Priority } from "./gen/index.ts";
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
    .procs.Test()
    .handle(async ({ input }) => {
      return { c: input.c, p: input.p };
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
      const testCases: { color: Color; priority: Priority }[] = [
        { color: ColorValues.Red, priority: PriorityValues.High },
        { color: ColorValues.Blue, priority: PriorityValues.Low },
      ];

      for (const tc of testCases) {
        const response = await client.procs
          .serviceTest()
          .execute({ c: tc.color, p: tc.priority });

        if (response.c !== tc.color) {
          console.error(`Expected color ${tc.color}, got ${response.c}`);
          process.exit(1);
        }
        if (response.p !== tc.priority) {
          console.error(`Expected priority ${tc.priority}, got ${response.p}`);
          process.exit(1);
        }
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

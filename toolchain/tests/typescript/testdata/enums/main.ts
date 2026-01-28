// Verifies enum serialization: both string enums and int enums are echoed correctly.
import { Server, NewClient, ColorValues, PriorityValues } from "./gen/index.ts";
import { createNodeHandler } from "./gen/adapters/node.ts";
import type { Color, Priority } from "./gen/index.ts";
import { createServer } from "http";

async function main() {
  const server = new Server();

  server.rpcs
    .service()
    .procs.Test()
    .handle(async ({ input }) => {
      return { c: input.c, p: input.p };
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
}

main().catch((err) => {
  console.error(err);
  process.exit(1);
});

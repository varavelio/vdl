// Verifies primitive type serialization: int, float, bool, string, and datetime.
import { Server, NewClient } from "./gen/index.ts";
import { createNodeHandler } from "./gen/adapters/node.ts";
import { createServer } from "http";

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
}

main().catch((err) => {
  console.error(err);
  process.exit(1);
});

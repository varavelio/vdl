// Verifies multiple RPC blocks with different procs work correctly.
import { Server, NewClient } from "./gen/index.ts";
import { createNodeHandler } from "./gen/adapters/node.ts";
import { createServer } from "http";

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

  // Test A.X
  await client.procs.aX().execute({});

  // Test B.Y
  await client.procs.bY().execute({});

  console.log("Success");
  httpServer.close();
  process.exit(0);
}

main().catch((err) => {
  console.error(err);
  process.exit(1);
});

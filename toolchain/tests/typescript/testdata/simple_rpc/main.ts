import { createServer } from "http";
import { createNodeHandler } from "./gen/adapters/node.ts";
import { NewClient, Server } from "./gen/index.ts";

async function main() {
  const server = new Server();

  server.rpcs
    .greeter()
    .procs.Hello()
    .handle(async ({ input }) => {
      return { result: `Hello ${input.name}!` };
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
    const response = await client.procs.greeterHello().execute({ name: "World" });
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
}

main().catch((err) => {
  console.error(err);
  process.exit(1);
});

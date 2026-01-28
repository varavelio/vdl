// Verifies that nested structures and arrays serialize/deserialize correctly.
import { Server, NewClient } from "./gen/index.ts";
import { createNodeHandler } from "./gen/adapters/node.ts";
import { createServer } from "http";

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
}

main().catch((err) => {
  console.error(err);
  process.exit(1);
});

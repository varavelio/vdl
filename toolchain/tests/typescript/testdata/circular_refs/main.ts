import { createNodeHandler } from "./gen/adapters/node.ts";
import {
  Server,
  NewClient,
  SelfReferencing,
  NodeA,
  FullyOptionalA,
} from "./gen/index.ts";
import { createServer } from "http";

async function main() {
  const server = new Server();

  server.rpcs
    .testService()
    .procs.TestCircular()
    .handle(async ({ input }) => {
      return {
        self: input.self,
        chain: input.chain,
        optional: input.optional,
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
    const selfRef: SelfReferencing = {
      id: "root",
      name: "Root Node",
      parent: {
        id: "parent",
        name: "Parent Node",
        parent: {
          id: "grandparent",
          name: "Grandparent Node",
          parent: undefined,
        },
      },
    };

    const chain: NodeA = {
      value: "A",
      nodeB: {
        value: "B",
        nodeC: {
          value: "C",
          nodeD: {
            value: "D",
            nodeE: {
              value: "E",
              backToA: undefined,
            },
          },
        },
      },
    };

    const optional: FullyOptionalA = {
      id: "A",
      b: {
        id: "B",
        c: {
          id: "C",
          d: {
            id: "D",
            a: undefined,
          },
        },
      },
    };

    const response = await client.procs.testServiceTestCircular().execute({
      self: selfRef,
      chain,
      optional,
    });

    if (JSON.stringify(response.self) !== JSON.stringify(selfRef)) {
      console.error("Self-referencing data mismatch");
      process.exit(1);
    }

    if (JSON.stringify(response.chain) !== JSON.stringify(chain)) {
      console.error("Chain data mismatch");
      process.exit(1);
    }

    if (JSON.stringify(response.optional) !== JSON.stringify(optional)) {
      console.error("Fully optional data mismatch");
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

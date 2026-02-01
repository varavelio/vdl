// Verifies complex type serialization: deeply nested structures, maps of arrays,
// arrays of maps, nested objects, and multi-dimensional arrays.
import { createNodeHandler } from "./gen/adapters/node.ts";
import { Server, NewClient } from "./gen/index.ts";
import type {
  Container,
  User,
  Address,
  Level1,
  Level2,
  Level3,
  Point,
  Score,
} from "./gen/index.ts";
import { createServer } from "http";

function deepEqual(a: unknown, b: unknown): boolean {
  return JSON.stringify(a) === JSON.stringify(b);
}

function buildComplexInput(): Container {
  const address: Address = {
    street: "123 Main St",
    city: "Wonderland",
    zip: 90210,
  };
  const user: User = {
    id: 123,
    username: "alice",
    isActive: true,
    tags: ["admin", "editor", "viewer"],
    metadata: { role: "superuser", level: "9000" },
    address: address,
  };

  const level3: Level3 = { value: "deepest" };
  const level2: Level2 = { name: "level2", level3: level3 };
  const level1: Level1 = { id: 1, level2: level2 };

  const points: Point[] = [
    { x: 10, y: 20 },
    { x: 30, y: 40 },
  ];

  const scores: Record<string, Score> = {
    player1: { name: "Alice", value: 95.5 },
    player2: { name: "Bob", value: 87.3 },
  };

  return {
    user: user,
    matrix: [
      [1, 2, 3],
      [4, 5, 6],
      [7, 8, 9],
    ],
    nestedArrays: [
      [
        ["a", "b"],
        ["c", "d"],
      ],
      [
        ["e", "f"],
        ["g", "h"],
      ],
    ],
    mapOfArrays: {
      primes: [2, 3, 5, 7, 11],
      fibonacci: [1, 1, 2, 3, 5, 8],
    },
    arrayOfMaps: [{ key1: "value1", key2: "value2" }, { key3: "value3" }],
    mapOfObjects: {
      alice: {
        id: 1,
        username: "alice",
        isActive: true,
        tags: ["a"],
        metadata: {},
      },
      bob: {
        id: 2,
        username: "bob",
        isActive: false,
        tags: ["b"],
        metadata: { x: "y" },
      },
    },
    deepNest: level1,
    points: points,
    scores: scores,
    arrayOfMapOfArrays: [
      { set1: [1, 2, 3], set2: [4, 5] },
      { set3: [6, 7, 8, 9] },
    ],
  };
}

async function main() {
  const server = new Server();

  server.rpcs
    .service()
    .procs.Echo()
    .handle(async ({ input }) => {
      return { data: input.data };
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
    const complexInput = buildComplexInput();
    const response = await client.procs
      .serviceEcho()
      .execute({ data: complexInput });

    if (!deepEqual(response.data, complexInput)) {
      console.error("Complex types mismatch");
      console.error("Sent:", JSON.stringify(complexInput, null, 2));
      console.error("Got:", JSON.stringify(response.data, null, 2));
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

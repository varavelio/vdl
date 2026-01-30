// Verifies forward compatibility: the client should ignore unknown fields
// returned by the server (simulating a newer server version).
import { NewClient } from "./gen/index.ts";
import { createServer } from "http";

async function main() {
  // Create a custom server that returns extra fields not in the schema
  const httpServer = createServer(async (req, res) => {
    if (req.method !== "POST") {
      res.writeHead(405);
      res.end();
      return;
    }

    // Server returns extra fields that the client doesn't know about
    const response = {
      ok: true,
      output: {
        name: "test",
        value: 42,
        unknownField: "this field doesn't exist in schema",
        extraNested: {
          foo: "bar",
          baz: 123,
        },
        extraArray: ["a", "b", "c"],
      },
    };

    res.writeHead(200, { "Content-Type": "application/json" });
    res.end(JSON.stringify(response));
  });

  await new Promise<void>((resolve) => {
    httpServer.listen(0, resolve);
  });

  const addr = httpServer.address() as any;
  const port = addr.port;
  const baseUrl = `http://localhost:${port}/rpc`;

  const client = NewClient(baseUrl).build();

  try {
    const result = await client.procs.serviceGetData().execute({});

    // Verify the known fields are correctly parsed
    if (result.name !== "test") {
      console.error(`expected name='test', got: ${result.name}`);
      process.exit(1);
    }
    if (result.value !== 42) {
      console.error(`expected value=42, got: ${result.value}`);
      process.exit(1);
    }

    // The test passes if we got here - unknown fields were ignored without error
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

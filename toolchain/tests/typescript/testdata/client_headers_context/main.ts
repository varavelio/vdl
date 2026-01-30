// Verifies header passing through context: client-level and request-level headers
// are extracted by the HTTP handler and returned by the server.
import { NewClient } from "./gen/index.ts";
import { createServer, IncomingMessage, ServerResponse } from "http";

async function main() {
  // Create a custom server that returns headers from the request
  const httpServer = createServer(
    async (req: IncomingMessage, res: ServerResponse) => {
      if (req.method !== "POST") {
        res.writeHead(405);
        res.end();
        return;
      }

      // Extract specific headers and return them
      const values: Record<string, string> = {
        "X-Custom": (req.headers["x-custom"] as string) || "",
        Authorization: (req.headers["authorization"] as string) || "",
      };

      const response = {
        ok: true,
        output: { values },
      };

      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify(response));
    },
  );

  await new Promise<void>((resolve) => {
    httpServer.listen(0, resolve);
  });

  const addr = httpServer.address() as any;
  const port = addr.port;
  const baseUrl = `http://localhost:${port}/rpc`;

  // Create client with global header (Authorization)
  const client = NewClient(baseUrl)
    .withGlobalHeader("Authorization", "Bearer secret")
    .build();

  try {
    // Execute with request-level header (X-Custom)
    const result = await client.procs
      .serviceGetHeaders()
      .withHeader("X-Custom", "123")
      .execute({});

    // Verify both headers were received
    if (result.values["Authorization"] !== "Bearer secret") {
      console.error(
        `missing Authorization header, got: ${result.values["Authorization"]}`,
      );
      process.exit(1);
    }
    if (result.values["X-Custom"] !== "123") {
      console.error(
        `missing X-Custom header, got: ${result.values["X-Custom"]}`,
      );
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

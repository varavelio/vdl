// Verifies that import_extension: ".js" generates correct import paths
import { createNodeHandler } from "./gen/adapters/node.ts";
import { Server, NewClient } from "./gen/index.ts";
import { readFileSync } from "fs";
import { createServer } from "http";

async function main() {
  // =========================================================================
  // Step 1: Verify generated imports have .js extension
  // =========================================================================
  // Tests run from the case directory, so ./gen is correct
  const genDir = "./gen";

  const filesToCheck = ["index.ts", "client.ts", "server.ts", "types.ts"];

  for (const file of filesToCheck) {
    const content = readFileSync(`${genDir}/${file}`, "utf-8");

    // Check for .js imports
    const importMatches = content.match(/from\s+["']\.\/[^"']+["']/g) || [];

    for (const imp of importMatches) {
      if (!imp.includes(".js")) {
        throw new Error(
          `${file}: expected .js extension in import, got: ${imp}`,
        );
      }
      if (imp.includes(".ts")) {
        throw new Error(`${file}: unexpected .ts extension in import: ${imp}`);
      }
    }

    if (importMatches.length === 0 && file !== "types.ts") {
      throw new Error(`${file}: no relative imports found to verify`);
    }
  }

  console.log("Import extension verification passed: all imports use .js");

  // =========================================================================
  // Step 2: Verify runtime functionality
  // =========================================================================
  const server = new Server();

  server.rpcs
    .echo()
    .procs.Say()
    .handle(async ({ input }) => {
      return { message: input.message };
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
    const response = await client.procs.echoSay().execute({ message: "test" });

    if (response.message !== "test") {
      throw new Error(`expected "test", got "${response.message}"`);
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

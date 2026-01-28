// Verifies middleware execution order and coverage across all levels:
// Global middleware applies to all RPCs, RPC middleware applies to all procs in that RPC,
// and Proc middleware applies only to specific procs.
import { Server, NewClient } from "./gen/index.ts";
import { createNodeHandler } from "./gen/adapters/node.ts";
import { createServer, IncomingMessage, ServerResponse } from "http";

interface AppProps {
  trace: string[];
}

async function main() {
  const server = new Server<AppProps>();

  // Global middleware - runs for all RPCs
  server.use((next) => async (c) => {
    c.props.trace.push("Global");
    return next(c);
  });

  // ServiceA RPC middleware
  server.rpcs.serviceA().use((next) => async (c) => {
    c.props.trace.push("RpcA");
    return next(c);
  });

  // ServiceB RPC middleware
  server.rpcs.serviceB().use((next) => async (c) => {
    c.props.trace.push("RpcB");
    return next(c);
  });

  // ServiceA.Proc1 specific middleware
  server.rpcs
    .serviceA()
    .procs.Proc1()
    .use((next) => async (c) => {
      c.props.trace.push("ProcA1");
      return next(c);
    });

  // ServiceA.Proc1 handler
  server.rpcs
    .serviceA()
    .procs.Proc1()
    .handle(async (c) => {
      c.props.trace.push("HandlerA1");
      return { trace: c.props.trace };
    });

  // ServiceA.Proc2 handler (no proc middleware)
  server.rpcs
    .serviceA()
    .procs.Proc2()
    .handle(async (c) => {
      c.props.trace.push("HandlerA2");
      return { trace: c.props.trace };
    });

  // ServiceB.Proc1 handler
  server.rpcs
    .serviceB()
    .procs.Proc1()
    .handle(async (c) => {
      c.props.trace.push("HandlerB1");
      return { trace: c.props.trace };
    });

  // Create handler that initializes trace for each request
  const handler = createNodeHandler<AppProps>(server, () => ({ trace: [] }), {
    prefix: "/rpc",
  });

  const httpServer = createServer(
    async (req: IncomingMessage, res: ServerResponse) => {
      if (req.method !== "POST") {
        res.writeHead(405);
        res.end();
        return;
      }
      await handler(req, res);
    },
  );

  await new Promise<void>((resolve) => {
    httpServer.listen(0, resolve);
  });

  const addr = httpServer.address() as any;
  const baseUrl = `http://localhost:${addr.port}/rpc`;
  const client = NewClient(baseUrl).build();

  const testCases = [
    {
      name: "ServiceA.Proc1 (Global + RpcA + ProcA1)",
      call: async () => {
        const res = await client.procs.serviceAProc1().execute({});
        return res.trace;
      },
      expected: ["Global", "RpcA", "ProcA1", "HandlerA1"],
    },
    {
      name: "ServiceA.Proc2 (Global + RpcA, no proc middleware)",
      call: async () => {
        const res = await client.procs.serviceAProc2().execute({});
        return res.trace;
      },
      expected: ["Global", "RpcA", "HandlerA2"],
    },
    {
      name: "ServiceB.Proc1 (Global + RpcB, no proc middleware)",
      call: async () => {
        const res = await client.procs.serviceBProc1().execute({});
        return res.trace;
      },
      expected: ["Global", "RpcB", "HandlerB1"],
    },
  ];

  for (const tc of testCases) {
    const trace = await tc.call();
    const pass =
      trace.length === tc.expected.length &&
      trace.every((v, i) => v === tc.expected[i]);

    if (!pass) {
      console.error(
        `${tc.name}: expected [${tc.expected.join(", ")}], got [${trace.join(", ")}]`,
      );
      process.exit(1);
    }
  }

  console.log("Success");
  httpServer.close();
  process.exit(0);
}

main().catch((err) => {
  console.error(err);
  process.exit(1);
});

// This imports are just to prevent errors in the IDE when developing, this imports
// are handled in the generator for the generated code

import { HTTPAdapter } from "../server";
type Server<_> = any;

/** START FROM HERE **/

import type { IncomingMessage, ServerResponse } from "node:http";

// -----------------------------------------------------------------------------
// Node.js Adapter - For Express, Fastify, native http, etc.
// -----------------------------------------------------------------------------

/**
 * NodeAdapter implements HTTPAdapter for Node.js HTTP environments.
 *
 * This adapter works with Node.js `IncomingMessage` and `ServerResponse` objects
 * used by Express, Fastify, native `http.createServer`, and similar frameworks.
 *
 * It supports:
 * - JSON body parsing (manual buffering or pre-parsed by middleware)
 * - Streaming responses (SSE)
 * - Connection close detection
 */
export class NodeAdapter implements HTTPAdapter {
  private req: IncomingMessage;
  private res: ServerResponse;
  private parsedBody: unknown;
  private bodyPromise: Promise<unknown> | null;
  private closeCallbacks: (() => void)[];
  private closed: boolean;

  /**
   * Creates a new NodeAdapter.
   *
   * @param req - The Node.js IncomingMessage (request)
   * @param res - The Node.js ServerResponse
   * @param parsedBody - Optional pre-parsed body (from Express body-parser, etc.)
   */
  constructor(req: IncomingMessage, res: ServerResponse, parsedBody?: unknown) {
    this.req = req;
    this.res = res;
    this.parsedBody = parsedBody;
    this.bodyPromise = null;
    this.closeCallbacks = [];
    this.closed = false;

    // Listen for premature client disconnect (aborted connection)
    // The 'close' event on res fires when the underlying connection is closed
    // We use 'aborted' on req to detect if the client closed the connection early
    req.on("aborted", () => {
      this.closed = true;
      this.closeCallbacks.forEach((cb) => cb());
    });

    // The 'close' event on res is fired when the response has been sent
    // OR when the underlying connection is closed before finishing
    // We only care about premature closes, not normal completion
    res.on("close", () => {
      // Only mark closed if we haven't finished writing
      if (!this.res.writableFinished) {
        this.closed = true;
        this.closeCallbacks.forEach((cb) => cb());
      }
    });

    // If no pre-parsed body, start buffering immediately to avoid missing data events
    if (parsedBody === undefined) {
      this.startBodyBuffering();
    }
  }

  /**
   * Starts buffering the request body immediately.
   * This must be called in the constructor to avoid missing data events.
   */
  private startBodyBuffering(): void {
    const chunks: Buffer[] = [];

    this.bodyPromise = new Promise<unknown>((resolve, reject) => {
      this.req.on("data", (chunk: Buffer) => {
        chunks.push(chunk);
      });

      this.req.on("end", () => {
        try {
          const bodyString = Buffer.concat(chunks).toString("utf-8");
          if (!bodyString || bodyString.trim() === "") {
            resolve({});
          } else {
            resolve(JSON.parse(bodyString));
          }
        } catch (err) {
          reject(new Error(`Failed to parse JSON body: ${err}`));
        }
      });

      this.req.on("error", (err) => {
        reject(err);
      });
    });
  }

  /**
   * Returns the parsed JSON body of the request.
   *
   * If a pre-parsed body was provided (e.g., from Express body-parser),
   * it returns that directly. Otherwise, it returns the buffered and parsed
   * request body.
   */
  async json(): Promise<unknown> {
    // If pre-parsed body was provided, use it
    if (this.parsedBody !== undefined) {
      return this.parsedBody;
    }

    // Return the body promise (started in constructor)
    if (this.bodyPromise) {
      return this.bodyPromise;
    }

    // Fallback: if somehow bodyPromise wasn't created, return empty object
    return {};
  }

  /**
   * Sets a response header.
   */
  setHeader(key: string, value: string): void {
    if (!this.res.headersSent) {
      this.res.setHeader(key, value);
    }
  }

  /**
   * Writes data to the response.
   */
  write(data: string): void {
    if (this.closed || this.res.writableEnded) return;
    this.res.write(data);
  }

  /**
   * Flushes buffered response data to the client.
   * Important for SSE to ensure real-time delivery.
   */
  flush(): void {
    if (this.closed || this.res.writableEnded) return;

    // Node.js ServerResponse may have a flush method when using compression
    // or be wrapped by frameworks. We attempt to call it if available.
    const resAny = this.res as any;
    if (typeof resAny.flush === "function") {
      resAny.flush();
    }
    // For the native http module, writes are typically unbuffered
    // but we ensure cork/uncork behavior is handled
    if (typeof resAny.flushHeaders === "function" && !this.res.headersSent) {
      resAny.flushHeaders();
    }
  }

  /**
   * Signals that the response is complete.
   */
  end(): void {
    if (!this.res.writableEnded) {
      this.res.end();
    }
  }

  /**
   * Registers a callback for when the connection closes.
   */
  onClose(callback: () => void): void {
    this.closeCallbacks.push(callback);
    // If already closed, call immediately
    if (this.closed) {
      callback();
    }
  }
}

/**
 * Extracts RPC name and operation name from a URL path.
 *
 * @param url - The request URL (e.g., "/rpc/Service/Echo")
 * @param prefix - Optional path prefix to strip (e.g., "/rpc")
 * @returns Object with rpcName and operationName, or null if parsing fails
 */
export function parseNodeRpcPath(
  url: string,
  prefix?: string,
): { rpcName: string; operationName: string } | null {
  // Extract pathname (strip query string)
  let pathname = url.split("?")[0] || url;

  // Remove prefix if specified
  if (prefix) {
    const normalizedPrefix = prefix.startsWith("/") ? prefix : "/" + prefix;
    if (pathname.startsWith(normalizedPrefix)) {
      pathname = pathname.slice(normalizedPrefix.length);
    }
  }

  // Remove leading/trailing slashes and split
  const segments = pathname.replace(/^\/+|\/+$/g, "").split("/");

  // We need at least 2 segments: rpcName and operationName
  if (segments.length < 2) {
    return null;
  }

  // Take the last two segments as rpcName and operationName
  const operationName = segments.pop()!;
  const rpcName = segments.pop()!;

  if (!rpcName || !operationName) {
    return null;
  }

  return { rpcName, operationName };
}

/**
 * Options for createNodeHandler.
 */
export interface NodeHandlerOptions<T> {
  /**
   * URL path prefix to strip before parsing RPC/operation names.
   * Example: "/rpc" or "/api/v1"
   */
  prefix?: string;
}

/**
 * Creates a Node.js HTTP request handler for a VDL Server.
 *
 * This handler can be used with:
 * - Native `http.createServer()`
 * - Express: `app.use('/rpc', handler)` or `app.all('/rpc/*', handler)`
 * - Fastify: `fastify.all('/rpc/*', handler)`
 * - Any Node.js HTTP framework
 *
 * @example
 * ```typescript
 * import { createServer } from "http";
 * import { Server } from "./gen/server";
 * import { createNodeHandler } from "./gen/adapters/node";
 *
 * const server = new Server<MyContext>();
 * // ... register handlers ...
 *
 * const handler = createNodeHandler(server, {
 *   prefix: "/rpc",
 * });
 *
 * // Native http
 * createServer(async (req, res) => {
 *   await handler(req, res);
 * }).listen(3000);
 *
 * // Express
 * app.use('/rpc', async (req, res, next) => {
 *   try {
 *     await handler(req, res);
 *   } catch (err) {
 *     next(err);
 *   }
 * });
 * ```
 *
 * @param server - The VDL Server instance
 * @param createContext - Optional function to create the context (props) from req/res
 * @param options - Handler options including path prefix
 * @returns An async function that handles Node.js HTTP requests
 */
export function createNodeHandler<T = unknown>(
  server: Server<T>,
  createContext?: (req: IncomingMessage, res: ServerResponse) => T | Promise<T>,
  options?: NodeHandlerOptions<T>,
): (req: IncomingMessage, res: ServerResponse) => Promise<void> {
  const prefix = options?.prefix;

  return async (req: IncomingMessage, res: ServerResponse): Promise<void> => {
    const url = req.url || "/";

    // Parse the URL to extract RPC and operation names
    const parsed = parseNodeRpcPath(url, prefix);

    if (!parsed) {
      if (!res.headersSent) {
        res.writeHead(404, { "Content-Type": "application/json" });
      }
      res.end(
        JSON.stringify({
          ok: false,
          error: {
            code: "NOT_FOUND",
            message:
              "Invalid RPC path. Expected: /[prefix/]rpcName/operationName",
          },
        }),
      );
      return;
    }

    const { rpcName, operationName } = parsed;

    // Check if the body was already parsed by middleware (Express body-parser, etc.)
    // We need to check this BEFORE creating the adapter
    const reqAny = req as any;
    const parsedBody = reqAny.body;

    // Create the adapter immediately - it will start buffering the body if needed
    const adapter = new NodeAdapter(req, res, parsedBody);

    // Create context
    let props: T;
    try {
      props = createContext
        ? await createContext(req, res)
        : (undefined as unknown as T);
    } catch (err) {
      if (!res.headersSent) {
        res.writeHead(500, { "Content-Type": "application/json" });
      }
      res.end(
        JSON.stringify({
          ok: false,
          error: {
            code: "CONTEXT_ERROR",
            message: "Failed to create request context",
            details: { originalError: String(err) },
          },
        }),
      );
      return;
    }

    // Process the request
    try {
      await server.handleRequest(props, rpcName, operationName, adapter);
    } catch (err) {
      // Fatal error - return 500 if headers not sent
      if (!res.headersSent) {
        res.writeHead(500, { "Content-Type": "application/json" });
      }
      if (!res.writableEnded) {
        res.end(
          JSON.stringify({
            ok: false,
            error: {
              code: "INTERNAL_ERROR",
              message: "Internal server error",
              details: { originalError: String(err) },
            },
          }),
        );
      }
    }
  };
}

// This imports are just to prevent errors in the IDE when developing, this imports
// are handled in the generator for the generated code

import type { HTTPAdapter } from "../server";
type Server<_> = any;

/** START FROM HERE **/

// -----------------------------------------------------------------------------
// Fetch Adapter - Universal Web Standards (Bun, Deno, Cloudflare Workers, etc.)
// -----------------------------------------------------------------------------

/**
 * FetchAdapter implements HTTPAdapter for Web Standards environments.
 *
 * This adapter works with the standard `Request` and `Response` objects
 * available in Bun, Deno, Cloudflare Workers, and other edge runtimes.
 *
 * It supports:
 * - JSON body parsing
 * - Streaming responses (SSE via ReadableStream)
 * - Connection abort handling via AbortSignal
 */
export class FetchAdapter implements HTTPAdapter {
  private request: Request;
  private headers: Map<string, string>;
  private chunks: string[];
  private streamController: ReadableStreamDefaultController<Uint8Array> | null;
  private encoder: TextEncoder;
  private closeCallbacks: (() => void)[];
  private aborted: boolean;

  constructor(request: Request) {
    this.request = request;
    this.headers = new Map();
    this.chunks = [];
    this.streamController = null;
    this.encoder = new TextEncoder();
    this.closeCallbacks = [];
    this.aborted = false;

    // Monitor abort signal if available
    if (request.signal) {
      request.signal.addEventListener("abort", () => {
        this.aborted = true;
        this.closeCallbacks.forEach((cb) => cb());
      });
    }
  }

  /**
   * Returns the parsed JSON body of the request.
   */
  async json(): Promise<unknown> {
    return this.request.json();
  }

  /**
   * Sets a response header.
   */
  setHeader(key: string, value: string): void {
    this.headers.set(key, value);
  }

  /**
   * Writes data to the response buffer or stream.
   */
  write(data: string): void {
    if (this.aborted) return;

    if (this.streamController) {
      // Streaming mode: push directly to the stream
      this.streamController.enqueue(this.encoder.encode(data));
    } else {
      // Buffered mode: accumulate chunks
      this.chunks.push(data);
    }
  }

  /**
   * Flushes buffered data.
   * In streaming mode, data is already sent immediately.
   */
  flush(): void {
    // No-op for FetchAdapter as writes are immediate in streaming mode
  }

  /**
   * Signals that the response is complete.
   */
  end(): void {
    if (this.streamController) {
      try {
        this.streamController.close();
      } catch {
        // Controller may already be closed
      }
    }
    this.closeCallbacks.forEach((cb) => cb());
  }

  /**
   * Registers a callback for when the connection closes.
   */
  onClose(callback: () => void): void {
    this.closeCallbacks.push(callback);
    // If already aborted, call immediately
    if (this.aborted) {
      callback();
    }
  }

  /**
   * Builds the final Response object.
   * Call this after handleRequest completes for procedures.
   */
  toResponse(): Response {
    const body = this.chunks.join("");
    return new Response(body, {
      status: 200,
      headers: Object.fromEntries(this.headers),
    });
  }

  /**
   * Creates a streaming Response for SSE.
   * Call this before handleRequest for streams.
   */
  toStreamingResponse(): Response {
    const stream = new ReadableStream<Uint8Array>({
      start: (controller) => {
        this.streamController = controller;
      },
      cancel: () => {
        this.aborted = true;
        this.closeCallbacks.forEach((cb) => cb());
      },
    });

    return new Response(stream, {
      status: 200,
      headers: Object.fromEntries(this.headers),
    });
  }
}

/**
 * Extracts RPC name and operation name from a URL path.
 *
 * Assumes the path format: `/prefix/rpcName/operationName` or `/rpcName/operationName`
 *
 * @param url - The request URL
 * @param prefix - Optional path prefix (e.g., "/rpc", "/api/v1")
 * @returns Object with rpcName and operationName, or null if parsing fails
 */
export function parseRpcPath(
  url: string | URL,
  prefix?: string,
): { rpcName: string; operationName: string } | null {
  const urlObj = typeof url === "string" ? new URL(url) : url;
  let pathname = urlObj.pathname;

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
 * Options for createFetchHandler.
 */
export interface FetchHandlerOptions<T> {
  /**
   * URL path prefix to strip before parsing RPC/operation names.
   * Example: "/rpc" or "/api/v1"
   */
  prefix?: string;
}

/**
 * Creates a Fetch API compatible request handler for a VDL Server.
 *
 * This handler can be used directly with:
 * - Bun.serve()
 * - Deno.serve()
 * - Cloudflare Workers fetch handler
 * - Any Web Standards compatible runtime
 *
 * @example
 * ```typescript
 * const server = new Server<MyContext>();
 * // ... register handlers ...
 *
 * const handler = createFetchHandler(server, {
 *   prefix: "/rpc",
 * });
 *
 * // Bun
 * Bun.serve({ fetch: handler });
 *
 * // Deno
 * Deno.serve(handler);
 *
 * // Cloudflare Workers
 * export default { fetch: handler };
 * ```
 *
 * @param server - The VDL Server instance
 * @param createContext - Optional function to create the context (props) from the request
 * @param options - Handler options including path prefix
 * @returns A function that handles Fetch API requests
 */
export function createFetchHandler<T = unknown>(
  server: Server<T>,
  createContext?: (req: Request) => T | Promise<T>,
  options?: FetchHandlerOptions<T>,
): (req: Request) => Promise<Response> {
  const prefix = options?.prefix;

  return async (req: Request): Promise<Response> => {
    // Parse the URL to extract RPC and operation names
    const parsed = parseRpcPath(req.url, prefix);

    if (!parsed) {
      return new Response(
        JSON.stringify({
          ok: false,
          error: {
            code: "NOT_FOUND",
            message:
              "Invalid RPC path. Expected: /[prefix/]rpcName/operationName",
          },
        }),
        {
          status: 404,
          headers: { "Content-Type": "application/json" },
        },
      );
    }

    const { rpcName, operationName } = parsed;

    // Create context
    let props: T;
    try {
      props = createContext
        ? await createContext(req)
        : (undefined as unknown as T);
    } catch (err) {
      return new Response(
        JSON.stringify({
          ok: false,
          error: {
            code: "CONTEXT_ERROR",
            message: "Failed to create request context",
            details: { originalError: String(err) },
          },
        }),
        {
          status: 500,
          headers: { "Content-Type": "application/json" },
        },
      );
    }

    // Create the adapter
    const adapter = new FetchAdapter(req);

    // Determine if this is likely a stream request based on Accept header
    const acceptHeader = req.headers.get("Accept") || "";
    const isStreamRequest = acceptHeader.includes("text/event-stream");

    if (isStreamRequest) {
      // For streams, we need to return the streaming response immediately
      // and process the request in the background
      const response = adapter.toStreamingResponse();

      // Process request in background (don't await)
      server.handleRequest(props, rpcName, operationName, adapter).catch(() => {
        // Error handling is done inside handleRequest
        adapter.end();
      });

      return response;
    }

    // For procedures, wait for completion
    try {
      await server.handleRequest(props, rpcName, operationName, adapter);
    } catch (err) {
      // Fatal error - return 500
      return new Response(
        JSON.stringify({
          ok: false,
          error: {
            code: "INTERNAL_ERROR",
            message: "Internal server error",
            details: { originalError: String(err) },
          },
        }),
        {
          status: 500,
          headers: { "Content-Type": "application/json" },
        },
      );
    }

    return adapter.toResponse();
  };
}

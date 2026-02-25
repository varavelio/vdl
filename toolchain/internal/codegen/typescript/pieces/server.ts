// This imports are just to prevent errors in the IDE when developing, this imports
// are handled in the generator for the generated code
import {
  asError,
  type OperationDefinition,
  type OperationType,
  type Response,
  VdlError,
} from "./core";

/** START FROM HERE **/

// -----------------------------------------------------------------------------
// Server Types
// -----------------------------------------------------------------------------

/**
 * HTTPAdapter defines the interface required by VDL server to handle
 * incoming HTTP requests and write responses to clients.
 *
 * This abstraction allows the server to work with different HTTP frameworks
 * (Express, Fastify, standard http, etc.) while maintaining the same core functionality.
 */
export interface HTTPAdapter {
  /**
   * Returns the parsed JSON body of the request.
   * The server expects the body to be a JSON object containing the RPC input.
   */
  json(): Promise<unknown>;

  /**
   * Sets a response header with the specified key-value pair.
   */
  setHeader(key: string, value: string): void;

  /**
   * Writes data to the response.
   * For procedures, this is called once with the full response.
   * For streams, this is called for each chunk.
   */
  write(data: string): void;

  /**
   * Flushes any buffered response data to the client.
   * Crucial for streaming responses to ensure real-time delivery.
   * Optional if the underlying framework handles this automatically.
   */
  flush?(): void;

  /**
   * Signals that the response is complete.
   */
  end(): void;

  /**
   * Registers a callback to be invoked when the connection is closed by the client.
   * Used to stop stream processing when the client disconnects.
   */
  onClose(callback: () => void): void;
}

/**
 * HandlerContext is the unified container for all request information and state
 * that flows through the entire request processing pipeline.
 *
 * @typeParam T - The application context type containing dependencies and request-scoped data.
 * @typeParam I - The input payload type for this operation.
 */
export class HandlerContext<T, I> {
  /** User-defined container for application dependencies and request data. */
  public props: T;

  /** The request input, already deserialized and typed. */
  public input: I;

  /** Signal for cancellation (client disconnects). */
  public signal: AbortSignal;

  /** Details of the RPC operation. */
  public readonly operation: OperationDefinition;

  constructor(props: T, input: I, signal: AbortSignal, operation: OperationDefinition) {
    this.props = props;
    this.input = input;
    this.signal = signal;
    this.operation = operation;
  }

  get rpcName(): string {
    return this.operation.rpcName;
  }
  get operationName(): string {
    return this.operation.name;
  }
  get operationType(): OperationType {
    return this.operation.type;
  }
}

// -----------------------------------------------------------------------------
// Middleware Types
// -----------------------------------------------------------------------------

/**
 * A handler function for global middleware.
 * At this level, input and output types are unknown.
 *
 * @typeParam T - The application context type (props).
 */
export type GlobalHandlerFunc<T> = (c: HandlerContext<T, unknown>) => Promise<unknown>;

/**
 * Middleware that applies to all requests (Procs and Streams).
 * Wraps the next handler in the chain.
 *
 * @typeParam T - The application context type (props).
 */
export type GlobalMiddlewareFunc<T> = (next: GlobalHandlerFunc<T>) => GlobalHandlerFunc<T>;

/**
 * The core logic for a Procedure.
 * Receives context with typed input and returns a Promise with typed output.
 *
 * @typeParam T - The application context type (props).
 * @typeParam I - The input payload type.
 * @typeParam O - The output payload type.
 */
export type ProcHandlerFunc<T, I, O> = (c: HandlerContext<T, I>) => Promise<O>;

/**
 * Middleware specific to Procedures.
 * Can inspect/modify typed input and output.
 *
 * @typeParam T - The application context type (props).
 * @typeParam I - The input payload type.
 * @typeParam O - The output payload type.
 */
export type ProcMiddlewareFunc<T, I, O> = (
  next: ProcHandlerFunc<T, I, O>,
) => ProcHandlerFunc<T, I, O>;

/**
 * Function to emit an event to a stream.
 *
 * @typeParam T - The application context type (props).
 * @typeParam I - The input payload type of the stream.
 * @typeParam O - The output payload type of the emitted event.
 */
export type EmitFunc<T, I, O> = (c: HandlerContext<T, I>, output: O) => Promise<void>;

/**
 * Middleware that wraps the emit function of a stream.
 * Can be used to transform outgoing events or handle backpressure.
 *
 * @typeParam T - The application context type (props).
 * @typeParam I - The input payload type of the stream.
 * @typeParam O - The output payload type of the emitted event.
 */
export type EmitMiddlewareFunc<T, I, O> = (next: EmitFunc<T, I, O>) => EmitFunc<T, I, O>;

/**
 * The core logic for a Stream.
 * Receives context with typed input and an emit function to send events.
 *
 * @typeParam T - The application context type (props).
 * @typeParam I - The input payload type.
 * @typeParam O - The output event type.
 */
export type StreamHandlerFunc<T, I, O> = (
  c: HandlerContext<T, I>,
  emit: EmitFunc<T, I, O>,
) => Promise<void>;

/**
 * Middleware specific to Stream handlers.
 * Wraps the execution of the stream function itself.
 *
 * @typeParam T - The application context type (props).
 * @typeParam I - The input payload type.
 * @typeParam O - The output event type.
 */
export type StreamMiddlewareFunc<T, I, O> = (
  next: StreamHandlerFunc<T, I, O>,
) => StreamHandlerFunc<T, I, O>;

/**
 * Internal function to deserialize raw input (JSON) into typed objects.
 */
export type DeserializerFunc = (raw: unknown) => Promise<unknown>;

/**
 * Custom error handler to transform errors into VdlError responses.
 *
 * @typeParam T - The application context type (props).
 */
export type ErrorHandlerFunc<T> = (c: HandlerContext<T, unknown>, error: unknown) => VdlError;

/**
 * Configuration for stream behavior (Server-Sent Events).
 */
export interface StreamConfig {
  /**
   * Interval in milliseconds at which ping events are sent to the client.
   * Used to keep the connection alive and detect disconnected clients.
   *
   * Default: 30000 (30s).
   */
  pingIntervalMs?: number;
}

// -----------------------------------------------------------------------------
// Server Internal Implementation
// -----------------------------------------------------------------------------

/**
 * The core server engine used by generated VDL server wrappers.
 *
 * This class manages:
 * - Request routing (RPCs, Procs, Streams)
 * - Middleware execution chains
 * - Input deserialization
 * - Error handling
 * - Response formatting (JSON for Procs, SSE for Streams)
 *
 * **Do not instantiate directly.** Use the generated `NewServer` function.
 *
 * @typeParam T - The application context type (props) containing dependencies.
 */
export class InternalServer<T> {
  private operationDefs: Map<string, Map<string, OperationType>>;

  // Handlers
  private procHandlers: Map<string, Map<string, ProcHandlerFunc<T, any, any>>>;
  private streamHandlers: Map<string, Map<string, StreamHandlerFunc<T, any, any>>>;

  // Middlewares
  private globalMiddlewares: GlobalMiddlewareFunc<T>[];
  private rpcMiddlewares: Map<string, GlobalMiddlewareFunc<T>[]>;
  private procMiddlewares: Map<string, Map<string, ProcMiddlewareFunc<T, any, any>[]>>;
  private streamMiddlewares: Map<string, Map<string, StreamMiddlewareFunc<T, any, any>[]>>;
  private streamEmitMiddlewares: Map<string, Map<string, EmitMiddlewareFunc<T, any, any>[]>>;

  // Configs & Helpers
  private procDeserializers: Map<string, Map<string, DeserializerFunc>>;
  private streamDeserializers: Map<string, Map<string, DeserializerFunc>>;
  private globalStreamConfig: StreamConfig;
  private rpcStreamConfigs: Map<string, StreamConfig>;
  private streamConfigs: Map<string, Map<string, StreamConfig>>;
  private globalErrorHandler?: ErrorHandlerFunc<T>;
  private rpcErrorHandlers: Map<string, ErrorHandlerFunc<T>>;

  /**
   * Creates a new internal server.
   *
   * @param procDefs - Procedure definitions from the schema.
   * @param streamDefs - Stream definitions from the schema.
   */
  constructor(procDefs: OperationDefinition[], streamDefs: OperationDefinition[]) {
    this.operationDefs = new Map();
    this.procHandlers = new Map();
    this.streamHandlers = new Map();
    this.globalMiddlewares = [];
    this.rpcMiddlewares = new Map();
    this.procMiddlewares = new Map();
    this.streamMiddlewares = new Map();
    this.streamEmitMiddlewares = new Map();
    this.procDeserializers = new Map();
    this.streamDeserializers = new Map();
    this.globalStreamConfig = { pingIntervalMs: 30000 };
    this.rpcStreamConfigs = new Map();
    this.streamConfigs = new Map();
    this.rpcErrorHandlers = new Map();

    const ensureRPC = (rpcName: string) => {
      if (!this.operationDefs.has(rpcName)) {
        this.operationDefs.set(rpcName, new Map());
        this.procHandlers.set(rpcName, new Map());
        this.streamHandlers.set(rpcName, new Map());
        this.rpcMiddlewares.set(rpcName, []);
        this.procMiddlewares.set(rpcName, new Map());
        this.streamMiddlewares.set(rpcName, new Map());
        this.streamEmitMiddlewares.set(rpcName, new Map());
        this.procDeserializers.set(rpcName, new Map());
        this.streamDeserializers.set(rpcName, new Map());
        this.rpcStreamConfigs.set(rpcName, {});
        this.streamConfigs.set(rpcName, new Map());
      }
    };

    for (const def of procDefs) {
      ensureRPC(def.rpcName);
      this.operationDefs.get(def.rpcName)!.set(def.name, def.type);
    }
    for (const def of streamDefs) {
      ensureRPC(def.rpcName);
      this.operationDefs.get(def.rpcName)!.set(def.name, def.type);
    }
  }

  // Registration Methods

  /** Registers a global middleware applied to all requests. */
  addGlobalMiddleware(mw: GlobalMiddlewareFunc<T>) {
    this.globalMiddlewares.push(mw);
  }

  /** Registers a middleware applied to all requests within a specific RPC group. */
  addRPCMiddleware(rpcName: string, mw: GlobalMiddlewareFunc<T>) {
    const list = this.rpcMiddlewares.get(rpcName);
    if (list) list.push(mw);
  }

  /** Registers a middleware for a specific procedure. */
  addProcMiddleware(rpcName: string, procName: string, mw: ProcMiddlewareFunc<T, any, any>) {
    const rpcMap = this.procMiddlewares.get(rpcName);
    if (rpcMap) {
      if (!rpcMap.has(procName)) rpcMap.set(procName, []);
      rpcMap.get(procName)!.push(mw);
    }
  }

  /** Registers a middleware for a specific stream handler. */
  addStreamMiddleware(rpcName: string, streamName: string, mw: StreamMiddlewareFunc<T, any, any>) {
    const rpcMap = this.streamMiddlewares.get(rpcName);
    if (rpcMap) {
      if (!rpcMap.has(streamName)) rpcMap.set(streamName, []);
      rpcMap.get(streamName)!.push(mw);
    }
  }

  /** Registers a middleware for a specific stream's emit function. */
  addStreamEmitMiddleware(
    rpcName: string,
    streamName: string,
    mw: EmitMiddlewareFunc<T, any, any>,
  ) {
    const rpcMap = this.streamEmitMiddlewares.get(rpcName);
    if (rpcMap) {
      if (!rpcMap.has(streamName)) rpcMap.set(streamName, []);
      rpcMap.get(streamName)!.push(mw);
    }
  }

  /** Sets the global configuration for streams. */
  setGlobalStreamConfig(cfg: StreamConfig) {
    this.globalStreamConfig = { ...this.globalStreamConfig, ...cfg };
    if (!this.globalStreamConfig.pingIntervalMs || this.globalStreamConfig.pingIntervalMs <= 0) {
      this.globalStreamConfig.pingIntervalMs = 30000;
    }
  }

  /** Sets the stream configuration for a specific RPC group. */
  setRPCStreamConfig(rpcName: string, cfg: StreamConfig) {
    this.rpcStreamConfigs.set(rpcName, cfg);
  }

  /** Sets the configuration for a specific stream. */
  setStreamConfig(rpcName: string, streamName: string, cfg: StreamConfig) {
    const rpcMap = this.streamConfigs.get(rpcName);
    if (rpcMap) rpcMap.set(streamName, cfg);
  }

  /** Sets a custom global error handler. */
  setGlobalErrorHandler(handler: ErrorHandlerFunc<T>) {
    this.globalErrorHandler = handler;
  }

  /** Sets a custom error handler for a specific RPC group. */
  setRPCErrorHandler(rpcName: string, handler: ErrorHandlerFunc<T>) {
    this.rpcErrorHandlers.set(rpcName, handler);
  }

  /**
   * Registers a handler implementation for a procedure.
   * Called by the generated code.
   */
  setProcHandler(
    rpcName: string,
    procName: string,
    handler: ProcHandlerFunc<T, any, any>,
    deserializer: DeserializerFunc,
  ) {
    const rpcHandlers = this.procHandlers.get(rpcName);
    if (rpcHandlers) {
      if (rpcHandlers.has(procName)) {
        throw new Error(`Procedure handler for ${rpcName}.${procName} is already registered`);
      }
      rpcHandlers.set(procName, handler);
    }
    const rpcDeserializers = this.procDeserializers.get(rpcName);
    if (rpcDeserializers) rpcDeserializers.set(procName, deserializer);
  }

  /**
   * Registers a handler implementation for a stream.
   * Called by the generated code.
   */
  setStreamHandler(
    rpcName: string,
    streamName: string,
    handler: StreamHandlerFunc<T, any, any>,
    deserializer: DeserializerFunc,
  ) {
    const rpcHandlers = this.streamHandlers.get(rpcName);
    if (rpcHandlers) {
      if (rpcHandlers.has(streamName)) {
        throw new Error(`Stream handler for ${rpcName}.${streamName} is already registered`);
      }
      rpcHandlers.set(streamName, handler);
    }
    const rpcDeserializers = this.streamDeserializers.get(rpcName);
    if (rpcDeserializers) rpcDeserializers.set(streamName, deserializer);
  }

  // Runtime Logic

  /**
   * Main entry point for handling an incoming HTTP request.
   *
   * 1. Validates the request path (RPC/Operation).
   * 2. Deserializes the input.
   * 3. Executes the middleware chain.
   * 4. Invokes the registered handler.
   * 5. Writes the response via the adapter.
   */
  async handleRequest(
    props: T,
    rpcName: string,
    operationName: string,
    adapter: HTTPAdapter,
  ): Promise<void> {
    if (!adapter) throw new Error("HTTPAdapter is required");

    let rawInput: unknown;
    try {
      rawInput = await adapter.json();
    } catch (err) {
      return this.writeProcResponse(adapter, {
        ok: false,
        error: new VdlError({
          message: "Invalid request body",
          details: { originalError: String(err) },
        }),
      });
    }

    const rpcOps = this.operationDefs.get(rpcName);
    const opType = rpcOps?.get(operationName);

    if (!opType) {
      return this.writeProcResponse(adapter, {
        ok: false,
        error: new VdlError({
          message: `Invalid operation: ${rpcName}.${operationName}`,
        }),
      });
    }

    const abortController = new AbortController();
    adapter.onClose(() => abortController.abort());

    const ctx = new HandlerContext<T, unknown>(props, rawInput, abortController.signal, {
      rpcName,
      name: operationName,
      type: opType,
    });

    if (opType === "stream") {
      await this.handleStreamRequest(ctx, rpcName, operationName, adapter);
      return;
    }

    // Procedure
    let response: Response<unknown>;
    try {
      const output = await this.handleProcRequest(ctx, rpcName, operationName);
      response = { ok: true, output };
    } catch (err) {
      const handler = this.resolveErrorHandler(rpcName);
      response = { ok: false, error: handler(ctx, err) };
    }
    return this.writeProcResponse(adapter, response);
  }

  /**
   * Resolves the appropriate error handler for an RPC group.
   * Priority: RPC-specific > Global > Default.
   */
  private resolveErrorHandler(rpcName: string): ErrorHandlerFunc<T> {
    const rpcHandler = this.rpcErrorHandlers.get(rpcName);
    if (rpcHandler) return rpcHandler;
    if (this.globalErrorHandler) return this.globalErrorHandler;

    // Default passthrough
    return (_, err) => asError(err);
  }

  /**
   * Executes a procedure call including all middleware layers.
   */
  private async handleProcRequest(
    c: HandlerContext<T, unknown>,
    rpcName: string,
    procName: string,
  ): Promise<unknown> {
    const baseHandler = this.procHandlers.get(rpcName)?.get(procName);
    const mws = this.procMiddlewares.get(rpcName)?.get(procName);
    const rpcMws = this.rpcMiddlewares.get(rpcName);
    const deserialize = this.procDeserializers.get(rpcName)?.get(procName);

    if (!baseHandler || !deserialize) {
      throw new Error(`${rpcName}.${procName} procedure not implemented`);
    }

    // Deserialize
    c.input = await deserialize(c.input);

    // Chain: Global -> RPC -> Proc -> Handler
    // Middlewares wrap 'next'. So we build from inside out (Handler -> Proc -> RPC -> Global)
    // Actually, middleware signature is (next) -> next.
    // So `wrapped = mw(next)`.
    // Order: Global MW runs first, calls next...
    // So we apply global MWs in reverse order to the inner chain.

    let next: GlobalHandlerFunc<T> = (ctx) => baseHandler(ctx as HandlerContext<T, any>);

    // Apply Proc MWs
    if (mws && mws.length > 0) {
      // We need to adapt ProcHandlerFunc to GlobalHandlerFunc temporarily or cast
      // The logic is: baseHandler is ProcHandlerFunc.
      // We build the proc chain first.
      let procNext = baseHandler;
      for (let i = mws.length - 1; i >= 0; i--) {
        procNext = mws[i](procNext);
      }
      next = (ctx) => procNext(ctx as HandlerContext<T, any>);
    }

    // Apply RPC MWs
    if (rpcMws && rpcMws.length > 0) {
      for (let i = rpcMws.length - 1; i >= 0; i--) {
        next = rpcMws[i](next);
      }
    }

    // Apply Global MWs
    if (this.globalMiddlewares.length > 0) {
      for (let i = this.globalMiddlewares.length - 1; i >= 0; i--) {
        next = this.globalMiddlewares[i](next);
      }
    }

    return next(c);
  }

  /**
   * Executes a stream request including connection setup, pings, and middleware.
   */
  private async handleStreamRequest(
    c: HandlerContext<T, unknown>,
    rpcName: string,
    streamName: string,
    adapter: HTTPAdapter,
  ): Promise<void> {
    const baseHandler = this.streamHandlers.get(rpcName)?.get(streamName);
    const streamMws = this.streamMiddlewares.get(rpcName)?.get(streamName);
    const emitMws = this.streamEmitMiddlewares.get(rpcName)?.get(streamName);
    const rpcMws = this.rpcMiddlewares.get(rpcName);
    const deserialize = this.streamDeserializers.get(rpcName)?.get(streamName);

    // Config
    let pingInterval = this.globalStreamConfig.pingIntervalMs || 30000;
    pingInterval = pingInterval <= 0 ? 3000 : pingInterval;
    const rpcCfg = this.rpcStreamConfigs.get(rpcName);
    if (rpcCfg?.pingIntervalMs) pingInterval = rpcCfg.pingIntervalMs;
    const streamCfg = this.streamConfigs.get(rpcName)?.get(streamName);
    if (streamCfg?.pingIntervalMs) pingInterval = streamCfg.pingIntervalMs;

    // Headers
    adapter.setHeader("Content-Type", "text/event-stream");
    adapter.setHeader("Cache-Control", "no-cache");
    adapter.setHeader("Connection", "keep-alive");

    if (!baseHandler || !deserialize) {
      // Send error event
      const res: Response<unknown> = {
        ok: false,
        error: new VdlError({
          message: `${rpcName}.${streamName} not implemented`,
        }),
      };
      adapter.write(`data: ${JSON.stringify(res)}\n\n`);
      if (adapter.flush) adapter.flush();
      adapter.end();
      return;
    }

    try {
      c.input = await deserialize(c.input);
    } catch (err) {
      const res: Response<unknown> = {
        ok: false,
        error: asError(err),
      };
      adapter.write(`data: ${JSON.stringify(res)}\n\n`);
      if (adapter.flush) adapter.flush();
      adapter.end();
      return;
    }

    let closed = false;
    adapter.onClose(() => {
      closed = true;
    });

    const safeWrite = (data: string) => {
      if (closed || c.signal.aborted) return;
      adapter.write(data);
      if (adapter.flush) adapter.flush();
    };

    // Ping Loop
    const pingTimer = setInterval(() => {
      if (closed || c.signal.aborted) {
        clearInterval(pingTimer);
        return;
      }
      safeWrite(": ping\n\n");
    }, pingInterval);

    // Cleanup on finish
    const cleanup = () => {
      clearInterval(pingTimer);
      closed = true;
      adapter.end();
    };

    // Base Emit
    const baseEmit: EmitFunc<T, any, unknown> = async (ctx, output) => {
      const res: Response<unknown> = { ok: true, output };
      safeWrite(`data: ${JSON.stringify(res)}\n\n`);
    };

    // Compose Emit Chain
    let emitFinal = baseEmit;
    if (emitMws && emitMws.length > 0) {
      for (let i = emitMws.length - 1; i >= 0; i--) {
        emitFinal = emitMws[i](emitFinal);
      }
    }

    // Compose Handler Chain (Stream MWs + RPC/Global MWs)
    // Stream MWs wrap the stream handler signature: (c, emit) => void
    let finalHandler = baseHandler;
    if (streamMws && streamMws.length > 0) {
      for (let i = streamMws.length - 1; i >= 0; i--) {
        finalHandler = streamMws[i](finalHandler);
      }
    }

    // RPC/Global MWs wrap (c) => Promise<any>.
    // We adapt the stream handler to this signature.
    // The RPC/Global MWs are executed *before* the stream handler starts.
    // However, in Go, the Global MWs wrap the *initiation* of the stream.
    // "exec := func(c) { return nil, final(c, emitFinal) }"

    let exec: GlobalHandlerFunc<T> = async (ctx) => {
      await finalHandler(ctx as HandlerContext<T, any>, emitFinal);
      return null;
    };

    if (rpcMws && rpcMws.length > 0) {
      for (let i = rpcMws.length - 1; i >= 0; i--) {
        exec = rpcMws[i](exec);
      }
    }

    if (this.globalMiddlewares.length > 0) {
      for (let i = this.globalMiddlewares.length - 1; i >= 0; i--) {
        exec = this.globalMiddlewares[i](exec);
      }
    }

    // Execute
    try {
      await exec(c);
    } catch (err) {
      if (!closed && !c.signal.aborted) {
        const errorHandler = this.resolveErrorHandler(rpcName);
        const res: Response<unknown> = {
          ok: false,
          error: errorHandler(c, err),
        };
        safeWrite(`data: ${JSON.stringify(res)}\n\n`);
      }
    } finally {
      cleanup();
    }
  }

  /**
   * Helper to write a standard JSON response for procedures.
   */
  private writeProcResponse(adapter: HTTPAdapter, response: Response<unknown>) {
    adapter.setHeader("Content-Type", "application/json");
    adapter.write(JSON.stringify(response));
    adapter.end();
  }
}

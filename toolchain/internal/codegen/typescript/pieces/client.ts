// This imports are just to prevent errors in the IDE when developing, this imports
// are handled in the generator for the generated code

import {
  Response,
  VdlError,
  asError,
  sleep,
  OperationType,
  OperationDefinition,
} from "./coreTypes";

/** START FROM HERE **/

// =============================================================================
// Configuration Types
// =============================================================================

/**
 * Controls automatic retry behavior for procedure calls.
 *
 * When a request fails due to network errors, timeouts, or 5xx HTTP status codes,
 * the client will automatically retry up to `maxAttempts` times with exponential backoff.
 *
 * @example
 * ```ts
 * const config: RetryConfig = {
 *   maxAttempts: 3,
 *   initialDelayMs: 100,
 *   maxDelayMs: 5000,
 *   delayMultiplier: 2.0,
 *   jitter: 0.2,
 * };
 * ```
 */
interface RetryConfig {
  /** How many times to attempt the request. Set to 1 for no retries. Default: 1. */
  maxAttempts: number;
  /** Delay before the first retry (in milliseconds). Default: 0. */
  initialDelayMs: number;
  /** Maximum delay between retries, caps exponential growth (in milliseconds). Default: 0. */
  maxDelayMs: number;
  /** Multiplier applied to delay after each retry. E.g., 2.0 doubles the delay. Default: 1.0. */
  delayMultiplier: number;
  /** Random variance added to delays to prevent thundering herd. 0.2 = ±20%. Default: 0.2. */
  jitter: number;
}

/**
 * Controls request timeout behavior.
 *
 * If the server doesn't respond within `timeoutMs`, the request is aborted
 * and may be retried according to {@link RetryConfig}.
 */
interface TimeoutConfig {
  /** Maximum time to wait for a response (in milliseconds). Default: 30000 (30s). */
  timeoutMs: number;
}

/**
 * Controls automatic reconnection behavior for streams.
 *
 * When a stream connection is lost due to network issues or server errors,
 * the client will automatically attempt to reconnect with exponential backoff.
 *
 * @example
 * ```ts
 * const config: ReconnectConfig = {
 *   maxAttempts: 30,
 *   initialDelayMs: 1000,
 *   maxDelayMs: 30000,
 *   delayMultiplier: 1.5,
 *   jitter: 0.2,
 * };
 * ```
 */
interface ReconnectConfig {
  /** How many reconnection attempts before giving up. Default: 30. */
  maxAttempts: number;
  /** Delay before the first reconnection attempt (in milliseconds). Default: 1000 (1s). */
  initialDelayMs: number;
  /** Maximum delay between reconnection attempts (in milliseconds). Default: 30000 (30s). */
  maxDelayMs: number;
  /** Multiplier applied to delay after each reconnection attempt. Default: 1.5. */
  delayMultiplier: number;
  /** Random variance added to delays to prevent thundering herd. 0.2 = ±20%. Default: 0.2. */
  jitter: number;
}

// =============================================================================
// Header Provider & Interceptor Types
// =============================================================================

/**
 * A function that adds or modifies headers before each request.
 *
 * Header providers are called in order: Global → RPC-level → Operation-level.
 * They run before every request attempt, including retries.
 *
 * @example
 * ```ts
 * // Add auth token dynamically
 * const authProvider: HeaderProvider = (headers) => {
 *   headers["Authorization"] = `Bearer ${getToken()}`;
 * };
 *
 * // Async provider (e.g., refresh token if expired)
 * const asyncProvider: HeaderProvider = async (headers) => {
 *   const token = await refreshTokenIfNeeded();
 *   headers["Authorization"] = `Bearer ${token}`;
 * };
 * ```
 *
 * @throws If the provider throws, the request is aborted immediately.
 */
type HeaderProvider = (headers: Record<string, string>) => void | Promise<void>;

/**
 * Metadata about the current RPC request, passed to interceptors.
 */
interface RequestInfo {
  /** The RPC group name (e.g., "users", "orders"). */
  rpcName: string;
  /** The operation name within the RPC (e.g., "getUser", "createOrder"). */
  operationName: string;
  /** The input payload being sent. */
  input: unknown;
  /** Whether this is a "proc" (request-response) or "stream" (SSE). */
  type: OperationType;
}

/**
 * The final function in the interceptor chain that performs the actual HTTP request.
 */
type Invoker = (req: RequestInfo) => Promise<Response<unknown>>;

/**
 * Middleware that wraps request execution.
 *
 * Interceptors can inspect/modify requests, handle responses, add logging,
 * implement caching, or add custom error handling.
 *
 * @example
 * ```ts
 * // Logging interceptor
 * const loggingInterceptor: Interceptor = async (req, next) => {
 *   console.log(`→ ${req.rpcName}.${req.operationName}`);
 *   const start = Date.now();
 *   const result = await next(req);
 *   console.log(`← ${req.rpcName}.${req.operationName} (${Date.now() - start}ms)`);
 *   return result;
 * };
 *
 * // Error handling interceptor
 * const errorInterceptor: Interceptor = async (req, next) => {
 *   const result = await next(req);
 *   if (!result.ok && result.error.code === "UNAUTHORIZED") {
 *     await refreshAuth();
 *     return next(req); // Retry with fresh auth
 *   }
 *   return result;
 * };
 * ```
 */
type Interceptor = (
  req: RequestInfo,
  next: Invoker,
) => Promise<Response<unknown>>;

// =============================================================================
// Fetch Interface
// =============================================================================

/**
 * Minimal interface for a fetch response.
 * Compatible with the standard Fetch API and most polyfills.
 */
interface FetchLikeResponse {
  ok: boolean;
  status: number;
  body?: ReadableStream<Uint8Array> | undefined | null;
  json(): Promise<any>;
  text(): Promise<string>;
}

/**
 * Minimal interface for a fetch function.
 * Compatible with the standard Fetch API, node-fetch, and most polyfills.
 */
type FetchLike = (input: any, init?: any) => Promise<FetchLikeResponse>;

// =============================================================================
// Default Configuration Values
// =============================================================================

/**
 * Default retry config: no retries (single attempt).
 * These values match the Go client defaults.
 */
const defaultRetryConfig: RetryConfig = {
  maxAttempts: 1,
  initialDelayMs: 0,
  maxDelayMs: 0,
  delayMultiplier: 1.0,
  jitter: 0.2,
};

/**
 * Default timeout: 30 seconds.
 */
const defaultTimeoutConfig: TimeoutConfig = {
  timeoutMs: 30000,
};

/**
 * Default reconnect config for streams.
 * Aggressive reconnection with up to 30 attempts over ~5 minutes.
 */
const defaultReconnectConfig: ReconnectConfig = {
  maxAttempts: 30,
  initialDelayMs: 1000,
  maxDelayMs: 30000,
  delayMultiplier: 1.5,
  jitter: 0.2,
};

/** Default maximum message size for streams: 4MB. */
const defaultMaxMessageSize = 4 * 1024 * 1024;

// =============================================================================
// Backoff Utilities
// =============================================================================

/**
 * Adds random jitter to a delay to prevent thundering herd.
 *
 * @param delayMs - Base delay in milliseconds.
 * @param jitterFactor - Variance factor (0.2 = ±20% of the delay).
 * @returns Delay with random jitter applied.
 *
 * @example
 * ```ts
 * applyJitter(1000, 0.2); // Returns 800-1200ms randomly
 * ```
 */
function applyJitter(delayMs: number, jitterFactor: number): number {
  if (jitterFactor <= 0) {
    return delayMs;
  }

  // Clamp to [0, 1] for safety
  jitterFactor = Math.min(jitterFactor, 1.0);

  // Calculate range: [delay * (1 - jitter), delay * (1 + jitter)]
  const delta = delayMs * jitterFactor;
  const min = Math.max(0, delayMs - delta);
  const max = delayMs + delta;

  return min + Math.random() * (max - min);
}

/**
 * Calculates exponential backoff delay for retry attempts.
 *
 * Formula: `min(initialDelay * multiplier^(attempt-1), maxDelay) ± jitter`
 *
 * @param config - Retry configuration.
 * @param attempt - Current attempt number (1-based).
 * @returns Delay in milliseconds before the next retry.
 */
function calculateBackoff(config: RetryConfig, attempt: number): number {
  let delay = config.initialDelayMs;

  // Apply multiplier for each previous attempt
  for (let i = 1; i < attempt; i++) {
    delay = delay * config.delayMultiplier;
  }

  // Cap at maximum delay
  if (delay > config.maxDelayMs) {
    delay = config.maxDelayMs;
  }

  return applyJitter(delay, config.jitter);
}

/**
 * Calculates exponential backoff delay for stream reconnection attempts.
 *
 * @param config - Reconnect configuration.
 * @param attempt - Current attempt number (1-based).
 * @returns Delay in milliseconds before the next reconnection.
 */
function calculateReconnectBackoff(
  config: ReconnectConfig,
  attempt: number,
): number {
  let delay = config.initialDelayMs;

  for (let i = 1; i < attempt; i++) {
    delay = delay * config.delayMultiplier;
  }

  if (delay > config.maxDelayMs) {
    delay = config.maxDelayMs;
  }

  return applyJitter(delay, config.jitter);
}

// =============================================================================
// Internal Client
// =============================================================================

/**
 * Core HTTP client engine used by generated VDL client wrappers.
 *
 * This class handles all the low-level HTTP communication, including:
 * - Request/response serialization
 * - Retry logic with exponential backoff
 * - Timeout handling
 * - Stream (SSE) connections with auto-reconnect
 * - Header providers and interceptors
 *
 * **Do not instantiate directly.** Use {@link clientBuilder} instead.
 *
 * @internal
 */
class internalClient {
  /** Base URL for all requests (trailing slashes are stripped). */
  private baseURL: string;

  /** The fetch implementation used for HTTP requests. */
  private fetchFn: FetchLike;

  /**
   * Registry of valid operations.
   * Structure: rpcName → operationName → OperationType
   * Used to validate requests and fail fast on typos.
   */
  private operationDefs: Map<string, Map<string, OperationType>>;

  // ---------------------------------------------------------------------------
  // Dynamic Components
  // ---------------------------------------------------------------------------

  /** Global header providers applied to all requests. */
  private headerProviders: HeaderProvider[] = [];

  /** Interceptors applied to all requests (in registration order). */
  private interceptors: Interceptor[] = [];

  /** RPC-specific header providers. Key: rpcName. */
  private rpcHeaderProviders: Map<string, HeaderProvider[]> = new Map();

  // ---------------------------------------------------------------------------
  // Global Default Configurations
  // ---------------------------------------------------------------------------

  private globalRetryConf: RetryConfig | null = null;
  private globalTimeoutConf: TimeoutConfig | null = null;
  private globalReconnectConf: ReconnectConfig | null = null;
  private globalMaxMessageSize: number = 0;

  // ---------------------------------------------------------------------------
  // Per-RPC Default Configurations
  // ---------------------------------------------------------------------------

  private rpcRetryConf: Map<string, RetryConfig> = new Map();
  private rpcTimeoutConf: Map<string, TimeoutConfig> = new Map();
  private rpcReconnectConf: Map<string, ReconnectConfig> = new Map();
  private rpcMaxMessageSize: Map<string, number> = new Map();

  // ---------------------------------------------------------------------------
  // Constructor
  // ---------------------------------------------------------------------------

  /**
   * Creates a new internal client.
   *
   * @param baseURL - Base URL for the VDL server.
   * @param procDefs - List of procedure definitions from the schema.
   * @param streamDefs - List of stream definitions from the schema.
   * @param opts - Configuration options to apply.
   *
   * @throws If required runtime dependencies are missing (AbortController, etc.).
   * @throws If no fetch implementation is available.
   */
  constructor(
    baseURL: string,
    procDefs: OperationDefinition[],
    streamDefs: OperationDefinition[],
    opts: internalClientOption[],
  ) {
    this.verifyRuntimeDeps();

    // Strip trailing slashes for consistent URL building
    this.baseURL = baseURL.replace(/\/+$/, "");

    // Build operation registry for validation
    this.operationDefs = new Map();
    const ensureRPC = (rpcName: string) => {
      if (!this.operationDefs.has(rpcName)) {
        this.operationDefs.set(rpcName, new Map());
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

    // Default to globalThis.fetch if available
    this.fetchFn = (globalThis.fetch ?? null) as FetchLike;

    // Apply all configuration options
    opts.forEach((optFn) => optFn(this));

    if (!this.fetchFn) {
      throw new Error(
        "globalThis.fetch is undefined - please supply a custom fetch using withFetch()",
      );
    }
  }

  /**
   * Verifies that required browser/runtime APIs are available.
   * @throws If any required API is missing.
   */
  private verifyRuntimeDeps() {
    const missing: string[] = [];

    if (typeof AbortController !== "function") {
      missing.push("AbortController");
    }
    if (typeof ReadableStream === "undefined") {
      missing.push("ReadableStream");
    }
    if (typeof TextDecoder !== "function") {
      missing.push("TextDecoder");
    }

    if (missing.length > 0) {
      throw new Error(
        `Missing required runtime dependencies: ${missing.join(", ")}. ` +
          `Install the necessary polyfills or use a compatible environment.`,
      );
    }
  }

  // ---------------------------------------------------------------------------
  // Configuration Setters (called by builder options)
  // ---------------------------------------------------------------------------

  /** Sets a custom fetch implementation. */
  setFetch(fetchFn: FetchLike) {
    this.fetchFn = fetchFn;
  }

  /** Adds a global header provider. */
  addHeaderProvider(provider: HeaderProvider) {
    this.headerProviders.push(provider);
  }

  /** Adds an interceptor to the chain. */
  addInterceptor(interceptor: Interceptor) {
    this.interceptors.push(interceptor);
  }

  /** Sets the global default retry configuration. */
  setGlobalRetryConfig(conf: RetryConfig) {
    this.globalRetryConf = conf;
  }

  /** Sets the global default timeout configuration. */
  setGlobalTimeoutConfig(conf: TimeoutConfig) {
    this.globalTimeoutConf = conf;
  }

  /** Sets the global default reconnect configuration for streams. */
  setGlobalReconnectConfig(conf: ReconnectConfig) {
    this.globalReconnectConf = conf;
  }

  /** Sets the global default maximum message size for streams. */
  setGlobalMaxMessageSize(size: number) {
    this.globalMaxMessageSize = size;
  }

  // ---------------------------------------------------------------------------
  // Per-RPC Configuration Setters
  // ---------------------------------------------------------------------------

  /** Sets the default retry config for a specific RPC. */
  setRPCRetryConfig(rpcName: string, conf: RetryConfig) {
    this.rpcRetryConf.set(rpcName, conf);
  }

  /** Sets the default timeout config for a specific RPC. */
  setRPCTimeoutConfig(rpcName: string, conf: TimeoutConfig) {
    this.rpcTimeoutConf.set(rpcName, conf);
  }

  /** Sets the default reconnect config for a specific RPC. */
  setRPCReconnectConfig(rpcName: string, conf: ReconnectConfig) {
    this.rpcReconnectConf.set(rpcName, conf);
  }

  /** Sets the default max message size for a specific RPC. */
  setRPCMaxMessageSize(rpcName: string, size: number) {
    this.rpcMaxMessageSize.set(rpcName, size);
  }

  /** Adds a header provider for a specific RPC. */
  setRPCHeaderProvider(rpcName: string, provider: HeaderProvider) {
    if (!this.rpcHeaderProviders.has(rpcName)) {
      this.rpcHeaderProviders.set(rpcName, []);
    }
    this.rpcHeaderProviders.get(rpcName)!.push(provider);
  }

  // ---------------------------------------------------------------------------
  // Configuration Merging
  // Priority: Operation-level > RPC-level > Global > Built-in defaults
  // ---------------------------------------------------------------------------

  /**
   * Resolves the effective retry configuration for a request.
   */
  private mergeRetryConfig(
    rpcName: string,
    opConf: RetryConfig | undefined,
  ): RetryConfig {
    if (opConf) return opConf;
    if (this.rpcRetryConf.has(rpcName)) return this.rpcRetryConf.get(rpcName)!;
    if (this.globalRetryConf) return this.globalRetryConf;
    return defaultRetryConfig;
  }

  /**
   * Resolves the effective timeout configuration for a request.
   */
  private mergeTimeoutConfig(
    rpcName: string,
    opConf: TimeoutConfig | undefined,
  ): TimeoutConfig {
    if (opConf) return opConf;
    if (this.rpcTimeoutConf.has(rpcName))
      return this.rpcTimeoutConf.get(rpcName)!;
    if (this.globalTimeoutConf) return this.globalTimeoutConf;
    return defaultTimeoutConfig;
  }

  /**
   * Resolves the effective reconnect configuration for a stream.
   */
  private mergeReconnectConfig(
    rpcName: string,
    opConf: ReconnectConfig | undefined,
  ): ReconnectConfig {
    if (opConf) return opConf;
    if (this.rpcReconnectConf.has(rpcName))
      return this.rpcReconnectConf.get(rpcName)!;
    if (this.globalReconnectConf) return this.globalReconnectConf;
    return defaultReconnectConfig;
  }

  /**
   * Resolves the effective max message size for a stream.
   */
  private mergeMaxMessageSize(rpcName: string, opSize: number): number {
    if (opSize > 0) return opSize;

    if (this.rpcMaxMessageSize.has(rpcName)) {
      const rpcSize = this.rpcMaxMessageSize.get(rpcName)!;
      if (rpcSize > 0) return rpcSize;
    }

    if (this.globalMaxMessageSize > 0) return this.globalMaxMessageSize;

    return defaultMaxMessageSize;
  }

  // ---------------------------------------------------------------------------
  // Interceptor Chain Execution
  // ---------------------------------------------------------------------------

  /**
   * Builds and executes the interceptor chain.
   *
   * Interceptors are wrapped in reverse order so they execute in registration order.
   * The chain is: interceptor[0] → interceptor[1] → ... → invoker
   */
  private async executeChain(
    req: RequestInfo,
    final: Invoker,
  ): Promise<Response<unknown>> {
    let chain = final;

    // Wrap interceptors in reverse order
    for (let i = this.interceptors.length - 1; i >= 0; i--) {
      const mw = this.interceptors[i];
      const next = chain;
      chain = (req: RequestInfo) => mw(req, next);
    }

    return chain(req);
  }

  // ---------------------------------------------------------------------------
  // Header Application
  // ---------------------------------------------------------------------------

  /**
   * Applies all header providers to a request's headers.
   *
   * Order: Global providers → RPC providers → Operation providers
   *
   * @throws If any provider throws, propagates the error.
   */
  private async applyHeaders(
    rpcName: string,
    headers: Record<string, string>,
    opHeaderProviders: HeaderProvider[],
  ): Promise<void> {
    // 1. Global providers
    for (const provider of this.headerProviders) {
      await provider(headers);
    }

    // 2. RPC-specific providers
    const rpcProviders = this.rpcHeaderProviders.get(rpcName) ?? [];
    for (const provider of rpcProviders) {
      await provider(headers);
    }

    // 3. Operation-specific providers
    for (const provider of opHeaderProviders) {
      await provider(headers);
    }
  }

  // ---------------------------------------------------------------------------
  // Procedure Call (Request-Response)
  // ---------------------------------------------------------------------------

  /**
   * Executes a procedure call with automatic retry and timeout handling.
   *
   * @param rpcName - The RPC group name.
   * @param procName - The procedure name.
   * @param input - The request payload.
   * @param opHeaderProviders - Operation-specific header providers.
   * @param opRetryConf - Operation-specific retry configuration (optional).
   * @param opTimeoutConf - Operation-specific timeout configuration (optional).
   * @returns The response from the server.
   */
  async callProc(
    rpcName: string,
    procName: string,
    input: unknown,
    opHeaderProviders: HeaderProvider[],
    opRetryConf?: RetryConfig,
    opTimeoutConf?: TimeoutConfig,
  ): Promise<Response<any>> {
    const reqInfo: RequestInfo = {
      rpcName,
      operationName: procName,
      input,
      type: "proc",
    };

    /**
     * The invoker contains the actual HTTP request logic.
     * It's wrapped by interceptors before execution.
     */
    const invoker: Invoker = async (req) => {
      // Validate that the operation exists in the schema
      const rpcOps = this.operationDefs.get(req.rpcName);
      if (!rpcOps || !rpcOps.has(req.operationName)) {
        return {
          ok: false,
          error: new VdlError({
            message: `${req.rpcName}.${req.operationName} procedure not found in schema`,
            category: "ClientError",
            code: "INVALID_PROC",
            details: { rpc: req.rpcName, procedure: req.operationName },
          }),
        };
      }

      // Resolve effective configurations
      const retryConf = this.mergeRetryConfig(req.rpcName, opRetryConf);
      const timeoutConf = this.mergeTimeoutConfig(req.rpcName, opTimeoutConf);

      // Serialize input to JSON
      let payload: string;
      try {
        payload = req.input == null ? "{}" : JSON.stringify(req.input);
      } catch (err) {
        return {
          ok: false,
          error: new VdlError({
            message: `failed to marshal input for ${req.rpcName}.${req.operationName}: ${err}`,
            category: "ClientError",
            code: "ENCODE_INPUT",
          }),
        };
      }

      const url = `${this.baseURL}/${req.rpcName}/${req.operationName}`;
      let lastError: VdlError | null = null;

      // Retry loop
      for (let attempt = 1; attempt <= retryConf.maxAttempts; attempt++) {
        const abortController = new AbortController();
        let timeoutId: ReturnType<typeof setTimeout> | undefined;

        try {
          // Set up request timeout
          if (timeoutConf.timeoutMs > 0) {
            timeoutId = setTimeout(() => {
              abortController.abort();
            }, timeoutConf.timeoutMs);
          }

          // Build request headers
          const hdrs: Record<string, string> = {
            "content-type": "application/json",
            accept: "application/json",
          };

          // Apply header providers (may throw)
          try {
            await this.applyHeaders(req.rpcName, hdrs, opHeaderProviders);
          } catch (err) {
            if (timeoutId !== undefined) clearTimeout(timeoutId);
            return { ok: false, error: asError(err) };
          }

          // Execute the HTTP request
          const fetchResp = await this.fetchFn(url, {
            method: "POST",
            headers: hdrs,
            body: payload,
            signal: abortController.signal,
          });

          if (timeoutId !== undefined) clearTimeout(timeoutId);

          // Server error (5xx): retry if attempts remaining
          if (fetchResp.status >= 500) {
            const error = new VdlError({
              message: `unexpected HTTP status: ${fetchResp.status}`,
              category: "HTTPError",
              code: "BAD_STATUS",
              details: { status: fetchResp.status },
            });

            if (attempt < retryConf.maxAttempts) {
              lastError = error;
              await sleep(calculateBackoff(retryConf, attempt));
              continue;
            }

            return { ok: false, error };
          }

          // Client error (4xx) or other non-OK: don't retry
          if (!fetchResp.ok) {
            return {
              ok: false,
              error: new VdlError({
                message: `unexpected HTTP status: ${fetchResp.status}`,
                category: "HTTPError",
                code: "BAD_STATUS",
                details: { status: fetchResp.status },
              }),
            };
          }

          // Parse and return the response
          try {
            return await fetchResp.json();
          } catch (parseErr) {
            return {
              ok: false,
              error: new VdlError({
                message: `failed to decode VDL response: ${parseErr}`,
                category: "ClientError",
                code: "DECODE_RESPONSE",
              }),
            };
          }
        } catch (err) {
          if (timeoutId !== undefined) clearTimeout(timeoutId);

          // Timeout error: retry if attempts remaining
          if (abortController.signal.aborted && timeoutConf.timeoutMs > 0) {
            const timeoutError = new VdlError({
              message: "Request timeout",
              category: "TimeoutError",
              code: "REQUEST_TIMEOUT",
              details: { attempt },
            });

            if (attempt < retryConf.maxAttempts) {
              lastError = timeoutError;
              await sleep(calculateBackoff(retryConf, attempt));
              continue;
            }

            return { ok: false, error: timeoutError };
          }

          // Network error: retry if attempts remaining
          const error = asError(err);
          if (attempt < retryConf.maxAttempts) {
            lastError = error;
            await sleep(calculateBackoff(retryConf, attempt));
            continue;
          }

          return { ok: false, error };
        }
      }

      // All retries exhausted
      return {
        ok: false,
        error:
          lastError ||
          new VdlError({
            message: "Unknown error",
            category: "ClientError",
            code: "UNKNOWN",
          }),
      };
    };

    // Execute through the interceptor chain
    return this.executeChain(reqInfo, invoker) as Promise<Response<any>>;
  }

  // ---------------------------------------------------------------------------
  // Stream Call (Server-Sent Events)
  // ---------------------------------------------------------------------------

  /**
   * Opens a streaming connection with automatic reconnection.
   *
   * Uses Server-Sent Events (SSE) protocol. The stream will automatically
   * attempt to reconnect on connection failures.
   *
   * @param rpcName - The RPC group name.
   * @param streamName - The stream name.
   * @param input - The request payload.
   * @param opHeaderProviders - Operation-specific header providers.
   * @param opReconnectConf - Operation-specific reconnect configuration.
   * @param opMaxMessageSize - Maximum allowed message size in bytes.
   * @param onConnect - Called when the stream successfully connects.
   * @param onDisconnect - Called when the stream permanently disconnects.
   * @param onReconnect - Called before each reconnection attempt.
   * @returns An object with the stream generator and a cancel function.
   */
  callStream(
    rpcName: string,
    streamName: string,
    input: unknown,
    opHeaderProviders: HeaderProvider[],
    opReconnectConf?: ReconnectConfig,
    opMaxMessageSize?: number,
    onConnect?: () => void,
    onDisconnect?: (error: Error | null) => void,
    onReconnect?: (attempt: number, delayMs: number) => void,
  ): {
    stream: AsyncGenerator<Response<any>, void, unknown>;
    cancel: () => void;
  } {
    const self = this;
    let isCancelled = false;
    let currentAbortController: AbortController | null = null;

    /** Cancels the stream and aborts any pending request. */
    const cancel = () => {
      isCancelled = true;
      currentAbortController?.abort();
    };

    const reqInfo: RequestInfo = {
      rpcName,
      operationName: streamName,
      input,
      type: "stream",
    };

    /**
     * The main stream generator that handles SSE parsing and reconnection.
     */
    async function* generator(): AsyncGenerator<Response<any>, void, unknown> {
      let finalError: Error | null = null;

      try {
        // Validate that the stream exists in the schema
        const rpcOps = self.operationDefs.get(rpcName);
        if (!rpcOps || !rpcOps.has(streamName)) {
          yield {
            ok: false,
            error: new VdlError({
              message: `${rpcName}.${streamName} stream not found in schema`,
              category: "ClientError",
              code: "INVALID_STREAM",
              details: { rpc: rpcName, stream: streamName },
            }),
          };
          return;
        }

        // Resolve effective configurations
        const reconnectConf = self.mergeReconnectConfig(
          rpcName,
          opReconnectConf,
        );
        const maxMessageSize = self.mergeMaxMessageSize(
          rpcName,
          opMaxMessageSize ?? 0,
        );

        // Serialize input to JSON
        let payload: string;
        try {
          payload = input == null ? "{}" : JSON.stringify(input);
        } catch (err) {
          finalError = err as Error;
          yield { ok: false, error: asError(err) };
          return;
        }

        const url = `${self.baseURL}/${rpcName}/${streamName}`;
        let reconnectAttempt = 0;

        // Main connection loop (handles reconnection)
        while (!isCancelled) {
          currentAbortController = new AbortController();

          try {
            // Build request headers
            const hdrs: Record<string, string> = {
              "content-type": "application/json",
              accept: "text/event-stream",
            };

            // Apply header providers (may throw)
            try {
              await self.applyHeaders(rpcName, hdrs, opHeaderProviders);
            } catch (err) {
              finalError = err as Error;
              yield { ok: false, error: asError(err) };
              return;
            }

            // Execute the HTTP request
            const fetchResp = await self.fetchFn(url, {
              method: "POST",
              headers: hdrs,
              body: payload,
              signal: currentAbortController.signal,
            });

            // Server error (5xx): attempt reconnection
            if (
              fetchResp.status >= 500 &&
              reconnectAttempt < reconnectConf.maxAttempts
            ) {
              const delayMs = calculateReconnectBackoff(
                reconnectConf,
                reconnectAttempt + 1,
              );
              if (onReconnect) onReconnect(reconnectAttempt + 1, delayMs);
              reconnectAttempt++;
              await sleep(delayMs);
              continue;
            }

            // Non-OK response: report error and stop
            if (!fetchResp.ok) {
              const error = new VdlError({
                message: `unexpected HTTP status: ${fetchResp.status}`,
                category: "HTTPError",
                code: "BAD_STATUS",
                details: { status: fetchResp.status },
              });
              finalError = error;
              yield { ok: false, error };
              return;
            }

            // No response body: cannot stream
            if (!fetchResp.body) {
              const error = new VdlError({
                message: "Missing response body for stream",
                category: "ConnectionError",
                code: "STREAM_CONNECT_FAILED",
              });
              finalError = error;
              yield { ok: false, error };
              return;
            }

            // Successfully connected
            if (onConnect) onConnect();
            reconnectAttempt = 0;

            // Set up SSE parsing
            const reader = fetchResp.body.getReader();
            const decoder = new TextDecoder();
            let buffer = "";

            try {
              // Read and parse SSE events
              while (!isCancelled) {
                const { done, value } = await reader.read();
                if (done) break;

                buffer += decoder.decode(value, { stream: true });

                // SSE events are separated by double newlines
                let idx: number;
                while ((idx = buffer.indexOf("\n\n")) >= 0) {
                  const line = buffer.slice(0, idx).trim();
                  buffer = buffer.slice(idx + 2);

                  // Skip empty lines
                  if (line === "") continue;

                  // Skip SSE comments (lines starting with :)
                  if (line.startsWith(":")) continue;

                  // Parse data lines
                  if (line.startsWith("data:")) {
                    const jsonStr = line.slice(5).trim();

                    // Check message size limit
                    if (jsonStr.length > maxMessageSize) {
                      yield {
                        ok: false,
                        error: new VdlError({
                          message: `Stream message exceeded maximum size of ${maxMessageSize} bytes`,
                          category: "ProtocolError",
                          code: "MESSAGE_TOO_LARGE",
                        }),
                      };
                      return; // Fatal error, no reconnect
                    }

                    // Parse and yield the event
                    try {
                      const evt = JSON.parse(jsonStr) as Response<any>;
                      yield evt;
                    } catch (err) {
                      yield {
                        ok: false,
                        error: new VdlError({
                          message: `received invalid SSE payload: ${err}`,
                          category: "ProtocolError",
                          code: "INVALID_PAYLOAD",
                        }),
                      };
                      return; // Fatal error, no reconnect
                    }
                  }
                }

                // Check buffer size to prevent memory exhaustion
                if (buffer.length > maxMessageSize) {
                  yield {
                    ok: false,
                    error: new VdlError({
                      message: `Stream message accumulation exceeded maximum size of ${maxMessageSize} bytes`,
                      category: "ProtocolError",
                      code: "MESSAGE_TOO_LARGE",
                    }),
                  };
                  return; // Fatal error, no reconnect
                }
              }

              // Stream ended normally
              if (!isCancelled) return;
            } catch (readError) {
              // Connection interrupted during read: attempt reconnection
              if (
                !isCancelled &&
                reconnectAttempt < reconnectConf.maxAttempts
              ) {
                const delayMs = calculateReconnectBackoff(
                  reconnectConf,
                  reconnectAttempt + 1,
                );
                if (onReconnect) onReconnect(reconnectAttempt + 1, delayMs);
                reconnectAttempt++;
                await sleep(delayMs);
                continue;
              }

              // Max reconnection attempts reached
              if (!isCancelled) {
                finalError = readError as Error;
                yield { ok: false, error: asError(readError) };
              }
              return;
            }
          } catch (fetchError) {
            // Connection failed: attempt reconnection
            if (!isCancelled && reconnectAttempt < reconnectConf.maxAttempts) {
              const delayMs = calculateReconnectBackoff(
                reconnectConf,
                reconnectAttempt + 1,
              );
              if (onReconnect) onReconnect(reconnectAttempt + 1, delayMs);
              reconnectAttempt++;
              await sleep(delayMs);
              continue;
            }

            // Max reconnection attempts reached
            if (!isCancelled) {
              finalError = fetchError as Error;
              yield { ok: false, error: asError(fetchError) };
            }
            return;
          }
        }
      } finally {
        // Always call onDisconnect when the stream ends
        if (onDisconnect) onDisconnect(finalError);
      }
    }

    /**
     * Wraps the generator to run through the interceptor chain.
     *
     * For streams, interceptors only wrap the initial connection.
     * They don't wrap individual events.
     */
    const wrappedGenerator = async function* () {
      const invoker: Invoker = async () => {
        // Streams use a different flow - just return success to pass interceptors
        return { ok: true, output: null };
      };

      // Run interceptor chain (for validation, logging, etc.)
      const result = await self.executeChain(reqInfo, invoker);
      if (!result.ok) {
        yield result as Response<any>;
        return;
      }

      // Yield all events from the actual stream generator
      yield* generator();
    };

    return { stream: wrappedGenerator(), cancel };
  }
}

// =============================================================================
// Builder Option Functions
// =============================================================================

/** A function that configures an internalClient. */
type internalClientOption = (c: internalClient) => void;

/** Sets a custom fetch implementation. */
function withFetch(fetchFn: FetchLike): internalClientOption {
  return (c) => c.setFetch(fetchFn);
}

/** Adds a static global header to all requests. */
function withGlobalHeader(key: string, value: string): internalClientOption {
  return (c) =>
    c.addHeaderProvider((headers) => {
      headers[key] = value;
    });
}

/** Adds a dynamic header provider to all requests. */
function withHeaderProvider(provider: HeaderProvider): internalClientOption {
  return (c) => c.addHeaderProvider(provider);
}

/** Adds an interceptor to the request chain. */
function withInterceptor(interceptor: Interceptor): internalClientOption {
  return (c) => c.addInterceptor(interceptor);
}

/** Sets the global default retry configuration. */
function withGlobalRetryConfig(conf: RetryConfig): internalClientOption {
  return (c) => c.setGlobalRetryConfig(conf);
}

/** Sets the global default timeout configuration. */
function withGlobalTimeoutConfig(conf: TimeoutConfig): internalClientOption {
  return (c) => c.setGlobalTimeoutConfig(conf);
}

/** Sets the global default reconnect configuration for streams. */
function withGlobalReconnectConfig(
  conf: ReconnectConfig,
): internalClientOption {
  return (c) => c.setGlobalReconnectConfig(conf);
}

/** Sets the global default maximum message size for streams. */
function withGlobalMaxMessageSize(size: number): internalClientOption {
  return (c) => c.setGlobalMaxMessageSize(size);
}

// =============================================================================
// Client Builder
// =============================================================================

/**
 * Fluent builder for creating a VDL client.
 *
 * Used by generated code to construct clients with the proper schema information.
 *
 * @example
 * ```ts
 * const client = new clientBuilder(baseURL, procDefs, streamDefs)
 *   .withGlobalHeader("X-API-Key", apiKey)
 *   .withGlobalRetryConfig({ maxAttempts: 3, ... })
 *   .withInterceptor(loggingInterceptor)
 *   .build();
 * ```
 */
class clientBuilder {
  private baseURL: string;
  private procDefs: OperationDefinition[] = [];
  private streamDefs: OperationDefinition[] = [];
  private opts: internalClientOption[] = [];

  /**
   * Creates a new client builder.
   *
   * @param baseURL - Base URL for the VDL server.
   * @param procDefs - Procedure definitions from the schema.
   * @param streamDefs - Stream definitions from the schema.
   */
  constructor(
    baseURL: string,
    procDefs: OperationDefinition[],
    streamDefs: OperationDefinition[],
  ) {
    this.baseURL = baseURL;
    this.procDefs = procDefs;
    this.streamDefs = streamDefs;
  }

  /** Sets a custom fetch implementation. */
  withFetch(fetchFn: FetchLike): clientBuilder {
    this.opts.push(withFetch(fetchFn));
    return this;
  }

  /** Adds a static header to all requests. */
  withGlobalHeader(key: string, value: string): clientBuilder {
    this.opts.push(withGlobalHeader(key, value));
    return this;
  }

  /** Adds a dynamic header provider to all requests. */
  withHeaderProvider(provider: HeaderProvider): clientBuilder {
    this.opts.push(withHeaderProvider(provider));
    return this;
  }

  /** Adds an interceptor to the request chain. */
  withInterceptor(interceptor: Interceptor): clientBuilder {
    this.opts.push(withInterceptor(interceptor));
    return this;
  }

  /** Sets the global default retry configuration. */
  withGlobalRetryConfig(conf: RetryConfig): clientBuilder {
    this.opts.push(withGlobalRetryConfig(conf));
    return this;
  }

  /** Sets the global default timeout configuration. */
  withGlobalTimeoutConfig(conf: TimeoutConfig): clientBuilder {
    this.opts.push(withGlobalTimeoutConfig(conf));
    return this;
  }

  /** Sets the global default reconnect configuration for streams. */
  withGlobalReconnectConfig(conf: ReconnectConfig): clientBuilder {
    this.opts.push(withGlobalReconnectConfig(conf));
    return this;
  }

  /** Sets the global default maximum message size for streams. */
  withGlobalMaxMessageSize(size: number): clientBuilder {
    this.opts.push(withGlobalMaxMessageSize(size));
    return this;
  }

  /** Builds and returns the configured client. */
  build(): internalClient {
    return new internalClient(
      this.baseURL,
      this.procDefs,
      this.streamDefs,
      this.opts,
    );
  }
}

// =============================================================================
// Procedure Call Builder
// =============================================================================

/**
 * Fluent builder for configuring a single procedure call.
 *
 * Allows per-call customization of headers, retry, and timeout settings.
 *
 * @example
 * ```ts
 * const response = await client.someRpc.someProc({ id: 123 })
 *   .withHeader("X-Request-Id", requestId)
 *   .withRetryConfig({ maxAttempts: 5, ... })
 *   .withTimeoutConfig({ timeoutMs: 60000 })
 *   .execute();
 * ```
 */
class procCallBuilder {
  private client: internalClient;
  private rpcName: string;
  private name: string;
  private input: unknown;
  private headerProviders: HeaderProvider[] = [];
  private retryConf?: RetryConfig;
  private timeoutConf?: TimeoutConfig;

  constructor(
    client: internalClient,
    rpcName: string,
    name: string,
    input: unknown,
  ) {
    this.client = client;
    this.rpcName = rpcName;
    this.name = name;
    this.input = input;
  }

  /** Adds a static header to this request. */
  withHeader(key: string, value: string): procCallBuilder {
    this.headerProviders.push((headers) => {
      headers[key] = value;
    });
    return this;
  }

  /** Adds a dynamic header provider to this request. */
  withHeaderProvider(provider: HeaderProvider): procCallBuilder {
    this.headerProviders.push(provider);
    return this;
  }

  /** Sets the retry configuration for this request. */
  withRetryConfig(retryConfig: RetryConfig): procCallBuilder {
    this.retryConf = retryConfig;
    return this;
  }

  /** Sets the timeout configuration for this request. */
  withTimeoutConfig(timeoutConfig: TimeoutConfig): procCallBuilder {
    this.timeoutConf = timeoutConfig;
    return this;
  }

  /** Executes the procedure call and returns the response. */
  execute(): Promise<Response<any>> {
    return this.client.callProc(
      this.rpcName,
      this.name,
      this.input,
      this.headerProviders,
      this.retryConf,
      this.timeoutConf,
    );
  }
}

// =============================================================================
// Stream Call Builder
// =============================================================================

/**
 * Fluent builder for configuring a single stream connection.
 *
 * Allows per-stream customization of headers, reconnection, and callbacks.
 *
 * @example
 * ```ts
 * const { stream, cancel } = client.someRpc.someStream({ filter: "all" })
 *   .withHeader("X-Request-Id", requestId)
 *   .withReconnectConfig({ maxAttempts: 10, ... })
 *   .withOnConnect(() => console.log("Connected!"))
 *   .withOnDisconnect((err) => console.log("Disconnected:", err))
 *   .execute();
 *
 * for await (const event of stream) {
 *   console.log(event);
 * }
 * ```
 */
class streamCallBuilder {
  private client: internalClient;
  private rpcName: string;
  private name: string;
  private input: unknown;
  private headerProviders: HeaderProvider[] = [];
  private reconnectConf?: ReconnectConfig;
  private maxMessageSize: number = 0;
  private onConnectCb?: () => void;
  private onDisconnectCb?: (error: Error | null) => void;
  private onReconnectCb?: (attempt: number, delayMs: number) => void;

  constructor(
    client: internalClient,
    rpcName: string,
    name: string,
    input: unknown,
  ) {
    this.client = client;
    this.rpcName = rpcName;
    this.name = name;
    this.input = input;
  }

  /** Adds a static header to this stream request. */
  withHeader(key: string, value: string): streamCallBuilder {
    this.headerProviders.push((headers) => {
      headers[key] = value;
    });
    return this;
  }

  /** Adds a dynamic header provider to this stream request. */
  withHeaderProvider(provider: HeaderProvider): streamCallBuilder {
    this.headerProviders.push(provider);
    return this;
  }

  /** Sets the reconnection configuration for this stream. */
  withReconnectConfig(reconnectConfig: ReconnectConfig): streamCallBuilder {
    this.reconnectConf = reconnectConfig;
    return this;
  }

  /** Sets the maximum allowed message size for this stream. */
  withMaxMessageSize(size: number): streamCallBuilder {
    this.maxMessageSize = size;
    return this;
  }

  /** Sets a callback invoked when the stream successfully connects. */
  withOnConnect(cb: () => void): streamCallBuilder {
    this.onConnectCb = cb;
    return this;
  }

  /** Sets a callback invoked when the stream permanently disconnects. */
  withOnDisconnect(cb: (error: Error | null) => void): streamCallBuilder {
    this.onDisconnectCb = cb;
    return this;
  }

  /** Sets a callback invoked before each reconnection attempt. */
  withOnReconnect(
    cb: (attempt: number, delayMs: number) => void,
  ): streamCallBuilder {
    this.onReconnectCb = cb;
    return this;
  }

  /** Opens the stream and returns the event generator and cancel function. */
  execute(): {
    stream: AsyncGenerator<Response<any>, void, unknown>;
    cancel: () => void;
  } {
    return this.client.callStream(
      this.rpcName,
      this.name,
      this.input,
      this.headerProviders,
      this.reconnectConf,
      this.maxMessageSize,
      this.onConnectCb,
      this.onDisconnectCb,
      this.onReconnectCb,
    );
  }
}

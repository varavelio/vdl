import { Response, UfoError } from "./core_types";

/**
 * Mocks for the parts that are generated but not exported
 */

function asError(err: unknown): UfoError {
  return err as UfoError;
}

function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

/** START FROM HERE **/

// -----------------------------------------------------------------------------
// Configuration Types
// -----------------------------------------------------------------------------

/**
 * Configuration for automatic retry behavior in procedures.
 */
interface RetryConfig {
  /** Maximum number of retry attempts (default: 3) */
  maxAttempts: number;
  /** Initial delay between retries in milliseconds (default: 1000) */
  initialDelayMs: number;
  /** Maximum delay between retries in milliseconds (default: 5000) */
  maxDelayMs: number;
  /** Multiplier for exponential backoff (default: 2.0) */
  delayMultiplier: number;
}

/**
 * Configuration for timeout behavior in procedures.
 */
interface TimeoutConfig {
  /** Timeout for each individual attempt in milliseconds (default: 30000) */
  timeoutMs: number;
}

/**
 * Configuration for automatic reconnection behavior in streams.
 */
interface ReconnectConfig {
  /** Maximum number of reconnection attempts (default: 5) */
  maxAttempts: number;
  /** Initial delay between reconnection attempts in milliseconds (default: 1000) */
  initialDelayMs: number;
  /** Maximum delay between reconnection attempts in milliseconds (default: 5000) */
  maxDelayMs: number;
  /** Multiplier for exponential backoff (default: 2.0) */
  delayMultiplier: number;
}

// -----------------------------------------------------------------------------
// Internal Client
// -----------------------------------------------------------------------------

/**
 * FetchLikeResponse is a minimal interface that a fetch response implementation must
 * satisfy.
 */
interface FetchLikeResponse {
  ok: boolean;
  status: number;
  body?: ReadableStream<Uint8Array> | undefined | null;
  json(): Promise<any>;
  text(): Promise<string>;
}

/**
 * FetchLike is a minimal interface that a fetch implementation must satisfy.
 */
type FetchLike = (input: any, init?: any) => Promise<FetchLikeResponse>;

/**
 * internalClient is the engine used by the generated façade. All identifiers
 * are deliberately un-exported because user code should interact only with the
 * generated wrappers.
 */
class internalClient {
  private baseURL: string;
  private fetchFn: FetchLike;
  private globalHeaders: Record<string, string> = {};
  private procSet: Set<string>;
  private streamSet: Set<string>;

  constructor(
    baseURL: string,
    procNames: string[],
    streamNames: string[],
    opts: internalClientOption[],
  ) {
    this.verifyRuntimeDeps();

    this.baseURL = baseURL.replace(/\/+$/, "");
    this.procSet = new Set(procNames);
    this.streamSet = new Set(streamNames);
    this.fetchFn = (globalThis.fetch ?? null) as FetchLike;

    opts.forEach((optFn) => optFn(this));

    if (!this.fetchFn) {
      throw new Error(
        "globalThis.fetch is undefined - please supply a custom fetch using WithFetch()",
      );
    }
  }

  /**
   * verifyRuntimeDeps checks if the runtime dependencies used by the client are present.
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
      const missingStr = missing.join(", ");
      throw new Error(
        `Missing required runtime dependencies: ${missingStr}. Install the necessary polyfills or use a compatible environment.`,
      );
    }
  }

  async callProc(
    name: string,
    input: unknown,
    headers: Record<string, string>,
    retryConfig?: RetryConfig,
    timeoutConfig?: TimeoutConfig,
  ): Promise<Response<any>> {
    const retryConf = retryConfig ?? {
      maxAttempts: 3,
      initialDelayMs: 1000,
      maxDelayMs: 5000,
      delayMultiplier: 2.0,
    };

    const timeoutConf = timeoutConfig ?? {
      timeoutMs: 30000,
    };

    if (!this.procSet.has(name)) {
      return {
        ok: false,
        error: new UfoError({
          message: `${name} procedure not found in schema`,
          category: "ClientError",
          code: "INVALID_PROC",
        }),
      };
    }

    let payload: string;
    try {
      payload = input == null ? "{}" : JSON.stringify(input);
    } catch (err) {
      return {
        ok: false,
        error: asError(err),
      };
    }

    const url = `${this.baseURL}/${name}`;
    const hdrs: Record<string, string> = {
      "content-type": "application/json",
      accept: "application/json",
      ...this.globalHeaders,
      ...headers,
    };

    let lastError: UfoError | null = null;
    for (let attempt = 1; attempt <= retryConf.maxAttempts; attempt++) {
      // Create AbortController for this attempt's timeout
      const abortController = new AbortController();
      let timeoutId: ReturnType<typeof setTimeout> | undefined = undefined;

      try {
        // Set up timeout if configured
        if (timeoutConf?.timeoutMs) {
          timeoutId = setTimeout(() => {
            abortController.abort();
          }, timeoutConf.timeoutMs);
        }

        const fetchResp = await this.fetchFn(url, {
          method: "POST",
          headers: hdrs,
          body: payload,
          signal: abortController.signal,
        });

        // Clear timeout on successful response
        if (timeoutId !== undefined) {
          clearTimeout(timeoutId);
        }

        if (!fetchResp.ok) {
          const error = new UfoError({
            message: `Unexpected HTTP status: ${fetchResp.status}`,
            category: "HTTPError",
            code: "BAD_STATUS",
            details: { status: fetchResp.status, attempt },
          });

          // Only retry on 5xx errors or network issues
          if (fetchResp.status >= 500 && attempt < retryConf.maxAttempts) {
            lastError = error;
            const backoffMs = Math.min(
              retryConf.initialDelayMs *
                Math.pow(retryConf.delayMultiplier, attempt - 1),
              retryConf.maxDelayMs,
            );
            await sleep(backoffMs);
            continue;
          }

          return { ok: false, error };
        }

        try {
          return await fetchResp.json();
        } catch (parseErr) {
          return { ok: false, error: asError(parseErr) } as Response<any>;
        }
      } catch (err) {
        // Clear timeout on error
        if (timeoutId !== undefined) {
          clearTimeout(timeoutId);
        }

        const error = asError(err);

        // Check if this was a timeout error
        if (abortController.signal.aborted && timeoutConf?.timeoutMs) {
          const timeoutError = new UfoError({
            message: `Request timeout after ${timeoutConf.timeoutMs}ms`,
            category: "TimeoutError",
            code: "REQUEST_TIMEOUT",
            details: { timeoutMs: timeoutConf.timeoutMs, attempt },
          });

          // Retry on timeout if we have attempts left
          if (attempt < retryConf.maxAttempts) {
            lastError = timeoutError;
            const backoffMs = Math.min(
              retryConf.initialDelayMs *
                Math.pow(retryConf.delayMultiplier, attempt - 1),
              retryConf.maxDelayMs,
            );
            await sleep(backoffMs);
            continue;
          }

          return { ok: false, error: timeoutError };
        }

        // Retry on network errors
        if (attempt < retryConf.maxAttempts) {
          lastError = error;
          const backoffMs = Math.min(
            retryConf.initialDelayMs *
              Math.pow(retryConf.delayMultiplier, attempt - 1),
            retryConf.maxDelayMs,
          );
          await sleep(backoffMs);
          continue;
        }

        return { ok: false, error };
      }
    }

    // This should never be reached, but just in case
    return {
      ok: false,
      error:
        lastError ||
        new UfoError({
          message: "Unknown error",
          category: "ClientError",
          code: "UNKNOWN",
        }),
    };
  }

  callStream(
    name: string,
    input: unknown,
    headers: Record<string, string>,
    reconnectConfig?: ReconnectConfig,
  ): {
    stream: AsyncGenerator<Response<any>, void, unknown>;
    cancel: () => void;
  } {
    const reconnectConf = reconnectConfig ?? {
      maxAttempts: 5,
      initialDelayMs: 1000,
      maxDelayMs: 5000,
      delayMultiplier: 2.0,
    };

    const self = this;
    let isCancelled = false;
    let currentAbortController: AbortController | null = null;

    const cancel = () => {
      isCancelled = true;
      currentAbortController?.abort();
    };

    async function* generator() {
      if (!self.streamSet.has(name)) {
        yield {
          ok: false,
          error: new UfoError({
            message: `${name} stream not found in schema`,
            category: "ClientError",
            code: "INVALID_STREAM",
          }),
        } as Response<any>;
        return;
      }

      let payload: string;
      try {
        payload = input == null ? "{}" : JSON.stringify(input);
      } catch (err) {
        yield { ok: false, error: asError(err) } as Response<any>;
        return;
      }

      const url = `${self.baseURL}/${name}`;
      const hdrs: Record<string, string> = {
        "content-type": "application/json",
        accept: "text/event-stream",
        ...self.globalHeaders,
        ...headers,
      };

      let reconnectAttempt = 0;
      while (!isCancelled) {
        currentAbortController = new AbortController();

        try {
          const fetchResp = await self.fetchFn(url, {
            method: "POST",
            headers: hdrs,
            body: payload,
            signal: currentAbortController.signal,
          });

          if (!fetchResp.ok) {
            const error = new UfoError({
              message: `Unexpected HTTP status: ${fetchResp.status}`,
              category: "HTTPError",
              code: "BAD_STATUS",
              details: { status: fetchResp.status, reconnectAttempt },
            });

            if (
              fetchResp.status >= 500 &&
              reconnectConf &&
              !isCancelled &&
              reconnectAttempt < reconnectConf.maxAttempts
            ) {
              yield { ok: false, error } as Response<any>;
              reconnectAttempt++;

              const delayMs = Math.min(
                reconnectConf.initialDelayMs *
                  Math.pow(reconnectConf.delayMultiplier, reconnectAttempt - 1),
                reconnectConf.maxDelayMs,
              );

              await sleep(delayMs);
              continue;
            }

            yield { ok: false, error } as Response<any>;
            return;
          }

          if (!fetchResp.body) {
            const error = new UfoError({
              message: "Missing response body for stream",
              category: "ConnectionError",
              code: "STREAM_CONNECT_FAILED",
              details: { reconnectAttempt },
            });
            // Do not reconnect; treat as non-retriable
            yield { ok: false, error } as Response<any>;
            return;
          }

          // Reset reconnect attempt counter on successful connection
          reconnectAttempt = 0;

          const reader = fetchResp.body.getReader();
          const decoder = new TextDecoder();
          let buffer = "";

          try {
            while (!isCancelled) {
              const { done, value } = await reader.read();
              if (done) break;

              buffer += decoder.decode(value, { stream: true });
              const idx = buffer.indexOf("\n\n");
              if (idx < 0) continue;

              const line = buffer.slice(0, idx).trim();
              buffer = buffer.slice(idx + 2);
              if (line === "") continue;

              if (line.startsWith("data:")) {
                const jsonStr = line.slice(5).trim();
                try {
                  const evt = JSON.parse(jsonStr) as Response<any>;
                  yield evt;
                } catch (err) {
                  yield { ok: false, error: asError(err) } as Response<any>;
                  return;
                }
              }
            }

            // If we reach here and weren't cancelled, the stream ended naturally
            if (!isCancelled) {
              return;
            }
          } catch (readError) {
            // Connection was interrupted, try to reconnect if enabled
            if (
              reconnectConf &&
              !isCancelled &&
              reconnectAttempt < reconnectConf.maxAttempts
            ) {
              yield {
                ok: false,
                error: new UfoError({
                  message: `Stream connection lost, attempting reconnect (${
                    reconnectAttempt + 1
                  }/${reconnectConf.maxAttempts})`,
                  category: "ConnectionError",
                  code: "STREAM_INTERRUPTED",
                  details: { reconnectAttempt: reconnectAttempt + 1 },
                }),
              } as Response<any>;

              reconnectAttempt++;
              const delayMs = Math.min(
                reconnectConf.initialDelayMs *
                  Math.pow(reconnectConf.delayMultiplier, reconnectAttempt - 1),
                reconnectConf.maxDelayMs,
              );

              await sleep(delayMs);
              continue;
            }

            // No more reconnect attempts or manually cancelled
            if (!isCancelled) {
              yield { ok: false, error: asError(readError) } as Response<any>;
            }
            return;
          }
        } catch (fetchError) {
          // Initial connection failed
          if (
            reconnectConf &&
            !isCancelled &&
            reconnectAttempt < reconnectConf.maxAttempts
          ) {
            yield {
              ok: false,
              error: new UfoError({
                message: `Failed to connect to stream, attempting reconnect (${
                  reconnectAttempt + 1
                }/${reconnectConf.maxAttempts})`,
                category: "ConnectionError",
                code: "STREAM_CONNECT_FAILED",
                details: { reconnectAttempt: reconnectAttempt + 1 },
              }),
            } as Response<any>;

            reconnectAttempt++;
            const delayMs = Math.min(
              reconnectConf.initialDelayMs *
                Math.pow(reconnectConf.delayMultiplier, reconnectAttempt - 1),
              reconnectConf.maxDelayMs,
            );

            await sleep(delayMs);
            continue;
          }

          // No more reconnect attempts or manually cancelled
          if (!isCancelled) {
            yield { ok: false, error: asError(fetchError) } as Response<any>;
          }
          return;
        }
      }
    }

    return { stream: generator(), cancel };
  }

  // Exposed mutators from builder
  setFetch(fetchFn: FetchLike) {
    this.fetchFn = fetchFn;
  }

  addGlobalHeader(k: string, v: string) {
    this.globalHeaders[k] = v;
  }
}

// -----------------------------------------------------------------------------
// Builder Helpers
// -----------------------------------------------------------------------------

type internalClientOption = (c: internalClient) => void;

function withFetch(fetchFn: FetchLike): internalClientOption {
  return (c) => c.setFetch(fetchFn);
}

function withGlobalHeader(key: string, value: string): internalClientOption {
  return (c) => c.addGlobalHeader(key, value);
}

// -----------------------------------------------------------------------------
// Fluent Builders exposed to generated wrappers
// -----------------------------------------------------------------------------

class clientBuilder {
  private baseURL: string;
  private opts: internalClientOption[] = [];

  constructor(baseURL: string) {
    this.baseURL = baseURL;
  }

  withFetch(fetchFn: FetchLike): clientBuilder {
    this.opts.push(withFetch(fetchFn));
    return this;
  }

  withGlobalHeader(key: string, value: string): clientBuilder {
    this.opts.push(withGlobalHeader(key, value));
    return this;
  }

  build(procNames: string[], streamNames: string[]): internalClient {
    return new internalClient(this.baseURL, procNames, streamNames, this.opts);
  }
}

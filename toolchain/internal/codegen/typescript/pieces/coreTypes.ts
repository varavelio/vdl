/** START FROM HERE **/

// -----------------------------------------------------------------------------
// Core Types
// -----------------------------------------------------------------------------

export type OperationType = "proc" | "stream";

export interface OperationDefinition {
  rpcName: string;
  name: string;
  type: OperationType;
}

/**
 * Response represents the response of a VDL call.
 *
 * @typeParam T - The concrete type of the successful response output.
 */
export type Response<T> =
  /** Successful response */
  | {
      /** Indicates whether the RPC call was successful. */
      ok: true;
      /** The successful output (only present if `ok === true`). */
      output: T;
      /** Not present when the call is successful. */
      error?: never;
    }
  /** Error response */
  | {
      /** Indicates whether the RPC call was successful. */
      ok: false;
      /** Structured error (only present if `ok === false`). */
      error: VdlError;
      /** Not present when the call fails. */
      output?: never;
    };

/**
 * Structured error type used throughout the VDL ecosystem.
 *
 * It provides structured information about errors that occur within the system,
 * enabling consistent error handling across servers and clients.
 *
 * Fields:
 *   - Message: A human-readable description of the error.
 *   - Category: Optional. Categorizes the error by its nature or source (e.g., "ValidationError", "DatabaseError").
 *   - Code: Optional. A machine-readable identifier for the specific error condition (e.g., "INVALID_EMAIL").
 *   - Details: Optional. Additional information about the error.
 */
export class VdlError extends Error {
  public readonly category?: string;
  public readonly code?: string;
  public readonly details?: Record<string, unknown>;

  constructor(options: {
    /**
     * Message provides a human-readable description of the error.
     *
     * This message can be displayed to end-users or used for logging and debugging purposes.
     *
     * Use Cases:
     *   1. If localization is not implemented, Message can be directly shown to the user to inform them of the issue.
     *   2. Developers can use Message in logs to diagnose problems during development or in production.
     */
    message: string;
    /**
     * Category categorizes the error by its nature or source.
     *
     * Examples:
     *   - "ValidationError" for input validation errors.
     *   - "DatabaseError" for errors originating from database operations.
     *   - "AuthenticationError" for authentication-related issues.
     *
     * Use Cases:
     *   1. In middleware, you can use Category to determine how to handle the error.
     *      For instance, you might log "InternalError" types and return a generic message to the client.
     *   2. Clients can inspect the Category to decide whether to prompt the user for action,
     *      such as re-authentication if the Category is "AuthenticationError".
     */
    category?: string;
    /**
     * Code is a machine-readable identifier for the specific error condition.
     *
     * Examples:
     *   - "INVALID_EMAIL" when an email address fails validation.
     *   - "USER_NOT_FOUND" when a requested user does not exist.
     *   - "RATE_LIMIT_EXCEEDED" when a client has made too many requests.
     *
     * Use Cases:
     *   1. Clients can map Codes to localized error messages for internationalization (i18n),
     *      displaying appropriate messages based on the user's language settings.
     *   2. Clients or middleware can implement specific logic based on the Code,
     *      such as retry mechanisms for "TEMPORARY_FAILURE" or showing captcha for "RATE_LIMIT_EXCEEDED".
     */
    code?: string;
    /**
     * Details contains optional additional information about the error.
     *
     * This field can include any relevant data that provides more context about the error.
     * The contents should be serializable to JSON.
     *
     * Use Cases:
     *   1. Providing field-level validation errors, e.g., Details could be:
     *      {"fields": {"email": "Email is invalid", "password": "Password is too short"}}
     *   2. Including diagnostic information such as timestamps, request IDs, or stack traces
     *      (ensure sensitive information is not exposed to clients).
     */
    details?: Record<string, unknown>;
  }) {
    super(options.message);
    this.name = "VdlError"; // Easier stack-trace filtering.
    this.category = options.category;
    this.code = options.code;
    this.details = options.details;
    Object.setPrototypeOf(this, new.target.prototype); // Maintains proper prototype chain when targeting ES5.
  }

  /** Serialises the error including all optional fields. */
  toJSON(): Record<string, unknown> {
    return {
      message: this.message,
      ...(this.category ? { category: this.category } : {}),
      ...(this.code ? { code: this.code } : {}),
      ...(this.details ? { details: this.details } : {}),
    };
  }
}

/**
 * Normalises arbitrary error inputs into a `VdlError` instance.
 */
export function asError(err: unknown): VdlError {
  if (err instanceof VdlError) {
    return err;
  }

  if (err instanceof Error) {
    return new VdlError({ message: err.message });
  }

  // Fallback for non-error values.
  return new VdlError({ message: String(err) });
}

/**
 * Convenience helper for missing-field validation errors.
 */
export function errorMissingRequiredField(message: string): VdlError {
  return new VdlError({
    message,
    category: "ValidationError",
    code: "MISSING_REQUIRED_FIELD",
  });
}

/**
 * Sleep for the given number of milliseconds.
 */
export function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

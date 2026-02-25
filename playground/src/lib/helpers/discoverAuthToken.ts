/**
 * Information about a discovered authentication token
 */
export interface TokenInfo {
  /** The original key name where the token was found */
  key: string;
  /** The token value */
  value: string;
  /** The full path to the token (e.g., "data.auth.accessToken") */
  path: string;
  /** The depth level where the token was found (0 = root level) */
  depth: number;
}

/**
 * Configuration options for token discovery
 */
export interface DiscoverTokenOptions {
  /** Maximum depth to search recursively (default: 10) */
  maxDepth?: number;
  /** Custom regex pattern to match token keys (overrides default patterns) */
  customPattern?: RegExp;
  /** Whether to include tokens found in arrays (default: true) */
  includeArrays?: boolean;
  /** Minimum token length to consider valid (default: 1) */
  minTokenLength?: number;
}

/**
 * Internal interface with all required options
 */
interface RequiredDiscoverTokenOptions {
  maxDepth: number;
  customPattern: RegExp | null;
  includeArrays: boolean;
  minTokenLength: number;
}

/**
 * Default regex patterns for detecting authentication token keys
 * Matches common authentication token naming conventions:
 * - token, authToken, auth_token, userToken, etc.
 * - jwt, jwtToken
 * - apiKey, api_key, authKey, auth_key
 * - accessToken, access_token, refreshToken, refresh_token
 * - bearer, bearerToken
 * - sessionToken, session_token
 */
const DEFAULT_TOKEN_PATTERNS = [
  /.*token.*/i, // Any key containing "token"
  /^jwt$/i, // Exact match for "jwt"
  /^bearer$/i, // Exact match for "bearer"
  /.*key$/i, // Any key ending with "key" (apiKey, authKey, etc.)
  /^auth$/i, // Exact match for "auth"
  /^access$/i, // Exact match for "access"
  /^refresh$/i, // Exact match for "refresh"
  /^session$/i, // Exact match for "session"
];

/**
 * Checks if a key matches any of the token patterns
 */
function isTokenKey(key: string, customPattern: RegExp | null): boolean {
  if (customPattern) {
    return customPattern.test(key);
  }

  const normalizedKey = key.toLowerCase();
  return DEFAULT_TOKEN_PATTERNS.some((pattern) => pattern.test(normalizedKey));
}

/**
 * Validates if a value appears to be a valid token
 */
function isValidTokenValue(value: unknown, minLength: number): value is string {
  return typeof value === "string" && value.trim().length >= minLength && value.trim().length > 0;
}

/**
 * Recursively searches for authentication tokens in an object or array
 */
function searchTokensRecursively(
  obj: unknown,
  options: RequiredDiscoverTokenOptions,
  path = "",
  depth = 0,
): TokenInfo[] {
  const tokens: TokenInfo[] = [];

  // Stop if we've reached maximum depth
  if (depth > options.maxDepth) {
    return tokens;
  }

  // Handle null/undefined
  if (obj === null || obj === undefined) {
    return tokens;
  }

  // Handle arrays
  if (Array.isArray(obj)) {
    if (!options.includeArrays) {
      return tokens;
    }

    obj.forEach((item, index) => {
      const currentPath = path ? `${path}[${index}]` : `[${index}]`;
      tokens.push(...searchTokensRecursively(item, options, currentPath, depth + 1));
    });
    return tokens;
  }

  // Handle objects
  if (typeof obj === "object") {
    const entries = Object.entries(obj);

    for (const [key, value] of entries) {
      const currentPath = path ? `${path}.${key}` : key;

      // Check if this key matches token patterns
      if (isTokenKey(key, options.customPattern)) {
        if (isValidTokenValue(value, options.minTokenLength)) {
          tokens.push({
            key,
            value,
            path: currentPath,
            depth,
          });
        }
      }

      // Recursively search nested objects/arrays
      if (typeof value === "object" && value !== null) {
        tokens.push(...searchTokensRecursively(value, options, currentPath, depth + 1));
      }
    }
  }

  return tokens;
}

/**
 * Discovers authentication tokens in a given object or array
 *
 * @param data - The object or array to search for tokens
 * @param options - Configuration options for the search
 * @returns Array of discovered tokens, sorted by depth (shallowest first)
 *
 * @example
 * ```typescript
 * const response = {
 *   token: "abc123",
 *   user: {
 *     authToken: "def456"
 *   }
 * };
 *
 * const tokens = discoverAuthToken(response);
 * // Returns: [
 * //   { key: "token", value: "abc123", path: "token", depth: 0 },
 * //   { key: "authToken", value: "def456", path: "user.authToken", depth: 1 }
 * // ]
 * ```
 */
export function discoverAuthToken(data: unknown, options: DiscoverTokenOptions = {}): TokenInfo[] {
  if (data === null || (typeof data !== "string" && typeof data !== "object")) {
    return [];
  }

  let dataObject: object = {};
  if (typeof data === "string") {
    try {
      dataObject = JSON.parse(data);
    } catch (_error) {
      return [];
    }
  } else {
    dataObject = data;
  }

  const defaultOptions: RequiredDiscoverTokenOptions = {
    maxDepth: options.maxDepth ?? 10,
    customPattern: options.customPattern ?? null,
    includeArrays: options.includeArrays ?? true,
    minTokenLength: options.minTokenLength ?? 1,
  };

  const tokens = searchTokensRecursively(dataObject, defaultOptions, "", 0);

  // Sort by depth (shallowest first), then by path alphabetically
  return tokens.sort((a, b) => {
    if (a.depth !== b.depth) {
      return a.depth - b.depth;
    }
    return a.path.localeCompare(b.path);
  });
}

/**
 * Convenience function to get the first (shallowest) token found
 *
 * @param data - The object or array to search for tokens
 * @param options - Configuration options for the search
 * @returns The first token found, or null if no tokens are discovered
 */
export function discoverFirstAuthToken(
  data: unknown,
  options: DiscoverTokenOptions = {},
): TokenInfo | null {
  const tokens = discoverAuthToken(data, options);
  return tokens.length > 0 ? tokens[0] : null;
}

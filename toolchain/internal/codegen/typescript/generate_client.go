package typescript

import (
	_ "embed"
	"fmt"

	"github.com/varavelio/gen"
	"github.com/varavelio/vdl/toolchain/internal/codegen/config"
	"github.com/varavelio/vdl/toolchain/internal/codegen/config/configtypes"
	"github.com/varavelio/vdl/toolchain/internal/core/ir/irtypes"
	"github.com/varavelio/vdl/toolchain/internal/util/strutil"
)

//go:embed pieces/client.ts
var clientRawPiece string

func generateClient(schema *irtypes.IrSchema, cfg *configtypes.TypeScriptTargetConfig) (string, error) {
	if !config.ShouldGenClient(cfg.GenClient) {
		return "", nil
	}

	piece := strutil.GetStrAfter(clientRawPiece, "/** START FROM HERE **/")
	if piece == "" {
		return "", fmt.Errorf("client.ts: could not find start delimiter")
	}

	g := gen.New().WithSpaces(2)

	generateImport(g, []string{"Response", "OperationType", "OperationDefinition"}, "./core", true, cfg)
	generateImport(g, []string{"VdlError", "asError", "sleep"}, "./core", false, cfg)
	generateImport(g, []string{"VDLProcedures", "VDLStreams"}, "./catalog", false, cfg)
	generateImportAll(g, "vdlTypes", "./types", cfg)
	g.Break()

	g.Raw(piece)
	g.Break()

	g.Line("// =============================================================================")
	g.Line("// Generated Client Implementation")
	g.Line("// =============================================================================")
	g.Break()

	// Generate main client builder function
	generateClientBuilder(g)
	g.Break()

	// Generate main Client class
	generateClientClass(g)
	g.Break()

	// Generate procedure registry and builders
	generateProcedureImplementation(g, schema)
	g.Break()

	// Generate stream registry and builders
	generateStreamImplementation(g, schema)

	return g.String(), nil
}

// generateClientBuilder creates the main NewClient function
func generateClientBuilder(g *gen.Generator) {
	g.Line("/**")
	g.Line(" * Creates a new VDL RPC client builder.")
	g.Line(" *")
	g.Line(" * @param baseURL - The base URL for the RPC endpoint")
	g.Line(" * @returns A fluent builder for configuring the client")
	g.Line(" *")
	g.Line(" * @example")
	g.Line(" * ```typescript")
	g.Line(" * const client = NewClient(\"https://api.example.com/v1/rpc\")")
	g.Line(" *   .withGlobalHeader(\"Authorization\", \"Bearer token\")")
	g.Line(" *   .build();")
	g.Line(" * ```")
	g.Line(" */")
	g.Line("export function NewClient(baseURL: string): ClientBuilder {")
	g.Block(func() {
		g.Line("return new ClientBuilder(baseURL);")
	})
	g.Line("}")
	g.Break()

	g.Line("/**")
	g.Line(" * Fluent builder for configuring VDL RPC client options.")
	g.Line(" */")
	g.Line("class ClientBuilder {")
	g.Block(func() {
		g.Line("private builder: clientBuilder;")
		g.Break()

		g.Line("constructor(baseURL: string) {")
		g.Block(func() {
			g.Line("this.builder = new clientBuilder(baseURL, VDLProcedures, VDLStreams);")
		})
		g.Line("}")
		g.Break()

		g.Line("/**")
		g.Line(" * Sets a custom fetch function for HTTP requests.")
		g.Line(" * Useful for environments without global fetch or for custom configurations.")
		g.Line(" */")
		g.Line("withCustomFetch(fetchFn: FetchLike): ClientBuilder {")
		g.Block(func() {
			g.Line("this.builder.withFetch(fetchFn);")
			g.Line("return this;")
		})
		g.Line("}")
		g.Break()

		g.Line("/**")
		g.Line(" * Adds a global header that will be sent with every request.")
		g.Line(" * Can be called multiple times to set different headers.")
		g.Line(" */")
		g.Line("withGlobalHeader(key: string, value: string): ClientBuilder {")
		g.Block(func() {
			g.Line("this.builder.withGlobalHeader(key, value);")
			g.Line("return this;")
		})
		g.Line("}")
		g.Break()

		g.Line("/**")
		g.Line(" * Builds the configured client instance.")
		g.Line(" * @returns A fully configured Client ready for use")
		g.Line(" */")
		g.Line("build(): Client {")
		g.Block(func() {
			g.Line("const intClient = this.builder.build();")
			g.Line("return new Client(intClient);")
		})
		g.Line("}")
	})
	g.Line("}")
}

// generateClientClass creates the main Client class
func generateClientClass(g *gen.Generator) {
	g.Line("/**")
	g.Line(" * Main VDL RPC client providing type-safe access to procedures and streams.")
	g.Line(" */")
	g.Line("export class Client {")
	g.Block(func() {
		g.Line("private intClient: internalClient;")
		g.Break()

		g.Line("/** Registry for accessing RPC procedures */")
		g.Line("public readonly procs: ProcRegistry;")
		g.Break()
		g.Line("/** Registry for accessing RPC streams */")
		g.Line("public readonly streams: StreamRegistry;")
		g.Break()

		g.Line("constructor(intClient: internalClient) {")
		g.Block(func() {
			g.Line("this.intClient = intClient;")
			g.Line("this.procs = new ProcRegistry(intClient);")
			g.Line("this.streams = new StreamRegistry(intClient);")
		})
		g.Line("}")
	})
	g.Line("}")
}

// generateProcedureImplementation generates all procedure-related code
func generateProcedureImplementation(g *gen.Generator, schema *irtypes.IrSchema) {
	g.Line("// =============================================================================")
	g.Line("// Procedure Implementation")
	g.Line("// =============================================================================")
	g.Break()

	// Generate procedure registry
	g.Line("/**")
	g.Line(" * Registry providing access to all RPC procedures.")
	g.Line(" */")
	g.Line("class ProcRegistry {")
	g.Line("private intClient: internalClient;")
	g.Break()
	g.Block(func() {
		g.Line("constructor(intClient: internalClient) {")
		g.Block(func() {
			g.Line("this.intClient = intClient;")
		})
		g.Line("}")
		g.Break()

		// Generate method for each procedure
		for _, proc := range schema.Procedures {
			fullName := strutil.ToPascalCase(proc.RpcName) + strutil.ToPascalCase(proc.Name)
			builderName := fmt.Sprintf("builder%s", fullName)
			methodName := strutil.ToCamelCase(proc.RpcName) + strutil.ToPascalCase(proc.Name)

			g.Linef("/**")
			g.Linef(" * Creates a call builder for the %s procedure.", fullName)
			if proc.Deprecated != nil {
				renderDeprecated(g, proc.Deprecated)
			}
			g.Linef(" */")
			g.Linef("%s(): %s {", methodName, builderName)
			g.Block(func() {
				g.Linef("return new %s(this.intClient, \"%s\", \"%s\");", builderName, proc.RpcName, proc.Name)
			})
			g.Line("}")
			g.Break()
		}
	})
	g.Line("}")
	g.Break()

	// Generate individual procedure builders
	for _, proc := range schema.Procedures {
		fullName := strutil.ToPascalCase(proc.RpcName) + strutil.ToPascalCase(proc.Name)
		builderName := fmt.Sprintf("builder%s", fullName)
		hydrateFuncName := fmt.Sprintf("hydrate%sOutput", fullName)
		inputType := fmt.Sprintf("%sInput", fullName)
		outputType := fmt.Sprintf("%sOutput", fullName)
		methodName := strutil.ToCamelCase(proc.RpcName) + strutil.ToPascalCase(proc.Name)

		g.Linef("/**")
		g.Linef(" * Fluent builder for the %s procedure.", fullName)
		if proc.Deprecated != nil && *proc.Deprecated != "" {
			g.Linef(" * @deprecated %s", *proc.Deprecated)
		}
		g.Linef(" */")
		g.Linef("class %s {", builderName)
		g.Block(func() {
			g.Line("private intClient: internalClient;")
			g.Line("private rpcName: string;")
			g.Line("private procName: string;")
			g.Line("private headers: Record<string, string> = {};")
			g.Line("private retryConfig?: RetryConfig;")
			g.Line("private timeoutConfig?: TimeoutConfig;")
			g.Line("private signal?: AbortSignal;")
			g.Break()

			g.Line("constructor(")
			g.Block(func() {
				g.Line("intClient: internalClient,")
				g.Line("rpcName: string,")
				g.Line("procName: string")
			})
			g.Line(") {")
			g.Block(func() {
				g.Line("this.intClient = intClient;")
				g.Line("this.rpcName = rpcName;")
				g.Line("this.procName = procName;")
			})
			g.Line("}")
			g.Break()

			g.Line("/**")
			g.Linef(" * Adds a custom header to the %s request.", fullName)
			g.Line(" * Can be called multiple times to set different headers.")
			g.Line(" */")
			g.Linef("withHeader(key: string, value: string): %s {", builderName)
			g.Block(func() {
				g.Line("this.headers[key] = value;")
				g.Line("return this;")
			})
			g.Line("}")
			g.Break()

			g.Line("/**")
			g.Linef(" * Configures automatic retry behavior for the %s procedure.", fullName)
			g.Line(" * Retries are performed with exponential backoff on 5xx errors and network failures.")
			g.Line(" *")
			g.Line(" * @param config - Retry configuration object")
			g.Line(" * @param config.maxAttempts - Maximum number of retry attempts (default: 3)")
			g.Line(" * @param config.initialDelayMs - Initial delay between retries in milliseconds (default: 1000)")
			g.Line(" * @param config.maxDelayMs - Maximum delay between retries in milliseconds (default: 5000)")
			g.Line(" * @param config.delayMultiplier - Cumulative multiplier applied to initialDelayMs on each retry (default: 2.0)")
			g.Line(" * @returns The builder instance for method chaining")
			g.Line(" *")
			g.Line(" * @example")
			g.Line(" * ```typescript")
			g.Line(" * // Basic retry configuration")
			g.Linef(" * const result = await client.procs.%s()", methodName)
			g.Line(" *   .withRetries({ maxAttempts: 3 })")
			g.Line(" *   .execute(input);")
			g.Line(" *")
			g.Line(" * // Advanced retry configuration")
			g.Linef(" * const result = await client.procs.%s()", methodName)
			g.Line(" *   .withRetries({")
			g.Line(" *     maxAttempts: 5,")
			g.Line(" *     initialDelayMs: 500,")
			g.Line(" *     maxDelayMs: 30000,")
			g.Line(" *     delayMultiplier: 1.5")
			g.Line(" *   })")
			g.Line(" *   .execute(input);")
			g.Line(" * ```")
			g.Line(" */")
			g.Linef("withRetries(config: Partial<RetryConfig>): %s {", builderName)
			g.Block(func() {
				g.Line("this.retryConfig = {")
				g.Block(func() {
					g.Line("maxAttempts: Math.max(0, config.maxAttempts ?? 3),")
					g.Line("initialDelayMs: Math.max(100, config.initialDelayMs ?? 1000),")
					g.Line("maxDelayMs: Math.max(100, config.maxDelayMs ?? 5000),")
					g.Line("delayMultiplier: Math.max(1, config.delayMultiplier ?? 2.0),")
					g.Line("jitter: Math.max(0, Math.min(1, config.jitter ?? 0.2)),")
				})
				g.Line("};")
				g.Line("return this;")
			})
			g.Line("}")
			g.Break()

			g.Line("/**")
			g.Linef(" * Configures timeout for each individual attempt of the %s procedure.", fullName)
			g.Line(" * Each retry attempt will be cancelled if it exceeds the specified timeout.")
			g.Line(" *")
			g.Line(" * @param config - Timeout configuration object")
			g.Line(" * @param config.timeoutMs - Timeout for each attempt in milliseconds (default: 30000)")
			g.Line(" * @returns The builder instance for method chaining")
			g.Line(" *")
			g.Line(" * @example")
			g.Line(" * ```typescript")
			g.Line(" * // Set 10 second timeout per attempt")
			g.Linef(" * const result = await client.procs.%s()", methodName)
			g.Line(" *   .withTimeout({ timeoutMs: 10000 })")
			g.Line(" *   .withRetries({ maxAttempts: 3 })")
			g.Line(" *   .execute(input);")
			g.Line(" *")
			g.Line(" * // Each of the 3 attempts will timeout after 10 seconds")
			g.Line(" * // Total maximum time: 3 attempts × 10s + retry delays")
			g.Line(" * ```")
			g.Line(" */")
			g.Linef("withTimeout(config: Partial<TimeoutConfig>): %s {", builderName)
			g.Block(func() {
				g.Line("this.timeoutConfig = {")
				g.Block(func() {
					g.Line("timeoutMs: Math.max(100, config.timeoutMs ?? 30000)")
				})
				g.Line("};")
				g.Line("return this;")
			})
			g.Line("}")
			g.Break()

			g.Line("/**")
			g.Linef(" * Sets an external AbortSignal to cancel the %s request.", fullName)
			g.Line(" *")
			g.Line(" * When the signal is aborted, the request is immediately cancelled and returns")
			g.Line(" * an error with code \"ABORTED\". Any in-progress retries are also stopped.")
			g.Line(" *")
			g.Line(" * @param signal - An AbortSignal from an AbortController")
			g.Line(" * @returns The builder instance for method chaining")
			g.Line(" *")
			g.Line(" * @example")
			g.Line(" * ```typescript")
			g.Line(" * // React useEffect cleanup pattern")
			g.Line(" * useEffect(() => {")
			g.Line(" *   const controller = new AbortController();")
			g.Linef(" *   client.procs.%s()", methodName)
			g.Line(" *     .withSignal(controller.signal)")
			g.Line(" *     .execute(input)")
			g.Line(" *     .then(setData)")
			g.Line(" *     .catch((e) => { if (e.code !== 'ABORTED') throw e; });")
			g.Line(" *   return () => controller.abort();")
			g.Line(" * }, []);")
			g.Line(" * ```")
			g.Line(" */")
			g.Linef("withSignal(signal: AbortSignal): %s {", builderName)
			g.Block(func() {
				g.Line("this.signal = signal;")
				g.Line("return this;")
			})
			g.Line("}")
			g.Break()

			g.Line("/**")
			g.Linef(" * Executes the %s procedure.", fullName)
			g.Line(" *")
			g.Linef(" * @param input - The %s input parameters", fullName)
			g.Linef(" * @returns Promise resolving to %s or throws VdlError if something went wrong", outputType)
			g.Line(" */")
			g.Linef("async execute(input: vdlTypes.%s): Promise<vdlTypes.%s> {", inputType, outputType)
			g.Block(func() {
				// Add client-side input validation
				validateFuncName := fmt.Sprintf("vdlTypes.validate%s", inputType)
				g.Linef("const validationError = %s(input);", validateFuncName)
				g.Line("if (validationError !== null) {")
				g.Block(func() {
					g.Line("throw new VdlError({")
					g.Block(func() {
						g.Line("message: validationError,")
						g.Line("code: \"INVALID_INPUT\",")
					})
					g.Line("});")
				})
				g.Line("}")
				g.Break()

				g.Line("const headerProvider: HeaderProvider = (h) => { Object.assign(h, this.headers); };")
				g.Line("const rawResponse = await this.intClient.callProc(")
				g.Block(func() {
					g.Line("this.rpcName,")
					g.Line("this.procName,")
					g.Line("input,")
					g.Line("[headerProvider],")
					g.Line("this.retryConfig,")
					g.Line("this.timeoutConfig,")
					g.Line("this.signal")
				})
				g.Line(");")

				g.Line("if (!rawResponse.ok) throw rawResponse.error;")
				g.Linef("return vdlTypes.%s(rawResponse.output);", hydrateFuncName)
			})
			g.Line("}")
		})
		g.Line("}")
		g.Break()
	}
}

// generateStreamImplementation generates all stream-related code
func generateStreamImplementation(g *gen.Generator, schema *irtypes.IrSchema) {
	g.Line("// =============================================================================")
	g.Line("// Stream Implementation")
	g.Line("// =============================================================================")
	g.Break()

	// Generate stream registry
	g.Line("/**")
	g.Line(" * Registry providing access to all RPC streams.")
	g.Line(" */")
	g.Line("class StreamRegistry {")
	g.Block(func() {
		g.Line("private intClient: internalClient;")
		g.Break()

		g.Line("constructor(intClient: internalClient) {")
		g.Block(func() {
			g.Line("this.intClient = intClient;")
		})
		g.Line("}")
		g.Break()

		// Generate method for each stream
		for _, stream := range schema.Streams {
			fullName := strutil.ToPascalCase(stream.RpcName) + strutil.ToPascalCase(stream.Name)
			builderName := fmt.Sprintf("builder%sStream", fullName)
			methodName := strutil.ToCamelCase(stream.RpcName) + strutil.ToPascalCase(stream.Name)

			g.Linef("/**")
			g.Linef(" * Creates a stream builder for the %s stream.", fullName)
			if stream.Deprecated != nil {
				renderDeprecated(g, stream.Deprecated)
			}
			g.Linef(" */")
			g.Linef("%s(): %s {", methodName, builderName)
			g.Block(func() {
				g.Linef("return new %s(this.intClient, \"%s\", \"%s\");", builderName, stream.RpcName, stream.Name)
			})
			g.Line("}")
			g.Break()
		}
	})
	g.Line("}")
	g.Break()

	// Generate individual stream builders
	for _, stream := range schema.Streams {
		fullName := strutil.ToPascalCase(stream.RpcName) + strutil.ToPascalCase(stream.Name)
		builderName := fmt.Sprintf("builder%sStream", fullName)
		hydrateFuncName := fmt.Sprintf("hydrate%sOutput", fullName)
		inputType := fmt.Sprintf("%sInput", fullName)
		outputType := fmt.Sprintf("%sOutput", fullName)
		responseType := fmt.Sprintf("%sResponse", fullName)
		methodName := strutil.ToCamelCase(stream.RpcName) + strutil.ToPascalCase(stream.Name)

		g.Linef("/**")
		g.Linef(" * Fluent builder for the %s stream.", fullName)
		if stream.Deprecated != nil && *stream.Deprecated != "" {
			g.Linef(" * @deprecated %s", *stream.Deprecated)
		}
		g.Linef(" */")
		g.Linef("class %s {", builderName)
		g.Block(func() {
			g.Line("private intClient: internalClient;")
			g.Line("private rpcName: string;")
			g.Line("private streamName: string;")
			g.Line("private headers: Record<string, string> = {};")
			g.Line("private reconnectConfig?: ReconnectConfig;")
			g.Line("private signal?: AbortSignal;")
			g.Line("private onConnectCb?: () => void;")
			g.Line("private onDisconnectCb?: (error: Error | null) => void;")
			g.Line("private onReconnectCb?: (attempt: number, delayMs: number) => void;")
			g.Break()

			g.Line("constructor(")
			g.Block(func() {
				g.Line("intClient: internalClient,")
				g.Line("rpcName: string,")
				g.Line("streamName: string")
			})
			g.Line(") {")
			g.Block(func() {
				g.Line("this.intClient = intClient;")
				g.Line("this.rpcName = rpcName;")
				g.Line("this.streamName = streamName;")
			})
			g.Line("}")
			g.Break()

			g.Line("/**")
			g.Linef(" * Adds a custom header to the %s stream request.", fullName)
			g.Line(" * Can be called multiple times to set different headers.")
			g.Line(" */")
			g.Linef("withHeader(key: string, value: string): %s {", builderName)
			g.Block(func() {
				g.Line("this.headers[key] = value;")
				g.Line("return this;")
			})
			g.Line("}")
			g.Break()

			g.Line("/**")
			g.Linef(" * Configures automatic reconnection for the %s stream.", fullName)
			g.Line(" * The stream will automatically attempt to reconnect when the connection is lost,")
			g.Line(" * but NOT when manually cancelled using the cancel() function.")
			g.Line(" *")
			g.Line(" * @param config - Reconnection configuration object")
			g.Line(" * @param config.maxAttempts - Maximum number of reconnection attempts (default: 5)")
			g.Line(" * @param config.initialDelayMs - Initial delay between reconnection attempts in milliseconds (default: 1000)")
			g.Line(" * @param config.maxDelayMs - Maximum delay between reconnection attempts in milliseconds (default: 5000)")
			g.Line(" * @param config.delayMultiplier - Cumulative multiplier applied to initialDelayMs on each retry (default: 2.0)")
			g.Line(" * @returns The builder instance for method chaining")
			g.Line(" *")
			g.Line(" * @example")
			g.Line(" * ```typescript")
			g.Line(" * // Basic reconnection configuration")
			g.Linef(" * const { stream, cancel } = client.streams.%s()", methodName)
			g.Line(" *   .withReconnect({ maxAttempts: 5 })")
			g.Line(" *   .execute(input);")
			g.Line(" *")
			g.Line(" * // Advanced reconnection configuration")
			g.Linef(" * const { stream, cancel } = client.streams.%s()", methodName)
			g.Line(" *   .withReconnect({")
			g.Line(" *     maxAttempts: 10,")
			g.Line(" *     initialDelayMs: 500,")
			g.Line(" *     maxDelayMs: 60000,")
			g.Line(" *     delayMultiplier: 1.5")
			g.Line(" *   })")
			g.Line(" *   .execute(input);")
			g.Line(" * ```")
			g.Line(" */")
			g.Linef("withReconnect(config: Partial<ReconnectConfig>): %s {", builderName)
			g.Block(func() {
				g.Line("this.reconnectConfig = {")
				g.Block(func() {
					g.Line("maxAttempts: Math.max(0, config.maxAttempts ?? 5),")
					g.Line("initialDelayMs: Math.max(100, config.initialDelayMs ?? 1000),")
					g.Line("maxDelayMs: Math.max(100, config.maxDelayMs ?? 5000),")
					g.Line("delayMultiplier: Math.max(1, config.delayMultiplier ?? 2.0),")
					g.Line("jitter: Math.max(0, Math.min(1, config.jitter ?? 0.2)),")
				})
				g.Line("};")
				g.Line("return this;")
			})
			g.Line("}")
			g.Break()

			g.Line("/**")
			g.Linef(" * Sets an external AbortSignal to cancel the %s stream.", fullName)
			g.Line(" *")
			g.Line(" * When the signal is aborted, the stream is immediately closed and no further")
			g.Line(" * reconnection attempts are made. This is equivalent to calling cancel().")
			g.Line(" *")
			g.Line(" * @param signal - An AbortSignal from an AbortController")
			g.Line(" * @returns The builder instance for method chaining")
			g.Line(" *")
			g.Line(" * @example")
			g.Line(" * ```typescript")
			g.Line(" * // React useEffect cleanup pattern")
			g.Line(" * useEffect(() => {")
			g.Line(" *   const controller = new AbortController();")
			g.Linef(" *   const { stream } = client.streams.%s()", methodName)
			g.Line(" *     .withSignal(controller.signal)")
			g.Line(" *     .execute(input);")
			g.Line(" *")
			g.Line(" *   (async () => {")
			g.Line(" *     for await (const event of stream) {")
			g.Line(" *       setData(event);")
			g.Line(" *     }")
			g.Line(" *   })();")
			g.Line(" *")
			g.Line(" *   return () => controller.abort();")
			g.Line(" * }, []);")
			g.Line(" * ```")
			g.Line(" */")
			g.Linef("withSignal(signal: AbortSignal): %s {", builderName)
			g.Block(func() {
				g.Line("this.signal = signal;")
				g.Line("return this;")
			})
			g.Line("}")
			g.Break()

			// withOnConnect method
			g.Line("/**")
			g.Linef(" * Sets a callback invoked when the %s stream successfully connects.", fullName)
			g.Line(" *")
			g.Line(" * This callback is called each time the connection is established, including")
			g.Line(" * after successful reconnection attempts.")
			g.Line(" *")
			g.Line(" * @param cb - Callback function to invoke on connection")
			g.Line(" * @returns The builder instance for method chaining")
			g.Line(" *")
			g.Line(" * @example")
			g.Line(" * ```typescript")
			g.Linef(" * const { stream } = client.streams.%s()", methodName)
			g.Line(" *   .withOnConnect(() => {")
			g.Line(" *     console.log('Stream connected!');")
			g.Line(" *     setConnectionStatus('connected');")
			g.Line(" *   })")
			g.Line(" *   .execute(input);")
			g.Line(" * ```")
			g.Line(" */")
			g.Linef("onConnect(cb: () => void): %s {", builderName)
			g.Block(func() {
				g.Line("this.onConnectCb = cb;")
				g.Line("return this;")
			})
			g.Line("}")
			g.Break()

			// withOnDisconnect method
			g.Line("/**")
			g.Linef(" * Sets a callback invoked when the %s stream permanently disconnects.", fullName)
			g.Line(" *")
			g.Line(" * This callback is called when the stream ends, either due to:")
			g.Line(" * - Normal completion (error will be null)")
			g.Line(" * - Maximum reconnection attempts exhausted")
			g.Line(" * - Manual cancellation via cancel() or AbortSignal")
			g.Line(" * - Unrecoverable error")
			g.Line(" *")
			g.Line(" * @param cb - Callback function receiving the final error (if any)")
			g.Line(" * @returns The builder instance for method chaining")
			g.Line(" *")
			g.Line(" * @example")
			g.Line(" * ```typescript")
			g.Linef(" * const { stream } = client.streams.%s()", methodName)
			g.Line(" *   .withOnDisconnect((error) => {")
			g.Line(" *     if (error) {")
			g.Line(" *       console.error('Stream disconnected with error:', error);")
			g.Line(" *     } else {")
			g.Line(" *       console.log('Stream ended normally');")
			g.Line(" *     }")
			g.Line(" *     setConnectionStatus('disconnected');")
			g.Line(" *   })")
			g.Line(" *   .execute(input);")
			g.Line(" * ```")
			g.Line(" */")
			g.Linef("onDisconnect(cb: (error: Error | null) => void): %s {", builderName)
			g.Block(func() {
				g.Line("this.onDisconnectCb = cb;")
				g.Line("return this;")
			})
			g.Line("}")
			g.Break()

			// withOnReconnect method
			g.Line("/**")
			g.Linef(" * Sets a callback invoked before each reconnection attempt for the %s stream.", fullName)
			g.Line(" *")
			g.Line(" * This callback is called when the connection is lost and the client is about")
			g.Line(" * to attempt reconnection. Use this to show reconnection status to users.")
			g.Line(" *")
			g.Line(" * @param cb - Callback function receiving the attempt number (1-based) and delay in ms")
			g.Line(" * @returns The builder instance for method chaining")
			g.Line(" *")
			g.Line(" * @example")
			g.Line(" * ```typescript")
			g.Linef(" * const { stream } = client.streams.%s()", methodName)
			g.Line(" *   .withReconnect({ maxAttempts: 10 })")
			g.Line(" *   .withOnReconnect((attempt, delayMs) => {")
			g.Line(" *     console.log(`Reconnecting (attempt ${attempt}) in ${delayMs}ms...`);")
			g.Line(" *     setStatus(`Reconnecting... (${attempt}/10)`);")
			g.Line(" *   })")
			g.Line(" *   .execute(input);")
			g.Line(" * ```")
			g.Line(" */")
			g.Linef("onReconnect(cb: (attempt: number, delayMs: number) => void): %s {", builderName)
			g.Block(func() {
				g.Line("this.onReconnectCb = cb;")
				g.Line("return this;")
			})
			g.Line("}")
			g.Break()

			g.Line("/**")
			g.Linef(" * Opens the %s Server-Sent Events stream.", fullName)
			g.Line(" *")
			g.Linef(" * @param input - The %s input parameters", fullName)
			g.Line(" * @returns Object containing:")
			g.Linef(" *   - stream: AsyncGenerator yielding Response<%s> events", outputType)
			g.Line(" *   - cancel: Function for cancelling the stream")
			g.Line(" *")
			g.Line(" * @example")
			g.Line(" * ```typescript")
			g.Linef(" * const { stream, cancel } = client.streams.%s().execute(input);", methodName)
			g.Line(" * ")
			g.Line(" * // All stream events are received here")
			g.Line(" * for await (const event of stream) {")
			g.Line(" *   if (event.ok) {")
			g.Line(" *     console.log('Received:', event.output);")
			g.Line(" *   } else {")
			g.Line(" *     console.error('Error:', event.error);")
			g.Line(" *   }")
			g.Line(" * }")
			g.Line(" * ")
			g.Line(" * // Cancel the stream when needed")
			g.Line(" * cancel();")
			g.Line(" * ```")
			g.Line(" */")
			g.Linef("execute(input: vdlTypes.%s): {", inputType)
			g.Block(func() {
				g.Linef("stream: AsyncGenerator<vdlTypes.%s, void, unknown>;", responseType)
				g.Line("cancel: () => void;")
			})
			g.Line("} {")
			g.Block(func() {
				// Add client-side input validation
				validateFuncName := fmt.Sprintf("validate%s", inputType)
				g.Linef("const validationError = vdlTypes.%s(input);", validateFuncName)
				g.Line("if (validationError !== null) {")
				g.Block(func() {
					g.Line("throw new VdlError({")
					g.Block(func() {
						g.Line("message: validationError,")
						g.Line("code: \"INVALID_INPUT\",")
					})
					g.Line("});")
				})
				g.Line("}")
				g.Break()

				g.Line("const headerProvider: HeaderProvider = (h) => { Object.assign(h, this.headers); };")
				g.Line("const { stream, cancel } = this.intClient.callStream(")
				g.Block(func() {
					g.Line("this.rpcName,")
					g.Line("this.streamName,")
					g.Line("input,")
					g.Line("[headerProvider],")
					g.Line("this.reconnectConfig,")
					g.Line("undefined, // maxMessageSize")
					g.Line("this.onConnectCb,")
					g.Line("this.onDisconnectCb,")
					g.Line("this.onReconnectCb,")
					g.Line("this.signal")
				})
				g.Line(");")

				g.Linef("const typedStream = async function* (): AsyncGenerator<vdlTypes.%s, void, unknown> {", responseType)
				g.Block(func() {
					g.Line("for await (const event of stream) {")
					g.Block(func() {
						g.Linef("const evt = event as vdlTypes.%s;", responseType)
						g.Linef("if (evt.ok) evt.output = vdlTypes.%s(evt.output);", hydrateFuncName)
						g.Line("yield evt;")
					})
					g.Line("}")
				})
				g.Line("};")

				g.Line("return {")
				g.Block(func() {
					g.Line("stream: typedStream(),")
					g.Line("cancel: cancel")
				})
				g.Line("};")
			})
			g.Line("}")
		})
		g.Line("}")
		g.Break()
	}
}

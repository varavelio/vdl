package typescript

import (
	_ "embed"
	"fmt"

	"github.com/uforg/ufogenkit"
	"github.com/uforg/uforpc/urpc/internal/schema"
	"github.com/uforg/uforpc/urpc/internal/util/strutil"
)

//go:embed pieces/client.ts
var clientRawPiece string

func generateClient(sch schema.Schema, config Config) (string, error) {
	if !config.IncludeClient {
		return "", nil
	}

	piece := strutil.GetStrAfter(clientRawPiece, "/** START FROM HERE **/")
	if piece == "" {
		return "", fmt.Errorf("client.ts: could not find start delimiter")
	}

	g := ufogenkit.NewGenKit().WithSpaces(2)

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
	generateProcedureImplementation(g, sch)
	g.Break()

	// Generate stream registry and builders
	generateStreamImplementation(g, sch)

	return g.String(), nil
}

// generateClientBuilder creates the main NewClient function
func generateClientBuilder(g *ufogenkit.GenKit) {
	g.Line("/**")
	g.Line(" * Creates a new UFO RPC client builder.")
	g.Line(" *")
	g.Line(" * @param baseURL - The base URL for the RPC endpoint")
	g.Line(" * @returns A fluent builder for configuring the client")
	g.Line(" *")
	g.Line(" * @example")
	g.Line(" * ```typescript")
	g.Line(" * const client = NewClient(\"https://api.example.com/v1/urpc\")")
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
	g.Line(" * Fluent builder for configuring UFO RPC client options.")
	g.Line(" */")
	g.Line("class ClientBuilder {")
	g.Block(func() {
		g.Line("private builder: clientBuilder;")
		g.Break()

		g.Line("constructor(baseURL: string) {")
		g.Block(func() {
			g.Line("this.builder = new clientBuilder(baseURL);")
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
			g.Line("const intClient = this.builder.build(ufoProcedureNames, ufoStreamNames);")
			g.Line("return new Client(intClient);")
		})
		g.Line("}")
	})
	g.Line("}")
}

// generateClientClass creates the main Client class
func generateClientClass(g *ufogenkit.GenKit) {
	g.Line("/**")
	g.Line(" * Main UFO RPC client providing type-safe access to procedures and streams.")
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
func generateProcedureImplementation(g *ufogenkit.GenKit, sch schema.Schema) {
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
		for _, procNode := range sch.GetProcNodes() {
			name := strutil.ToPascalCase(procNode.Name)
			builderName := fmt.Sprintf("builder%s", name)

			g.Linef("/**")
			g.Linef(" * Creates a call builder for the %s procedure.", name)
			renderDeprecated(g, procNode.Deprecated)
			g.Linef(" */")
			g.Linef("%s(): %s {", strutil.ToCamelCase(procNode.Name), builderName)
			g.Block(func() {
				g.Linef("return new %s(this.intClient, \"%s\");", builderName, procNode.Name)
			})
			g.Line("}")
			g.Break()
		}
	})
	g.Line("}")
	g.Break()

	// Generate individual procedure builders
	for _, procNode := range sch.GetProcNodes() {
		name := strutil.ToPascalCase(procNode.Name)
		builderName := fmt.Sprintf("builder%s", name)
		hydrateFuncName := fmt.Sprintf("hydrate%sOutput", name)
		inputType := fmt.Sprintf("%sInput", name)
		outputType := fmt.Sprintf("%sOutput", name)

		g.Linef("/**")
		g.Linef(" * Fluent builder for the %s procedure.", name)
		if procNode.Deprecated != nil && *procNode.Deprecated != "" {
			g.Linef(" * @deprecated %s", *procNode.Deprecated)
		}
		g.Linef(" */")
		g.Linef("class %s {", builderName)
		g.Block(func() {
			g.Line("private intClient: internalClient;")
			g.Line("private procName: string;")
			g.Line("private headers: Record<string, string> = {};")
			g.Line("private retryConfig?: RetryConfig;")
			g.Line("private timeoutConfig?: TimeoutConfig;")
			g.Break()

			g.Line("constructor(")
			g.Block(func() {
				g.Line("intClient: internalClient,")
				g.Line("procName: string")
			})
			g.Line(") {")
			g.Block(func() {
				g.Line("this.intClient = intClient;")
				g.Line("this.procName = procName;")
			})
			g.Line("}")
			g.Break()

			g.Line("/**")
			g.Linef(" * Adds a custom header to the %s request.", name)
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
			g.Linef(" * Configures automatic retry behavior for the %s procedure.", name)
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
			g.Linef(" * const result = await client.procs.%s()", strutil.ToCamelCase(procNode.Name))
			g.Line(" *   .withRetries({ maxAttempts: 3 })")
			g.Line(" *   .execute(input);")
			g.Line(" *")
			g.Line(" * // Advanced retry configuration")
			g.Linef(" * const result = await client.procs.%s()", strutil.ToCamelCase(procNode.Name))
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
					g.Line("delayMultiplier: Math.max(1, config.delayMultiplier ?? 2.0)")
				})
				g.Line("};")
				g.Line("return this;")
			})
			g.Line("}")
			g.Break()

			g.Line("/**")
			g.Linef(" * Configures timeout for each individual attempt of the %s procedure.", name)
			g.Line(" * Each retry attempt will be cancelled if it exceeds the specified timeout.")
			g.Line(" *")
			g.Line(" * @param config - Timeout configuration object")
			g.Line(" * @param config.timeoutMs - Timeout for each attempt in milliseconds (default: 30000)")
			g.Line(" * @returns The builder instance for method chaining")
			g.Line(" *")
			g.Line(" * @example")
			g.Line(" * ```typescript")
			g.Line(" * // Set 10 second timeout per attempt")
			g.Linef(" * const result = await client.procs.%s()", strutil.ToCamelCase(procNode.Name))
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
			g.Linef(" * Executes the %s procedure.", name)
			g.Line(" *")
			g.Linef(" * @param input - The %s input parameters", name)
			g.Linef(" * @returns Promise resolving to %s or throws UfoError if something went wrong", outputType)
			g.Line(" */")
			g.Linef("async execute(input: %s): Promise<%s> {", inputType, outputType)
			g.Block(func() {
				g.Line("const rawResponse = await this.intClient.callProc(")
				g.Block(func() {
					g.Line("this.procName,")
					g.Line("input,")
					g.Line("this.headers,")
					g.Line("this.retryConfig,")
					g.Line("this.timeoutConfig")
				})
				g.Line(");")

				g.Line("if (!rawResponse.ok) throw rawResponse.error;")
				g.Linef("return %s(rawResponse.output);", hydrateFuncName)
			})
			g.Line("}")
		})
		g.Line("}")
		g.Break()
	}
}

// generateStreamImplementation generates all stream-related code
func generateStreamImplementation(g *ufogenkit.GenKit, sch schema.Schema) {
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
		for _, streamNode := range sch.GetStreamNodes() {
			name := strutil.ToPascalCase(streamNode.Name)
			builderName := fmt.Sprintf("builder%sStream", name)

			g.Linef("/**")
			g.Linef(" * Creates a stream builder for the %s stream.", name)
			renderDeprecated(g, streamNode.Deprecated)
			g.Linef(" */")
			g.Linef("%s(): %s {", strutil.ToCamelCase(streamNode.Name), builderName)
			g.Block(func() {
				g.Linef("return new %s(this.intClient, \"%s\");", builderName, streamNode.Name)
			})
			g.Line("}")
			g.Break()
		}
	})
	g.Line("}")
	g.Break()

	// Generate individual stream builders
	for _, streamNode := range sch.GetStreamNodes() {
		name := strutil.ToPascalCase(streamNode.Name)
		builderName := fmt.Sprintf("builder%sStream", name)
		hydrateFuncName := fmt.Sprintf("hydrate%sOutput", name)
		inputType := fmt.Sprintf("%sInput", name)
		outputType := fmt.Sprintf("%sOutput", name)

		g.Linef("/**")
		g.Linef(" * Fluent builder for the %s stream.", name)
		if streamNode.Deprecated != nil && *streamNode.Deprecated != "" {
			g.Linef(" * @deprecated %s", *streamNode.Deprecated)
		}
		g.Linef(" */")
		g.Linef("class %s {", builderName)
		g.Block(func() {
			g.Line("private intClient: internalClient;")
			g.Line("private streamName: string;")
			g.Line("private headers: Record<string, string> = {};")
			g.Line("private reconnectConfig?: ReconnectConfig;")
			g.Break()

			g.Line("constructor(")
			g.Block(func() {
				g.Line("intClient: internalClient,")
				g.Line("streamName: string")
			})
			g.Line(") {")
			g.Block(func() {
				g.Line("this.intClient = intClient;")
				g.Line("this.streamName = streamName;")
			})
			g.Line("}")
			g.Break()

			g.Line("/**")
			g.Linef(" * Adds a custom header to the %s stream request.", name)
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
			g.Linef(" * Configures automatic reconnection for the %s stream.", name)
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
			g.Linef(" * const { stream, cancel } = client.streams.%s()", strutil.ToCamelCase(streamNode.Name))
			g.Line(" *   .withReconnect({ maxAttempts: 5 })")
			g.Line(" *   .execute(input);")
			g.Line(" *")
			g.Line(" * // Advanced reconnection configuration")
			g.Linef(" * const { stream, cancel } = client.streams.%s()", strutil.ToCamelCase(streamNode.Name))
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
					g.Line("delayMultiplier: Math.max(1, config.delayMultiplier ?? 2.0)")
				})
				g.Line("};")
				g.Line("return this;")
			})
			g.Line("}")
			g.Break()

			g.Line("/**")
			g.Linef(" * Opens the %s Server-Sent Events stream.", name)
			g.Line(" *")
			g.Linef(" * @param input - The %s input parameters", name)
			g.Line(" * @returns Object containing:")
			g.Linef(" *   - stream: AsyncGenerator yielding Response<%s> events", outputType)
			g.Line(" *   - cancel: Function for cancelling the stream")
			g.Line(" *")
			g.Line(" * @example")
			g.Line(" * ```typescript")
			g.Linef(" * const { stream, cancel } = client.streams.%s().execute(input);", strutil.ToCamelCase(streamNode.Name))
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
			g.Linef("execute(input: %s): {", inputType)
			g.Block(func() {
				g.Linef("stream: AsyncGenerator<Response<%s>, void, unknown>;", outputType)
				g.Line("cancel: () => void;")
			})
			g.Line("} {")
			g.Block(func() {
				g.Line("const { stream, cancel } = this.intClient.callStream(")
				g.Block(func() {
					g.Line("this.streamName,")
					g.Line("input,")
					g.Line("this.headers,")
					g.Line("this.reconnectConfig")
				})
				g.Line(");")

				g.Linef("const typedStream = async function* (): AsyncGenerator<Response<%s>, void, unknown> {", outputType)
				g.Block(func() {
					g.Line("for await (const event of stream) {")
					g.Block(func() {
						g.Linef("const evt = event as Response<%s>;", outputType)
						g.Linef("if (evt.ok) evt.output = %s(evt.output);", hydrateFuncName)
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

package dart

import (
	_ "embed"
	"fmt"

	"github.com/uforg/ufogenkit"
	"github.com/uforg/uforpc/urpc/internal/schema"
	"github.com/uforg/uforpc/urpc/internal/util/strutil"
)

func generateClient(sch schema.Schema, _ Config) (string, error) {
	g := ufogenkit.NewGenKit().WithSpaces(2)

	g.Line("// =============================================================================")
	g.Line("// Generated Client Implementation")
	g.Line("// =============================================================================")
	g.Break()

	generateClientBuilder(g)
	g.Break()

	generateClientClass(g)
	g.Break()

	generateProcedureImplementation(g, sch)
	g.Break()

	generateStreamImplementation(g, sch)
	g.Break()

	return g.String(), nil
}

func generateClientBuilder(g *ufogenkit.GenKit) {
	g.Line("/// Creates a new UFO RPC client builder.")
	g.Line("_ClientBuilder NewClient(String baseURL) => _ClientBuilder(baseURL);")
	g.Break()

	g.Line("/// Chainable builder for configuring UFO RPC client options (headers, etc.).")
	g.Line("class _ClientBuilder {")
	g.Block(func() {
		g.Line("final _InternalClientBuilder _builder;")
		g.Break()
		g.Line("/// Constructs a builder targeting the given base URL.")
		g.Line("_ClientBuilder(String baseURL) : _builder = _InternalClientBuilder(baseURL);")
		g.Break()
		g.Line("/// Adds a global header that will be sent with every request (procedures and streams). If the same header is set multiple times, the last value wins.")
		g.Line("_ClientBuilder withGlobalHeader(String key, String value) { _builder.withGlobalHeader(key, value); return this; }")
		g.Break()
		g.Line("/// Builds the configured client instance. Schema metadata is embedded to validate procedure/stream names at runtime.")
		g.Line("Client build() { final intClient = _builder.build(__ufoProcedureNames, __ufoStreamNames); return Client._internal(intClient); }")
	})
	g.Line("}")
}

func generateClientClass(g *ufogenkit.GenKit) {
	g.Line("/// Main UFO RPC client providing type-safe access to procedures and streams.")
	g.Line("class Client {")
	g.Block(func() {
		g.Line("final _ProcRegistry procs;")
		g.Line("final _StreamRegistry streams;")
		g.Break()
		g.Line("/// Internal constructor used by the builder.")
		g.Line("Client._internal(_InternalClient intClient) : procs = _ProcRegistry(intClient), streams = _StreamRegistry(intClient);")
	})
	g.Line("}")
}

func generateProcedureImplementation(g *ufogenkit.GenKit, sch schema.Schema) {
	g.Line("// =============================================================================")
	g.Line("// Procedure Implementation")
	g.Line("// =============================================================================")
	g.Break()

	g.Line("/// Registry providing access to all RPC procedures. Each method returns a fluent builder for configuring headers, retry and timeout settings.")
	g.Line("class _ProcRegistry {")
	g.Block(func() {
		g.Line("final _InternalClient _intClient;")
		g.Line("_ProcRegistry(this._intClient);")
		g.Break()
		for _, procNode := range sch.GetProcNodes() {
			name := strutil.ToPascalCase(procNode.Name)
			builderName := fmt.Sprintf("_Builder%s", name)
			g.Linef("/// Creates a call builder for the %s procedure.", name)
			renderDeprecatedDart(g, procNode.Deprecated)
			g.Linef("%s %s() => %s(_intClient, '%s');", builderName, strutil.ToCamelCase(procNode.Name), builderName, procNode.Name)
			g.Break()
		}
	})
	g.Line("}")
	g.Break()

	for _, procNode := range sch.GetProcNodes() {
		name := strutil.ToPascalCase(procNode.Name)
		builderName := fmt.Sprintf("_Builder%s", name)
		hydrateFuncName := fmt.Sprintf("%sOutput.fromJson", name)
		inputType := fmt.Sprintf("%sInput", name)
		outputType := fmt.Sprintf("%sOutput", name)

		g.Linef("/// Fluent builder for the %s procedure.", name)
		if procNode.Deprecated != nil && *procNode.Deprecated != "" {
			g.Linef("/// @deprecated %s", *procNode.Deprecated)
		}
		g.Linef("class %s {", builderName)
		g.Block(func() {
			g.Line("final _InternalClient _intClient;")
			g.Line("final String _procName;")
			g.Line("final Map<String, String> _headers = {};")
			g.Line("/// Per-call retry configuration. See RetryConfig for defaults and semantics.")
			g.Line("RetryConfig? retryConfig;")
			g.Line("/// Per-attempt timeout configuration. Applies to each retry attempt individually.")
			g.Line("TimeoutConfig? timeoutConfig;")
			g.Break()
			g.Linef("%s(this._intClient, this._procName);", builderName)
			g.Break()
			g.Linef("/// Adds a header for this specific call. Later calls with the same key override previous values.\n%s withHeader(String key, String value) { _headers[key] = value; return this; }", builderName)
			g.Break()
			g.Linef("/// Overrides the retry behavior for this call. Retries are attempted only for timeouts/network errors and HTTP 5xx responses.\n%s withRetries(RetryConfig config) { retryConfig = RetryConfig.sanitised(config); return this; }", builderName)
			g.Break()
			g.Linef("/// Sets a per-attempt timeout for this call. The timeout applies to each retry attempt separately.\n%s withTimeout(TimeoutConfig config) { timeoutConfig = TimeoutConfig.sanitised(config); return this; }", builderName)
			g.Break()
			g.Linef("/// Executes the %s procedure. Returns the typed output on success or throws a UfoError on failure.", name)
			g.Linef("Future<%s> execute(%s input) async {", outputType, inputType)
			g.Block(func() {
				g.Line("final rawResponse = await _intClient.callProc(_procName, input.toJson(), _headers, retryConfig, timeoutConfig);")
				g.Line("if (!rawResponse.ok) { throw rawResponse.error!; }")
				g.Linef("final out = %s((rawResponse.output as Map).cast<String, dynamic>());", hydrateFuncName)
				g.Line("return out;")
			})
			g.Line("}")
		})
		g.Line("}")
		g.Break()
	}
}

func generateStreamImplementation(g *ufogenkit.GenKit, sch schema.Schema) {
	g.Line("// =============================================================================")
	g.Line("// Stream Implementation")
	g.Line("// =============================================================================")
	g.Break()

	g.Line("/// Registry providing access to all RPC streams. Each method returns a fluent builder for configuring headers and reconnection settings.")
	g.Line("class _StreamRegistry {")
	g.Block(func() {
		g.Line("final _InternalClient _intClient;")
		g.Line("_StreamRegistry(this._intClient);")
		g.Break()
		for _, streamNode := range sch.GetStreamNodes() {
			name := strutil.ToPascalCase(streamNode.Name)
			builderName := fmt.Sprintf("_Builder%sStream", name)
			g.Linef("/// Creates a stream builder for the %s stream.", name)
			renderDeprecatedDart(g, streamNode.Deprecated)
			g.Linef("%s %s() => %s(_intClient, '%s');", builderName, strutil.ToCamelCase(streamNode.Name), builderName, streamNode.Name)
			g.Break()
		}
	})
	g.Line("}")
	g.Break()

	for _, streamNode := range sch.GetStreamNodes() {
		name := strutil.ToPascalCase(streamNode.Name)
		builderName := fmt.Sprintf("_Builder%sStream", name)
		hydrateFuncName := fmt.Sprintf("%sOutput.fromJson", name)
		inputType := fmt.Sprintf("%sInput", name)
		outputType := fmt.Sprintf("%sOutput", name)

		g.Linef("/// Fluent builder for the %s stream.", name)
		if streamNode.Deprecated != nil && *streamNode.Deprecated != "" {
			g.Linef("/// @deprecated %s", *streamNode.Deprecated)
		}
		g.Linef("class %s {", builderName)
		g.Block(func() {
			g.Line("final _InternalClient _intClient;")
			g.Line("final String _streamName;")
			g.Line("final Map<String, String> _headers = {};")
			g.Line("/// Per-stream reconnection configuration. See ReconnectConfig for defaults and semantics.")
			g.Line("ReconnectConfig? reconnectConfig;")
			g.Break()
			g.Linef("%s(this._intClient, this._streamName);", builderName)
			g.Break()
			g.Linef("/// Adds a header for this specific stream call. Later calls with the same key override previous values.\n%s withHeader(String key, String value) { _headers[key] = value; return this; }", builderName)
			g.Break()
			g.Linef("/// Overrides the reconnection behavior for this stream. Reconnects are attempted only on connection/read errors or HTTP 5xx at connect time.\n%s withReconnect(ReconnectConfig config) { reconnectConfig = ReconnectConfig.sanitised(config); return this; }", builderName)
			g.Break()
			g.Linef("/// Starts the %s stream and returns a typed stream handle with a cancel function.", name)
			g.Linef("_StreamHandle<%s> execute(%s input) {", outputType, inputType)
			g.Block(func() {
				g.Line("final handle = _intClient.callStream(_streamName, input.toJson(), _headers, reconnectConfig);")
				g.Linef("final typed = handle.stream.map((event) { if (event.ok) { final out = %s((event.output as Map).cast<String, dynamic>()); return Response<%s>.ok(out); } else { return Response<%s>.error(event.error!); } });", hydrateFuncName, outputType, outputType)
				g.Linef("return _StreamHandle<%s>(stream: typed, cancel: handle.cancel);", outputType)
			})
			g.Line("}")
		})
		g.Line("}")
		g.Break()
	}
}

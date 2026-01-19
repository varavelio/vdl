package dart

import (
	_ "embed"
	"fmt"

	"github.com/varavelio/gen"
	"github.com/varavelio/vdl/toolchain/internal/core/ir"
	"github.com/varavelio/vdl/toolchain/internal/util/strutil"
)

func generateClient(_ *ir.Schema, flat *flatSchema, _ Config) (string, error) {
	g := gen.New().WithSpaces(2)

	g.Line("// =============================================================================")
	g.Line("// Generated Client Implementation")
	g.Line("// =============================================================================")
	g.Break()

	generateClientBuilder(g)
	g.Break()

	generateClientClass(g)
	g.Break()

	generateProcedureImplementation(g, flat)
	g.Break()

	generateStreamImplementation(g, flat)
	g.Break()

	return g.String(), nil
}

func generateClientBuilder(g *gen.Generator) {
	g.Line("/// Creates a new VDL RPC client builder.")
	g.Line("_ClientBuilder NewClient(String baseURL) => _ClientBuilder(baseURL);")
	g.Break()

	g.Line("/// Chainable builder for configuring VDL RPC client options (headers, etc.).")
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
		g.Line("Client build() { final intClient = _builder.build(__vdlProcedureNames, __vdlStreamNames); return Client._internal(intClient); }")
	})
	g.Line("}")
}

func generateClientClass(g *gen.Generator) {
	g.Line("/// Main VDL RPC client providing type-safe access to procedures and streams.")
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

func generateProcedureImplementation(g *gen.Generator, flat *flatSchema) {
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
		for _, fp := range flat.Procedures {
			proc := fp.Procedure
			fullName := fullProcName(fp.RPCName, proc.Name)
			builderName := fmt.Sprintf("_Builder%s", fullName)
			methodName := strutil.ToCamelCase(fp.RPCName) + strutil.ToPascalCase(proc.Name)
			path := rpcProcPath(fp.RPCName, proc.Name)

			g.Linef("/// Creates a call builder for the %s procedure.", fullName)
			if proc.Deprecated != nil {
				renderDeprecatedDart(g, proc.Deprecated)
			}
			g.Linef("%s %s() => %s(_intClient, '%s');", builderName, methodName, builderName, path)
			g.Break()
		}
	})
	g.Line("}")
	g.Break()

	for _, fp := range flat.Procedures {
		proc := fp.Procedure
		fullName := fullProcName(fp.RPCName, proc.Name)
		builderName := fmt.Sprintf("_Builder%s", fullName)
		hydrateFuncName := fmt.Sprintf("%sOutput.fromJson", fullName)
		inputType := fmt.Sprintf("%sInput", fullName)
		outputType := fmt.Sprintf("%sOutput", fullName)

		g.Linef("/// Fluent builder for the %s procedure.", fullName)
		if proc.Deprecated != nil && proc.Deprecated.Message != "" {
			g.Linef("/// @deprecated %s", proc.Deprecated.Message)
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
			g.Linef("/// Executes the %s procedure. Returns the typed output on success or throws a UfoError on failure.", fullName)
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

func generateStreamImplementation(g *gen.Generator, flat *flatSchema) {
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
		for _, fs := range flat.Streams {
			stream := fs.Stream
			fullName := fullStreamName(fs.RPCName, stream.Name)
			builderName := fmt.Sprintf("_Builder%sStream", fullName)
			methodName := strutil.ToCamelCase(fs.RPCName) + strutil.ToPascalCase(stream.Name)
			path := rpcStreamPath(fs.RPCName, stream.Name)

			g.Linef("/// Creates a stream builder for the %s stream.", fullName)
			if stream.Deprecated != nil {
				renderDeprecatedDart(g, stream.Deprecated)
			}
			g.Linef("%s %s() => %s(_intClient, '%s');", builderName, methodName, builderName, path)
			g.Break()
		}
	})
	g.Line("}")
	g.Break()

	for _, fs := range flat.Streams {
		stream := fs.Stream
		fullName := fullStreamName(fs.RPCName, stream.Name)
		builderName := fmt.Sprintf("_Builder%sStream", fullName)
		hydrateFuncName := fmt.Sprintf("%sOutput.fromJson", fullName)
		inputType := fmt.Sprintf("%sInput", fullName)
		outputType := fmt.Sprintf("%sOutput", fullName)

		g.Linef("/// Fluent builder for the %s stream.", fullName)
		if stream.Deprecated != nil && stream.Deprecated.Message != "" {
			g.Linef("/// @deprecated %s", stream.Deprecated.Message)
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
			g.Linef("/// Starts the %s stream and returns a typed stream handle with a cancel function.", fullName)
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

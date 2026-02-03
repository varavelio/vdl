package golang

import (
	_ "embed"
	"fmt"

	"github.com/varavelio/gen"
	"github.com/varavelio/vdl/toolchain/internal/codegen/config"
	"github.com/varavelio/vdl/toolchain/internal/core/ir/irtypes"
	"github.com/varavelio/vdl/toolchain/internal/util/strutil"
)

//go:embed pieces/client.go
var clientRawPiece string

// generateClientCore generates the core client implementation (rpc_client.go).
func generateClientCore(_ *irtypes.IrSchema, config *config.GoConfig) (string, error) {
	if !config.GenClient {
		return "", nil
	}

	piece := strutil.GetStrAfter(clientRawPiece, "/** START FROM HERE **/")
	if piece == "" {
		return "", fmt.Errorf("client.go: could not find start delimiter")
	}

	g := gen.New().WithTabs()

	g.Raw(piece)
	g.Break()

	g.Line("// -----------------------------------------------------------------------------")
	g.Line("// Client generated implementation")
	g.Line("// -----------------------------------------------------------------------------")
	g.Break()

	g.Line("// clientBuilder provides a fluent API for configuring the VDL RPC client before instantiation.")
	g.Line("//")
	g.Line("// A builder is obtained by calling NewClient(baseURL), then optional")
	g.Line("// configuration methods can be chained before calling Build() to obtain a *Client ready for use.")
	g.Line("type clientBuilder struct {")
	g.Block(func() {
		g.Line("baseURL string")
		g.Line("opts    []internalClientOption")
	})
	g.Line("}")
	g.Break()

	g.Line("// NewClient instantiates a fluent builder for the VDL RPC client.")
	g.Line("//")
	g.Line("// The baseURL argument must point to the HTTP endpoint that handles VDL RPC")
	g.Line("// requests, for example: \"https://api.example.com/v1/rpc\".")
	g.Line("//")
	g.Line("// Example usage:")
	g.Line("//   client := NewClient(\"https://api.example.com/v1/rpc\").Build()")
	g.Line("func NewClient(baseURL string) *clientBuilder {")
	g.Block(func() {
		g.Line("return &clientBuilder{baseURL: baseURL, opts: []internalClientOption{}}")
	})
	g.Line("}")
	g.Break()

	g.Line("// WithHTTPClient supplies a custom *http.Client (e.g., with timeouts or custom transport).")
	g.Line("func (b *clientBuilder) WithHTTPClient(hc *http.Client) *clientBuilder {")
	g.Block(func() {
		g.Line("b.opts = append(b.opts, withHTTPClient(hc))")
		g.Line("return b")
	})
	g.Line("}")
	g.Break()

	g.Line("// WithGlobalHeader sets a header that will be sent with every request.")
	g.Line("// This is a convenience method that adds a static header provider.")
	g.Line("func (b *clientBuilder) WithHeader(key, value string) *clientBuilder {")
	g.Block(func() {
		g.Line("b.opts = append(b.opts, withGlobalHeader(key, value))")
		g.Line("return b")
	})
	g.Line("}")
	g.Break()

	g.Line("// WithHeaderProvider adds a dynamic header provider that is called before every request.")
	g.Line("// This is useful for injecting dynamic tokens (e.g. Auth) that might expire.")
	g.Line("// The provider can mutate the header map in-place.")
	g.Line("func (b *clientBuilder) WithHeaderProvider(provider HeaderProvider) *clientBuilder {")
	g.Block(func() {
		g.Line("b.opts = append(b.opts, withHeaderProvider(provider))")
		g.Line("return b")
	})
	g.Line("}")
	g.Break()

	g.Line("// WithInterceptor adds a global interceptor (middleware) that wraps every request.")
	g.Line("// Interceptors are executed in the order they are added.")
	g.Line("func (b *clientBuilder) WithInterceptor(interceptor Interceptor) *clientBuilder {")
	g.Block(func() {
		g.Line("b.opts = append(b.opts, withInterceptor(interceptor))")
		g.Line("return b")
	})
	g.Line("}")
	g.Break()

	g.Line("// WithRetryConfig sets the default retry configuration for all procedures.")
	g.Line("//")
	g.Line("// Parameters:")
	g.Line("//   - retryConfig.maxAttempts: Maximum number of retry attempts (default: 1)")
	g.Line("//   - retryConfig.initialDelay: Initial delay between retries (default: 0)")
	g.Line("//   - retryConfig.maxDelay: Maximum delay between retries (default: 0)")
	g.Line("//   - retryConfig.delayMultiplier: Cumulative multiplier applied to initialDelay on each retry (default: 1.0)")
	g.Line("//   - retryConfig.jitter: Randomness factor to prevent synchronized retries (thundering herd). Range: 0.0-1.0 (default: 0.2)")
	g.Line("func (b *clientBuilder) WithRetryConfig(conf RetryConfig) *clientBuilder {")
	g.Block(func() {
		g.Line("b.opts = append(b.opts, withGlobalRetryConfig(conf))")
		g.Line("return b")
	})
	g.Line("}")
	g.Break()

	g.Line("// WithTimeoutConfig sets the default timeout configuration for all procedures.")
	g.Line("func (b *clientBuilder) WithTimeoutConfig(conf TimeoutConfig) *clientBuilder {")
	g.Block(func() {
		g.Line("b.opts = append(b.opts, withGlobalTimeoutConfig(conf))")
		g.Line("return b")
	})
	g.Line("}")
	g.Break()

	g.Line("// WithReconnectConfig sets the default reconnection configuration for all streams.")
	g.Line("//")
	g.Line("// Parameters:")
	g.Line("//   - reconnectConfig.maxAttempts: Maximum number of reconnection attempts (default: 30)")
	g.Line("//   - reconnectConfig.initialDelay: Initial delay between reconnection attempts (default: 1 second)")
	g.Line("//   - reconnectConfig.maxDelay: Maximum delay between reconnection attempts (default: 30 seconds)")
	g.Line("//   - reconnectConfig.delayMultiplier: Cumulative multiplier applied to initialDelay on each retry (default: 1.5)")
	g.Line("//   - reconnectConfig.jitter: Randomness factor to prevent synchronized retries (thundering herd). Range: 0.0-1.0 (default: 0.2)")
	g.Line("func (b *clientBuilder) WithReconnectConfig(conf ReconnectConfig) *clientBuilder {")
	g.Block(func() {
		g.Line("b.opts = append(b.opts, withGlobalReconnectConfig(conf))")
		g.Line("return b")
	})
	g.Line("}")
	g.Break()

	g.Line("// WithMaxMessageSize sets the default maximum message size for all streams.")
	g.Line("//")
	g.Line("// The default value is 4MB.")
	g.Line("func (b *clientBuilder) WithMaxMessageSize(size int64) *clientBuilder {")
	g.Block(func() {
		g.Line("b.opts = append(b.opts, withGlobalMaxMessageSize(size))")
		g.Line("return b")
	})
	g.Line("}")
	g.Break()

	g.Line("// Build constructs the *Client using the configured options.")
	g.Line("func (b *clientBuilder) Build() *Client {")
	g.Block(func() {
		g.Line("intClient := newInternalClient(b.baseURL, VDLProcedures, VDLStreams, b.opts...)")
		g.Line("return &Client{RPCs: &clientRPCRegistry{intClient: intClient}}")
	})
	g.Line("}")
	g.Break()

	g.Line("// Client provides a high-level, type-safe interface for invoking RPC procedures and streams.")
	g.Line("type Client struct {")
	g.Block(func() {
		g.Line("RPCs *clientRPCRegistry")
	})
	g.Line("}")
	g.Break()

	g.Line("type clientRPCRegistry struct {")
	g.Block(func() {
		g.Line("intClient *internalClient")
	})
	g.Line("}")
	g.Break()

	return g.String(), nil
}

// generateClientRPC generates the client implementation for a specific RPC (rpc_{rpcName}_client.go).
func generateClientRPC(rpcName string, procs []irtypes.ProcedureDef, streams []irtypes.StreamDef, config *config.GoConfig) (string, error) {
	if !config.GenClient {
		return "", nil
	}

	g := gen.New().WithTabs()

	rpcStructName := fmt.Sprintf("client%sRPC", rpcName)
	procsStructName := fmt.Sprintf("client%sProcs", rpcName)
	streamsStructName := fmt.Sprintf("client%sStreams", rpcName)

	// 1. Method on clientRPCRegistry to get this RPC
	g.Linef("// %s returns the client registry for the %s RPC service.", rpcName, rpcName)
	g.Linef("func (r *clientRPCRegistry) %s() *%s {", rpcName, rpcStructName)
	g.Block(func() {
		g.Linef("return &%s{", rpcStructName)
		g.Linef("Procs: &%s{intClient: r.intClient},", procsStructName)
		g.Linef("Streams: &%s{intClient: r.intClient},", streamsStructName)
		g.Line("intClient: r.intClient,")
		g.Line("}")
	})
	g.Line("}")
	g.Break()

	// 2. Struct for this RPC
	g.Linef("type %s struct {", rpcStructName)
	g.Block(func() {
		g.Line("intClient *internalClient")
		g.Linef("Procs     *%s", procsStructName)
		g.Linef("Streams   *%s", streamsStructName)
	})
	g.Line("}")
	g.Break()

	// RPC Config setters
	g.Linef("// WithRetryConfig sets the default retry configuration for all procedures in the %s RPC.", rpcName)
	g.Line("//")
	g.Line("// This configuration overrides the global defaults but can be overridden by operation-specific configurations.")
	g.Line("//")
	g.Line("// Parameters:")
	g.Line("//   - retryConfig.maxAttempts: Maximum number of retry attempts (default: 1)")
	g.Line("//   - retryConfig.initialDelay: Initial delay between retries (default: 0)")
	g.Line("//   - retryConfig.maxDelay: Maximum delay between retries (default: 0)")
	g.Line("//   - retryConfig.delayMultiplier: Cumulative multiplier applied to initialDelay on each retry (default: 1.0)")
	g.Line("//   - retryConfig.jitter: Randomness factor to prevent synchronized retries (thundering herd). Range: 0.0-1.0 (default: 0.2)")
	g.Linef("func (r *%s) WithRetryConfig(conf RetryConfig) *%s {", rpcStructName, rpcStructName)
	g.Block(func() {
		g.Linef("r.intClient.setRPCRetryConfig(%q, conf)", rpcName)
		g.Line("return r")
	})
	g.Line("}")
	g.Break()

	g.Linef("// WithTimeoutConfig sets the default timeout configuration for all procedures in the %s RPC.", rpcName)
	g.Line("//")
	g.Line("// This configuration overrides the global defaults but can be overridden by operation-specific configurations.")
	g.Linef("func (r *%s) WithTimeoutConfig(conf TimeoutConfig) *%s {", rpcStructName, rpcStructName)
	g.Block(func() {
		g.Linef("r.intClient.setRPCTimeoutConfig(%q, conf)", rpcName)
		g.Line("return r")
	})
	g.Line("}")
	g.Break()

	g.Linef("// WithReconnectConfig sets the default reconnection configuration for all streams in the %s RPC.", rpcName)
	g.Line("//")
	g.Line("// This configuration overrides the global defaults but can be overridden by operation-specific configurations.")
	g.Line("//")
	g.Line("// Parameters:")
	g.Line("//   - reconnectConfig.maxAttempts: Maximum number of reconnection attempts (default: 30)")
	g.Line("//   - reconnectConfig.initialDelay: Initial delay between reconnection attempts (default: 1 second)")
	g.Line("//   - reconnectConfig.maxDelay: Maximum delay between reconnection attempts (default: 30 seconds)")
	g.Line("//   - reconnectConfig.delayMultiplier: Cumulative multiplier applied to initialDelay on each retry (default: 1.5)")
	g.Line("//   - reconnectConfig.jitter: Randomness factor to prevent synchronized retries (thundering herd). Range: 0.0-1.0 (default: 0.2)")
	g.Linef("func (r *%s) WithReconnectConfig(conf ReconnectConfig) *%s {", rpcStructName, rpcStructName)
	g.Block(func() {
		g.Linef("r.intClient.setRPCReconnectConfig(%q, conf)", rpcName)
		g.Line("return r")
	})
	g.Line("}")
	g.Break()

	g.Linef("// WithMaxMessageSize sets the default maximum message size for all streams in the %s RPC.", rpcName)
	g.Line("//")
	g.Line("// This configuration overrides the global defaults but can be overridden by operation-specific configurations.")
	g.Linef("func (r *%s) WithMaxMessageSize(size int64) *%s {", rpcStructName, rpcStructName)
	g.Block(func() {
		g.Linef("r.intClient.setRPCMaxMessageSize(%q, size)", rpcName)
		g.Line("return r")
	})
	g.Line("}")
	g.Break()

	g.Linef("// WithHeaderProvider adds a header provider for all operations in the %s RPC.", rpcName)
	g.Line("//")
	g.Line("// RPC-level providers are executed after global providers and before operation-specific providers.")
	g.Linef("func (r *%s) WithHeaderProvider(provider HeaderProvider) *%s {", rpcStructName, rpcStructName)
	g.Block(func() {
		g.Linef("r.intClient.setRPCHeaderProvider(%q, provider)", rpcName)
		g.Line("return r")
	})
	g.Line("}")
	g.Break()

	g.Linef("// WithHeader adds a static header for all operations in the %s RPC.", rpcName)
	g.Line("//")
	g.Line("// This is a convenience method that adds a static header provider at the RPC level.")
	g.Linef("func (r *%s) WithHeader(key, value string) *%s {", rpcStructName, rpcStructName)
	g.Block(func() {
		g.Linef("r.intClient.setRPCHeaderProvider(%q, func(_ context.Context, h http.Header) error {", rpcName)
		g.Line("h.Set(key, value)")
		g.Line("return nil")
		g.Line("})")
		g.Line("return r")
	})
	g.Line("}")
	g.Break()

	// 3. Procs Registry Struct
	g.Linef("type %s struct {", procsStructName)
	g.Block(func() {
		g.Line("intClient *internalClient")
	})
	g.Line("}")
	g.Break()

	// 4. Streams Registry Struct
	g.Linef("type %s struct {", streamsStructName)
	g.Block(func() {
		g.Line("intClient *internalClient")
	})
	g.Line("}")
	g.Break()

	for _, proc := range procs {
		uniqueName := rpcName + proc.Name
		builderName := "clientBuilder" + uniqueName

		// Client method to create builder
		g.Linef("// %s creates a call builder for the %s.%s procedure.", proc.Name, rpcName, proc.Name)
		if proc.GetDoc() != "" {
			renderDoc(g, proc.GetDoc(), true)
		}
		renderDeprecated(g, proc.Deprecation)
		g.Linef("func (registry *%s) %s() *%s {", procsStructName, proc.Name, builderName)
		g.Block(func() {
			g.Linef("return &%s{client: registry.intClient, headerProviders: []HeaderProvider{}, rpcName: %q, name: %q}", builderName, rpcName, proc.Name)
		})
		g.Line("}")
		g.Break()

		// Builder struct
		g.Linef("// %s represents a fluent call builder for the %s procedure.", builderName, uniqueName)
		g.Linef("type %s struct {", builderName)
		g.Block(func() {
			g.Line("rpcName     string")
			g.Line("name        string")
			g.Line("client      *internalClient")
			g.Line("input       any")
			g.Line("headerProviders []HeaderProvider")
			g.Line("retryConf   *RetryConfig")
			g.Line("timeoutConf *TimeoutConfig")
		})
		g.Line("}")
		g.Break()

		// WithHeader method
		g.Linef("// WithHeader adds a single HTTP header to the %s invocation.", uniqueName)
		g.Line("//")
		g.Line("// This header is applied after global and RPC-level headers, potentially overriding them.")
		g.Linef("func (b *%s) WithHeader(key, value string) *%s {", builderName, builderName)
		g.Block(func() {
			g.Line("b.headerProviders = append(b.headerProviders, func(_ context.Context, h http.Header) error {")
			g.Line("h.Set(key, value)")
			g.Line("return nil")
			g.Line("})")
			g.Line("return b")
		})
		g.Line("}")
		g.Break()

		// WithHeaderProvider method
		g.Linef("// WithHeaderProvider adds a dynamic header provider to the %s invocation.", uniqueName)
		g.Line("//")
		g.Line("// The provider is executed after global and RPC-level providers.")
		g.Linef("func (b *%s) WithHeaderProvider(provider HeaderProvider) *%s {", builderName, builderName)
		g.Block(func() {
			g.Line("b.headerProviders = append(b.headerProviders, provider)")
			g.Line("return b")
		})
		g.Line("}")
		g.Break()

		// WithRetryConfig method
		g.Linef("// WithRetryConfig sets the retry configuration for the %s procedure.", uniqueName)
		g.Line("//")
		g.Line("// This configuration overrides both global and RPC-level defaults.")
		g.Line("//")
		g.Line("// Parameters:")
		g.Line("//   - retryConfig.maxAttempts: Maximum number of retry attempts (default: 1)")
		g.Line("//   - retryConfig.initialDelay: Initial delay between retries (default: 0)")
		g.Line("//   - retryConfig.maxDelay: Maximum delay between retries (default: 0)")
		g.Line("//   - retryConfig.delayMultiplier: Cumulative multiplier applied to initialDelay on each retry (default: 1.0)")
		g.Line("//   - retryConfig.jitter: Randomness factor to prevent synchronized retries (thundering herd). Range: 0.0-1.0 (default: 0.2)")
		g.Linef("func (b *%s) WithRetryConfig(retryConfig RetryConfig) *%s {", builderName, builderName)
		g.Block(func() {
			g.Line("b.retryConf = &retryConfig")
			g.Line("return b")
		})
		g.Line("}")
		g.Break()

		// WithTimeoutConfig method
		g.Linef("// WithTimeoutConfig sets the timeout configuration for the %s procedure.", uniqueName)
		g.Line("//")
		g.Line("// This configuration overrides both global and RPC-level defaults.")
		g.Line("//")
		g.Line("// Parameters:")
		g.Line("//   - timeoutConfig.timeout: Request timeout (default: 30 seconds)")
		g.Linef("func (b *%s) WithTimeoutConfig(timeoutConfig TimeoutConfig) *%s {", builderName, builderName)
		g.Block(func() {
			g.Line("b.timeoutConf = &timeoutConfig")
			g.Line("return b")
		})
		g.Line("}")
		g.Break()

		// Execute method
		g.Linef("// Execute sends a request to the %s procedure.", uniqueName)
		g.Line("//")
		g.Line("// Returns:")
		g.Linef("//   1. The parsed %sOutput value on success.", uniqueName)
		g.Line("//   2. The error when the server responds with Ok=false or a transport/JSON error occurs.")
		g.Linef("func (b *%s) Execute(ctx context.Context, input %sInput) (%sOutput, error) {", builderName, uniqueName, uniqueName)
		g.Block(func() {
			g.Line("raw := b.client.proc(ctx, b.rpcName, b.name, input, b.headerProviders, b.retryConf, b.timeoutConf)")

			g.Line("if !raw.Ok {")
			g.Block(func() {
				g.Linef("return %sOutput{}, raw.Error", uniqueName)
			})
			g.Line("}")

			g.Linef("var out %sOutput", uniqueName)
			g.Line("if err := json.Unmarshal(raw.Output, &out); err != nil {")
			g.Block(func() {
				g.Linef("return %sOutput{}, Error{Message: fmt.Sprintf(\"failed to decode %s output: %%v\", err)}", uniqueName, uniqueName)
			})
			g.Line("}")

			g.Line("return out, nil")
		})
		g.Line("}")
		g.Break()
	}

	for _, stream := range streams {
		uniqueName := rpcName + stream.Name
		builderStream := "clientBuilder" + uniqueName + "Stream"

		// Client method to create stream builder
		g.Linef("// %s creates a stream builder for the %s.%s stream.", stream.Name, rpcName, stream.Name)
		if stream.GetDoc() != "" {
			renderDoc(g, stream.GetDoc(), true)
		}
		renderDeprecated(g, stream.Deprecation)
		g.Linef("func (registry *%s) %s() *%s {", streamsStructName, stream.Name, builderStream)
		g.Block(func() {
			g.Linef("return &%s{client: registry.intClient, headerProviders: []HeaderProvider{}, rpcName: %q, name: %q}", builderStream, rpcName, stream.Name)
		})
		g.Line("}")
		g.Break()

		// Builder struct
		g.Linef("// %s represents a fluent call builder for the %s stream.", builderStream, uniqueName)
		g.Linef("type %s struct {", builderStream)
		g.Block(func() {
			g.Line("rpcName       string")
			g.Line("name          string")
			g.Line("client        *internalClient")
			g.Line("input         any")
			g.Line("headerProviders []HeaderProvider")
			g.Line("reconnectConf *ReconnectConfig")
			g.Line("maxMessageSize int64")
			g.Line("onConnect     func()")
			g.Line("onDisconnect  func(error)")
			g.Line("onReconnect   func(int, time.Duration)")
		})
		g.Line("}")
		g.Break()

		// WithHeader
		g.Linef("// WithHeader adds a single HTTP header to the %s stream subscription.", uniqueName)
		g.Line("//")
		g.Line("// This header is applied after global and RPC-level headers, potentially overriding them.")
		g.Linef("func (b *%s) WithHeader(key, value string) *%s {", builderStream, builderStream)
		g.Block(func() {
			g.Line("b.headerProviders = append(b.headerProviders, func(_ context.Context, h http.Header) error {")
			g.Line("h.Set(key, value)")
			g.Line("return nil")
			g.Line("})")
			g.Line("return b")
		})
		g.Line("}")
		g.Break()

		// WithHeaderProvider method
		g.Linef("// WithHeaderProvider adds a dynamic header provider to the %s stream subscription.", uniqueName)
		g.Line("//")
		g.Line("// The provider is executed after global and RPC-level providers.")
		g.Linef("func (b *%s) WithHeaderProvider(provider HeaderProvider) *%s {", builderStream, builderStream)
		g.Block(func() {
			g.Line("b.headerProviders = append(b.headerProviders, provider)")
			g.Line("return b")
		})
		g.Line("}")
		g.Break()

		// WithReconnectConfig method
		g.Linef("// WithReconnectConfig sets the reconnection configuration for the %s stream.", uniqueName)
		g.Line("//")
		g.Line("// This configuration overrides both global and RPC-level defaults.")
		g.Line("//")
		g.Line("// Parameters:")
		g.Line("//   - reconnectConfig.maxAttempts: Maximum number of reconnection attempts (default: 30)")
		g.Line("//   - reconnectConfig.initialDelay: Initial delay between reconnection attempts (default: 1 second)")
		g.Line("//   - reconnectConfig.maxDelay: Maximum delay between reconnection attempts (default: 30 seconds)")
		g.Line("//   - reconnectConfig.delayMultiplier: Cumulative multiplier applied to initialDelay on each retry (default: 1.5)")
		g.Line("//   - reconnectConfig.jitter: Randomness factor to prevent synchronized retries (thundering herd). Range: 0.0-1.0 (default: 0.2)")
		g.Linef("func (b *%s) WithReconnectConfig(reconnectConfig ReconnectConfig) *%s {", builderStream, builderStream)
		g.Block(func() {
			g.Line("b.reconnectConf = &reconnectConfig")
			g.Line("return b")
		})
		g.Line("}")
		g.Break()

		// WithMaxMessageSize method
		g.Linef("// WithMaxMessageSize sets the maximum message size for the %s stream.", uniqueName)
		g.Line("//")
		g.Line("// This configuration overrides both global and RPC-level defaults.")
		g.Linef("func (b *%s) WithMaxMessageSize(size int64) *%s {", builderStream, builderStream)
		g.Block(func() {
			g.Line("b.maxMessageSize = size")
			g.Line("return b")
		})
		g.Line("}")
		g.Break()

		// Hooks
		g.Linef("// OnConnect registers a callback that is invoked when the stream is successfully connected.")
		g.Linef("func (b *%s) OnConnect(cb func()) *%s {", builderStream, builderStream)
		g.Block(func() {
			g.Line("b.onConnect = cb")
			g.Line("return b")
		})
		g.Line("}")
		g.Break()

		g.Linef("// OnDisconnect registers a callback that is invoked when the stream is disconnected.")
		g.Linef("func (b *%s) OnDisconnect(cb func(error)) *%s {", builderStream, builderStream)
		g.Block(func() {
			g.Line("b.onDisconnect = cb")
			g.Line("return b")
		})
		g.Line("}")
		g.Break()

		g.Linef("// OnReconnect registers a callback that is invoked when the stream is attempting to reconnect.")
		g.Linef("func (b *%s) OnReconnect(cb func(int, time.Duration)) *%s {", builderStream, builderStream)
		g.Block(func() {
			g.Line("b.onReconnect = cb")
			g.Line("return b")
		})
		g.Line("}")
		g.Break()

		// Execute
		g.Linef("// Execute opens the %s Server-Sent Events stream.", uniqueName)
		g.Line("//")
		g.Linef("// It returns a read-only channel of Response[%sOutput].", uniqueName)
		g.Line("//")
		g.Line("// Each event on the channel follows these rules:")
		g.Linef("//   - Ok=true  ⇒ Output contains a %sOutput value.", uniqueName)
		g.Linef("//   - Ok=false ⇒ Error describes either a server sent or transport error.")
		g.Line("//")
		g.Line("// The caller should cancel the supplied context to terminate the stream and must")
		g.Line("// drain the channel until it is closed.")
		g.Linef("func (b *%s) Execute(ctx context.Context, input %sInput) <-chan Response[%sOutput] {", builderStream, uniqueName, uniqueName)
		g.Block(func() {
			g.Line("rawCh := b.client.stream(ctx, b.rpcName, b.name, input, b.headerProviders, b.reconnectConf, b.maxMessageSize, b.onConnect, b.onDisconnect, b.onReconnect)")
			g.Linef("outCh := make(chan Response[%sOutput])", uniqueName)
			g.Line("go func() {")
			g.Block(func() {
				g.Line("for evt := range rawCh {")
				g.Block(func() {
					g.Line("if !evt.Ok {")
					g.Block(func() {
						g.Linef("outCh <- Response[%sOutput]{Ok: false, Error: evt.Error}", uniqueName)
					})
					g.Line("continue")
					g.Line("}")
					g.Linef("var out %sOutput", uniqueName)
					g.Line("if err := json.Unmarshal(evt.Output, &out); err != nil {")
					g.Block(func() {
						g.Linef("outCh <- Response[%sOutput]{Ok: false, Error: Error{Message: fmt.Sprintf(\"failed to decode %s output: %%v\", err)}}", uniqueName, uniqueName)
					})
					g.Line("continue")
					g.Line("}")
					g.Linef("outCh <- Response[%sOutput]{Ok: true, Output: out}", uniqueName)
				})
				g.Line("}")
				g.Line("close(outCh)")
			})
			g.Line("}()")
			g.Linef("return outCh")
		})
		g.Line("}")
		g.Break()
	}

	return g.String(), nil
}

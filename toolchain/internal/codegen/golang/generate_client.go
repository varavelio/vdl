package golang

import (
	_ "embed"
	"fmt"

	"github.com/uforg/ufogenkit"
	"github.com/uforg/uforpc/urpc/internal/schema"
	"github.com/uforg/uforpc/urpc/internal/util/strutil"
)

//go:embed pieces/client.go
var clientRawPiece string

func generateClient(sch schema.Schema, config Config) (string, error) {
	if !config.IncludeClient {
		return "", nil
	}

	piece := strutil.GetStrAfter(clientRawPiece, "/** START FROM HERE **/")
	if piece == "" {
		return "", fmt.Errorf("client.go: could not find start delimiter")
	}

	g := ufogenkit.NewGenKit().WithTabs()

	g.Raw(piece)
	g.Break()

	g.Line("// -----------------------------------------------------------------------------")
	g.Line("// Client generated implementation")
	g.Line("// -----------------------------------------------------------------------------")
	g.Break()

	g.Line("// clientBuilder provides a fluent API for configuring the UFO RPC client before instantiation.")
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

	g.Line("// NewClient instantiates a fluent builder for the UFO RPC client.")
	g.Line("//")
	g.Line("// The baseURL argument must point to the HTTP endpoint that handles UFO RPC")
	g.Line("// requests, for example: \"https://api.example.com/v1/urpc\".")
	g.Line("//")
	g.Line("// Example usage:")
	g.Line("//   client := NewClient(\"https://api.example.com/v1/urpc\").Build()")
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
	g.Line("func (b *clientBuilder) WithGlobalHeader(key, value string) *clientBuilder {")
	g.Block(func() {
		g.Line("b.opts = append(b.opts, withGlobalHeader(key, value))")
		g.Line("return b")
	})
	g.Line("}")
	g.Break()

	g.Line("// Build constructs the *Client using the configured options.")
	g.Line("func (b *clientBuilder) Build() *Client {")
	g.Block(func() {
		g.Line("intClient := newInternalClient(b.baseURL, ufoProcedureNames, ufoStreamNames, b.opts...)")
		g.Line("return &Client{Procs: &clientProcRegistry{intClient: intClient}, Streams: &clientStreamRegistry{intClient: intClient}}")
	})
	g.Line("}")
	g.Break()

	g.Line("// Client provides a high-level, type-safe interface for invoking RPC procedures and streams.")
	g.Line("type Client struct {")
	g.Block(func() {
		g.Line("Procs     *clientProcRegistry")
		g.Line("Streams   *clientStreamRegistry")
	})
	g.Line("}")
	g.Break()

	// -----------------------------------------------------------------------------
	// Generate procedure wrappers
	// -----------------------------------------------------------------------------

	g.Line("type clientProcRegistry struct {")
	g.Block(func() {
		g.Line("intClient *internalClient")
	})
	g.Line("}")
	g.Break()

	for _, procNode := range sch.GetProcNodes() {
		name := strutil.ToPascalCase(procNode.Name)
		builderName := "clientBuilder" + name

		// Client method to create builder
		g.Linef("// %s creates a call builder for the %s procedure.", name, name)
		renderDoc(g, procNode.Doc, true)
		renderDeprecated(g, procNode.Deprecated)
		g.Linef("func (registry *clientProcRegistry) %s() *%s {", name, builderName)
		g.Block(func() {
			g.Linef("return &%s{client: registry.intClient, headers: map[string]string{}, name: \"%s\"}", builderName, name)
		})
		g.Line("}")
		g.Break()

		// Builder struct
		g.Linef("// %s represents a fluent call builder for the %s procedure.", builderName, name)
		g.Linef("type %s struct {", builderName)
		g.Block(func() {
			g.Line("name        string")
			g.Line("client      *internalClient")
			g.Line("headers     map[string]string")
			g.Line("retryConf   *RetryConfig")
			g.Line("timeoutConf *TimeoutConfig")
		})
		g.Line("}")
		g.Break()

		// WithHeader method
		g.Linef("// WithHeader adds a single HTTP header to the %s invocation.", name)
		g.Linef("func (b *%s) WithHeader(key, value string) *%s {", builderName, builderName)
		g.Block(func() {
			g.Line("b.headers[key] = value")
			g.Line("return b")
		})
		g.Line("}")
		g.Break()

		// WithRetryConfig method
		g.Linef("// WithRetryConfig sets the retry configuration for the %s procedure.", name)
		g.Line("//")
		g.Line("// Parameters:")
		g.Line("//   - retryConfig.maxAttempts: Maximum number of retry attempts (default: 3)")
		g.Line("//   - retryConfig.initialDelay: Initial delay between retries (default: 1 second)")
		g.Line("//   - retryConfig.maxDelay: Maximum delay between retries (default: 5 seconds)")
		g.Line("//   - retryConfig.delayMultiplier: Cumulative multiplier applied to initialDelay on each retry (default: 2.0)")
		g.Linef("func (b *%s) WithRetryConfig(retryConfig RetryConfig) *%s {", builderName, builderName)
		g.Block(func() {
			g.Line("b.retryConf = &retryConfig")
			g.Line("return b")
		})
		g.Line("}")
		g.Break()

		// WithTimeoutConfig method
		g.Linef("// WithTimeoutConfig sets the timeout configuration for the %s procedure.", name)
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
		g.Linef("// Execute sends a request to the %s procedure.", name)
		g.Line("//")
		g.Line("// Returns:")
		g.Linef("//   1. The parsed %sOutput value on success.", name)
		g.Line("//   2. The error when the server responds with Ok=false or a transport/JSON error occurs.")
		g.Linef("func (b *%s) Execute(ctx context.Context, input %sInput) (%sOutput, error) {", builderName, name, name)
		g.Block(func() {
			g.Line("raw := b.client.proc(ctx, b.name, input, b.headers, b.retryConf, b.timeoutConf)")

			g.Line("if !raw.Ok {")
			g.Block(func() {
				g.Linef("return %sOutput{}, raw.Error", name)
			})
			g.Line("}")

			g.Linef("var out %sOutput", name)
			g.Line("if err := json.Unmarshal(raw.Output, &out); err != nil {")
			g.Block(func() {
				g.Linef("return %sOutput{}, Error{Message: fmt.Sprintf(\"failed to decode %s output: %%v\", err)}", name, name)
			})
			g.Line("}")

			g.Line("return out, nil")
		})
		g.Line("}")
		g.Break()
	}

	// -----------------------------------------------------------------------------
	// Generate stream wrappers
	// -----------------------------------------------------------------------------

	g.Line("type clientStreamRegistry struct {")
	g.Block(func() {
		g.Line("intClient *internalClient")
	})
	g.Line("}")
	g.Break()

	for _, streamNode := range sch.GetStreamNodes() {
		name := strutil.ToPascalCase(streamNode.Name)
		builderStream := "clientBuilder" + name + "Stream"

		// Client method to create stream builder
		g.Linef("// %s creates a stream builder for the %s stream.", name, name)
		renderDoc(g, streamNode.Doc, true)
		renderDeprecated(g, streamNode.Deprecated)
		g.Linef("func (registry *clientStreamRegistry) %s() *%s {", name, builderStream)
		g.Block(func() {
			g.Linef("return &%s{client: registry.intClient, headers: map[string]string{}, name: \"%s\"}", builderStream, name)
		})
		g.Line("}")
		g.Break()

		// Builder struct
		g.Linef("// %s represents a fluent call builder for the %s stream.", builderStream, name)
		g.Linef("type %s struct {", builderStream)
		g.Block(func() {
			g.Line("name          string")
			g.Line("client        *internalClient")
			g.Line("headers       map[string]string")
			g.Line("reconnectConf *ReconnectConfig")
		})
		g.Line("}")
		g.Break()

		// WithHeader
		g.Linef("// WithHeader adds a single HTTP header to the %s stream subscription.", name)
		g.Linef("func (b *%s) WithHeader(key, value string) *%s {", builderStream, builderStream)
		g.Block(func() {
			g.Line("b.headers[key] = value")
			g.Line("return b")
		})
		g.Line("}")
		g.Break()

		// WithReconnectConfig method
		g.Linef("// WithReconnectConfig sets the reconnection configuration for the %s stream.", name)
		g.Line("//")
		g.Line("// Parameters:")
		g.Line("//   - reconnectConfig.maxAttempts: Maximum number of reconnection attempts (default: 5)")
		g.Line("//   - reconnectConfig.initialDelay: Initial delay between reconnection attempts (default: 1 second)")
		g.Line("//   - reconnectConfig.maxDelay: Maximum delay between reconnection attempts (default: 5 seconds)")
		g.Line("//   - reconnectConfig.delayMultiplier: Cumulative multiplier applied to initialDelay on each retry (default: 2.0)")
		g.Linef("func (b *%s) WithReconnectConfig(reconnectConfig ReconnectConfig) *%s {", builderStream, builderStream)
		g.Block(func() {
			g.Line("b.reconnectConf = &reconnectConfig")
			g.Line("return b")
		})
		g.Line("}")
		g.Break()

		// Execute
		g.Linef("// Execute opens the %s Server-Sent Events stream.", name)
		g.Line("//")
		g.Linef("// It returns a read-only channel of Response[%sOutput].", name)
		g.Line("//")
		g.Line("// Each event on the channel follows these rules:")
		g.Linef("//   - Ok=true  ⇒ Output contains a %sOutput value.", name)
		g.Line("//   - Ok=false ⇒ Error describes either a server sent or transport error.")
		g.Line("//")
		g.Line("// The caller should cancel the supplied context to terminate the stream and must")
		g.Line("// drain the channel until it is closed.")
		g.Linef("func (b *%s) Execute(ctx context.Context, input %sInput) <-chan Response[%sOutput] {", builderStream, name, name)
		g.Block(func() {
			g.Line("rawCh := b.client.stream(ctx, b.name, input, b.headers, b.reconnectConf)")
			g.Linef("outCh := make(chan Response[%sOutput])", name)
			g.Line("go func() {")
			g.Block(func() {
				g.Line("for evt := range rawCh {")
				g.Block(func() {
					g.Line("if !evt.Ok {")
					g.Block(func() {
						g.Linef("outCh <- Response[%sOutput]{Ok: false, Error: evt.Error}", name)
					})
					g.Line("continue")
					g.Line("}")
					g.Linef("var out %sOutput", name)
					g.Line("if err := json.Unmarshal(evt.Output, &out); err != nil {")
					g.Block(func() {
						g.Linef("outCh <- Response[%sOutput]{Ok: false, Error: Error{Message: fmt.Sprintf(\"failed to decode %s output: %%v\", err)}}", name, name)
					})
					g.Line("continue")
					g.Line("}")
					g.Linef("outCh <- Response[%sOutput]{Ok: true, Output: out}", name)
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

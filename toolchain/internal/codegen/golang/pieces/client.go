//nolint:unused
package pieces

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"
)

/** START FROM HERE **/

// -----------------------------------------------------------------------------
// Client types
// -----------------------------------------------------------------------------

// RetryConfig defines retry behavior for procedure calls.
type RetryConfig struct {
	// Maximum number of retry attempts (default: 1)
	MaxAttempts int
	// Initial delay between retries (default: 0)
	InitialDelay time.Duration
	// Maximum delay between retries (default: 0)
	MaxDelay time.Duration
	// Cumulative multiplier applied to initialDelayMs on each retry (default: 1.0)
	DelayMultiplier float64
	// Jitter introduces randomness to the delay to avoid thundering herd (default: 0.2)
	Jitter float64
}

// TimeoutConfig defines timeout behavior for procedure calls.
type TimeoutConfig struct {
	// Request Timeout (default: 30 seconds)
	Timeout time.Duration
}

// ReconnectConfig defines reconnection behavior for stream calls.
type ReconnectConfig struct {
	// Maximum number of reconnection attempts (default: 30)
	MaxAttempts int
	// Initial delay between reconnection attempts (default: 1 second)
	InitialDelay time.Duration
	// Maximum delay between reconnection attempts (default: 30 seconds)
	MaxDelay time.Duration
	// Cumulative multiplier applied to initialDelayMs on each retry (default: 1.5)
	DelayMultiplier float64
	// Jitter introduces randomness to the delay to avoid thundering herd (default: 0.2)
	Jitter float64
}

// HeaderProvider receives the current headers and mutates them in place.
// It is called before every request (including retries).
// If an error is returned, the request is aborted.
type HeaderProvider func(ctx context.Context, h http.Header) error

// RequestInfo contains information about the RPC request.
type RequestInfo struct {
	RPCName       string
	OperationName string
	Input         any
	Type          OperationType // "proc" | "stream"
}

// Invoker is the final step in the interceptor chain that performs the actual request.
type Invoker func(ctx context.Context, req RequestInfo) (Response[json.RawMessage], error)

// Interceptor is a middleware that wraps the request execution.
type Interceptor func(ctx context.Context, req RequestInfo, next Invoker) (Response[json.RawMessage], error)

// internalClient is the core engine used by the generated client abstraction. It is
// thread-safe and can be reused across concurrent requests.
//
// The client validates the requested operation name against the schema to fail
// fast when a typo occurs in the generated wrapper.
//
// The zero value is not usable – use newInternalClient to construct one.
type internalClient struct {
	// Immutable after construction.
	baseURL       string
	httpClient    *http.Client
	operationDefs map[string]map[string]OperationType
	// header configuration (global on every request)
	// globalHeaders map[string]string -> Removed in favor of headerProviders

	// Dynamic components
	headerProviders []HeaderProvider
	interceptors    []Interceptor

	// RPC specific header providers
	rpcHeaderProviders map[string][]HeaderProvider

	// Default Configurations
	globalRetryConf      *RetryConfig
	globalTimeoutConf    *TimeoutConfig
	globalReconnectConf  *ReconnectConfig
	globalMaxMessageSize int64

	// Per-RPC Default Configurations
	rpcRetryConf      map[string]*RetryConfig
	rpcTimeoutConf    map[string]*TimeoutConfig
	rpcReconnectConf  map[string]*ReconnectConfig
	rpcMaxMessageSize map[string]int64

	// mu protects concurrent access to the configuration maps
	mu sync.RWMutex
}

// internalClientOption represents a configuration option for internalClient.
type internalClientOption func(*internalClient)

// withHTTPClient supplies a custom *http.Client. If nil, the default client is
// used. Callers can leverage this to inject time-outs, proxies, or a transport
// with advanced TLS settings.
func withHTTPClient(hc *http.Client) internalClientOption {
	return func(c *internalClient) {
		if hc != nil {
			c.httpClient = hc
		}
	}
}

// withGlobalHeader sets a header that will be attached to every request.
func withGlobalHeader(key string, value string) internalClientOption {
	return func(c *internalClient) {
		c.headerProviders = append(c.headerProviders, func(_ context.Context, h http.Header) error {
			h.Set(key, value)
			return nil
		})
	}
}

// withHeaderProvider adds a dynamic header provider.
func withHeaderProvider(provider HeaderProvider) internalClientOption {
	return func(c *internalClient) {
		c.headerProviders = append(c.headerProviders, provider)
	}
}

// withInterceptor adds an interceptor to the chain.
func withInterceptor(interceptor Interceptor) internalClientOption {
	return func(c *internalClient) {
		c.interceptors = append(c.interceptors, interceptor)
	}
}

// withGlobalRetryConfig sets the global default retry configuration.
func withGlobalRetryConfig(conf RetryConfig) internalClientOption {
	return func(c *internalClient) {
		c.globalRetryConf = &conf
	}
}

// withGlobalTimeoutConfig sets the global default timeout configuration.
func withGlobalTimeoutConfig(conf TimeoutConfig) internalClientOption {
	return func(c *internalClient) {
		c.globalTimeoutConf = &conf
	}
}

// withGlobalReconnectConfig sets the global default reconnection configuration.
func withGlobalReconnectConfig(conf ReconnectConfig) internalClientOption {
	return func(c *internalClient) {
		c.globalReconnectConf = &conf
	}
}

// withGlobalMaxMessageSize sets the global maximum message size for streams.
func withGlobalMaxMessageSize(size int64) internalClientOption {
	return func(c *internalClient) {
		c.globalMaxMessageSize = size
	}
}

// newInternalClient creates a new internalClient capable of talking to the VDL
// server described by procNames and streamNames.
//
// The caller can optionally pass functional options to tweak the configuration
// (base URL, custom *http.Client, …).
func newInternalClient(
	baseURL string,
	procDefs []OperationDefinition,
	streamDefs []OperationDefinition,
	opts ...internalClientOption,
) *internalClient {
	operationDefs := make(map[string]map[string]OperationType)

	ensureRPC := func(rpcName string) {
		if _, ok := operationDefs[rpcName]; !ok {
			operationDefs[rpcName] = make(map[string]OperationType)
		}
	}

	for _, def := range procDefs {
		ensureRPC(def.RPCName)
		operationDefs[def.RPCName][def.Name] = def.Type
	}
	for _, def := range streamDefs {
		ensureRPC(def.RPCName)
		operationDefs[def.RPCName][def.Name] = def.Type
	}

	cli := &internalClient{
		baseURL:            strings.TrimRight(baseURL, "/"),
		httpClient:         http.DefaultClient,
		operationDefs:      operationDefs,
		headerProviders:    []HeaderProvider{},
		rpcHeaderProviders: make(map[string][]HeaderProvider),
		interceptors:       []Interceptor{},
		rpcRetryConf:       make(map[string]*RetryConfig),
		rpcTimeoutConf:     make(map[string]*TimeoutConfig),
		rpcReconnectConf:   make(map[string]*ReconnectConfig),
		rpcMaxMessageSize:  make(map[string]int64),
	}

	// Apply functional options.
	for _, opt := range opts {
		opt(cli)
	}

	return cli
}

/* Internal client builder */

// internalClientBuilder helps constructing an internalClient using chained
// configuration methods before calling Build().
type internalClientBuilder struct {
	baseURL    string
	procDefs   []OperationDefinition
	streamDefs []OperationDefinition
	opts       []internalClientOption
}

// newClientBuilder creates a builder with the schema information (procedure and
// stream names). Generated code will pass the automatically produced slices.
func newClientBuilder(baseURL string, procDefs, streamDefs []OperationDefinition) *internalClientBuilder {
	return &internalClientBuilder{
		baseURL:    baseURL,
		procDefs:   procDefs,
		streamDefs: streamDefs,
		opts:       []internalClientOption{},
	}
}

// withHTTPClient sets a custom *http.Client.
func (b *internalClientBuilder) withHTTPClient(hc *http.Client) *internalClientBuilder {
	b.opts = append(b.opts, withHTTPClient(hc))
	return b
}

// withGlobalHeader adds a global header that will be sent with every request.
func (b *internalClientBuilder) withGlobalHeader(key, value string) *internalClientBuilder {
	b.opts = append(b.opts, withGlobalHeader(key, value))
	return b
}

// withHeaderProvider adds a dynamic header provider.
func (b *internalClientBuilder) withHeaderProvider(provider HeaderProvider) *internalClientBuilder {
	b.opts = append(b.opts, withHeaderProvider(provider))
	return b
}

// withInterceptor adds an interceptor.
func (b *internalClientBuilder) withInterceptor(interceptor Interceptor) *internalClientBuilder {
	b.opts = append(b.opts, withInterceptor(interceptor))
	return b
}

// withGlobalRetryConfig sets the global default retry configuration.
func (b *internalClientBuilder) withGlobalRetryConfig(conf RetryConfig) *internalClientBuilder {
	b.opts = append(b.opts, withGlobalRetryConfig(conf))
	return b
}

// withGlobalTimeoutConfig sets the global default timeout configuration.
func (b *internalClientBuilder) withGlobalTimeoutConfig(conf TimeoutConfig) *internalClientBuilder {
	b.opts = append(b.opts, withGlobalTimeoutConfig(conf))
	return b
}

// withGlobalReconnectConfig sets the global default reconnection configuration.
func (b *internalClientBuilder) withGlobalReconnectConfig(conf ReconnectConfig) *internalClientBuilder {
	b.opts = append(b.opts, withGlobalReconnectConfig(conf))
	return b
}

// withGlobalMaxMessageSize sets the global maximum message size for streams.
func (b *internalClientBuilder) withGlobalMaxMessageSize(size int64) *internalClientBuilder {
	b.opts = append(b.opts, withGlobalMaxMessageSize(size))
	return b
}

// Build creates the internalClient applying all accumulated options.
func (b *internalClientBuilder) Build() *internalClient {
	return newInternalClient(b.baseURL, b.procDefs, b.streamDefs, b.opts...)
}

// setRPCRetryConfig sets the default retry config for a specific RPC.
func (c *internalClient) setRPCRetryConfig(rpcName string, conf RetryConfig) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.rpcRetryConf[rpcName] = &conf
}

// setRPCTimeoutConfig sets the default timeout config for a specific RPC.
func (c *internalClient) setRPCTimeoutConfig(rpcName string, conf TimeoutConfig) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.rpcTimeoutConf[rpcName] = &conf
}

// setRPCReconnectConfig sets the default reconnect config for a specific RPC.
func (c *internalClient) setRPCReconnectConfig(rpcName string, conf ReconnectConfig) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.rpcReconnectConf[rpcName] = &conf
}

// setRPCMaxMessageSize sets the default max message size for a specific RPC.
func (c *internalClient) setRPCMaxMessageSize(rpcName string, size int64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.rpcMaxMessageSize[rpcName] = size
}

// setRPCHeaderProvider adds a header provider for a specific RPC.
func (c *internalClient) setRPCHeaderProvider(rpcName string, provider HeaderProvider) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.rpcHeaderProviders[rpcName] = append(c.rpcHeaderProviders[rpcName], provider)
}

// mergeRetryConfig resolves the retry configuration based on precedence:
// Operation > RPC > Global > Default.
func (c *internalClient) mergeRetryConfig(rpcName string, opConf *RetryConfig) *RetryConfig {
	if opConf != nil {
		return opConf
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	if rpcConf, ok := c.rpcRetryConf[rpcName]; ok {
		return rpcConf
	}
	if c.globalRetryConf != nil {
		return c.globalRetryConf
	}
	return &RetryConfig{
		MaxAttempts:     1,
		InitialDelay:    0,
		MaxDelay:        0,
		DelayMultiplier: 1.0,
		Jitter:          0.2,
	}
}

// mergeTimeoutConfig resolves the timeout configuration based on precedence:
// Operation > RPC > Global > Default.
func (c *internalClient) mergeTimeoutConfig(rpcName string, opConf *TimeoutConfig) *TimeoutConfig {
	if opConf != nil {
		return opConf
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	if rpcConf, ok := c.rpcTimeoutConf[rpcName]; ok {
		return rpcConf
	}
	if c.globalTimeoutConf != nil {
		return c.globalTimeoutConf
	}
	return &TimeoutConfig{
		Timeout: 30 * time.Second,
	}
}

// mergeReconnectConfig resolves the reconnection configuration based on precedence:
// Operation > RPC > Global > Default.
func (c *internalClient) mergeReconnectConfig(rpcName string, opConf *ReconnectConfig) *ReconnectConfig {
	if opConf != nil {
		return opConf
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	if rpcConf, ok := c.rpcReconnectConf[rpcName]; ok {
		return rpcConf
	}
	if c.globalReconnectConf != nil {
		return c.globalReconnectConf
	}
	return &ReconnectConfig{
		MaxAttempts:     30,
		InitialDelay:    1 * time.Second,
		MaxDelay:        30 * time.Second,
		DelayMultiplier: 1.5,
		Jitter:          0.2,
	}
}

// mergeMaxMessageSize resolves the max message size based on precedence:
// Operation > RPC > Global > Default (4MB).
func (c *internalClient) mergeMaxMessageSize(rpcName string, opSize int64) int64 {
	if opSize > 0 {
		return opSize
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	if rpcSize, ok := c.rpcMaxMessageSize[rpcName]; ok && rpcSize > 0 {
		return rpcSize
	}
	if c.globalMaxMessageSize > 0 {
		return c.globalMaxMessageSize
	}
	return 4 * 1024 * 1024 // Default: 4MB
}

// executeChain builds the interceptor chain and executes the final invoker.
func (c *internalClient) executeChain(ctx context.Context, req RequestInfo, final Invoker) (Response[json.RawMessage], error) {
	chain := final
	for i := len(c.interceptors) - 1; i >= 0; i-- {
		mw := c.interceptors[i]
		next := chain
		chain = func(ctx context.Context, req RequestInfo) (Response[json.RawMessage], error) {
			return mw(ctx, req, next)
		}
	}
	return chain(ctx, req)
}

// proc invokes the given procedure.
func (c *internalClient) proc(
	ctx context.Context,
	rpcName string,
	procName string,
	input any,
	opHeaderProviders []HeaderProvider,
	opRetryConf *RetryConfig,
	opTimeoutConf *TimeoutConfig,
) Response[json.RawMessage] {
	reqInfo := RequestInfo{
		RPCName:       rpcName,
		OperationName: procName,
		Input:         input,
		Type:          OperationTypeProc,
	}

	// The invoker encapsulates the actual HTTP logic
	invoker := func(ctx context.Context, req RequestInfo) (Response[json.RawMessage], error) {
		if _, ok := c.operationDefs[req.RPCName][req.OperationName]; !ok {
			return Response[json.RawMessage]{
				Ok: false,
				Error: Error{
					Category: "ClientError",
					Code:     "INVALID_PROC",
					Message:  fmt.Sprintf("%s.%s procedure not found in schema", req.RPCName, req.OperationName),
					Details:  map[string]any{"rpc": req.RPCName, "procedure": req.OperationName},
				},
			}, nil
		}

		// Resolve configurations
		retryConf := c.mergeRetryConfig(req.RPCName, opRetryConf)
		timeoutConf := c.mergeTimeoutConfig(req.RPCName, opTimeoutConf)

		// Encode input
		var payload []byte
		var err error
		if req.Input == nil {
			payload = []byte("{}")
		} else {
			payload, err = json.Marshal(req.Input)
			if err != nil {
				return Response[json.RawMessage]{
					Ok: false,
					Error: Error{
						Category: "ClientError",
						Code:     "ENCODE_INPUT",
						Message:  fmt.Sprintf("failed to marshal input for %s.%s: %v", req.RPCName, req.OperationName, err),
					},
				}, nil
			}
		}

		url := fmt.Sprintf("%s/%s/%s", c.baseURL, req.RPCName, req.OperationName)
		var lastError Error

		for attempt := 1; attempt <= retryConf.MaxAttempts; attempt++ {
			// Create context with timeout for this attempt
			attemptCtx := ctx
			var cancel context.CancelFunc
			if timeoutConf.Timeout > 0 {
				attemptCtx, cancel = context.WithTimeout(ctx, timeoutConf.Timeout)
			}

			// Perform the request
			res, shouldRetry, backoff := c.doRequest(attemptCtx, req.RPCName, url, payload, opHeaderProviders, attempt, retryConf)
			if cancel != nil {
				cancel()
			}

			if !shouldRetry {
				return res, nil
			}

			lastError = res.Error
			if attempt < retryConf.MaxAttempts {
				time.Sleep(backoff)
			}
		}

		return Response[json.RawMessage]{Ok: false, Error: lastError}, nil
	}

	res, err := c.executeChain(ctx, reqInfo, invoker)
	if err != nil {
		// Should usually not happen if interceptors behave correctly and return Response on error
		return Response[json.RawMessage]{Ok: false, Error: ToError(err)}
	}
	return res
}

// doRequest performs a single HTTP request attempt.
func (c *internalClient) doRequest(
	ctx context.Context,
	rpcName string,
	url string,
	payload []byte,
	opHeaderProviders []HeaderProvider,
	attempt int,
	retryConf *RetryConfig,
) (Response[json.RawMessage], bool, time.Duration) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return Response[json.RawMessage]{Ok: false, Error: ToError(fmt.Errorf("failed to create HTTP request: %w", err))}, false, 0
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Apply Providers in order: Global, RPC, Local (Op)

	// 1. Global
	for _, provider := range c.headerProviders {
		if err := provider(ctx, req.Header); err != nil {
			return Response[json.RawMessage]{Ok: false, Error: ToError(fmt.Errorf("global header provider failed: %w", err))}, false, 0
		}
	}

	// 2. RPC
	// Fetch providers under lock to avoid concurrent map read/write
	var rpcProviders []HeaderProvider
	c.mu.RLock()
	if providers, ok := c.rpcHeaderProviders[rpcName]; ok {
		// Copy the slice header (not the underlying array, which is fine as functions are immutable)
		// We copy to avoid holding the lock during provider execution.
		// NOTE: If providers are appended concurrently, we might miss new ones, which is acceptable for a "snapshot" read.
		rpcProviders = providers
	}
	c.mu.RUnlock()

	for _, provider := range rpcProviders {
		if err := provider(ctx, req.Header); err != nil {
			return Response[json.RawMessage]{Ok: false, Error: ToError(fmt.Errorf("rpc header provider failed: %w", err))}, false, 0
		}
	}

	// 3. Local (Op)
	for _, provider := range opHeaderProviders {
		if err := provider(ctx, req.Header); err != nil {
			return Response[json.RawMessage]{Ok: false, Error: ToError(fmt.Errorf("operation header provider failed: %w", err))}, false, 0
		}
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		var lastError Error
		if ctx.Err() == context.DeadlineExceeded {
			lastError = Error{
				Category: "TimeoutError",
				Code:     "REQUEST_TIMEOUT",
				Message:  "Request timeout",
				Details:  map[string]any{"attempt": attempt},
			}
		} else {
			lastError = ToError(fmt.Errorf("http request failed: %w", err))
		}
		return Response[json.RawMessage]{Ok: false, Error: lastError}, true, calculateBackoff(retryConf, attempt)
	}
	defer resp.Body.Close()

	// Retry on 5xx
	if resp.StatusCode >= 500 {
		return Response[json.RawMessage]{
			Ok: false,
			Error: Error{
				Category: "HTTPError",
				Code:     "BAD_STATUS",
				Message:  fmt.Sprintf("unexpected HTTP status: %s", resp.Status),
				Details:  map[string]any{"status": resp.StatusCode},
			},
		}, true, calculateBackoff(retryConf, attempt)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return Response[json.RawMessage]{
			Ok: false,
			Error: Error{
				Category: "HTTPError",
				Code:     "BAD_STATUS",
				Message:  fmt.Sprintf("unexpected HTTP status: %s", resp.Status),
				Details:  map[string]any{"status": resp.StatusCode},
			},
		}, false, 0
	}

	var raw struct {
		Ok     bool            `json:"ok"`
		Output json.RawMessage `json:"output"`
		Error  Error           `json:"error"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return Response[json.RawMessage]{
			Ok:    false,
			Error: ToError(fmt.Errorf("failed to decode VDL response: %w", err)),
		}, false, 0
	}

	if !raw.Ok {
		return Response[json.RawMessage]{Ok: false, Error: raw.Error}, false, 0
	}

	return Response[json.RawMessage]{Ok: true, Output: raw.Output}, false, 0
}

// stream establishes a SSE subscription.
func (c *internalClient) stream(
	ctx context.Context,
	rpcName string,
	streamName string,
	input any,
	opHeaderProviders []HeaderProvider,
	opReconnectConf *ReconnectConfig,
	opMaxMessageSize int64,
	onConnect func(),
	onDisconnect func(error),
	onReconnect func(int, time.Duration),
) <-chan Response[json.RawMessage] {
	events := make(chan Response[json.RawMessage])

	// We use the interceptor chain only for the initial connection attempt logic if we wanted to wrapping the whole lifecycle.
	// However, streams are long-lived and self-healing. Interceptors typically wrap the "request" concept.
	// For simplicity and to match the requirement "The interceptor wraps the initiation of the connection",
	// we will run the chain ONCE to validate the request and potentially modify it, but the actual loop runs inside the invoker.

	reqInfo := RequestInfo{
		RPCName:       rpcName,
		OperationName: streamName,
		Input:         input,
		Type:          OperationTypeStream,
	}

	invoker := func(ctx context.Context, req RequestInfo) (Response[json.RawMessage], error) {
		// This invoker spawns the goroutine that handles the stream.
		// It returns immediately with an empty response indicating the stream started "successfully" from the perspective of the interceptor.
		// Real errors are sent through the channel.

		if _, ok := c.operationDefs[req.RPCName][req.OperationName]; !ok {
			err := Error{
				Category: "ClientError",
				Code:     "INVALID_STREAM",
				Message:  fmt.Sprintf("%s.%s stream not found in schema", req.RPCName, req.OperationName),
				Details:  map[string]any{"rpc": req.RPCName, "stream": req.OperationName},
			}
			events <- Response[json.RawMessage]{Ok: false, Error: err}
			close(events)
			if onDisconnect != nil {
				onDisconnect(err)
			}
			return Response[json.RawMessage]{}, nil
		}

		reconnectConf := c.mergeReconnectConfig(req.RPCName, opReconnectConf)
		maxMessageSize := c.mergeMaxMessageSize(req.RPCName, opMaxMessageSize)

		// Encode input
		var payload []byte
		var err error
		if req.Input == nil {
			payload = []byte("{}")
		} else {
			payload, err = json.Marshal(req.Input)
			if err != nil {
				errVal := ToError(fmt.Errorf("failed to marshal input: %w", err))
				events <- Response[json.RawMessage]{Ok: false, Error: errVal}
				close(events)
				if onDisconnect != nil {
					onDisconnect(errVal)
				}
				return Response[json.RawMessage]{}, nil
			}
		}

		url := fmt.Sprintf("%s/%s/%s", c.baseURL, req.RPCName, req.OperationName)

		go func() {
			defer close(events)
			var finalErr error
			defer func() {
				if onDisconnect != nil {
					onDisconnect(finalErr)
				}
			}()

			reconnectAttempt := 0
			for {
				select {
				case <-ctx.Done():
					finalErr = ctx.Err()
					return
				default:
				}

				reqHTTP, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
				if err != nil {
					finalErr = err
					events <- Response[json.RawMessage]{Ok: false, Error: ToError(err)}
					return
				}

				reqHTTP.Header.Set("Content-Type", "application/json")
				reqHTTP.Header.Set("Accept", "text/event-stream")

				// Apply Providers in order: Global, RPC, Local (Op)

				// 1. Global
				for _, provider := range c.headerProviders {
					if err := provider(ctx, reqHTTP.Header); err != nil {
						finalErr = err
						events <- Response[json.RawMessage]{Ok: false, Error: ToError(err)}
						return
					}
				}

				// 2. RPC
				var rpcProviders []HeaderProvider
				c.mu.RLock()
				if providers, ok := c.rpcHeaderProviders[rpcName]; ok {
					rpcProviders = providers
				}
				c.mu.RUnlock()

				for _, provider := range rpcProviders {
					if err := provider(ctx, reqHTTP.Header); err != nil {
						finalErr = err
						events <- Response[json.RawMessage]{Ok: false, Error: ToError(err)}
						return
					}
				}

				// 3. Local (Op)
				for _, provider := range opHeaderProviders {
					if err := provider(ctx, reqHTTP.Header); err != nil {
						finalErr = err
						events <- Response[json.RawMessage]{Ok: false, Error: ToError(err)}
						return
					}
				}

				resp, err := c.httpClient.Do(reqHTTP)
				if err != nil {
					if reconnectAttempt < reconnectConf.MaxAttempts {
						if onReconnect != nil {
							onReconnect(reconnectAttempt+1, calculateReconnectBackoff(reconnectConf, reconnectAttempt+1))
						}
						reconnectAttempt++
						time.Sleep(calculateReconnectBackoff(reconnectConf, reconnectAttempt))
						continue
					}
					finalErr = err
					events <- Response[json.RawMessage]{Ok: false, Error: ToError(err)}
					return
				}

				if resp.StatusCode >= 500 && reconnectAttempt < reconnectConf.MaxAttempts {
					resp.Body.Close()
					if onReconnect != nil {
						onReconnect(reconnectAttempt+1, calculateReconnectBackoff(reconnectConf, reconnectAttempt+1))
					}
					reconnectAttempt++
					time.Sleep(calculateReconnectBackoff(reconnectConf, reconnectAttempt))
					continue
				}

				if resp.StatusCode < 200 || resp.StatusCode >= 300 {
					resp.Body.Close()
					finalErr = fmt.Errorf("bad status: %s", resp.Status)
					events <- Response[json.RawMessage]{
						Ok: false,
						Error: Error{
							Category: "HTTPError",
							Code:     "BAD_STATUS",
							Message:  fmt.Sprintf("unexpected HTTP status: %s", resp.Status),
							Details:  map[string]any{"status": resp.StatusCode},
						},
					}
					return
				}

				// Connected successfully
				if onConnect != nil {
					onConnect()
				}
				reconnectAttempt = 0

				hadError := handleStreamEvents(ctx, resp, maxMessageSize, events)
				resp.Body.Close()

				if hadError && reconnectAttempt < reconnectConf.MaxAttempts {
					if onReconnect != nil {
						onReconnect(reconnectAttempt+1, calculateReconnectBackoff(reconnectConf, reconnectAttempt+1))
					}
					reconnectAttempt++
					time.Sleep(calculateReconnectBackoff(reconnectConf, reconnectAttempt))
					continue
				}

				// Normal closure or max attempts reached
				if hadError {
					finalErr = fmt.Errorf("stream interrupted")
				}
				return
			}
		}()

		return Response[json.RawMessage]{Ok: true}, nil
	}

	_, _ = c.executeChain(ctx, reqInfo, invoker)
	return events
}

// handleStreamEvents handles the SSE stream processing.
func handleStreamEvents(
	ctx context.Context,
	resp *http.Response,
	maxMessageSize int64,
	events chan<- Response[json.RawMessage],
) (hadError bool) {
	scanner := bufio.NewScanner(resp.Body)
	// Set the scanner buffer to handle large tokens up to maxMessageSize
	scanner.Buffer(make([]byte, 4096), int(maxMessageSize))

	var dataBuf bytes.Buffer

	flush := func() {
		if dataBuf.Len() == 0 {
			return
		}
		var evt Response[json.RawMessage]
		if err := json.Unmarshal(dataBuf.Bytes(), &evt); err != nil {
			events <- Response[json.RawMessage]{
				Ok:    false,
				Error: ToError(fmt.Errorf("received invalid SSE payload: %v", err)),
			}
			return
		}
		select {
		case events <- evt:
		case <-ctx.Done():
		}
		dataBuf.Reset()
	}

	for {
		select {
		case <-ctx.Done():
			return false
		default:
		}

		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				// If the error is due to token size, it is a fatal error, do not reconnect
				if err == bufio.ErrTooLong {
					events <- Response[json.RawMessage]{
						Ok: false,
						Error: Error{
							Category: "ProtocolError",
							Code:     "MESSAGE_TOO_LARGE",
							Message:  fmt.Sprintf("Stream message exceeded maximum size of %d bytes", maxMessageSize),
						},
					}
					return false // Fatal error, no reconnect
				}
				hadError = true
			}
			return
		}
		line := scanner.Text()
		if line == "" {
			flush()
			continue
		}
		if strings.HasPrefix(line, ":") {
			continue
		}
		if strings.HasPrefix(line, "data:") {
			chunk := strings.TrimSpace(strings.TrimPrefix(line, "data:"))

			// Check if accumulating this chunk would exceed the limit
			if int64(dataBuf.Len()+len(chunk)) > maxMessageSize {
				events <- Response[json.RawMessage]{
					Ok: false,
					Error: Error{
						Category: "ProtocolError",
						Code:     "MESSAGE_TOO_LARGE",
						Message:  fmt.Sprintf("Stream message accumulation exceeded maximum size of %d bytes", maxMessageSize),
					},
				}
				return false // Fatal error, no reconnect
			}

			dataBuf.WriteString(chunk)
		}
	}
}

// calculateBackoff calculates the backoff delay for retry attempts.
func calculateBackoff(config *RetryConfig, attempt int) time.Duration {
	delay := config.InitialDelay
	for i := 1; i < attempt; i++ {
		delay = time.Duration(float64(delay) * config.DelayMultiplier)
	}
	if delay > config.MaxDelay {
		delay = config.MaxDelay
	}
	return applyJitter(delay, config.Jitter)
}

// calculateReconnectBackoff calculates the backoff delay for reconnection attempts.
func calculateReconnectBackoff(config *ReconnectConfig, attempt int) time.Duration {
	delay := config.InitialDelay
	for i := 1; i < attempt; i++ {
		delay = time.Duration(float64(delay) * config.DelayMultiplier)
	}
	if delay > config.MaxDelay {
		delay = config.MaxDelay
	}
	return applyJitter(delay, config.Jitter)
}

// applyJitter applies a random jitter to the duration.
// jitterFactor is a fraction (e.g., 0.2 for 20%).
func applyJitter(d time.Duration, jitterFactor float64) time.Duration {
	if jitterFactor <= 0 {
		return d
	}

	// Clamp jitterFactor to [0.0, 1.0] for safety
	jitterFactor = min(jitterFactor, 1.0)

	// Range: [d * (1 - jitter), d * (1 + jitter)]
	delta := float64(d) * jitterFactor
	min := float64(d) - delta
	max := float64(d) + delta

	// Ensure min is not negative (though d should be positive)
	if min < 0 {
		min = 0
	}

	random := rand.Float64() // 0.0 to 1.0
	result := min + (random * (max - min))
	return time.Duration(result)
}

// procCallBuilder is a fluent builder for invoking a procedure.
type procCallBuilder struct {
	client          *internalClient
	rpcName         string
	name            string
	input           any
	headerProviders []HeaderProvider
	retryConf       *RetryConfig
	timeoutConf     *TimeoutConfig
}

// withHeader adds a header to this procedure invocation.
func (p *procCallBuilder) withHeader(key, value string) *procCallBuilder {
	p.headerProviders = append(p.headerProviders, func(_ context.Context, h http.Header) error {
		h.Set(key, value)
		return nil
	})
	return p
}

// withHeaderProvider adds a dynamic header provider to this procedure invocation.
func (p *procCallBuilder) withHeaderProvider(provider HeaderProvider) *procCallBuilder {
	p.headerProviders = append(p.headerProviders, provider)
	return p
}

// withRetryConfig sets the retry configuration for this procedure call.
func (p *procCallBuilder) withRetryConfig(retryConfig RetryConfig) *procCallBuilder {
	p.retryConf = &retryConfig
	return p
}

// withTimeoutConfig sets the timeout configuration for this procedure call.
func (p *procCallBuilder) withTimeoutConfig(timeoutConfig TimeoutConfig) *procCallBuilder {
	p.timeoutConf = &timeoutConfig
	return p
}

// execute performs the RPC call and returns the Response.
func (p *procCallBuilder) execute(ctx context.Context) Response[json.RawMessage] {
	return p.client.proc(ctx, p.rpcName, p.name, p.input, p.headerProviders, p.retryConf, p.timeoutConf)
}

// newProcCallBuilder creates a builder for calling the given procedure.
func (c *internalClient) newProcCallBuilder(rpcName, name string, input any) *procCallBuilder {
	return &procCallBuilder{
		client:          c,
		rpcName:         rpcName,
		name:            name,
		input:           input,
		headerProviders: []HeaderProvider{},
		retryConf:       nil,
		timeoutConf:     nil,
	}
}

// streamCall is a fluent builder for SSE subscriptions.
type streamCall struct {
	client          *internalClient
	rpcName         string
	name            string
	input           any
	headerProviders []HeaderProvider
	reconnectConf   *ReconnectConfig
	maxMessageSize  int64
	onConnect       func()
	onDisconnect    func(error)
	onReconnect     func(int, time.Duration)
}

// withHeader adds a header to this stream invocation.
func (s *streamCall) withHeader(key, value string) *streamCall {
	s.headerProviders = append(s.headerProviders, func(_ context.Context, h http.Header) error {
		h.Set(key, value)
		return nil
	})
	return s
}

// withHeaderProvider adds a dynamic header provider to this stream invocation.
func (s *streamCall) withHeaderProvider(provider HeaderProvider) *streamCall {
	s.headerProviders = append(s.headerProviders, provider)
	return s
}

// withReconnectConfig sets the reconnection configuration for this stream.
func (s *streamCall) withReconnectConfig(reconnectConfig ReconnectConfig) *streamCall {
	s.reconnectConf = &reconnectConfig
	return s
}

// withMaxMessageSize sets the maximum message size for this stream.
func (s *streamCall) withMaxMessageSize(size int64) *streamCall {
	s.maxMessageSize = size
	return s
}

// withOnConnect sets the callback for when the stream connects successfully.
func (s *streamCall) withOnConnect(cb func()) *streamCall {
	s.onConnect = cb
	return s
}

// withOnDisconnect sets the callback for when the stream disconnects permanently.
func (s *streamCall) withOnDisconnect(cb func(error)) *streamCall {
	s.onDisconnect = cb
	return s
}

// withOnReconnect sets the callback for when the stream is about to retry a connection.
func (s *streamCall) withOnReconnect(cb func(int, time.Duration)) *streamCall {
	s.onReconnect = cb
	return s
}

// execute starts the stream and returns the channel of events.
func (s *streamCall) execute(ctx context.Context) <-chan Response[json.RawMessage] {
	return s.client.stream(
		ctx,
		s.rpcName,
		s.name,
		s.input,
		s.headerProviders,
		s.reconnectConf,
		s.maxMessageSize,
		s.onConnect,
		s.onDisconnect,
		s.onReconnect,
	)
}

// newStreamCallBuilder creates a builder for the given stream.
func (c *internalClient) newStreamCallBuilder(rpcName, name string, input any) *streamCall {
	return &streamCall{
		client:          c,
		rpcName:         rpcName,
		name:            name,
		input:           input,
		headerProviders: []HeaderProvider{},
		reconnectConf:   nil,
		maxMessageSize:  0,
	}
}

//nolint:unused
package pieces

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

/** START FROM HERE **/

// -----------------------------------------------------------------------------
// Client types
// -----------------------------------------------------------------------------

// RetryConfig defines retry behavior for procedure calls.
type RetryConfig struct {
	// Maximum number of retry attempts (default: 3)
	MaxAttempts int
	// Initial delay between retries (default: 1 second)
	InitialDelay time.Duration
	// Maximum delay between retries (default: 5 seconds)
	MaxDelay time.Duration
	// Cumulative multiplier applied to initialDelayMs on each retry (default: 2.0)
	DelayMultiplier float64
}

// TimeoutConfig defines timeout behavior for procedure calls.
type TimeoutConfig struct {
	// Request Timeout (default: 30 seconds)
	Timeout time.Duration
}

// ReconnectConfig defines reconnection behavior for stream calls.
type ReconnectConfig struct {
	// Maximum number of reconnection attempts (default: 5)
	MaxAttempts int
	// Initial delay between reconnection attempts (default: 1 second)
	InitialDelay time.Duration
	// Maximum delay between reconnection attempts (default: 5 seconds)
	MaxDelay time.Duration
	// Cumulative multiplier applied to initialDelayMs on each retry (default: 2.0)
	DelayMultiplier float64
}

// internalClient is the core engine used by the generated client abstraction. It is
// thread-safe and can be reused across concurrent requests.
//
// The client validates the requested operation name against the schema to fail
// fast when a typo occurs in the generated wrapper.
//
// The zero value is not usable – use newInternalClient to construct one.
type internalClient struct {
	// Immutable after construction.
	baseURL        string
	httpClient     *http.Client
	procNames      []string
	procNamesMap   map[string]bool
	streamNames    []string
	streamNamesMap map[string]bool

	// header configuration (global on every request)
	globalHeaders map[string]string
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
		c.globalHeaders[key] = value
	}
}

// newInternalClient creates a new internalClient capable of talking to the UFO
// RPC server described by procNames and streamNames.
//
// The caller can optionally pass functional options to tweak the configuration
// (base URL, custom *http.Client, …).
func newInternalClient(
	baseURL string,
	procNames []string,
	streamNames []string,
	opts ...internalClientOption,
) *internalClient {
	procMap := make(map[string]bool, len(procNames))
	for _, n := range procNames {
		procMap[n] = true
	}
	streamMap := make(map[string]bool, len(streamNames))
	for _, n := range streamNames {
		streamMap[n] = true
	}

	cli := &internalClient{
		baseURL:        strings.TrimRight(baseURL, "/"),
		httpClient:     http.DefaultClient,
		procNames:      procNames,
		procNamesMap:   procMap,
		streamNames:    streamNames,
		streamNamesMap: streamMap,
		globalHeaders:  map[string]string{},
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
	baseURL     string
	procNames   []string
	streamNames []string
	opts        []internalClientOption
}

// newClientBuilder creates a builder with the schema information (procedure and
// stream names). Generated code will pass the automatically produced slices.
func newClientBuilder(baseURL string, procNames, streamNames []string) *internalClientBuilder {
	return &internalClientBuilder{
		baseURL:     baseURL,
		procNames:   procNames,
		streamNames: streamNames,
		opts:        []internalClientOption{},
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

// Build creates the internalClient applying all accumulated options.
func (b *internalClientBuilder) Build() *internalClient {
	return newInternalClient(b.baseURL, b.procNames, b.streamNames, b.opts...)
}

// proc invokes the given procedure with the provided input and returns the
// raw JSON response from the server wrapped in a Response object.
//
// Any transport or decoding error is converted into a Response with Ok set to
// false and the Error field describing the failure.
//
// This method implements retry logic with exponential backoff and timeout handling.
func (c *internalClient) proc(
	ctx context.Context,
	procName string,
	input any,
	extraHeaders map[string]string,
	retryConf *RetryConfig,
	timeoutConf *TimeoutConfig,
) Response[json.RawMessage] {
	if !c.procNamesMap[procName] {
		return Response[json.RawMessage]{
			Ok: false,
			Error: Error{
				Category: "ClientError",
				Code:     "INVALID_PROC",
				Message:  fmt.Sprintf("%s procedure not found in schema", procName),
				Details:  map[string]any{"procedure": procName},
			},
		}
	}

	// Default configurations if not provided
	if retryConf == nil {
		retryConf = &RetryConfig{
			MaxAttempts:     3,
			InitialDelay:    1 * time.Second,
			MaxDelay:        5 * time.Second,
			DelayMultiplier: 2.0,
		}
	}
	if timeoutConf == nil {
		timeoutConf = &TimeoutConfig{
			Timeout: 30 * time.Second,
		}
	}

	// Encode the input.
	var payload []byte
	var err error
	if input == nil {
		payload = []byte("{}")
	} else {
		payload, err = json.Marshal(input)
		if err != nil {
			return Response[json.RawMessage]{
				Ok: false,
				Error: Error{
					Category: "ClientError",
					Code:     "ENCODE_INPUT",
					Message:  fmt.Sprintf("failed to marshal input for %s: %v", procName, err),
				},
			}
		}
	}

	// Build URL – <baseURL>/<procName> . Leading slash added if missing.
	url := c.baseURL + "/" + procName

	var lastError Error
	for attempt := 1; attempt <= retryConf.MaxAttempts; attempt++ {
		// Create context with timeout for this attempt
		attemptCtx := ctx
		var cancel context.CancelFunc
		if timeoutConf.Timeout > 0 {
			attemptCtx, cancel = context.WithTimeout(ctx, timeoutConf.Timeout)
		}

		req, err := http.NewRequestWithContext(attemptCtx, http.MethodPost, url, bytes.NewReader(payload))
		if err != nil {
			if cancel != nil {
				cancel()
			}
			return Response[json.RawMessage]{
				Ok:    false,
				Error: asError(fmt.Errorf("failed to create HTTP request: %w", err)),
			}
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")

		// Apply headers: global + per-call extras.
		for key, value := range c.globalHeaders {
			req.Header.Set(key, value)
		}
		for key, value := range extraHeaders {
			req.Header.Set(key, value)
		}

		resp, err := c.httpClient.Do(req)
		if cancel != nil {
			cancel()
		}

		if err != nil {
			// Check if this was a timeout error
			if attemptCtx.Err() == context.DeadlineExceeded {
				lastError = Error{
					Category: "TimeoutError",
					Code:     "REQUEST_TIMEOUT",
					Message:  fmt.Sprintf("Request timeout after %s", timeoutConf.Timeout),
					Details:  map[string]any{"timeout": timeoutConf.Timeout, "attempt": attempt},
				}
			} else {
				lastError = asError(fmt.Errorf("http request failed: %w", err))
			}

			// Retry on timeout or network errors if we have attempts left
			if attempt < retryConf.MaxAttempts {
				backoff := calculateBackoff(retryConf, attempt)
				time.Sleep(backoff)
				continue
			}

			return Response[json.RawMessage]{
				Ok:    false,
				Error: lastError,
			}
		}

		// Check status code
		if resp.StatusCode >= 500 && attempt < retryConf.MaxAttempts {
			// Retry on 5xx errors
			resp.Body.Close()
			lastError = Error{
				Category: "HTTPError",
				Code:     "BAD_STATUS",
				Message:  fmt.Sprintf("unexpected HTTP status: %s", resp.Status),
				Details:  map[string]any{"status": resp.StatusCode},
			}
			backoff := calculateBackoff(retryConf, attempt)
			time.Sleep(backoff)
			continue
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			resp.Body.Close()
			return Response[json.RawMessage]{
				Ok: false,
				Error: Error{
					Category: "HTTPError",
					Code:     "BAD_STATUS",
					Message:  fmt.Sprintf("unexpected HTTP status: %s", resp.Status),
					Details:  map[string]any{"status": resp.StatusCode},
				},
			}
		}

		// Decode the generic response first so that we can decide what to do next.
		var raw struct {
			Ok     bool            `json:"ok"`
			Output json.RawMessage `json:"output"`
			Error  Error           `json:"error"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
			resp.Body.Close()
			return Response[json.RawMessage]{
				Ok:    false,
				Error: asError(fmt.Errorf("failed to decode UFO RPC response: %w", err)),
			}
		}
		resp.Body.Close()

		if !raw.Ok {
			return Response[json.RawMessage]{
				Ok:    false,
				Error: raw.Error,
			}
		}

		return Response[json.RawMessage]{
			Ok:     true,
			Output: raw.Output,
		}
	}

	// This should never be reached, but just in case
	return Response[json.RawMessage]{
		Ok:    false,
		Error: lastError,
	}
}

// calculateBackoff calculates the backoff delay for retry attempts.
func calculateBackoff(config *RetryConfig, attempt int) time.Duration {
	delay := config.InitialDelay
	for i := 1; i < attempt; i++ {
		delay = time.Duration(float64(delay) * config.DelayMultiplier)
	}
	if delay > config.MaxDelay {
		return config.MaxDelay
	}
	return delay
}

// calculateReconnectBackoff calculates the backoff delay for reconnection attempts.
func calculateReconnectBackoff(config *ReconnectConfig, attempt int) time.Duration {
	delay := config.InitialDelay
	for i := 1; i < attempt; i++ {
		delay = time.Duration(float64(delay) * config.DelayMultiplier)
	}
	if delay > config.MaxDelay {
		return config.MaxDelay
	}
	return delay
}

// procCallBuilder is a fluent builder for invoking a procedure.
type procCallBuilder struct {
	client      *internalClient
	name        string
	input       any
	headers     map[string]string
	retryConf   *RetryConfig
	timeoutConf *TimeoutConfig
}

// withHeader adds a header to this procedure invocation.
func (p *procCallBuilder) withHeader(key, value string) *procCallBuilder {
	p.headers[key] = value
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
	return p.client.proc(ctx, p.name, p.input, p.headers, p.retryConf, p.timeoutConf)
}

// newProcCallBuilder creates a builder for calling the given procedure.
func (c *internalClient) newProcCallBuilder(name string, input any) *procCallBuilder {
	return &procCallBuilder{
		client:      c,
		name:        name,
		input:       input,
		headers:     map[string]string{},
		retryConf:   nil,
		timeoutConf: nil,
	}
}

// stream establishes a Server-Sent Events subscription for the given stream
// name. Each received event is forwarded on the returned channel until ctx is
// cancelled or the server closes the connection.
//
// The channel is closed on termination and MUST be fully drained by the caller
// to avoid goroutine leaks.
//
// This method implements automatic reconnection with exponential backoff.
func (c *internalClient) stream(
	ctx context.Context,
	streamName string,
	input any,
	extraHeaders map[string]string,
	reconnectConf *ReconnectConfig,
) <-chan Response[json.RawMessage] {
	if !c.streamNamesMap[streamName] {
		ch := make(chan Response[json.RawMessage], 1)
		ch <- Response[json.RawMessage]{
			Ok: false,
			Error: Error{
				Category: "ClientError",
				Code:     "INVALID_STREAM",
				Message:  fmt.Sprintf("%s stream not found in schema", streamName),
				Details:  map[string]any{"stream": streamName},
			},
		}
		close(ch)
		return ch
	}

	// Encode input.
	var payload []byte
	var err error
	if input == nil {
		payload = []byte("{}")
	} else {
		payload, err = json.Marshal(input)
		if err != nil {
			ch := make(chan Response[json.RawMessage], 1)
			ch <- Response[json.RawMessage]{
				Ok:    false,
				Error: asError(fmt.Errorf("failed to marshal input for %s: %w", streamName, err)),
			}
			close(ch)
			return ch
		}
	}

	// Default configuration if not provided
	if reconnectConf == nil {
		reconnectConf = &ReconnectConfig{
			MaxAttempts:     5,
			InitialDelay:    1 * time.Second,
			MaxDelay:        5 * time.Second,
			DelayMultiplier: 2.0,
		}
	}

	// Build URL – <baseURL>/<streamName>
	url := c.baseURL + "/" + streamName

	// Channel for events.
	events := make(chan Response[json.RawMessage])

	go func() {
		defer close(events)

		reconnectAttempt := 0
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
			if err != nil {
				events <- Response[json.RawMessage]{
					Ok:    false,
					Error: asError(fmt.Errorf("failed to create HTTP request: %w", err)),
				}
				return
			}

			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Accept", "text/event-stream")

			// Apply headers: global + per-call extras.
			for key, value := range c.globalHeaders {
				req.Header.Set(key, value)
			}
			for key, value := range extraHeaders {
				req.Header.Set(key, value)
			}

			resp, err := c.httpClient.Do(req)
			if err != nil {
				// Try to reconnect if enabled and not manually cancelled
				if reconnectConf != nil && reconnectAttempt < reconnectConf.MaxAttempts {
					events <- Response[json.RawMessage]{
						Ok: false,
						Error: Error{
							Category: "ConnectionError",
							Code:     "STREAM_CONNECT_FAILED",
							Message:  fmt.Sprintf("Failed to connect to stream, attempting reconnect (%d/%d)", reconnectAttempt+1, reconnectConf.MaxAttempts),
							Details:  map[string]any{"stream": streamName, "attempt": reconnectAttempt + 1},
						},
					}
					reconnectAttempt++
					backoff := calculateReconnectBackoff(reconnectConf, reconnectAttempt)
					time.Sleep(backoff)
					continue
				}

				events <- Response[json.RawMessage]{
					Ok:    false,
					Error: asError(fmt.Errorf("stream request failed: %w", err)),
				}
				return
			}

			if resp.StatusCode < 200 || resp.StatusCode >= 300 {
				resp.Body.Close()

				// Reconnect only on 5xx HTTP errors
				if resp.StatusCode >= 500 && reconnectConf != nil && reconnectAttempt < reconnectConf.MaxAttempts {
					events <- Response[json.RawMessage]{
						Ok: false,
						Error: Error{
							Category: "HTTPError",
							Code:     "BAD_STATUS",
							Message:  fmt.Sprintf("Stream connection failed with status %s, attempting reconnect (%d/%d)", resp.Status, reconnectAttempt+1, reconnectConf.MaxAttempts),
							Details:  map[string]any{"status": resp.StatusCode, "attempt": reconnectAttempt + 1},
						},
					}
					reconnectAttempt++
					backoff := calculateReconnectBackoff(reconnectConf, reconnectAttempt)
					time.Sleep(backoff)
					continue
				}

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

			// Reset reconnect attempt counter on successful connection
			reconnectAttempt = 0

			// Process the stream
			hadError := handleStreamEvents(ctx, resp, events)
			resp.Body.Close()

			// If we reach here, the stream ended. Reconnect only on network/read errors.
			if reconnectConf != nil && hadError && reconnectAttempt < reconnectConf.MaxAttempts {
				events <- Response[json.RawMessage]{
					Ok: false,
					Error: Error{
						Category: "ConnectionError",
						Code:     "STREAM_INTERRUPTED",
						Message:  fmt.Sprintf("Stream connection lost, attempting reconnect (%d/%d)", reconnectAttempt+1, reconnectConf.MaxAttempts),
						Details:  map[string]any{"stream": streamName, "attempt": reconnectAttempt + 1},
					},
				}
				reconnectAttempt++
				backoff := calculateReconnectBackoff(reconnectConf, reconnectAttempt)
				time.Sleep(backoff)
				continue
			}

			// No more reconnect attempts, exit
			return
		}
	}()

	return events
}

// handleStreamEvents handles the SSE stream processing without size limitations.
func handleStreamEvents(
	ctx context.Context,
	resp *http.Response,
	events chan<- Response[json.RawMessage],
) (hadError bool) {
	// Use a large buffer with no maximum size limit for SSE events
	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 64*1024), bufio.MaxScanTokenSize)

	var dataBuf bytes.Buffer

	flush := func() {
		if dataBuf.Len() == 0 {
			return
		}
		var evt Response[json.RawMessage]
		if err := json.Unmarshal(dataBuf.Bytes(), &evt); err != nil {
			// Protocol violation – stop the stream without reconnect
			events <- Response[json.RawMessage]{
				Ok:    false,
				Error: asError(fmt.Errorf("received invalid SSE payload: %v", err)),
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
			return
		default:
		}

		if !scanner.Scan() {
			// EOF or error
			if err := scanner.Err(); err != nil {
				// Network/read error: indicate reconnection
				hadError = true
			} else {
				// Normal EOF: no reconnection
				hadError = false
			}
			return
		}
		line := scanner.Text()
		if line == "" { // Blank line marks end of event.
			flush()
			continue
		}
		if strings.HasPrefix(line, "data:") {
			// Strip the "data:" prefix and optional leading space.
			chunk := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
			dataBuf.WriteString(chunk)
		}
		// Everything else is ignored (e.g., id:, retry:, …).
	}
}

// streamCall is a fluent builder for SSE subscriptions.
type streamCall struct {
	client        *internalClient
	name          string
	input         any
	headers       map[string]string
	reconnectConf *ReconnectConfig
}

// withHeader adds a header to this stream invocation.
func (s *streamCall) withHeader(key, value string) *streamCall {
	s.headers[key] = value
	return s
}

// withReconnectConfig sets the reconnection configuration for this stream.
func (s *streamCall) withReconnectConfig(reconnectConfig ReconnectConfig) *streamCall {
	s.reconnectConf = &reconnectConfig
	return s
}

// execute starts the stream and returns the channel of events.
func (s *streamCall) execute(ctx context.Context) <-chan Response[json.RawMessage] {
	return s.client.stream(ctx, s.name, s.input, s.headers, s.reconnectConf)
}

// newStreamCallBuilder creates a builder for the given stream.
func (c *internalClient) newStreamCallBuilder(name string, input any) *streamCall {
	return &streamCall{
		client:        c,
		name:          name,
		input:         input,
		headers:       map[string]string{},
		reconnectConf: nil,
	}
}

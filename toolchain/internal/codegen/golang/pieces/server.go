//nolint:unused
package pieces

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

/** START FROM HERE **/

// -----------------------------------------------------------------------------
// Server Types
// -----------------------------------------------------------------------------

// HTTPAdapter defines the interface required by VDL server to handle
// incoming HTTP requests and write responses to clients. This abstraction allows
// the server to work with different HTTP frameworks while maintaining the same
// core functionality.
//
// Implementations must provide methods to read request bodies, set response headers,
// write response data, and flush the response buffer to ensure immediate delivery
// to the client.
type HTTPAdapter interface {
	// RequestBody returns the body reader for the incoming HTTP request.
	// The returned io.Reader allows the server to read the request payload
	// containing RPC call data.
	RequestBody() io.Reader

	// SetHeader sets a response header with the specified key-value pair.
	// This is used to configure response headers like Content-Type and
	// caching directives for both procedure and stream responses.
	SetHeader(key, value string)

	// Write writes the provided data to the response body.
	// Returns the number of bytes written and any error encountered.
	// For procedures, this writes the complete JSON response. For streams,
	// this writes individual Server-Sent Events data chunks.
	Write(data []byte) (int, error)

	// Flush immediately sends any buffered response data to the client.
	// This is crucial for streaming responses to ensure real-time delivery
	// of events. Returns an error if the flush operation fails.
	Flush() error
}

// NetHTTPAdapter implements HTTPAdapter for Go's standard net/http package.
// This adapter bridges the VDL server with the standard HTTP library, allowing
// seamless integration with existing HTTP servers and middlewares.
type NetHTTPAdapter struct {
	responseWriter http.ResponseWriter
	request        *http.Request
}

// NewNetHTTPAdapter creates a new NetHTTPAdapter that implements the
// HTTPAdapter interface for net/http.
//
// Parameters:
//   - w: The http.ResponseWriter to write responses to
//   - r: The *http.Request containing the incoming request data
//
// Returns a HTTPAdapter implementation ready for use with VDL server.
func NewNetHTTPAdapter(w http.ResponseWriter, r *http.Request) HTTPAdapter {
	return &NetHTTPAdapter{
		responseWriter: w,
		request:        r,
	}
}

// RequestBody returns the body reader for the HTTP request.
// This provides access to the request payload containing the RPC call data.
func (r *NetHTTPAdapter) RequestBody() io.Reader {
	return r.request.Body
}

// SetHeader sets a response header with the specified key-value pair.
// This configures headers for the HTTP response, such as Content-Type
// for JSON responses or streaming-specific headers.
func (r *NetHTTPAdapter) SetHeader(key, value string) {
	r.responseWriter.Header().Set(key, value)
}

// Write writes the provided data to the HTTP response body.
// Returns the number of bytes written and any error encountered during writing.
func (r *NetHTTPAdapter) Write(data []byte) (int, error) {
	return r.responseWriter.Write(data)
}

// Flush immediately sends any buffered response data to the client.
// For streaming responses, this ensures real-time delivery of events.
// If the underlying ResponseWriter doesn't support flushing, this is a no-op.
func (r *NetHTTPAdapter) Flush() error {
	if f, ok := r.responseWriter.(http.Flusher); ok {
		f.Flush()
	}
	return nil
}

// -----------------------------------------------------------------------------
// Middleware-based Server Architecture
// -----------------------------------------------------------------------------

// HandlerContext is the unified container for all request information and state
// that flows through the entire request processing pipeline.
//
// The generic type P represents the user-defined container for application
// dependencies and request data (e.g., UserID, DB connection, etc.).
//
// The generic type I represents the input type, which can be any type depending
// on the operation.
type HandlerContext[T any, I any] struct {
	// Props is the user-defined container, created per request,
	// for application dependencies and request data (e.g., UserID).
	Props T

	// Input contains the request body, already deserialized and typed.
	// For global middlewares, the type I will be any.
	Input I

	// Context is the standard Go context.Context for cancellations and deadlines.
	Context context.Context

	// operation is the details of the RPC operation including it's RPC name, operation name and type.
	operation OperationDefinition
}

// RPCName returns the name of the RPC service
func (h *HandlerContext[T, I]) RPCName() string { return h.operation.RPCName }

// OperationName returns the name of the operation (e.g. "CreateUser", "GetPost", etc.)
func (h *HandlerContext[T, I]) OperationName() string { return h.operation.Name }

// OperationType returns the type of operation, can be [OperationTypeProc] ("proc") or [OperationTypeStream] ("stream")
func (h *HandlerContext[T, I]) OperationType() OperationType { return h.operation.Type }

// GlobalHandlerFunc is the signature for a global handler function.
// Both for procedures and streams
type GlobalHandlerFunc[T any] func(
	c *HandlerContext[T, any],
) (any, error)

// GlobalMiddlewareFunc is the signature for a middleware applied to all requests.
type GlobalMiddlewareFunc[T any] func(
	next GlobalHandlerFunc[T],
) GlobalHandlerFunc[T]

// ProcHandlerFunc is the signature of the final business handler for a proc.
type ProcHandlerFunc[T any, I any, O any] func(
	c *HandlerContext[T, I],
) (O, error)

// ProcMiddlewareFunc is the signature for a proc-specific typed middleware.
// It uses a wrapper pattern for a clean composition.
//
// This is the same as [GlobalMiddlewareFunc] but for specific procedures and with types.
type ProcMiddlewareFunc[T any, I any, O any] func(
	next ProcHandlerFunc[T, I, O],
) ProcHandlerFunc[T, I, O]

// StreamHandlerFunc is the signature of the main handler that initializes a stream.
type StreamHandlerFunc[T any, I any, O any] func(
	c *HandlerContext[T, I],
	emit EmitFunc[T, I, O],
) error

// StreamMiddlewareFunc is the signature for a middleware that wraps the main stream handler.
type StreamMiddlewareFunc[T any, I any, O any] func(
	next StreamHandlerFunc[T, I, O],
) StreamHandlerFunc[T, I, O]

// EmitFunc is the signature for emitting events from a stream.
type EmitFunc[T any, I any, O any] func(
	c *HandlerContext[T, I],
	output O,
) error

// EmitMiddlewareFunc is the signature for a middleware that wraps each call to emit.
type EmitMiddlewareFunc[T any, I any, O any] func(
	next EmitFunc[T, I, O],
) EmitFunc[T, I, O]

// DeserializerFunc function convert raw JSON input into typed input prior to handler execution.
type DeserializerFunc func(raw json.RawMessage) (any, error)

// ErrorHandlerFunc transforms an internal Go error into a public VDL Error.
//
// The handler receives the full request context, enabling structured logging
// with access to Props, RPC Name, Operation Name, and other metadata.
type ErrorHandlerFunc[T any] func(c *HandlerContext[T, any], err error) Error

// StreamConfig allows configuring the behavior of streams.
type StreamConfig struct {
	// PingInterval is the interval at which ping events are sent to the client
	// to keep the connection alive. Defaults to 30 seconds.
	PingInterval time.Duration
}

// -----------------------------------------------------------------------------
// Server Internal Implementation
// -----------------------------------------------------------------------------

// internalServer manages RPC request handling and middleware execution for
// both procedures and streams. It maintains handler registrations, middleware
// chains, and coordinates the complete request lifecycle.
//
// The generic type P represents the user context type, allowing users to pass
// custom data (authentication info, user sessions, etc.) through the entire
// request processing pipeline.
type internalServer[T any] struct {
	// handlersMu protects all handler maps and middleware slices from concurrent access
	handlersMu sync.RWMutex

	// operationDefs contains the definition of all registered operations
	// Map format: rpcName -> operationName -> OperationType ("proc" or "stream")
	operationDefs map[string]map[string]OperationType

	// procHandlers stores the final implementation functions for procedures
	// Map format: rpcName -> procName -> handler
	procHandlers map[string]map[string]ProcHandlerFunc[T, any, any]

	// streamHandlers stores the final implementation functions for streams
	// Map format: rpcName -> streamName -> handler
	streamHandlers map[string]map[string]StreamHandlerFunc[T, any, any]

	// globalMiddlewares contains middlewares that run for every request (both procs and streams)
	globalMiddlewares []GlobalMiddlewareFunc[T]

	// rpcMiddlewares contains middlewares that run for every request within a specific RPC
	// Map format: rpcName -> middlewares
	rpcMiddlewares map[string][]GlobalMiddlewareFunc[T]

	// procMiddlewares contains per-procedure middlewares
	// Map format: rpcName -> procName -> middlewares
	procMiddlewares map[string]map[string][]ProcMiddlewareFunc[T, any, any]

	// streamMiddlewares contains per-stream middlewares
	// Map format: rpcName -> streamName -> middlewares
	streamMiddlewares map[string]map[string][]StreamMiddlewareFunc[T, any, any]

	// streamEmitMiddlewares contains per-stream emit middlewares
	// Map format: rpcName -> streamName -> middlewares
	streamEmitMiddlewares map[string]map[string][]EmitMiddlewareFunc[T, any, any]

	// procDeserializers contains per-procedure input deserializers
	// Map format: rpcName -> procName -> deserializer
	procDeserializers map[string]map[string]DeserializerFunc

	// streamDeserializers contains per-stream input deserializers
	// Map format: rpcName -> streamName -> deserializer
	streamDeserializers map[string]map[string]DeserializerFunc

	// globalStreamConfig contains the global configuration for streams
	globalStreamConfig StreamConfig

	// rpcStreamConfigs contains per-RPC configuration for streams
	// Map format: rpcName -> config
	rpcStreamConfigs map[string]StreamConfig

	// streamConfigs contains per-stream configuration
	// Map format: rpcName -> streamName -> config
	streamConfigs map[string]map[string]StreamConfig

	// globalErrorHandler is the global error handler
	globalErrorHandler ErrorHandlerFunc[T]

	// rpcErrorHandlers contains per-RPC error handlers
	// Map format: rpcName -> handler
	rpcErrorHandlers map[string]ErrorHandlerFunc[T]
}

// newInternalServer creates a new VDL server instance with the specified
// procedure and stream definitions. The server is initialized with empty handler
// maps and middleware slices, ready for registration.
//
// The generic type T represents the user context type, used to pass additional
// data to handlers, such as authentication information, user sessions, or any
// other request-scoped data.
//
// Parameters:
//   - procDefs: List of procedure definitions that this server will handle
//   - streamDefs: List of stream definitions that this server will handle
//
// Returns a new internalServer instance ready for handler and middleware registration.
func newInternalServer[T any](
	procDefs []OperationDefinition,
	streamDefs []OperationDefinition,
) *internalServer[T] {
	operationDefs := make(map[string]map[string]OperationType)

	// Initialize maps
	procHandlers := make(map[string]map[string]ProcHandlerFunc[T, any, any])
	streamHandlers := make(map[string]map[string]StreamHandlerFunc[T, any, any])
	rpcMiddlewares := make(map[string][]GlobalMiddlewareFunc[T])
	procMiddlewares := make(map[string]map[string][]ProcMiddlewareFunc[T, any, any])
	streamMiddlewares := make(map[string]map[string][]StreamMiddlewareFunc[T, any, any])
	streamEmitMiddlewares := make(map[string]map[string][]EmitMiddlewareFunc[T, any, any])
	procDeserializers := make(map[string]map[string]DeserializerFunc)
	streamDeserializers := make(map[string]map[string]DeserializerFunc)
	rpcStreamConfigs := make(map[string]StreamConfig)
	streamConfigs := make(map[string]map[string]StreamConfig)
	rpcErrorHandlers := make(map[string]ErrorHandlerFunc[T])

	// Helper to ensure RPC map existence
	ensureRPC := func(rpcName string) {
		if _, ok := operationDefs[rpcName]; !ok {
			operationDefs[rpcName] = make(map[string]OperationType)
			procHandlers[rpcName] = make(map[string]ProcHandlerFunc[T, any, any])
			streamHandlers[rpcName] = make(map[string]StreamHandlerFunc[T, any, any])
			procMiddlewares[rpcName] = make(map[string][]ProcMiddlewareFunc[T, any, any])
			streamMiddlewares[rpcName] = make(map[string][]StreamMiddlewareFunc[T, any, any])
			streamEmitMiddlewares[rpcName] = make(map[string][]EmitMiddlewareFunc[T, any, any])
			procDeserializers[rpcName] = make(map[string]DeserializerFunc)
			streamDeserializers[rpcName] = make(map[string]DeserializerFunc)
			streamConfigs[rpcName] = make(map[string]StreamConfig)
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

	return &internalServer[T]{
		handlersMu:            sync.RWMutex{},
		operationDefs:         operationDefs,
		procHandlers:          procHandlers,
		streamHandlers:        streamHandlers,
		globalMiddlewares:     []GlobalMiddlewareFunc[T]{},
		rpcMiddlewares:        rpcMiddlewares,
		procMiddlewares:       procMiddlewares,
		streamMiddlewares:     streamMiddlewares,
		streamEmitMiddlewares: streamEmitMiddlewares,
		procDeserializers:     procDeserializers,
		streamDeserializers:   streamDeserializers,
		globalStreamConfig:    StreamConfig{PingInterval: 30 * time.Second},
		rpcStreamConfigs:      rpcStreamConfigs,
		streamConfigs:         streamConfigs,
		globalErrorHandler:    nil,
		rpcErrorHandlers:      rpcErrorHandlers,
	}
}

// addGlobalMiddleware registers a global middleware that executes for every request (proc and stream).
// Middlewares are executed in the order they were registered.
func (s *internalServer[T]) addGlobalMiddleware(
	mw GlobalMiddlewareFunc[T],
) *internalServer[T] {
	s.handlersMu.Lock()
	defer s.handlersMu.Unlock()
	s.globalMiddlewares = append(s.globalMiddlewares, mw)
	return s
}

// addRPCMiddleware registers a middleware that executes for every request within a specific RPC.
// Middlewares are executed in the order they were registered.
func (s *internalServer[T]) addRPCMiddleware(
	rpcName string,
	mw GlobalMiddlewareFunc[T],
) *internalServer[T] {
	s.handlersMu.Lock()
	defer s.handlersMu.Unlock()
	s.rpcMiddlewares[rpcName] = append(s.rpcMiddlewares[rpcName], mw)
	return s
}

// addProcMiddleware registers a wrapper middleware for a specific procedure.
// Middlewares are executed in the order they were registered.
func (s *internalServer[T]) addProcMiddleware(
	rpcName string,
	procName string,
	mw ProcMiddlewareFunc[T, any, any],
) *internalServer[T] {
	s.handlersMu.Lock()
	defer s.handlersMu.Unlock()
	s.procMiddlewares[rpcName][procName] = append(s.procMiddlewares[rpcName][procName], mw)
	return s
}

// addStreamMiddleware registers a wrapper middleware for a specific stream.
// Middlewares are executed in the order they were registered.
func (s *internalServer[T]) addStreamMiddleware(
	rpcName string,
	streamName string,
	mw StreamMiddlewareFunc[T, any, any],
) *internalServer[T] {
	s.handlersMu.Lock()
	defer s.handlersMu.Unlock()
	s.streamMiddlewares[rpcName][streamName] = append(s.streamMiddlewares[rpcName][streamName], mw)
	return s
}

// addStreamEmitMiddleware registers an emit wrapper middleware for a specific stream.
// Middlewares are executed in the order they were registered.
func (s *internalServer[T]) addStreamEmitMiddleware(
	rpcName string,
	streamName string,
	mw EmitMiddlewareFunc[T, any, any],
) *internalServer[T] {
	s.handlersMu.Lock()
	defer s.handlersMu.Unlock()
	s.streamEmitMiddlewares[rpcName][streamName] = append(s.streamEmitMiddlewares[rpcName][streamName], mw)
	return s
}

// setGlobalStreamConfig sets the global configuration for streams.
func (s *internalServer[T]) setGlobalStreamConfig(cfg StreamConfig) *internalServer[T] {
	s.handlersMu.Lock()
	defer s.handlersMu.Unlock()
	if cfg.PingInterval <= 0 {
		cfg.PingInterval = 30 * time.Second
	}
	s.globalStreamConfig = cfg
	return s
}

// setRPCStreamConfig sets the configuration for streams within a specific RPC.
func (s *internalServer[T]) setRPCStreamConfig(rpcName string, cfg StreamConfig) *internalServer[T] {
	s.handlersMu.Lock()
	defer s.handlersMu.Unlock()
	s.rpcStreamConfigs[rpcName] = cfg
	return s
}

// setStreamConfig sets the configuration for a specific stream.
func (s *internalServer[T]) setStreamConfig(rpcName, streamName string, cfg StreamConfig) *internalServer[T] {
	s.handlersMu.Lock()
	defer s.handlersMu.Unlock()
	s.streamConfigs[rpcName][streamName] = cfg
	return s
}

// setGlobalErrorHandler sets the global error handler.
func (s *internalServer[T]) setGlobalErrorHandler(handler ErrorHandlerFunc[T]) *internalServer[T] {
	s.handlersMu.Lock()
	defer s.handlersMu.Unlock()
	s.globalErrorHandler = handler
	return s
}

// setRPCErrorHandler sets the error handler for a specific RPC.
func (s *internalServer[T]) setRPCErrorHandler(rpcName string, handler ErrorHandlerFunc[T]) *internalServer[T] {
	s.handlersMu.Lock()
	defer s.handlersMu.Unlock()
	s.rpcErrorHandlers[rpcName] = handler
	return s
}

// resolveErrorHandler returns the appropriate error handler for the given RPC name.
// Precedence: RPC > Global > Default (Passthrough)
func (s *internalServer[T]) resolveErrorHandler(rpcName string) ErrorHandlerFunc[T] {
	s.handlersMu.RLock()
	defer s.handlersMu.RUnlock()

	if handler, ok := s.rpcErrorHandlers[rpcName]; ok && handler != nil {
		return handler
	}
	if s.globalErrorHandler != nil {
		return s.globalErrorHandler
	}

	// Default: Passthrough
	return func(c *HandlerContext[T, any], err error) Error {
		return ToError(err)
	}
}

// setProcHandler registers the final implementation function and deserializer for the specified procedure name.
// The provided functions are stored as-is. Middlewares are composed at request time.
//
// Panics if a handler is already registered for the given procedure name.
func (s *internalServer[T]) setProcHandler(
	rpcName string,
	procName string,
	handler ProcHandlerFunc[T, any, any],
	deserializer DeserializerFunc,
) *internalServer[T] {
	s.handlersMu.Lock()
	defer s.handlersMu.Unlock()
	if _, exists := s.procHandlers[rpcName][procName]; exists {
		panic(fmt.Sprintf("the procedure handler for %s.%s is already registered", rpcName, procName))
	}
	s.procHandlers[rpcName][procName] = handler
	s.procDeserializers[rpcName][procName] = deserializer
	return s
}

// setStreamHandler registers the final implementation function and deserializer for the specified stream name.
// The provided functions are stored as-is. Middlewares are composed at request time.
//
// Panics if a handler is already registered for the given stream name.
func (s *internalServer[T]) setStreamHandler(
	rpcName string,
	streamName string,
	handler StreamHandlerFunc[T, any, any],
	deserializer DeserializerFunc,
) *internalServer[T] {
	s.handlersMu.Lock()
	defer s.handlersMu.Unlock()
	if _, exists := s.streamHandlers[rpcName][streamName]; exists {
		panic(fmt.Sprintf("the stream handler for %s.%s is already registered", rpcName, streamName))
	}
	s.streamHandlers[rpcName][streamName] = handler
	s.streamDeserializers[rpcName][streamName] = deserializer
	return s
}

// handleRequest processes an incoming RPC request by parsing the request body,
// building the global middleware chain, and dispatching to the appropriate
// adapter (procedure or stream).
//
// The request body must contain a JSON object with the input data for the handler.
//
// Parameters:
//   - ctx: The request context
//   - props: The UFO context containing user-defined data
//   - rpcName: The name of the RPC service to invoke
//   - operationName: The name of the procedure or stream to invoke
//   - httpAdapter: The HTTP adapter for reading requests and writing responses
//
// Returns an error if request processing fails at the transport level.
func (s *internalServer[T]) handleRequest(
	ctx context.Context,
	props T,
	rpcName string,
	operationName string,
	httpAdapter HTTPAdapter,
) error {
	if httpAdapter == nil {
		return fmt.Errorf("the HTTP adapter is nil, please provide a valid adapter")
	}

	// Decode the request body into a json.RawMessage as the initial input container
	var rawInput json.RawMessage
	if err := json.NewDecoder(httpAdapter.RequestBody()).Decode(&rawInput); err != nil {
		res := Response[any]{
			Ok:    false,
			Error: Error{Message: "Invalid request body"},
		}
		return s.writeProcResponse(httpAdapter, res)
	}

	operationType, operationExists := s.operationDefs[rpcName][operationName]
	if !operationExists {
		res := Response[any]{
			Ok:    false,
			Error: Error{Message: fmt.Sprintf("Invalid operation: %s.%s", rpcName, operationName)},
		}
		return s.writeProcResponse(httpAdapter, res)
	}

	// Build the unified handler context (raw input at this point).
	c := &HandlerContext[T, any]{
		Input:   rawInput,
		Props:   props,
		Context: ctx,
		operation: OperationDefinition{
			RPCName: rpcName,
			Name:    operationName,
			Type:    operationType,
		},
	}

	// Handle Stream
	if operationType == OperationTypeStream {
		err := s.handleStreamRequest(c, rpcName, operationName, rawInput, httpAdapter)

		// If no error, return without sending any response
		if err == nil {
			return nil
		}

		// Send an event with the error before closing the connection
		errorHandler := s.resolveErrorHandler(rpcName)
		response := Response[any]{
			Ok:    false,
			Error: errorHandler(c, err),
		}

		jsonData, marshalErr := json.Marshal(response)
		if marshalErr != nil {
			return fmt.Errorf("failed to marshal stream error: %w", marshalErr)
		}
		resPayload := fmt.Sprintf("data: %s\n\n", jsonData)
		if _, writeErr := httpAdapter.Write([]byte(resPayload)); writeErr != nil {
			return writeErr
		}
		if flushErr := httpAdapter.Flush(); flushErr != nil {
			return flushErr
		}
	}

	// Handle Procedure
	output, err := s.handleProcRequest(c, rpcName, operationName, rawInput)
	response := Response[any]{}
	if err != nil {
		response.Ok = false
		errorHandler := s.resolveErrorHandler(rpcName)
		response.Error = errorHandler(c, err)
	} else {
		response.Ok = true
		response.Output = output
	}

	return s.writeProcResponse(httpAdapter, response)
}

// handleProcRequest builds the per-request middleware chain for a procedure and executes it.
// It returns the procedure output (as any) and an error if the handler failed.
func (s *internalServer[T]) handleProcRequest(
	c *HandlerContext[T, any],
	rpcName string,
	procName string,
	rawInput json.RawMessage,
) (any, error) {
	// Snapshot handler, middlewares, and deserializer under read lock
	s.handlersMu.RLock()
	baseHandler, ok := s.procHandlers[rpcName][procName]
	mws := s.procMiddlewares[rpcName][procName]
	rpcMws := s.rpcMiddlewares[rpcName]
	deserialize := s.procDeserializers[rpcName][procName]
	s.handlersMu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("%s.%s procedure not implemented", rpcName, procName)
	}
	if deserialize == nil {
		return nil, fmt.Errorf("%s.%s procedure deserializer not registered", rpcName, procName)
	}

	// Deserialize, validate and transform input into its typed form
	typedInput, err := deserialize(rawInput)
	if err != nil {
		return nil, err
	}
	c.Input = typedInput

	// Compose specific per-proc middlewares around the base handler (reverse registration order)
	final := baseHandler
	if len(mws) > 0 {
		mwChain := append([]ProcMiddlewareFunc[T, any, any](nil), mws...)
		for i := len(mwChain) - 1; i >= 0; i-- {
			final = mwChain[i](final)
		}
	}

	// Wrap the specific chain with RPC-level middlewares (executed before specific ones)
	exec := func(c *HandlerContext[T, any]) (any, error) { return final(c) }
	if len(rpcMws) > 0 {
		mwChain := append([]GlobalMiddlewareFunc[T](nil), rpcMws...)
		for i := len(mwChain) - 1; i >= 0; i-- {
			exec = mwChain[i](exec)
		}
	}

	// Wrap the chain with global middlewares (executed before RPC-level ones)
	if len(s.globalMiddlewares) > 0 {
		mwChain := append([]GlobalMiddlewareFunc[T](nil), s.globalMiddlewares...)
		for i := len(mwChain) - 1; i >= 0; i-- {
			exec = mwChain[i](exec)
		}
	}

	return exec(c)
}

// handleStreamRequest builds the per-request middleware chain for a stream, sets up SSE,
// composes emit middlewares, and executes the stream handler.
func (s *internalServer[T]) handleStreamRequest(
	c *HandlerContext[T, any],
	rpcName string,
	streamName string,
	rawInput json.RawMessage,
	httpAdapter HTTPAdapter,
) error {
	// Snapshot handler, middlewares, emit middlewares and deserializer under read lock
	s.handlersMu.RLock()
	baseHandler, ok := s.streamHandlers[rpcName][streamName]
	streamMws := s.streamMiddlewares[rpcName][streamName]
	emitMws := s.streamEmitMiddlewares[rpcName][streamName]
	rpcMws := s.rpcMiddlewares[rpcName]
	deserialize := s.streamDeserializers[rpcName][streamName]

	// Determine configuration (Precedence: Stream > RPC > Global)
	pingInterval := s.globalStreamConfig.PingInterval
	if cfg, ok := s.rpcStreamConfigs[rpcName]; ok && cfg.PingInterval > 0 {
		pingInterval = cfg.PingInterval
	}
	if cfg, ok := s.streamConfigs[rpcName][streamName]; ok && cfg.PingInterval > 0 {
		pingInterval = cfg.PingInterval
	}

	s.handlersMu.RUnlock()

	// Set SSE headers to the response
	httpAdapter.SetHeader("Content-Type", "text/event-stream")
	httpAdapter.SetHeader("Cache-Control", "no-cache")
	httpAdapter.SetHeader("Connection", "keep-alive")

	if !ok {
		return fmt.Errorf("%s.%s stream not implemented", rpcName, streamName)
	}
	if deserialize == nil {
		return fmt.Errorf("%s.%s stream deserializer not registered", rpcName, streamName)
	}

	// Deserialize, validate and transform input into its typed form
	typedInput, err := deserialize(rawInput)
	if err != nil {
		return err
	}
	c.Input = typedInput

	// We need to synchronize writes to the httpAdapter because pings run in a separate goroutine
	var writeMu sync.Mutex
	safeWrite := func(parts ...[]byte) error {
		writeMu.Lock()
		defer writeMu.Unlock()
		for _, part := range parts {
			if _, err := httpAdapter.Write(part); err != nil {
				return err
			}
		}
		return httpAdapter.Flush()
	}

	// Start Ping Loop
	ctx, cancel := context.WithCancel(c.Context)
	defer cancel()

	go func() {
		ticker := time.NewTicker(pingInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				_ = safeWrite([]byte(": ping\n\n"))
			}
		}
	}()

	// Base emit writes SSE envelope with {ok:true, output}
	baseEmit := func(_ *HandlerContext[T, any], data any) error {
		response := Response[any]{
			Ok:     true,
			Output: data,
		}
		jsonData, err := json.Marshal(response)
		if err != nil {
			return fmt.Errorf("failed to marshal stream data: %w", err)
		}
		return safeWrite([]byte("data: "), jsonData, []byte("\n\n"))
	}

	// Compose emit middlewares (reverse registration order)
	emitFinal := baseEmit
	if len(emitMws) > 0 {
		mwChain := append([]EmitMiddlewareFunc[T, any, any](nil), emitMws...)
		for i := len(mwChain) - 1; i >= 0; i-- {
			emitFinal = mwChain[i](emitFinal)
		}
	}

	// Compose stream middlewares around the base handler (reverse order)
	final := baseHandler
	if len(streamMws) > 0 {
		mwChain := append([]StreamMiddlewareFunc[T, any, any](nil), streamMws...)
		for i := len(mwChain) - 1; i >= 0; i-- {
			final = mwChain[i](final)
		}
	}

	// Wrap the specific stream chain with RPC-level middlewares (executed before specific ones)
	exec := func(c *HandlerContext[T, any]) (any, error) { return nil, final(c, emitFinal) }
	if len(rpcMws) > 0 {
		mwChain := append([]GlobalMiddlewareFunc[T](nil), rpcMws...)
		for i := len(mwChain) - 1; i >= 0; i-- {
			exec = mwChain[i](exec)
		}
	}

	// Wrap with global middlewares (executed before RPC-level ones)
	if len(s.globalMiddlewares) > 0 {
		mwChain := append([]GlobalMiddlewareFunc[T](nil), s.globalMiddlewares...)
		for i := len(mwChain) - 1; i >= 0; i-- {
			exec = mwChain[i](exec)
		}
	}

	_, err = exec(c)
	return err
}

// writeProcResponse writes a procedure response to the client as JSON.
// This helper method sets the appropriate Content-Type header and marshals
// the response data before sending it to the client.
//
// Parameters:
//   - httpAdapter: The HTTP adapter for writing the response
//   - response: The response data to send to the client
//
// Returns an error if writing the response fails.
func (s *internalServer[T]) writeProcResponse(
	httpAdapter HTTPAdapter,
	response Response[any],
) error {
	httpAdapter.SetHeader("Content-Type", "application/json")
	_, err := httpAdapter.Write(response.Bytes())
	if err != nil {
		return fmt.Errorf("failed to write response: %w", err)
	}
	return nil
}

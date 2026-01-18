//nolint:unused
package pieces

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
)

/** START FROM HERE **/

// -----------------------------------------------------------------------------
// Server Types
// -----------------------------------------------------------------------------

const (
	OperationTypeProc   = "proc"
	OperationTypeStream = "stream"
)

// HTTPAdapter defines the interface required by UFO RPC server to handle
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
// This adapter bridges the UFO RPC server with the standard HTTP library, allowing
// seamless integration with existing HTTP servers and middleware.
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
// Returns a HTTPAdapter implementation ready for use with UFO RPC server.
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

	// operationName is the name of the invoked proc or stream (e.g., "CreateUser").
	operationName string

	// operationType is the type of operation ("proc" or "stream").
	operationType string
}

// OperationName returns the name of the operation (e.g. "CreateUser", "GetPost", etc.)
func (h *HandlerContext[T, I]) OperationName() string { return h.operationName }

// OperationType returns the type of operation (e.g. "proc" or "stream")
func (h *HandlerContext[T, I]) OperationType() string { return h.operationType }

// GlobalHandlerFunc is the signature for a global handler function.
// Both for procedures and streams
type GlobalHandlerFunc[T any] func(
	c *HandlerContext[T, any],
) (any, error)

// GlobalMiddleware is the signature for a middleware applied to all requests.
type GlobalMiddleware[T any] func(
	next GlobalHandlerFunc[T],
) GlobalHandlerFunc[T]

// ProcHandlerFunc is the signature of the final business handler for a proc.
type ProcHandlerFunc[T any, I any, O any] func(
	c *HandlerContext[T, I],
) (O, error)

// ProcMiddlewareFunc is the signature for a proc-specific typed middleware.
// It uses a wrapper pattern for a clean composition.
//
// This is the same as [GlobalMiddleware] but for specific procedures and with types.
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

// Deserializer function convert raw JSON input into typed input prior to handler execution.
type DeserializeFunc func(raw json.RawMessage) (any, error)

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
	// procNames contains the list of all registered procedure names
	procNames []string
	// procNamesMap contains the list of all registered procedure names
	procNamesMap map[string]bool
	// streamNames contains the list of all registered stream names
	streamNames []string
	// streamNamesMap contains the list of all registered stream names
	streamNamesMap map[string]bool
	// operationNamesMap contains the list of all registered operation names
	// and its corresponding type
	operationNamesMap map[string]string
	// handlersMu protects all handler maps and middleware slices from concurrent access
	handlersMu sync.RWMutex
	// procHandlers stores the final implementation functions for procedures
	procHandlers map[string]ProcHandlerFunc[T, any, any]
	// streamHandlers stores the final implementation functions for streams
	streamHandlers map[string]StreamHandlerFunc[T, any, any]
	// globalMiddlewares contains middlewares that run for every request (both procs and streams)
	globalMiddlewares []GlobalMiddleware[T]
	// procMiddlewares contains per-procedure middlewares
	procMiddlewares map[string][]ProcMiddlewareFunc[T, any, any]
	// streamMiddlewares contains per-stream middlewares
	streamMiddlewares map[string][]StreamMiddlewareFunc[T, any, any]
	// streamEmitMiddlewares contains per-stream emit middlewares
	streamEmitMiddlewares map[string][]EmitMiddlewareFunc[T, any, any]
	// procDeserializers contains per-procedure input deserializers
	procDeserializers map[string]DeserializeFunc
	// streamDeserializers contains per-stream input deserializers
	streamDeserializers map[string]DeserializeFunc
}

// newInternalServer creates a new UFO RPC server instance with the specified
// procedure and stream names. The server is initialized with empty handler
// maps and middleware slices, ready for registration.
//
// The generic type T represents the user context type, used to pass additional
// data to handlers, such as authentication information, user sessions, or any
// other request-scoped data.
//
// Parameters:
//   - procNames: List of procedure names that this server will handle
//   - streamNames: List of stream names that this server will handle
//
// Returns a new internalServer instance ready for handler and middleware registration.
func newInternalServer[T any](
	procNames []string,
	streamNames []string,
) *internalServer[T] {
	procNamesMap := make(map[string]bool)
	streamNamesMap := make(map[string]bool)
	operationNamesMap := make(map[string]string)
	for _, procName := range procNames {
		procNamesMap[procName] = true
		operationNamesMap[procName] = OperationTypeProc
	}
	for _, streamName := range streamNames {
		streamNamesMap[streamName] = true
		operationNamesMap[streamName] = OperationTypeStream
	}

	return &internalServer[T]{
		procNames:             procNames,
		procNamesMap:          procNamesMap,
		streamNames:           streamNames,
		streamNamesMap:        streamNamesMap,
		operationNamesMap:     operationNamesMap,
		handlersMu:            sync.RWMutex{},
		procHandlers:          map[string]ProcHandlerFunc[T, any, any]{},
		streamHandlers:        map[string]StreamHandlerFunc[T, any, any]{},
		globalMiddlewares:     []GlobalMiddleware[T]{},
		procMiddlewares:       map[string][]ProcMiddlewareFunc[T, any, any]{},
		streamMiddlewares:     map[string][]StreamMiddlewareFunc[T, any, any]{},
		streamEmitMiddlewares: map[string][]EmitMiddlewareFunc[T, any, any]{},
		procDeserializers:     map[string]DeserializeFunc{},
		streamDeserializers:   map[string]DeserializeFunc{},
	}
}

// addGlobalMiddleware registers a global middleware that executes for every request (proc and stream).
// Middlewares are executed in the order they were registered.
func (s *internalServer[T]) addGlobalMiddleware(
	mw GlobalMiddleware[T],
) *internalServer[T] {
	s.handlersMu.Lock()
	defer s.handlersMu.Unlock()
	s.globalMiddlewares = append(s.globalMiddlewares, mw)
	return s
}

// addProcMiddleware registers a wrapper middleware for a specific procedure.
// Middlewares are executed in the order they were registered.
func (s *internalServer[T]) addProcMiddleware(
	procName string,
	mw ProcMiddlewareFunc[T, any, any],
) *internalServer[T] {
	s.handlersMu.Lock()
	defer s.handlersMu.Unlock()
	s.procMiddlewares[procName] = append(s.procMiddlewares[procName], mw)
	return s
}

// addStreamMiddleware registers a wrapper middleware for a specific stream.
// Middlewares are executed in the order they were registered.
func (s *internalServer[T]) addStreamMiddleware(
	streamName string,
	mw StreamMiddlewareFunc[T, any, any],
) *internalServer[T] {
	s.handlersMu.Lock()
	defer s.handlersMu.Unlock()
	s.streamMiddlewares[streamName] = append(s.streamMiddlewares[streamName], mw)
	return s
}

// addStreamEmitMiddleware registers an emit wrapper middleware for a specific stream.
// Middlewares are executed in the order they were registered.
func (s *internalServer[T]) addStreamEmitMiddleware(
	streamName string,
	mw EmitMiddlewareFunc[T, any, any],
) *internalServer[T] {
	s.handlersMu.Lock()
	defer s.handlersMu.Unlock()
	s.streamEmitMiddlewares[streamName] = append(s.streamEmitMiddlewares[streamName], mw)
	return s
}

// setProcHandler registers the final implementation function and deserializer for the specified procedure name.
// The provided functions are stored as-is. Middlewares are composed at request time.
//
// Panics if a handler is already registered for the given procedure name.
func (s *internalServer[T]) setProcHandler(
	procName string,
	handler ProcHandlerFunc[T, any, any],
	deserializer DeserializeFunc,
) *internalServer[T] {
	s.handlersMu.Lock()
	defer s.handlersMu.Unlock()
	if _, exists := s.procHandlers[procName]; exists {
		panic(fmt.Sprintf("the procedure handler for %s is already registered", procName))
	}
	s.procHandlers[procName] = handler
	s.procDeserializers[procName] = deserializer
	return s
}

// setStreamHandler registers the final implementation function and deserializer for the specified stream name.
// The provided functions are stored as-is. Middlewares are composed at request time.
//
// Panics if a handler is already registered for the given stream name.
func (s *internalServer[T]) setStreamHandler(
	streamName string,
	handler StreamHandlerFunc[T, any, any],
	deserializer DeserializeFunc,
) *internalServer[T] {
	s.handlersMu.Lock()
	defer s.handlersMu.Unlock()
	if _, exists := s.streamHandlers[streamName]; exists {
		panic(fmt.Sprintf("the stream handler for %s is already registered", streamName))
	}
	s.streamHandlers[streamName] = handler
	s.streamDeserializers[streamName] = deserializer
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
//   - operationName: The name of the procedure or stream to invoke
//   - httpAdapter: The HTTP adapter for reading requests and writing responses
//
// Returns an error if request processing fails at the transport level.
func (s *internalServer[T]) handleRequest(
	ctx context.Context,
	props T,
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

	operationType, operationExists := s.operationNamesMap[operationName]
	if !operationExists {
		res := Response[any]{
			Ok:    false,
			Error: Error{Message: "Invalid operation name"},
		}
		return s.writeProcResponse(httpAdapter, res)
	}

	// Build the unified handler context (raw input at this point).
	c := &HandlerContext[T, any]{
		Input:         rawInput,
		Props:         props,
		Context:       ctx,
		operationName: operationName,
		operationType: operationType,
	}

	// Handle Stream
	if operationType == OperationTypeStream {
		err := s.handleStreamRequest(c, operationName, rawInput, httpAdapter)

		// If no error, return without sending any response
		if err == nil {
			return nil
		}

		// Send an event with the error before closing the connection
		response := Response[any]{
			Ok:    false,
			Error: asError(err),
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
	output, err := s.handleProcRequest(c, operationName, rawInput)
	response := Response[any]{}
	if err != nil {
		response.Ok = false
		response.Error = asError(err)
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
	procName string,
	rawInput json.RawMessage,
) (any, error) {
	// Snapshot handler, middlewares, and deserializer under read lock
	s.handlersMu.RLock()
	baseHandler, ok := s.procHandlers[procName]
	mws := s.procMiddlewares[procName]
	deserialize := s.procDeserializers[procName]
	s.handlersMu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("%s procedure not implemented", procName)
	}
	if deserialize == nil {
		return nil, fmt.Errorf("%s procedure deserializer not registered", procName)
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

	// Wrap the specific chain with global middlewares (executed before specific ones)
	exec := func(c *HandlerContext[T, any]) (any, error) { return final(c) }
	if len(s.globalMiddlewares) > 0 {
		mwChain := append([]GlobalMiddleware[T](nil), s.globalMiddlewares...)
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
	streamName string,
	rawInput json.RawMessage,
	httpAdapter HTTPAdapter,
) error {
	// Snapshot handler, middlewares, emit middlewares and deserializer under read lock
	s.handlersMu.RLock()
	baseHandler, ok := s.streamHandlers[streamName]
	streamMws := s.streamMiddlewares[streamName]
	emitMws := s.streamEmitMiddlewares[streamName]
	deserialize := s.streamDeserializers[streamName]
	s.handlersMu.RUnlock()

	// Set SSE headers to the response
	httpAdapter.SetHeader("Content-Type", "text/event-stream")
	httpAdapter.SetHeader("Cache-Control", "no-cache")
	httpAdapter.SetHeader("Connection", "keep-alive")

	if !ok {
		return fmt.Errorf("%s stream not implemented", streamName)
	}
	if deserialize == nil {
		return fmt.Errorf("%s stream deserializer not registered", streamName)
	}

	// Deserialize, validate and transform input into its typed form
	typedInput, err := deserialize(rawInput)
	if err != nil {
		return err
	}
	c.Input = typedInput

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
		resPayload := fmt.Sprintf("data: %s\n\n", jsonData)
		if _, err = httpAdapter.Write([]byte(resPayload)); err != nil {
			return err
		}
		if err := httpAdapter.Flush(); err != nil {
			return err
		}
		return nil
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

	// Wrap the specific stream chain with global middlewares (executed before specific ones)
	exec := func(c *HandlerContext[T, any]) (any, error) { return nil, final(c, emitFinal) }
	if len(s.globalMiddlewares) > 0 {
		mwChain := append([]GlobalMiddleware[T](nil), s.globalMiddlewares...)
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

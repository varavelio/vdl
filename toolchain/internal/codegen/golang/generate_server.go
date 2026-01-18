package golang

import (
	_ "embed"
	"fmt"

	"github.com/uforg/ufogenkit"
	"github.com/uforg/uforpc/urpc/internal/schema"
	"github.com/uforg/uforpc/urpc/internal/util/strutil"
)

//go:embed pieces/server.go
var serverRawPiece string

func generateServer(sch schema.Schema, config Config) (string, error) {
	if !config.IncludeServer {
		return "", nil
	}

	piece := strutil.GetStrAfter(serverRawPiece, "/** START FROM HERE **/")
	if piece == "" {
		return "", fmt.Errorf("server.go: could not find start delimiter")
	}

	g := ufogenkit.NewGenKit().WithTabs()

	// Core server piece (types + internal implementation)
	g.Raw(piece)
	g.Break()

	g.Line("// -----------------------------------------------------------------------------")
	g.Line("// Server generated implementation")
	g.Line("// -----------------------------------------------------------------------------")
	g.Break()

	// Server facade
	g.Line("// Server provides a high-level, type-safe API for building a UFO RPC server.")
	g.Line("// It exposes:")
	g.Line("//   - Procs: typed entries to register middlewares and the business handler per procedure")
	g.Line("//   - Streams: typed entries to register middlewares, emit middlewares and the handler per stream")
	g.Line("//   - Use: a global middleware API that runs for every operation (procedures and streams)")
	g.Line("//")
	g.Line("// The generic type parameter P is your application context (props) that flows through")
	g.Line("// the entire request lifecycle (authentication, per-request data, dependencies, etc.).")
	g.Line("type Server[T any] struct {")
	g.Block(func() {
		g.Line("intServer *internalServer[T]")
		g.Line("Procs     *serverProcRegistry[T]")
		g.Line("Streams   *serverStreamRegistry[T]")
	})
	g.Line("}")
	g.Break()

	g.Line("// NewServer creates a new UFO RPC server instance ready to handle all")
	g.Line("// defined procedures and streams using the middleware-based architecture.")
	g.Line("//")
	g.Line("// P is the application context type shared across the entire pipeline.")
	g.Line("//")
	g.Line("// Example:")
	g.Line("//   type AppProps struct {")
	g.Line("//       UserID string")
	g.Line("//   }")
	g.Line("//   s := NewServer[AppProps]()")
	g.Line("func NewServer[T any]() *Server[T] {")
	g.Block(func() {
		g.Line("intServer := newInternalServer[T](ufoProcedureNames, ufoStreamNames)")
		g.Line("return &Server[T]{")
		g.Block(func() {
			g.Line("intServer: intServer,")
			g.Line("Procs:     newServerProcRegistry(intServer),")
			g.Line("Streams:   newServerStreamRegistry(intServer),")
		})
		g.Line("}")
	})
	g.Line("}")
	g.Break()

	g.Line("// Use registers a global middleware that executes for every request (procedures and streams).")
	g.Line("//")
	g.Line("// Middlewares are executed in registration order and can:")
	g.Line("//   - read/augment the HandlerContext")
	g.Line("//   - short-circuit by returning an error")
	g.Line("//   - call next to continue the chain")
	g.Line("func (s *Server[T]) Use(mw GlobalMiddleware[T]) { s.intServer.addGlobalMiddleware(mw) }")
	g.Break()

	g.Line("// HandleRequest processes an incoming RPC request and drives the complete")
	g.Line("// request lifecycle (parsing, middleware chains, handler dispatch, response).")
	g.Line("//")
	g.Line("// operationName must be the last path segment of the request URL (e.g. /urpc/GetUser -> \"GetUser\").")
	g.Line("// httpAdapter bridges UFO RPC with your HTTP framework (use NewNetHTTPAdapter for net/http).")
	g.Line("//")
	g.Line("// Example (net/http):")
	g.Line("//   http.HandleFunc(\"POST /urpc/{operationName}\", func(w http.ResponseWriter, r *http.Request) {")
	g.Line("//       ctx := r.Context()")
	g.Line("//       props := AppProps{UserID: \"abc\"}")
	g.Line("//       op := r.PathValue(\"operationName\")")
	g.Line("//       adapter := NewNetHTTPAdapter(w, r)")
	g.Line("//       _ = server.HandleRequest(ctx, props, op, adapter)")
	g.Line("//   })")
	g.Line("func (s *Server[T]) HandleRequest(ctx context.Context, props T, operationName string, httpAdapter HTTPAdapter) error {")
	g.Block(func() {
		g.Line("return s.intServer.handleRequest(ctx, props, operationName, httpAdapter)")
	})
	g.Line("}")
	g.Break()

	// -----------------------------------------------------------------------------
	// Procedures registry and entries
	// -----------------------------------------------------------------------------
	g.Line("// serverProcRegistry groups all procedures and exposes typed entries to register")
	g.Line("// per-procedure middlewares and the final business handler. Input deserialization")
	g.Line("// and validation is handled automatically using generated pre* types.")
	g.Line("type serverProcRegistry[T any] struct {")
	g.Block(func() {
		g.Line("intServer *internalServer[T]")
		for _, procNode := range sch.GetProcNodes() {
			name := strutil.ToPascalCase(procNode.Name)
			g.Linef("%s proc%sEntry[T]", name, name)
		}
	})
	g.Line("}")
	g.Break()

	g.Line("func newServerProcRegistry[T any](intServer *internalServer[T]) *serverProcRegistry[T] {")
	g.Block(func() {
		g.Line("r := &serverProcRegistry[T]{intServer: intServer}")
		for _, procNode := range sch.GetProcNodes() {
			name := strutil.ToPascalCase(procNode.Name)
			g.Linef("r.%s = proc%sEntry[T]{intServer: intServer}", name, name)
		}
		g.Line("return r")
	})
	g.Line("}")
	g.Break()

	for _, procNode := range sch.GetProcNodes() {
		name := strutil.ToPascalCase(procNode.Name)

		g.Linef("// proc%sEntry contains the typed API for the %s procedure.", name, name)
		g.Linef("type proc%sEntry[T any] struct {", name)
		g.Block(func() {
			g.Line("intServer *internalServer[T]")
		})
		g.Line("}")
		g.Break()

		// Generate type aliases
		g.Linef("// Type aliases for %s procedure", name)
		g.Linef("type %sHandlerContext[T any] = HandlerContext[T, %sInput]", name, name)
		g.Linef("type %sHandlerFunc[T any] func(c *%sHandlerContext[T]) (%sOutput, error)", name, name, name)
		g.Linef("type %sMiddlewareFunc[T any] func(next %sHandlerFunc[T]) %sHandlerFunc[T]", name, name, name)
		g.Break()

		// Use (procedure middleware)
		g.Linef("// Use registers a typed middleware for the %s procedure.", name)
		g.Line("//")
		g.Line("// The middleware wraps the business handler registered with Handle, allowing you")
		g.Line("// to implement cross-cutting concerns such as validation, logging, auth, or")
		g.Line("// metrics in a type-safe way.")
		g.Line("//")
		g.Line("// Execution order: middlewares run in the order they were registered,")
		g.Line("// then the final handler is invoked.")
		renderDoc(g, procNode.Doc, true)
		renderDeprecated(g, procNode.Deprecated)
		g.Linef("func (e proc%sEntry[T]) Use(mw %sMiddlewareFunc[T]) {", name, name)
		g.Block(func() {
			g.Linef("adapted := func(next ProcHandlerFunc[T, any, any]) ProcHandlerFunc[T, any, any] {")
			g.Block(func() {
				g.Line("// This is the generic handler that will be executed by the server at runtime.")
				g.Linef("return func(cGeneric *HandlerContext[T, any]) (any, error) {")
				g.Block(func() {
					g.Line("// Create a type-safe 'next' function for the specific middleware to call.")
					g.Line("// This function acts as a bridge to translate the call back into the generic world.")
					g.Linef("typedNext := func(c *%sHandlerContext[T]) (%sOutput, error) {", name, name)
					g.Block(func() {
						g.Line("// Crucially, sync mutations from the specific context back to the generic")
						g.Line("// context before proceeding down the chain.")
						g.Line("cGeneric.Props = c.Props")
						g.Line("cGeneric.Input = c.Input")

						g.Line("// Call the original generic handler.")
						g.Line("genericOutput, err := next(cGeneric)")
						g.Line("if err != nil {")
						g.Block(func() {
							g.Line("// On error, return the zero value for the specific output type.")
							g.Linef("var zero %sOutput", name)
							g.Line("return zero, err")
						})
						g.Line("}")

						g.Line("// On success, assert the 'any' output to the specific output type.")
						g.Linef("specificOutput, _ := genericOutput.(%sOutput)", name)
						g.Line("return specificOutput, nil")
					})
					g.Line("}")

					g.Line("// Apply the user's middleware, giving it our typed bridge function.")
					g.Line("// The result is the complete, type-safe handler chain.")
					g.Line("typedChain := mw(typedNext)")

					g.Line("// Prepare the initial arguments for the typed chain by creating a")
					g.Line("// specific context from the generic one.")
					g.Linef("input, _ := cGeneric.Input.(%sInput)", name)
					g.Linef("cSpecific := &%sHandlerContext[T]{", name)
					g.Block(func() {
						g.Line("Input:         input,")
						g.Line("Props:         cGeneric.Props,")
						g.Line("Context:       cGeneric.Context,")
						g.Line("operationName: cGeneric.operationName,")
						g.Line("operationType: cGeneric.operationType,")
					})
					g.Line("}")

					g.Line("// Execute the fully composed, type-safe middleware chain.")
					g.Line("return typedChain(cSpecific)")
				})
				g.Line("}")
			})
			g.Line("}")
			g.Linef("e.intServer.addProcMiddleware(\"%s\", adapted)", name)
		})
		g.Line("}")
		g.Break()

		// Handle (procedure handler)
		g.Linef("// Handle registers the business handler for the %s procedure.", name)
		g.Line("//")
		g.Line("// The server will:")
		g.Line("//  1) Deserialize and validate the input using generated pre* types")
		g.Line("//  2) Build the procedure's middleware chain")
		g.Line("//  3) Invoke your handler with a typed context")
		renderDoc(g, procNode.Doc, true)
		renderDeprecated(g, procNode.Deprecated)
		g.Linef("func (e proc%sEntry[T]) Handle(handler %sHandlerFunc[T]) {", name, name)
		g.Block(func() {
			g.Linef("adaptedHandler := func(cGeneric *HandlerContext[T, any]) (any, error) {")
			g.Block(func() {
				g.Line("// Create the specific context from the generic one provided by the server.")
				g.Line("// It's assumed a higher layer guarantees the type is correct.")
				g.Linef("input, _ := cGeneric.Input.(%sInput)", name)
				g.Linef("cSpecific := &%sHandlerContext[T]{", name)
				g.Block(func() {
					g.Line("Input:         input,")
					g.Line("Props:         cGeneric.Props,")
					g.Line("Context:       cGeneric.Context,")
					g.Line("operationName: cGeneric.operationName,")
					g.Line("operationType: cGeneric.operationType,")
				})
				g.Line("}")

				g.Line("// Call the user-provided, type-safe handler with the adapted context.")
				g.Line("// The return values are compatible with (any, error).")
				g.Line("return handler(cSpecific)")
			})
			g.Line("}")

			g.Linef("deserializer := func(raw json.RawMessage) (any, error) {")
			g.Block(func() {
				g.Linef("var pre pre%sInput", name)
				g.Line("if err := json.Unmarshal(raw, &pre); err != nil {")
				g.Block(func() { g.Linef("return nil, fmt.Errorf(\"failed to unmarshal %s input: %%w\", err)", name) })
				g.Line("}")
				g.Line("if err := pre.validate(); err != nil { return nil, err }")
				g.Line("typed := pre.transform()")
				g.Line("return typed, nil")
			})
			g.Line("}")

			g.Linef("e.intServer.setProcHandler(\"%s\", adaptedHandler, deserializer)", name)
		})
		g.Line("}")
		g.Break()
	}

	// -----------------------------------------------------------------------------
	// Streams registry and entries
	// -----------------------------------------------------------------------------
	g.Line("// serverStreamRegistry groups all streams and exposes typed entries to register")
	g.Line("// per-stream middlewares, emit middlewares, and the final business handler.")
	g.Line("// Streaming uses Server-Sent Events and the middleware chain is composed per request.")
	g.Line("type serverStreamRegistry[T any] struct {")
	g.Block(func() {
		g.Line("intServer *internalServer[T]")
		for _, streamNode := range sch.GetStreamNodes() {
			name := strutil.ToPascalCase(streamNode.Name)
			g.Linef("%s stream%sEntry[T]", name, name)
		}
	})
	g.Line("}")
	g.Break()

	g.Line("func newServerStreamRegistry[T any](intServer *internalServer[T]) *serverStreamRegistry[T] {")
	g.Block(func() {
		g.Line("r := &serverStreamRegistry[T]{intServer: intServer}")
		for _, streamNode := range sch.GetStreamNodes() {
			name := strutil.ToPascalCase(streamNode.Name)
			g.Linef("r.%s = stream%sEntry[T]{intServer: intServer}", name, name)
		}
		g.Line("return r")
	})
	g.Line("}")
	g.Break()

	for _, streamNode := range sch.GetStreamNodes() {
		name := strutil.ToPascalCase(streamNode.Name)
		g.Linef("// stream%sEntry contains the typed API for the %s stream.", name, name)
		g.Linef("type stream%sEntry[T any] struct {", name)
		g.Block(func() {
			g.Line("intServer *internalServer[T]")
		})
		g.Line("}")
		g.Break()

		// Generate type aliases
		g.Linef("// Type aliases for %s stream", name)
		g.Linef("type %sHandlerContext[T any] = HandlerContext[T, %sInput]", name, name)
		g.Linef("type %sEmitFunc[T any] func(c *%sHandlerContext[T], output %sOutput) error", name, name, name)
		g.Linef("type %sHandlerFunc[T any] func(c *%sHandlerContext[T], emit %sEmitFunc[T]) error", name, name, name)
		g.Linef("type %sMiddlewareFunc[T any] func(next %sHandlerFunc[T]) %sHandlerFunc[T]", name, name, name)
		g.Linef("type %sEmitMiddlewareFunc[T any] func(next %sEmitFunc[T]) %sEmitFunc[T]", name, name, name)
		g.Break()

		// Generate Use (stream middleware)
		g.Linef("// Use registers a typed middleware for the %s stream.", name)
		g.Linef("//")
		g.Linef("// This function allows you to add a middleware specific to the %s stream.", name)
		g.Linef("// The middleware is applied to the stream's handler chain, enabling you to intercept,")
		g.Linef("// modify, or augment the handling of incoming stream requests for %s.", name)
		g.Line("//")
		g.Line("// Execution order: middlewares run in registration order, then the handler.")
		renderDoc(g, streamNode.Doc, true)
		renderDeprecated(g, streamNode.Deprecated)
		g.Linef("func (e stream%sEntry[T]) Use(mw %sMiddlewareFunc[T]) {", name, name)
		g.Block(func() {
			g.Linef("adapted := func(next StreamHandlerFunc[T, any, any]) StreamHandlerFunc[T, any, any] {")
			g.Block(func() {
				g.Line("// This is the generic handler that will be executed by the server at runtime.")
				g.Linef("return func(cGeneric *HandlerContext[T, any], emitGeneric EmitFunc[T, any, any]) error {")
				g.Block(func() {
					g.Line("// Create a type-safe 'next' function for the specific middleware to call.")
					g.Line("// This function acts as a bridge to translate the call back into the generic world.")
					g.Linef("typedNext := func(c *%sHandlerContext[T], emit %sEmitFunc[T]) error {", name, name)
					g.Block(func() {
						g.Line("// Crucially, sync mutations from the specific context back to the generic")
						g.Line("// context before proceeding down the chain.")
						g.Line("cGeneric.Props = c.Props")
						g.Line("cGeneric.Input = c.Input")

						g.Line("// Call the original generic handler.")
						g.Line("return next(cGeneric, emitGeneric)")
					})
					g.Line("}")

					g.Line("// Apply the user's middleware, giving it our typed bridge function.")
					g.Line("// The result is the complete, type-safe handler chain.")
					g.Line("typedChain := mw(typedNext)")

					g.Line("// Create a type-safe 'emit' function that delegates to the generic one.")
					g.Line("// It uses 'cGeneric' from the outer scope, which is the correct context.")
					g.Linef("emitSpecific := func(c *%sHandlerContext[T], output %sOutput) error {", name, name)
					g.Block(func() {
						g.Line("return emitGeneric(cGeneric, output)")
					})
					g.Line("}")

					g.Line("// Prepare the initial arguments for the typed chain by creating a")
					g.Line("// specific context from the generic one.")
					g.Linef("input, _ := cGeneric.Input.(%sInput)", name)
					g.Linef("cSpecific := &%sHandlerContext[T]{", name)
					g.Block(func() {
						g.Line("Input:         input,")
						g.Line("Props:         cGeneric.Props,")
						g.Line("Context:       cGeneric.Context,")
						g.Line("operationName: cGeneric.operationName,")
						g.Line("operationType: cGeneric.operationType,")
					})
					g.Line("}")

					g.Line("// Execute the fully composed, type-safe middleware chain.")
					g.Line("return typedChain(cSpecific, emitSpecific)")
				})
				g.Line("}")
			})
			g.Line("}")
			g.Linef("e.intServer.addStreamMiddleware(\"%s\", adapted)", name)
		})
		g.Line("}")
		g.Break()

		// UseEmit (emit middleware)
		g.Linef("// UseEmit registers a typed emit middleware for the %s stream.", name)
		g.Line("//")
		g.Line("// Emit middlewares wrap every call to emit inside your handler, allowing you to")
		g.Line("// transform, filter, decorate, or audit outgoing events in a type-safe way.")
		g.Line("//")
		g.Line("// Execution order: emit middlewares run in registration order for every event.")
		renderDoc(g, streamNode.Doc, true)
		renderDeprecated(g, streamNode.Deprecated)
		g.Linef("func (e stream%sEntry[T]) UseEmit(mw %sEmitMiddlewareFunc[T]) {", name, name)
		g.Block(func() {
			g.Linef("adapted := func(next EmitFunc[T, any, any]) EmitFunc[T, any, any] {")
			g.Block(func() {
				g.Line("// Return a new generic 'emit' function that wraps the logic for every event.")
				g.Line("return func(cGeneric *HandlerContext[T, any], outputGeneric any) error {")
				g.Block(func() {
					g.Line("// Create a type-safe 'next' function for the specific emit middleware to call.")
					g.Line("// This function acts as a bridge, calling the original generic 'next' function.")
					g.Linef("typedNext := func(c *%sHandlerContext[T], output %sOutput) error {", name, name)
					g.Block(func() {
						g.Line("// Crucially, sync mutations from the specific context back to the generic")
						g.Line("// context before proceeding down the chain.")
						g.Line("cGeneric.Props = c.Props")
						g.Line("cGeneric.Input = c.Input")

						g.Line("// Call the original generic 'next' function with the updated context.")
						g.Line("return next(cGeneric, output)")
					})
					g.Line("}")

					g.Line("// Apply the user's middleware, giving it our typed bridge function.")
					g.Line("// The result is the complete, type-safe emit chain.")
					g.Line("emitChain := mw(typedNext)")

					g.Line("// Prepare the arguments for the typed chain by creating a specific context")
					g.Line("// and asserting the output type.")
					g.Linef("input, _ := cGeneric.Input.(%sInput)", name)
					g.Linef("cSpecific := &%sHandlerContext[T]{", name)
					g.Block(func() {
						g.Line("Input:         input,")
						g.Line("Props:         cGeneric.Props,")
						g.Line("Context:       cGeneric.Context,")
						g.Line("operationName: cGeneric.operationName,")
						g.Line("operationType: cGeneric.operationType,")
					})
					g.Line("}")
					g.Linef("outputSpecific, _ := outputGeneric.(%sOutput)", name)

					g.Line("// Execute the fully composed, type-safe emit middleware chain.")
					g.Line("return emitChain(cSpecific, outputSpecific)")
				})
				g.Line("}")
			})
			g.Line("}")
			g.Linef("e.intServer.addStreamEmitMiddleware(\"%s\", adapted)", name)
		})
		g.Line("}")
		g.Break()

		// Handle (stream handler)
		g.Linef("// Handle registers the business handler for the %s stream.", name)
		g.Line("//")
		g.Line("// The server will:")
		g.Line("//  1) Deserialize and validate the input using generated pre* types")
		g.Line("//  2) Build the stream's middleware chain and the emit chain")
		g.Line("//  3) Provide a typed emit function and invoke your handler")
		renderDoc(g, streamNode.Doc, true)
		renderDeprecated(g, streamNode.Deprecated)
		g.Linef("func (e stream%sEntry[T]) Handle(handler %sHandlerFunc[T]) {", name, name)
		g.Block(func() {
			g.Linef("adaptedHandler := func(cGeneric *HandlerContext[T, any], emitGeneric EmitFunc[T, any, any]) error {")
			g.Block(func() {
				g.Line("// Create the specific, type-safe emit function by wrapping the generic one.")
				g.Line("// It uses 'cGeneric' from the outer scope, which has the correct type for the generic call.")
				g.Linef("emitSpecific := func(c *%sHandlerContext[T], output %sOutput) error {", name, name)
				g.Block(func() {
					g.Line("return emitGeneric(cGeneric, output)")
				})
				g.Line("}")

				g.Line("// Create the specific context from the generic one provided by the server.")
				g.Line("// It's assumed a higher layer guarantees the type is correct.")
				g.Linef("input, _ := cGeneric.Input.(%sInput)", name)
				g.Linef("cSpecific := &%sHandlerContext[T]{", name)
				g.Block(func() {
					g.Line("Input:         input,")
					g.Line("Props:         cGeneric.Props,")
					g.Line("Context:       cGeneric.Context,")
					g.Line("operationName: cGeneric.operationName,")
					g.Line("operationType: cGeneric.operationType,")
				})
				g.Line("}")

				g.Line("// Call the user-provided, type-safe handler with the adapted arguments.")
				g.Line("return handler(cSpecific, emitSpecific)")
			})
			g.Line("}")

			g.Linef("deserializer := func(raw json.RawMessage) (any, error) {")
			g.Block(func() {
				g.Linef("var pre pre%sInput", name)
				g.Line("if err := json.Unmarshal(raw, &pre); err != nil {")
				g.Block(func() { g.Linef("return nil, fmt.Errorf(\"failed to unmarshal %s input: %%w\", err)", name) })
				g.Line("}")
				g.Line("if err := pre.validate(); err != nil { return nil, err }")
				g.Line("typed := pre.transform()")
				g.Line("return typed, nil")
			})
			g.Line("}")

			g.Linef("e.intServer.setStreamHandler(\"%s\", adaptedHandler, deserializer)", name)
		})
		g.Line("}")
		g.Break()
	}

	return g.String(), nil
}

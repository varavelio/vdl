package golang

import (
	_ "embed"
	"fmt"

	"github.com/varavelio/gen"
	"github.com/varavelio/vdl/toolchain/internal/codegen/config"
	"github.com/varavelio/vdl/toolchain/internal/codegen/config/configtypes"
	"github.com/varavelio/vdl/toolchain/internal/core/ir/irtypes"
	"github.com/varavelio/vdl/toolchain/internal/util/strutil"
)

//go:embed pieces/server.go
var serverRawPiece string

// generateServerCore generates the core server implementation (rpc_server.go).
func generateServerCore(_ *irtypes.IrSchema, cfg *configtypes.GoConfig) (string, error) {
	if !config.ShouldGenServer(cfg.GenServer) {
		return "", nil
	}

	piece := strutil.GetStrAfter(serverRawPiece, "/** START FROM HERE **/")
	if piece == "" {
		return "", fmt.Errorf("server.go: could not find start delimiter")
	}

	g := gen.New().WithTabs()

	// Core server piece (types + internal implementation)
	g.Raw(piece)
	g.Break()

	g.Line("// -----------------------------------------------------------------------------")
	g.Line("// Server generated implementation")
	g.Line("// -----------------------------------------------------------------------------")
	g.Break()

	// Server facade
	g.Line("// Server provides a high-level, type-safe API for building a VDL RPC server.")
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
		g.Line("RPCs      *serverRPCRegistry[T]")
	})
	g.Line("}")
	g.Break()

	g.Line("// NewServer creates a new VDL RPC server instance ready to handle all")
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
		g.Line("intServer := newInternalServer[T](VDLProcedures, VDLStreams)")
		g.Line("return &Server[T]{")
		g.Block(func() {
			g.Line("intServer: intServer,")
			g.Line("RPCs:      newServerRPCRegistry(intServer),")
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
	g.Line("func (s *Server[T]) Use(mw GlobalMiddlewareFunc[T]) { s.intServer.addGlobalMiddleware(mw) }")
	g.Break()

	g.Line("// SetStreamConfig sets the global configuration for all streams.")
	g.Line("//")
	g.Line("// This applies to all streams unless overridden by RPC-level or stream-specific configurations (if set).")
	g.Line("func (s *Server[T]) SetStreamConfig(cfg StreamConfig) { s.intServer.setGlobalStreamConfig(cfg) }")
	g.Break()

	g.Line("// SetErrorHandler sets a global error handler that intercepts and transforms errors")
	g.Line("// from all RPCs before sending them to the client.")
	g.Line("//")
	g.Line("// This handler applies to all RPCs unless a specific handler is registered for an RPC.")
	g.Line("func (s *Server[T]) SetErrorHandler(fn ErrorHandlerFunc[T]) { s.intServer.setGlobalErrorHandler(fn) }")
	g.Break()

	g.Line("// HandleRequest processes an incoming RPC request and drives the complete")
	g.Line("// request lifecycle (parsing, middleware chains, handler dispatch, response).")
	g.Line("//")
	g.Line("// rpcName and operationName must be extracted from the request URL (e.g. /rpc/MyService/GetUser -> \"MyService\", \"GetUser\").")
	g.Line("// httpAdapter bridges VDL RPC with your HTTP framework (use NewNetHTTPAdapter for net/http).")
	g.Line("//")
	g.Line("// Example (net/http):")
	g.Line("//   http.HandleFunc(\"POST /rpc/{rpcName}/{operationName}\", func(w http.ResponseWriter, r *http.Request) {")
	g.Line("//       ctx := r.Context()")
	g.Line("//       props := AppProps{UserID: \"abc\"}")
	g.Line("//       adapter := NewNetHTTPAdapter(w, r)")
	g.Line("//       _ = server.HandleRequest(ctx, props, r.PathValue(\"rpcName\"), r.PathValue(\"operationName\"), adapter)")
	g.Line("//   })")
	g.Line("func (s *Server[T]) HandleRequest(ctx context.Context, props T, rpcName, operationName string, httpAdapter HTTPAdapter) error {")
	g.Block(func() {
		g.Line("return s.intServer.handleRequest(ctx, props, rpcName, operationName, httpAdapter)")
	})
	g.Line("}")
	g.Break()

	g.Line("type serverRPCRegistry[T any] struct {")
	g.Block(func() {
		g.Line("intServer *internalServer[T]")
	})
	g.Line("}")
	g.Break()

	g.Line("func newServerRPCRegistry[T any](intServer *internalServer[T]) *serverRPCRegistry[T] {")
	g.Block(func() {
		g.Line("return &serverRPCRegistry[T]{intServer: intServer}")
	})
	g.Line("}")
	g.Break()

	return g.String(), nil
}

// generateServerRPC generates the server implementation for a specific RPC (rpc_{rpcName}_server.go).
func generateServerRPC(rpcName string, procs []irtypes.ProcedureDef, streams []irtypes.StreamDef, cfg *configtypes.GoConfig) (string, error) {
	if !config.ShouldGenServer(cfg.GenServer) {
		return "", nil
	}

	g := gen.New().WithTabs()

	rpcStructName := fmt.Sprintf("server%sRPC", rpcName)
	procsStructName := fmt.Sprintf("server%sProcs", rpcName)
	streamsStructName := fmt.Sprintf("server%sStreams", rpcName)

	// 1. Method on serverRPCRegistry to get this RPC
	g.Linef("// %s returns the registry for the %s RPC service.", rpcName, rpcName)
	g.Linef("func (r *serverRPCRegistry[T]) %s() *%s[T] {", rpcName, rpcStructName)
	g.Block(func() {
		g.Linef("return &%s[T]{", rpcStructName)
		g.Line("intServer: r.intServer,")
		g.Linef("Procs: &%s[T]{intServer: r.intServer},", procsStructName)
		g.Linef("Streams: &%s[T]{intServer: r.intServer},", streamsStructName)
		g.Line("}")
	})
	g.Line("}")
	g.Break()

	// 2. Struct for this RPC
	g.Linef("type %s[T any] struct {", rpcStructName)
	g.Block(func() {
		g.Line("intServer *internalServer[T]")
		g.Linef("Procs     *%s[T]", procsStructName)
		g.Linef("Streams   *%s[T]", streamsStructName)
	})
	g.Line("}")
	g.Break()

	// 3. Use method for this RPC
	g.Linef("// Use registers a middleware that executes for every request within the %s RPC.", rpcName)
	g.Linef("func (r *%s[T]) Use(mw GlobalMiddlewareFunc[T]) {", rpcStructName)
	g.Block(func() {
		g.Linef("r.intServer.addRPCMiddleware(%q, mw)", rpcName)
	})
	g.Line("}")
	g.Break()

	g.Linef("// SetStreamConfig sets the configuration for all streams within the %s RPC.", rpcName)
	g.Line("//")
	g.Line("// This overrides the global configuration but is overridden by stream-specific configuration (if set).")
	g.Linef("func (r *%s[T]) SetStreamConfig(cfg StreamConfig) {", rpcStructName)
	g.Block(func() {
		g.Linef("r.intServer.setRPCStreamConfig(%q, cfg)", rpcName)
	})
	g.Line("}")
	g.Break()

	g.Linef("// SetErrorHandler sets an error handler specifically for the %s RPC.", rpcName)
	g.Line("//")
	g.Line("// This handler overrides the global error handler for all operations within this RPC.")
	g.Linef("func (r *%s[T]) SetErrorHandler(fn ErrorHandlerFunc[T]) {", rpcStructName)
	g.Block(func() {
		g.Linef("r.intServer.setRPCErrorHandler(%q, fn)", rpcName)
	})
	g.Line("}")
	g.Break()

	// 4. Procs Registry Struct
	g.Linef("type %s[T any] struct {", procsStructName)
	g.Block(func() {
		g.Line("intServer *internalServer[T]")
	})
	g.Line("}")
	g.Break()

	// 5. Streams Registry Struct
	g.Linef("type %s[T any] struct {", streamsStructName)
	g.Block(func() {
		g.Line("intServer *internalServer[T]")
	})
	g.Line("}")
	g.Break()

	// Procedures
	for _, proc := range procs {
		uniqueName := rpcName + proc.Name

		g.Linef("// Register the %s procedure.", proc.Name)
		g.Linef("func (r *%s[T]) %s() proc%sEntry[T] {", procsStructName, proc.Name, uniqueName)
		g.Block(func() {
			g.Linef("return proc%sEntry[T]{intServer: r.intServer}", uniqueName)
		})
		g.Line("}")
		g.Break()

		g.Linef("// proc%sEntry contains the typed API for the %s procedure.", uniqueName, uniqueName)
		g.Linef("type proc%sEntry[T any] struct {", uniqueName)
		g.Block(func() {
			g.Line("intServer *internalServer[T]")
		})
		g.Line("}")
		g.Break()

		// Generate type aliases
		g.Linef("// Type aliases for %s procedure", uniqueName)
		g.Linef("type %sHandlerContext[T any] = HandlerContext[T, %sInput]", uniqueName, uniqueName)
		g.Linef("type %sHandlerFunc[T any] func(c *%sHandlerContext[T]) (%sOutput, error)", uniqueName, uniqueName, uniqueName)
		g.Linef("type %sMiddlewareFunc[T any] func(next %sHandlerFunc[T]) %sHandlerFunc[T]", uniqueName, uniqueName, uniqueName)
		g.Break()

		// Use (procedure middleware)
		g.Linef("// Use registers a typed middleware for the %s procedure.", uniqueName)
		g.Line("//")
		g.Line("// The middleware wraps the business handler registered with Handle, allowing you")
		g.Line("// to implement cross-cutting concerns such as validation, logging, auth, or")
		g.Line("// metrics in a type-safe way.")
		g.Line("//")
		g.Line("// Execution order: middlewares run in the order they were registered,")
		g.Line("// then the final handler is invoked.")
		if proc.GetDoc() != "" {
			renderDoc(g, proc.GetDoc(), true)
		}
		renderDeprecated(g, proc.Deprecated)
		g.Linef("func (e proc%sEntry[T]) Use(mw %sMiddlewareFunc[T]) {", uniqueName, uniqueName)
		g.Block(func() {
			g.Linef("adapted := func(next ProcHandlerFunc[T, any, any]) ProcHandlerFunc[T, any, any] {")
			g.Block(func() {
				g.Line("// This is the generic handler that will be executed by the server at runtime.")
				g.Linef("return func(cGeneric *HandlerContext[T, any]) (any, error) {")
				g.Block(func() {
					g.Line("// Create a type-safe 'next' function for the specific middleware to call.")
					g.Line("// This function acts as a bridge to translate the call back into the generic world.")
					g.Linef("typedNext := func(c *%sHandlerContext[T]) (%sOutput, error) {", uniqueName, uniqueName)
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
							g.Linef("var zero %sOutput", uniqueName)
							g.Line("return zero, err")
						})
						g.Line("}")

						g.Line("// On success, assert the 'any' output to the specific output type.")
						g.Linef("specificOutput, _ := genericOutput.(%sOutput)", uniqueName)
						g.Line("return specificOutput, nil")
					})
					g.Line("}")

					g.Line("// Apply the user's middleware, giving it our typed bridge function.")
					g.Line("// The result is the complete, type-safe handler chain.")
					g.Line("typedChain := mw(typedNext)")

					g.Line("// Prepare the initial arguments for the typed chain by creating a")
					g.Line("// specific context from the generic one.")
					g.Linef("input, _ := cGeneric.Input.(%sInput)", uniqueName)
					g.Linef("cSpecific := &%sHandlerContext[T]{", uniqueName)
					g.Block(func() {
						g.Line("Input:   input,")
						g.Line("Props:   cGeneric.Props,")
						g.Line("Context: cGeneric.Context,")
						g.Line("operation: OperationDefinition{")
						g.Block(func() {
							g.Line("RPCName: cGeneric.RPCName(),")
							g.Line("Name:    cGeneric.OperationName(),")
							g.Line("Type:    cGeneric.OperationType(),")
						})
						g.Line("},")
					})
					g.Line("}")

					g.Line("// Execute the fully composed, type-safe middleware chain.")
					g.Line("return typedChain(cSpecific)")
				})
				g.Line("}")
			})
			g.Line("}")
			g.Linef("e.intServer.addProcMiddleware(%q, %q, adapted)", rpcName, proc.Name)
		})
		g.Line("}")
		g.Break()

		// Handle (procedure handler)
		g.Linef("// Handle registers the business handler for the %s procedure.", uniqueName)
		g.Line("//")
		g.Line("// The server will:")
		g.Line("//  1) Deserialize and validate the input using generated pre* types")
		g.Line("//  2) Build the procedure's middleware chain")
		g.Line("//  3) Invoke your handler with a typed context")
		if proc.GetDoc() != "" {
			renderDoc(g, proc.GetDoc(), true)
		}
		renderDeprecated(g, proc.Deprecated)
		g.Linef("func (e proc%sEntry[T]) Handle(handler %sHandlerFunc[T]) {", uniqueName, uniqueName)
		g.Block(func() {
			g.Linef("adaptedHandler := func(cGeneric *HandlerContext[T, any]) (any, error) {")
			g.Block(func() {
				g.Line("// Create the specific context from the generic one provided by the server.")
				g.Line("// It's assumed a higher layer guarantees the type is correct.")
				g.Linef("input, _ := cGeneric.Input.(%sInput)", uniqueName)
				g.Linef("cSpecific := &%sHandlerContext[T]{", uniqueName)
				g.Block(func() {
					g.Line("Input:   input,")
					g.Line("Props:   cGeneric.Props,")
					g.Line("Context: cGeneric.Context,")
					g.Line("operation: OperationDefinition{")
					g.Block(func() {
						g.Line("RPCName: cGeneric.RPCName(),")
						g.Line("Name:    cGeneric.OperationName(),")
						g.Line("Type:    cGeneric.OperationType(),")
					})
					g.Line("},")
				})
				g.Line("}")

				g.Line("// Call the user-provided, type-safe handler with the adapted context.")
				g.Line("// The return values are compatible with (any, error).")
				g.Line("return handler(cSpecific)")
			})
			g.Line("}")

			g.Linef("deserializer := func(raw json.RawMessage) (any, error) {")
			g.Block(func() {
				g.Linef("var pre pre%sInput", uniqueName)
				g.Line("if err := json.Unmarshal(raw, &pre); err != nil {")
				g.Block(func() { g.Linef("return nil, fmt.Errorf(\"failed to unmarshal %s input: %%w\", err)", uniqueName) })
				g.Line("}")
				g.Line("if err := pre.validate(); err != nil { return nil, err }")
				g.Line("typed := pre.transform()")
				g.Line("return typed, nil")
			})
			g.Line("}")

			g.Linef("e.intServer.setProcHandler(%q, %q, adaptedHandler, deserializer)", rpcName, proc.Name)
		})
		g.Line("}")
		g.Break()
	}

	// Streams
	for _, stream := range streams {
		uniqueName := rpcName + stream.Name

		g.Linef("// Register the %s stream.", stream.Name)
		g.Linef("func (r *%s[T]) %s() stream%sEntry[T] {", streamsStructName, stream.Name, uniqueName)
		g.Block(func() {
			g.Linef("return stream%sEntry[T]{intServer: r.intServer}", uniqueName)
		})
		g.Line("}")
		g.Break()

		g.Linef("// stream%sEntry contains the typed API for the %s stream.", uniqueName, uniqueName)
		g.Linef("type stream%sEntry[T any] struct {", uniqueName)
		g.Block(func() {
			g.Line("intServer *internalServer[T]")
		})
		g.Line("}")
		g.Break()

		// Generate type aliases
		g.Linef("// Type aliases for %s stream", uniqueName)
		g.Linef("type %sHandlerContext[T any] = HandlerContext[T, %sInput]", uniqueName, uniqueName)
		g.Linef("type %sEmitFunc[T any] func(c *%sHandlerContext[T], output %sOutput) error", uniqueName, uniqueName, uniqueName)
		g.Linef("type %sHandlerFunc[T any] func(c *%sHandlerContext[T], emit %sEmitFunc[T]) error", uniqueName, uniqueName, uniqueName)
		g.Linef("type %sMiddlewareFunc[T any] func(next %sHandlerFunc[T]) %sHandlerFunc[T]", uniqueName, uniqueName, uniqueName)
		g.Linef("type %sEmitMiddlewareFunc[T any] func(next %sEmitFunc[T]) %sEmitFunc[T]", uniqueName, uniqueName, uniqueName)
		g.Break()

		g.Linef("// SetConfig sets the configuration for the %s stream.", uniqueName)
		g.Line("//")
		g.Line("// This overrides both global and RPC-level configurations.")
		g.Linef("func (e stream%sEntry[T]) SetConfig(cfg StreamConfig) {", uniqueName)
		g.Block(func() {
			g.Linef("e.intServer.setStreamConfig(%q, %q, cfg)", rpcName, stream.Name)
		})
		g.Line("}")
		g.Break()

		// Generate Use (stream middleware)
		g.Linef("// Use registers a typed middleware for the %s stream.", uniqueName)
		g.Linef("//")
		g.Linef("// This function allows you to add a middleware specific to the %s stream.", uniqueName)
		g.Linef("// The middleware is applied to the stream's handler chain, enabling you to intercept,")
		g.Linef("// modify, or augment the handling of incoming stream requests for %s.", uniqueName)
		g.Line("//")
		g.Line("// Execution order: middlewares run in registration order, then the handler.")
		if stream.GetDoc() != "" {
			renderDoc(g, stream.GetDoc(), true)
		}
		renderDeprecated(g, stream.Deprecated)
		g.Linef("func (e stream%sEntry[T]) Use(mw %sMiddlewareFunc[T]) {", uniqueName, uniqueName)
		g.Block(func() {
			g.Linef("adapted := func(next StreamHandlerFunc[T, any, any]) StreamHandlerFunc[T, any, any] {")
			g.Block(func() {
				g.Line("// This is the generic handler that will be executed by the server at runtime.")
				g.Linef("return func(cGeneric *HandlerContext[T, any], emitGeneric EmitFunc[T, any, any]) error {")
				g.Block(func() {
					g.Line("// Create a type-safe 'next' function for the specific middleware to call.")
					g.Line("// This function acts as a bridge to translate the call back into the generic world.")
					g.Linef("typedNext := func(c *%sHandlerContext[T], emit %sEmitFunc[T]) error {", uniqueName, uniqueName)
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
					g.Linef("emitSpecific := func(c *%sHandlerContext[T], output %sOutput) error {", uniqueName, uniqueName)
					g.Block(func() {
						g.Line("return emitGeneric(cGeneric, output)")
					})
					g.Line("}")

					g.Line("// Prepare the initial arguments for the typed chain by creating a")
					g.Line("// specific context from the generic one.")
					g.Linef("input, _ := cGeneric.Input.(%sInput)", uniqueName)
					g.Linef("cSpecific := &%sHandlerContext[T]{", uniqueName)
					g.Block(func() {
						g.Line("Input:   input,")
						g.Line("Props:   cGeneric.Props,")
						g.Line("Context: cGeneric.Context,")
						g.Line("operation: OperationDefinition{")
						g.Block(func() {
							g.Line("RPCName: cGeneric.RPCName(),")
							g.Line("Name:    cGeneric.OperationName(),")
							g.Line("Type:    cGeneric.OperationType(),")
						})
						g.Line("},")
					})
					g.Line("}")

					g.Line("// Execute the fully composed, type-safe middleware chain.")
					g.Line("return typedChain(cSpecific, emitSpecific)")
				})
				g.Line("}")
			})
			g.Line("}")
			g.Linef("e.intServer.addStreamMiddleware(%q, %q, adapted)", rpcName, stream.Name)
		})
		g.Line("}")
		g.Break()

		// UseEmit (emit middleware)
		g.Linef("// UseEmit registers a typed emit middleware for the %s stream.", uniqueName)
		g.Line("//")
		g.Line("// Emit middlewares wrap every call to emit inside your handler, allowing you to")
		g.Line("// transform, filter, decorate, or audit outgoing events in a type-safe way.")
		g.Line("//")
		g.Line("// Execution order: emit middlewares run in registration order for every event.")
		if stream.GetDoc() != "" {
			renderDoc(g, stream.GetDoc(), true)
		}
		renderDeprecated(g, stream.Deprecated)
		g.Linef("func (e stream%sEntry[T]) UseEmit(mw %sEmitMiddlewareFunc[T]) {", uniqueName, uniqueName)
		g.Block(func() {
			g.Linef("adapted := func(next EmitFunc[T, any, any]) EmitFunc[T, any, any] {")
			g.Block(func() {
				g.Line("// Return a new generic 'emit' function that wraps the logic for every event.")
				g.Line("return func(cGeneric *HandlerContext[T, any], outputGeneric any) error {")
				g.Block(func() {
					g.Line("// Create a type-safe 'next' function for the specific emit middleware to call.")
					g.Line("// This function acts as a bridge, calling the original generic 'next' function.")
					g.Linef("typedNext := func(c *%sHandlerContext[T], output %sOutput) error {", uniqueName, uniqueName)
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
					g.Linef("input, _ := cGeneric.Input.(%sInput)", uniqueName)
					g.Linef("cSpecific := &%sHandlerContext[T]{", uniqueName)
					g.Block(func() {
						g.Line("Input:   input,")
						g.Line("Props:   cGeneric.Props,")
						g.Line("Context: cGeneric.Context,")
						g.Line("operation: OperationDefinition{")
						g.Block(func() {
							g.Line("RPCName: cGeneric.RPCName(),")
							g.Line("Name:    cGeneric.OperationName(),")
							g.Line("Type:    cGeneric.OperationType(),")
						})
						g.Line("},")
					})
					g.Line("}")
					g.Linef("outputSpecific, _ := outputGeneric.(%sOutput)", uniqueName)

					g.Line("// Execute the fully composed, type-safe emit middleware chain.")
					g.Line("return emitChain(cSpecific, outputSpecific)")
				})
				g.Line("}")
			})
			g.Line("}")
			g.Linef("e.intServer.addStreamEmitMiddleware(%q, %q, adapted)", rpcName, stream.Name)
		})
		g.Line("}")
		g.Break()

		// Handle (stream handler)
		g.Linef("// Handle registers the business handler for the %s stream.", uniqueName)
		g.Line("//")
		g.Line("// The server will:")
		g.Line("//  1) Deserialize and validate the input using generated pre* types")
		g.Line("//  2) Build the stream's middleware chain and the emit chain")
		g.Line("//  3) Provide a typed emit function and invoke your handler")
		if stream.GetDoc() != "" {
			renderDoc(g, stream.GetDoc(), true)
		}
		renderDeprecated(g, stream.Deprecated)
		g.Linef("func (e stream%sEntry[T]) Handle(handler %sHandlerFunc[T]) {", uniqueName, uniqueName)
		g.Block(func() {
			g.Linef("adaptedHandler := func(cGeneric *HandlerContext[T, any], emitGeneric EmitFunc[T, any, any]) error {")
			g.Block(func() {
				g.Line("// Create the specific, type-safe emit function by wrapping the generic one.")
				g.Line("// It uses 'cGeneric' from the outer scope, which has the correct type for the generic call.")
				g.Linef("emitSpecific := func(c *%sHandlerContext[T], output %sOutput) error {", uniqueName, uniqueName)
				g.Block(func() {
					g.Line("return emitGeneric(cGeneric, output)")
				})
				g.Line("}")

				g.Line("// Create the specific context from the generic one provided by the server.")
				g.Line("// It's assumed a higher layer guarantees the type is correct.")
				g.Linef("input, _ := cGeneric.Input.(%sInput)", uniqueName)
				g.Linef("cSpecific := &%sHandlerContext[T]{", uniqueName)
				g.Block(func() {
					g.Line("Input:   input,")
					g.Line("Props:   cGeneric.Props,")
					g.Line("Context: cGeneric.Context,")
					g.Line("operation: OperationDefinition{")
					g.Block(func() {
						g.Line("RPCName: cGeneric.RPCName(),")
						g.Line("Name:    cGeneric.OperationName(),")
						g.Line("Type:    cGeneric.OperationType(),")
					})
					g.Line("},")
				})
				g.Line("}")

				g.Line("// Call the user-provided, type-safe handler with the adapted arguments.")
				g.Line("return handler(cSpecific, emitSpecific)")
			})
			g.Line("}")

			g.Linef("deserializer := func(raw json.RawMessage) (any, error) {")
			g.Block(func() {
				g.Linef("var pre pre%sInput", uniqueName)
				g.Line("if err := json.Unmarshal(raw, &pre); err != nil {")
				g.Block(func() { g.Linef("return nil, fmt.Errorf(\"failed to unmarshal %s input: %%w\", err)", uniqueName) })
				g.Line("}")
				g.Line("if err := pre.validate(); err != nil { return nil, err }")
				g.Line("typed := pre.transform()")
				g.Line("return typed, nil")
			})
			g.Line("}")

			g.Linef("e.intServer.setStreamHandler(%q, %q, adaptedHandler, deserializer)", rpcName, stream.Name)
		})
		g.Line("}")
		g.Break()
	}

	return g.String(), nil
}

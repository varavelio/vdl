package typescript

import (
	_ "embed"
	"fmt"

	"github.com/varavelio/gen"
	"github.com/varavelio/vdl/toolchain/internal/codegen/config"
	"github.com/varavelio/vdl/toolchain/internal/codegen/config/configtypes"
	"github.com/varavelio/vdl/toolchain/internal/core/ir/irtypes"
	"github.com/varavelio/vdl/toolchain/internal/util/strutil"
)

//go:embed pieces/server.ts
var serverRawPiece string

// generateServer generates the complete server implementation.
func generateServer(schema *irtypes.IrSchema, cfg *configtypes.TypeScriptConfig) (string, error) {
	if !config.ShouldGenServer(cfg.GenServer) {
		return "", nil
	}

	g := gen.New().WithSpaces(2)

	generateImport(g, []string{"Response", "OperationDefinition", "OperationType"}, "./core", true, cfg)
	generateImport(g, []string{"VdlError", "asError"}, "./core", false, cfg)
	generateImport(g, []string{"VDLProcedures", "VDLStreams"}, "./catalog", false, cfg)
	generateImportAll(g, "vdlTypes", "./types", cfg)
	g.Break()

	core, err := generateServerCore(schema, cfg)
	if err != nil {
		return "", err
	}
	g.Raw(core)
	g.Break()

	for _, rpc := range schema.Rpcs {
		rpcCode, err := generateServerRPC(rpc, schema, cfg)
		if err != nil {
			return "", err
		}
		g.Raw(rpcCode)
		g.Break()
	}

	return g.String(), nil
}

// generateServerCore generates the core server implementation (server.ts).
func generateServerCore(schema *irtypes.IrSchema, cfg *configtypes.TypeScriptConfig) (string, error) {
	if !config.ShouldGenServer(cfg.GenServer) {
		return "", nil
	}

	piece := strutil.GetStrAfter(serverRawPiece, "/** START FROM HERE **/")
	if piece == "" {
		return "", fmt.Errorf("server.ts: could not find start delimiter")
	}

	g := gen.New().WithSpaces(2)

	// Core server piece (types + internal implementation)
	g.Raw(piece)
	g.Break()

	g.Line("// -----------------------------------------------------------------------------")
	g.Line("// Server generated implementation")
	g.Line("// -----------------------------------------------------------------------------")
	g.Break()

	// Server Facade
	g.Line("/**")
	g.Line(" * Server provides a high-level, type-safe API for building a VDL RPC server.")
	g.Line(" *")
	g.Line(" * It exposes:")
	g.Line(" *   - procs: typed entries to register middlewares and the business handler per procedure")
	g.Line(" *   - streams: typed entries to register middlewares, emit middlewares and the handler per stream")
	g.Line(" *   - use: a global middleware API that runs for every operation (procedures and streams)")
	g.Line(" *")
	g.Line(" * The generic type parameter T is your application context (props) that flows through")
	g.Line(" * the entire request lifecycle (authentication, per-request data, dependencies, etc.).")
	g.Line(" */")
	g.Line("export class Server<T = unknown> {")
	g.Block(func() {
		g.Line("private intServer: InternalServer<T>;")
		g.Line("public readonly rpcs: ServerRPCRegistry<T>;")
		g.Break()

		g.Line("constructor() {")
		g.Block(func() {
			g.Line("this.intServer = new InternalServer<T>(VDLProcedures, VDLStreams);")
			g.Line("this.rpcs = new ServerRPCRegistry(this.intServer);")
		})
		g.Line("}")
		g.Break()

		g.Line("/**")
		g.Line(" * Registers a global middleware that executes for every request (procedures and streams).")
		g.Line(" *")
		g.Line(" * Middlewares are executed in registration order and can:")
		g.Line(" *   - read/augment the HandlerContext")
		g.Line(" *   - short-circuit by returning an error")
		g.Line(" *   - call next to continue the chain")
		g.Line(" */")
		g.Line("use(mw: GlobalMiddlewareFunc<T>): void {")
		g.Block(func() {
			g.Line("this.intServer.addGlobalMiddleware(mw);")
		})
		g.Line("}")
		g.Break()

		g.Line("/**")
		g.Line(" * Sets the global configuration for all streams.")
		g.Line(" *")
		g.Line(" * This applies to all streams unless overridden by RPC-level or stream-specific configurations.")
		g.Line(" */")
		g.Line("setStreamConfig(cfg: StreamConfig): void {")
		g.Block(func() {
			g.Line("this.intServer.setGlobalStreamConfig(cfg);")
		})
		g.Line("}")
		g.Break()

		g.Line("/**")
		g.Line(" * Sets a global error handler that intercepts and transforms errors")
		g.Line(" * from all RPCs before sending them to the client.")
		g.Line(" *")
		g.Line(" * This handler applies to all RPCs unless a specific handler is registered for an RPC.")
		g.Line(" */")
		g.Line("setErrorHandler(fn: ErrorHandlerFunc<T>): void {")
		g.Block(func() {
			g.Line("this.intServer.setGlobalErrorHandler(fn);")
		})
		g.Line("}")
		g.Break()

		g.Line("/**")
		g.Line(" * HandleRequest processes an incoming RPC request and drives the complete")
		g.Line(" * request lifecycle (parsing, middleware chains, handler dispatch, response).")
		g.Line(" *")
		g.Line(" * rpcName and operationName must be extracted from the request URL.")
		g.Line(" * httpAdapter bridges VDL RPC with your HTTP framework.")
		g.Line(" */")
		g.Line("async handleRequest(")
		g.Block(func() {
			g.Line("props: T,")
			g.Line("rpcName: string,")
			g.Line("operationName: string,")
			g.Line("httpAdapter: HTTPAdapter")
		})
		g.Line("): Promise<void> {")
		g.Block(func() {
			g.Line("return this.intServer.handleRequest(props, rpcName, operationName, httpAdapter);")
		})
		g.Line("}")
	})
	g.Line("}")
	g.Break()

	// ServerRPCRegistry
	g.Line("export class ServerRPCRegistry<T> {")
	g.Block(func() {
		g.Line("private intServer: InternalServer<T>;")
		g.Line("constructor(intServer: InternalServer<T>) { this.intServer = intServer; }")
		g.Break()

		for _, rpc := range schema.Rpcs {
			rpcName := rpc.Name
			rpcPascal := strutil.ToPascalCase(rpcName)
			methodName := strutil.ToCamelCase(rpcName)
			structName := fmt.Sprintf("Server%sRPC", rpcPascal)

			g.Linef("/** Access the %s RPC. */", rpcName)
			g.Linef("%s(): %s<T> {", methodName, structName)
			g.Block(func() {
				g.Linef("return new %s(this.intServer);", structName)
			})
			g.Line("}")
			g.Break()
		}
	})
	g.Line("}")
	g.Break()

	return g.String(), nil
}

// generateServerRPC generates the server implementation for a specific RPC.
func generateServerRPC(rpc irtypes.RpcDef, schema *irtypes.IrSchema, cfg *configtypes.TypeScriptConfig) (string, error) {
	if !config.ShouldGenServer(cfg.GenServer) {
		return "", nil
	}

	g := gen.New().WithSpaces(2)

	rpcName := rpc.Name
	rpcPascal := strutil.ToPascalCase(rpcName)
	rpcStructName := fmt.Sprintf("Server%sRPC", rpcPascal)
	procsStructName := fmt.Sprintf("Server%sProcs", rpcPascal)
	streamsStructName := fmt.Sprintf("Server%sStreams", rpcPascal)

	// 1. Extend ServerRPCRegistry
	// Since we can't easily extend the existing class definition, we will generate the
	// specific RPC accessor on the ServerRPCRegistry class in generateServerCore (not here).
	// But we generate the RPC class itself here.

	// 2. Class for this RPC
	g.Linef("export class %s<T> {", rpcStructName)
	g.Block(func() {
		g.Line("private intServer: InternalServer<T>;")
		g.Linef("public readonly procs: %s<T>;", procsStructName)
		g.Linef("public readonly streams: %s<T>;", streamsStructName)
		g.Break()

		g.Line("constructor(intServer: InternalServer<T>) {")
		g.Block(func() {
			g.Line("this.intServer = intServer;")
			g.Linef("this.procs = new %s(intServer);", procsStructName)
			g.Linef("this.streams = new %s(intServer);", streamsStructName)
		})
		g.Line("}")
		g.Break()

		g.Linef("/** Registers a middleware that executes for every request within the %s RPC. */", rpcName)
		g.Line("use(mw: GlobalMiddlewareFunc<T>): void {")
		g.Block(func() {
			g.Linef("this.intServer.addRPCMiddleware(\"%s\", mw);", rpcName)
		})
		g.Line("}")
		g.Break()

		g.Linef("/** Sets the configuration for all streams within the %s RPC. */", rpcName)
		g.Line("setStreamConfig(cfg: StreamConfig): void {")
		g.Block(func() {
			g.Linef("this.intServer.setRPCStreamConfig(\"%s\", cfg);", rpcName)
		})
		g.Line("}")
		g.Break()

		g.Linef("/** Sets an error handler specifically for the %s RPC. */", rpcName)
		g.Line("setErrorHandler(fn: ErrorHandlerFunc<T>): void {")
		g.Block(func() {
			g.Linef("this.intServer.setRPCErrorHandler(\"%s\", fn);", rpcName)
		})
		g.Line("}")
	})
	g.Line("}")
	g.Break()

	// 3. Procs Registry Class
	g.Linef("export class %s<T> {", procsStructName)
	g.Block(func() {
		g.Line("private intServer: InternalServer<T>;")
		g.Line("constructor(intServer: InternalServer<T>) { this.intServer = intServer; }")
		g.Break()

		for _, proc := range schema.Procedures {
			if proc.RpcName != rpcName {
				continue
			}
			uniqueName := rpcPascal + strutil.ToPascalCase(proc.Name)
			entryName := fmt.Sprintf("Proc%sEntry", uniqueName)

			g.Linef("/** Register the %s procedure. */", proc.Name)
			g.Linef("%s(): %s<T> {", proc.Name, entryName)
			g.Block(func() {
				g.Linef("return new %s(this.intServer);", entryName)
			})
			g.Line("}")
			g.Break()
		}
	})
	g.Line("}")
	g.Break()

	// 4. Streams Registry Class
	g.Linef("export class %s<T> {", streamsStructName)
	g.Block(func() {
		g.Line("private intServer: InternalServer<T>;")
		g.Line("constructor(intServer: InternalServer<T>) { this.intServer = intServer; }")
		g.Break()

		for _, stream := range schema.Streams {
			if stream.RpcName != rpcName {
				continue
			}
			uniqueName := rpcPascal + strutil.ToPascalCase(stream.Name)
			entryName := fmt.Sprintf("Stream%sEntry", uniqueName)

			g.Linef("/** Register the %s stream. */", stream.Name)
			g.Linef("%s(): %s<T> {", stream.Name, entryName)
			g.Block(func() {
				g.Linef("return new %s(this.intServer);", entryName)
			})
			g.Line("}")
			g.Break()
		}
	})
	g.Line("}")
	g.Break()

	// 5. Procedure Entry Classes
	for _, proc := range schema.Procedures {
		if proc.RpcName != rpcName {
			continue
		}
		fullName := strutil.ToPascalCase(proc.RpcName) + strutil.ToPascalCase(proc.Name)
		uniqueName := rpcPascal + strutil.ToPascalCase(proc.Name)
		entryName := fmt.Sprintf("Proc%sEntry", uniqueName)
		inputName := fmt.Sprintf("%sInput", fullName)
		outputName := fmt.Sprintf("%sOutput", fullName)

		g.Linef("export class %s<T> {", entryName)
		g.Block(func() {
			g.Line("private intServer: InternalServer<T>;")
			g.Line("constructor(intServer: InternalServer<T>) { this.intServer = intServer; }")
			g.Break()

			g.Line("/**")
			g.Linef(" * Registers a typed middleware for the %s procedure.", uniqueName)
			g.Line(" *")
			g.Line(" * The middleware wraps the business handler, allowing you")
			g.Line(" * to implement cross-cutting concerns in a type-safe way.")
			g.Line(" */")
			g.Linef("use(mw: ProcMiddlewareFunc<T, vdlTypes.%s, vdlTypes.%s>): void {", inputName, outputName)
			g.Block(func() {
				g.Line("const adapted: ProcMiddlewareFunc<T, any, any> = (next) => {")
				g.Block(func() {
					g.Line("return async (cGeneric) => {")
					g.Block(func() {
						g.Linef("const typedNext: ProcHandlerFunc<T, vdlTypes.%s, vdlTypes.%s> = async (c) => {", inputName, outputName)
						g.Block(func() {
							g.Line("cGeneric.props = c.props;")
							g.Line("cGeneric.input = c.input;")
							g.Line("const genericOutput = await next(cGeneric);")
							g.Linef("return genericOutput as vdlTypes.%s;", outputName)
						})
						g.Line("};")
						g.Break()
						g.Line("const typedChain = mw(typedNext);")
						g.Break()
						g.Linef("const input = cGeneric.input as vdlTypes.%s;", inputName)
						g.Linef("const cSpecific = new HandlerContext<T, vdlTypes.%s>(", inputName)
						g.Line("cGeneric.props, input, cGeneric.signal, cGeneric.operation")
						g.Line(");")
						g.Break()
						g.Line("return typedChain(cSpecific);")
					})
					g.Line("};")
				})
				g.Line("};")
				g.Linef("this.intServer.addProcMiddleware(\"%s\", \"%s\", adapted);", rpcName, proc.Name)
			})
			g.Line("}")
			g.Break()

			g.Line("/**")
			g.Linef(" * Registers the business handler for the %s procedure.", uniqueName)
			g.Line(" */")
			g.Linef("handle(handler: ProcHandlerFunc<T, vdlTypes.%s, vdlTypes.%s>): void {", inputName, outputName)
			g.Block(func() {
				g.Line("const adaptedHandler: ProcHandlerFunc<T, any, any> = async (cGeneric) => {")
				g.Block(func() {
					g.Linef("const input = cGeneric.input as vdlTypes.%s;", inputName)
					g.Linef("const cSpecific = new HandlerContext<T, vdlTypes.%s>(", inputName)
					g.Line("cGeneric.props, input, cGeneric.signal, cGeneric.operation")
					g.Line(");")
					g.Line("return handler(cSpecific);")
				})
				g.Line("};")
				g.Break()
				g.Line("const deserializer: DeserializerFunc = async (raw) => {")
				g.Block(func() {
					validateFnName := fmt.Sprintf("validate%sInput", fullName)
					g.Linef("const err = vdlTypes.%s(raw);", validateFnName)
					g.Line("if (err !== null) {")
					g.Block(func() {
						g.Line("throw new VdlError({ message: err, code: \"INVALID_INPUT\", category: \"ValidationError\" });")
					})
					g.Line("}")
					g.Line("return raw;")
				})
				g.Line("};")
				g.Break()
				g.Linef("this.intServer.setProcHandler(\"%s\", \"%s\", adaptedHandler, deserializer);", rpcName, proc.Name)
			})
			g.Line("}")
		})
		g.Line("}")
		g.Break()
	}

	// 6. Stream Entry Classes
	for _, stream := range schema.Streams {
		if stream.RpcName != rpcName {
			continue
		}
		fullName := strutil.ToPascalCase(stream.RpcName) + strutil.ToPascalCase(stream.Name)
		uniqueName := rpcPascal + strutil.ToPascalCase(stream.Name)
		entryName := fmt.Sprintf("Stream%sEntry", uniqueName)
		inputName := fmt.Sprintf("%sInput", fullName)
		outputName := fmt.Sprintf("%sOutput", fullName)

		g.Linef("export class %s<T> {", entryName)
		g.Block(func() {
			g.Line("private intServer: InternalServer<T>;")
			g.Line("constructor(intServer: InternalServer<T>) { this.intServer = intServer; }")
			g.Break()

			g.Linef("/** Sets the configuration for the %s stream. */", uniqueName)
			g.Line("setConfig(cfg: StreamConfig): void {")
			g.Block(func() {
				g.Linef("this.intServer.setStreamConfig(\"%s\", \"%s\", cfg);", rpcName, stream.Name)
			})
			g.Line("}")
			g.Break()

			g.Linef("/** Registers a typed middleware for the %s stream. */", uniqueName)
			g.Linef("use(mw: StreamMiddlewareFunc<T, vdlTypes.%s, vdlTypes.%s>): void {", inputName, outputName)
			g.Block(func() {
				g.Line("const adapted: StreamMiddlewareFunc<T, any, any> = (next) => {")
				g.Block(func() {
					g.Line("return async (cGeneric, emitGeneric) => {")
					g.Block(func() {
						g.Linef("const typedNext: StreamHandlerFunc<T, vdlTypes.%s, vdlTypes.%s> = async (c, emit) => {", inputName, outputName)
						g.Block(func() {
							g.Line("cGeneric.props = c.props;")
							g.Line("cGeneric.input = c.input;")
							g.Line("return next(cGeneric, emitGeneric);")
						})
						g.Line("};")
						g.Break()
						g.Line("const typedChain = mw(typedNext);")
						g.Break()
						g.Linef("const emitSpecific: EmitFunc<T, vdlTypes.%s, vdlTypes.%s> = async (c, output) => {", inputName, outputName)
						g.Block(func() {
							g.Line("return emitGeneric(cGeneric, output);")
						})
						g.Line("};")
						g.Break()
						g.Linef("const input = cGeneric.input as vdlTypes.%s;", inputName)
						g.Linef("const cSpecific = new HandlerContext<T, vdlTypes.%s>(", inputName)
						g.Line("cGeneric.props, input, cGeneric.signal, cGeneric.operation")
						g.Line(");")
						g.Break()
						g.Line("return typedChain(cSpecific, emitSpecific);")
					})
					g.Line("};")
				})
				g.Line("};")
				g.Linef("this.intServer.addStreamMiddleware(\"%s\", \"%s\", adapted);", rpcName, stream.Name)
			})
			g.Line("}")
			g.Break()

			g.Linef("/** Registers a typed emit middleware for the %s stream. */", uniqueName)
			g.Linef("useEmit(mw: EmitMiddlewareFunc<T, vdlTypes.%s, vdlTypes.%s>): void {", inputName, outputName)
			g.Block(func() {
				g.Line("const adapted: EmitMiddlewareFunc<T, any, any> = (next) => {")
				g.Block(func() {
					g.Line("return async (cGeneric, outputGeneric) => {")
					g.Block(func() {
						g.Linef("const typedNext: EmitFunc<T, vdlTypes.%s, vdlTypes.%s> = async (c, output) => {", inputName, outputName)
						g.Block(func() {
							g.Line("cGeneric.props = c.props;")
							g.Line("cGeneric.input = c.input;")
							g.Line("return next(cGeneric, output);")
						})
						g.Line("};")
						g.Break()
						g.Line("const emitChain = mw(typedNext);")
						g.Break()
						g.Linef("const input = cGeneric.input as vdlTypes.%s;", inputName)
						g.Linef("const cSpecific = new HandlerContext<T, vdlTypes.%s>(", inputName)
						g.Line("cGeneric.props, input, cGeneric.signal, cGeneric.operation")
						g.Line(");")
						g.Linef("const outputSpecific = outputGeneric as vdlTypes.%s;", outputName)
						g.Break()
						g.Line("return emitChain(cSpecific, outputSpecific);")
					})
					g.Line("};")
				})
				g.Line("};")
				g.Linef("this.intServer.addStreamEmitMiddleware(\"%s\", \"%s\", adapted);", rpcName, stream.Name)
			})
			g.Line("}")
			g.Break()

			g.Linef("/** Registers the business handler for the %s stream. */", uniqueName)
			g.Linef("handle(handler: StreamHandlerFunc<T, vdlTypes.%s, vdlTypes.%s>): void {", inputName, outputName)
			g.Block(func() {
				g.Line("const adaptedHandler: StreamHandlerFunc<T, any, any> = async (cGeneric, emitGeneric) => {")
				g.Block(func() {
					g.Linef("const emitSpecific: EmitFunc<T, vdlTypes.%s, vdlTypes.%s> = async (c, output) => {", inputName, outputName)
					g.Block(func() {
						g.Line("return emitGeneric(cGeneric, output);")
					})
					g.Line("};")
					g.Break()
					g.Linef("const input = cGeneric.input as vdlTypes.%s;", inputName)
					g.Linef("const cSpecific = new HandlerContext<T, vdlTypes.%s>(", inputName)
					g.Line("cGeneric.props, input, cGeneric.signal, cGeneric.operation")
					g.Line(");")
					g.Break()
					g.Line("return handler(cSpecific, emitSpecific);")
				})
				g.Line("};")
				g.Break()
				g.Line("const deserializer: DeserializerFunc = async (raw) => {")
				g.Block(func() {
					validateFnName := fmt.Sprintf("validate%sInput", fullName)
					g.Linef("const err = vdlTypes.%s(raw);", validateFnName)
					g.Line("if (err !== null) {")
					g.Block(func() {
						g.Line("throw new VdlError({ message: err, code: \"INVALID_INPUT\", category: \"ValidationError\" });")
					})
					g.Line("}")
					g.Line("return raw;")
				})
				g.Line("};")
				g.Break()
				g.Linef("this.intServer.setStreamHandler(\"%s\", \"%s\", adaptedHandler, deserializer);", rpcName, stream.Name)
			})
			g.Line("}")
		})
		g.Line("}")
		g.Break()
	}

	return g.String(), nil
}

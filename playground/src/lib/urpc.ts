import type { Schema } from "./urpcTypes.ts";

/**
 * Docs: https://go.dev/wiki/WebAssembly
 */

/**
 * Load a script asynchronously
 *
 * @param src The script source
 * @returns A promise that resolves when the script has been loaded
 */
function loadScript(src: string): Promise<void> {
  return new Promise((resolve, reject) => {
    const s = document.createElement("script");
    s.src = src;
    s.onload = () => resolve();
    s.onerror = () => reject(new Error(`failed to load ${src}`));
    document.head.appendChild(s);
  });
}

/**
 * Check if the wasm has been initialized
 *
 * @returns True if the wasm has been initialized, false otherwise
 */
function isInitialized(): boolean {
  // biome-ignore lint/suspicious/noExplicitAny: it's a global value
  return (window as any).__urpcWasmReady;
}

/**
 * Wait until the wasm has been initialized
 *
 * @returns A promise that resolves when the wasm has been initialized
 */
function waitUntilInitialized(): Promise<void> {
  if (isInitialized()) return Promise.resolve();

  return new Promise((resolve) => {
    const interval = setInterval(() => {
      if (isInitialized()) {
        clearInterval(interval);
        resolve();
      }
    }, 100);
  });
}

/**
 * Initialize the wasm
 *
 * @returns A promise that resolves when the wasm has been initialized
 */
async function initWasm(): Promise<void> {
  const execURL = "./app/urpc/wasm_exec.js";
  const wasmURL = "./app/urpc/urpc.wasm";

  if (isInitialized()) return;
  await loadScript(execURL);

  // biome-ignore lint/suspicious/noExplicitAny: it's a global function
  const go = new (window as any).Go();
  const { instance } = await WebAssembly.instantiateStreaming(
    await fetch(wasmURL),
    go.importObject,
  );
  go.run(instance);

  // biome-ignore lint/suspicious/noExplicitAny: it's a global value
  (window as any).__urpcWasmReady = true;
}

/**
 * Format an URPC schema
 *
 * @param input The URPC schema to format
 * @returns The formatted URPC schema
 */
async function cmdFmt(input: string): Promise<string> {
  await waitUntilInitialized();
  // biome-ignore lint/suspicious/noExplicitAny: it's a global function
  return (window as any).cmdFmt(input);
}

/**
 * Transpile an URPC schema to JSON and vice versa based on the original
 * extension
 *
 * @param sourceExt The original extension of the file (.json or .urpc)
 * @param input The schema to transpile
 * @returns The transpiled schema
 */
async function cmdTranspile(sourceExt: string, input: string): Promise<string> {
  await waitUntilInitialized();
  // biome-ignore lint/suspicious/noExplicitAny: it's a global function
  return (window as any).cmdTranspile(sourceExt, input);
}

/**
 * Transpile an URPC schema to JSON
 *
 * @param input The URPC schema to transpile
 * @returns The transpiled JSON schema as a typed JSON object
 */
async function transpileUrpcToJson(input: string): Promise<Schema> {
  return JSON.parse(await cmdTranspile("urpc", input));
}

/**
 * Code generation types and command (WASM)
 */
export type CodegenGenerator =
  | "golang-server"
  | "golang-client"
  | "typescript-client"
  | "dart-client";

export interface CmdCodegenOptions {
  /** The generator to use */
  generator: CodegenGenerator;
  /** The URPC schema content */
  schemaInput: string;
  /** Required when generator is golang-server or golang-client */
  golangPackageName?: string;
  /** Required when generator is dart-client */
  dartPackageName?: string;
}

export interface CmdCodegenOutputFile {
  /** The path where the generated file should be saved */
  path: string;
  /** The content of the generated file */
  content: string;
}

export interface CmdCodegenOutput {
  /** The files that were generated */
  files: CmdCodegenOutputFile[];
}

/**
 * Run code generation inside the WASM module.
 * Accepts options matching the Go CmdCodegenOptions and returns a typed output.
 */
async function cmdCodegen(
  options: CmdCodegenOptions,
): Promise<CmdCodegenOutput> {
  await waitUntilInitialized();
  // biome-ignore lint/suspicious/noExplicitAny: it's a global function
  const json = await (window as any).cmdCodegen(JSON.stringify(options));
  return JSON.parse(json) as CmdCodegenOutput;
}

/**
 * Expand all type references in an URPC schema
 *
 * @param input The URPC schema to expand
 * @returns The expanded URPC schema with all type references inlined
 */
async function cmdExpandTypes(input: string): Promise<string> {
  await waitUntilInitialized();
  // biome-ignore lint/suspicious/noExplicitAny: it's a global function
  return (window as any).cmdExpandTypes(input);
}

/**
 * Extract a specific type declaration from an URPC schema
 *
 * @param input The URPC schema content
 * @param typeName The name of the type to extract
 * @returns The extracted type declaration as an URPC schema string
 */
async function cmdExtractType(
  input: string,
  typeName: string,
): Promise<string> {
  await waitUntilInitialized();
  // biome-ignore lint/suspicious/noExplicitAny: it's a global function
  return (window as any).cmdExtractType(input, typeName);
}

/**
 * Extract a specific proc declaration from an URPC schema
 *
 * @param input The URPC schema content
 * @param procName The name of the proc to extract
 * @returns The extracted proc declaration as an URPC schema string
 */
async function cmdExtractProc(
  input: string,
  procName: string,
): Promise<string> {
  await waitUntilInitialized();
  // biome-ignore lint/suspicious/noExplicitAny: it's a global function
  return (window as any).cmdExtractProc(input, procName);
}

/**
 * Extract a specific stream declaration from an URPC schema
 *
 * @param input The URPC schema content
 * @param streamName The name of the stream to extract
 * @returns The extracted stream declaration as an URPC schema string
 */
async function cmdExtractStream(
  input: string,
  streamName: string,
): Promise<string> {
  await waitUntilInitialized();
  // biome-ignore lint/suspicious/noExplicitAny: it's a global function
  return (window as any).cmdExtractStream(input, streamName);
}

export {
  cmdFmt,
  cmdTranspile,
  cmdCodegen,
  cmdExpandTypes,
  cmdExtractType,
  cmdExtractProc,
  cmdExtractStream,
  initWasm,
  isInitialized,
  transpileUrpcToJson,
  waitUntilInitialized,
};

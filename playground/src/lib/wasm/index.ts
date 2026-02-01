import * as vdl from "./wasmtypes";

// Docs: https://go.dev/wiki/WebAssembly

declare global {
  interface Window {
    // Global function exposed by Go
    wasmExecuteFunction: (jsonInput: string) => Promise<string>;
    // Flag to check if WASM is ready
    __vdlWasmReady?: boolean;
    // Go's wasm_exec.js loader
    Go: any;
  }
}

// Configuration constants
const WASM_EXEC_URL = "./app/vdl/wasm_exec.js";
const WASM_BINARY_URL = "./app/vdl/vdl.wasm";

export class WasmClient {
  /**
   * Initialize the WASM runtime.
   * Downloads the exec script and the binary, then instantiates the module.
   */
  async init(): Promise<void> {
    if (this.isInitialized()) return;

    // Load the Go WASM loader script
    await this.loadScript(WASM_EXEC_URL);

    // Instantiate and run WASM
    const go = new window.Go();
    const { instance } = await WebAssembly.instantiateStreaming(
      await fetch(WASM_BINARY_URL),
      go.importObject
    );
    
    // Run asynchronously
    go.run(instance);
    window.__vdlWasmReady = true;
  }

  /**
   * Check if WASM is fully loaded and ready to execute.
   */
  isInitialized(): boolean {
    return !!window.__vdlWasmReady;
  }

  /**
   * Generate code based on the provided configuration.
   */
  async codegen(input: vdl.CodegenInput): Promise<vdl.CodegenOutput> {
    const res = await this.call({
      functionName: "Codegen",
      codegen: input,
    });

    if (!res.codegen) {
      throw new Error("WASM Error: Output missing 'codegen' field");
    }
    return res.codegen;
  }

  /**
   * Expand all type references in a VDL schema to inline objects.
   */
  async expandTypes(vdlSchema: string): Promise<string> {
    const res = await this.call({
      functionName: "ExpandTypes",
      expandTypes: { vdlSchema },
    });

    if (!res.expandTypes) {
      throw new Error("WASM Error: Output missing 'expandTypes' field");
    }
    return res.expandTypes.expandedSchema;
  }

  /**
   * Extract a specific type declaration from the schema.
   */
  async extractType(vdlSchema: string, typeName: string): Promise<string> {
    const res = await this.call({
      functionName: "ExtractType",
      extractType: { vdlSchema, typeName },
    });

    if (!res.extractType) {
      throw new Error("WASM Error: Output missing 'extractType' field");
    }
    return res.extractType.typeSchema;
  }

  /**
   * Extract a specific procedure declaration from the schema.
   */
  async extractProc(vdlSchema: string, rpcName: string, procName: string): Promise<string> {
    const res = await this.call({
      functionName: "ExtractProc",
      extractProc: { vdlSchema, rpcName, procName },
    });

    if (!res.extractProc) {
      throw new Error("WASM Error: Output missing 'extractProc' field");
    }
    return res.extractProc.procSchema;
  }

  /**
   * Extract a specific stream declaration from the schema.
   */
  async extractStream(vdlSchema: string, rpcName: string, streamName: string): Promise<string> {
    const res = await this.call({
      functionName: "ExtractStream",
      extractStream: { vdlSchema, rpcName, streamName },
    });

    if (!res.extractStream) {
      throw new Error("WASM Error: Output missing 'extractStream' field");
    }
    return res.extractStream.streamSchema;
  }

  // --- Internal Helpers ---

  /**
   * Centralized method to handle initialization, JSON bridging, and execution.
   */
  private async call(input: vdl.WasmInput): Promise<vdl.WasmOutput> {
    await this.waitUntilInitialized();

    // Serialize input (Type-safe via VDL)
    const jsonInput = JSON.stringify(input);

    // Execute via the global Go hook
    const jsonOutput = await window.wasmExecuteFunction(jsonInput);

    // Deserialize output
    return JSON.parse(jsonOutput) as vdl.WasmOutput;
  }

  private loadScript(src: string): Promise<void> {
    return new Promise((resolve, reject) => {
      const s = document.createElement("script");
      s.src = src;
      s.onload = () => resolve();
      s.onerror = () => reject(new Error(`Failed to load ${src}`));
      document.head.appendChild(s);
    });
  }

  private waitUntilInitialized(): Promise<void> {
    if (this.isInitialized()) return Promise.resolve();

    return new Promise((resolve) => {
      const interval = setInterval(() => {
        if (this.isInitialized()) {
          clearInterval(interval);
          resolve();
        }
      }, 50);
    });
  }
}

// Export a singleton instance for ease of use
export const vdlWasm = new WasmClient();


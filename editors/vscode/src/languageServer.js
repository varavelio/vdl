const vscode = require("vscode");
const { LanguageClient } = require("vscode-languageclient/node");
const { TransportKind } = require("vscode-languageclient/node");
const { Executable } = require("vscode-languageclient/node");
const { LanguageClientOptions } = require("vscode-languageclient/node");

/**
 * The language server client instance.
 * @type {LanguageClient}
 */
let client;

/**
 * Starts the language server
 * @param {string} binaryPath The path to the binary
 */
async function startLanguageServer(binaryPath) {
  console.log(`VDL: Binary path: ${binaryPath}`);

  /**
   * Server Options
   * @type {Executable}
   */
  const serverOptions = {
    command: binaryPath,
    args: ["lsp"],
    transport: TransportKind.stdio,
  };

  /**
   * Client Options
   * @type {LanguageClientOptions}
   */
  const clientOptions = {
    documentSelector: [{ scheme: "file", language: "vdl" }],
  };

  // Create and Start the Client
  client = new LanguageClient(
    "vdlLanguageServer",
    "VDL Language Server",
    serverOptions,
    clientOptions,
  );

  console.log(`VDL: Starting Language Server: ${binaryPath}`);
  try {
    await client.start();
    console.log("VDL: Language Server started.");
  } catch (error) {
    console.error("VDL: Failed to start the language server:", error);
    vscode.window.showErrorMessage(
      `VDL: Failed to start Language Server: ${error}`,
    );
  }
}

/**
 * Stops the language server
 */
async function stopLanguageServer() {
  if (!client) {
    console.log("VDL: No Language Server running.");
    return;
  }

  console.log("VDL: Stopping Language Server...");
  try {
    await client.stop();
    console.log("VDL: Language Server stopped.");
  } catch (error) {
    console.error("VDL: Failed to stop the language server:", error);
  }
}

/**
 * Restarts the language server
 * @param {string} binaryPath The path to the binary
 */
async function restartLanguageServer(binaryPath) {
  console.log("VDL: Restarting language server...");
  await stopLanguageServer();
  await startLanguageServer(binaryPath);
  vscode.window.showInformationMessage("VDL: Language Server restarted");
}

module.exports = {
  startLanguageServer,
  stopLanguageServer,
  restartLanguageServer,
};

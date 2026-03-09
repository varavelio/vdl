const vscode = require("vscode");
const { LanguageClient, TransportKind } = require("vscode-languageclient/node");
const { getVdlCommand } = require("./vdlCommand.js");

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
  if (client) {
    console.log("VDL: Language Server already running.");
    return;
  }

  console.log(`VDL: Binary path: ${binaryPath}`);
  const command = getVdlCommand(binaryPath, ["lsp"]);

  /**
   * Server Options
   * @type {Executable}
   */
  const serverOptions = {
    command: command.command,
    args: command.args,
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
  const nextClient = new LanguageClient(
    "vdlLanguageServer",
    "VDL Language Server",
    serverOptions,
    clientOptions,
  );

  console.log(`VDL: Starting Language Server: ${binaryPath}`);
  try {
    await nextClient.start();
    client = nextClient;
    console.log("VDL: Language Server started.");
  } catch (error) {
    client = undefined;
    console.error("VDL: Failed to start the language server:", error);
    throw error;
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
  const currentClient = client;
  client = undefined;

  try {
    await currentClient.stop();
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

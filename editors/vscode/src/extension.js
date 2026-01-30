const vscode = require("vscode");
const getBinaryPath = require("./getBinaryPath.js");
const { initCommand } = require("./initCommand.js");
const {
  startLanguageServer,
  stopLanguageServer,
  restartLanguageServer,
} = require("./languageServer.js");

function isExtensionEnabled() {
  const config = vscode.workspace.getConfiguration("vdl");
  return config.get("enable") === true;
}

async function activate(context) {
  // Idempotent bootstrap function
  const bootstrap = async () => {
    if (!isExtensionEnabled()) {
      console.log("VDL: Extension disabled by settings.");
      return;
    }

    try {
      const binaryPath = getBinaryPath();
      await startLanguageServer(binaryPath);
    } catch (error) {
      vscode.window.showErrorMessage(error.message);
    }
  };

  // Initialization
  await bootstrap();

  // Command Registration (Lazy Loading of path for robustness)
  context.subscriptions.push(
    vscode.commands.registerCommand("vdl.init", () => {
      try {
        const path = getBinaryPath();
        initCommand(path);
      } catch (e) {
        vscode.window.showErrorMessage(e.message);
      }
    }),
  );

  context.subscriptions.push(
    vscode.commands.registerCommand("vdl.restart", async () => {
      try {
        const path = getBinaryPath();
        await restartLanguageServer(path);
      } catch (e) {
        vscode.window.showErrorMessage(e.message);
      }
    }),
  );

  // Configuration Watcher (Reactivity)
  context.subscriptions.push(
    vscode.workspace.onDidChangeConfiguration(async (e) => {
      if (
        e.affectsConfiguration("vdl.enable") ||
        e.affectsConfiguration("vdl.binaryPath")
      ) {
        console.log("VDL: Configuration changed. Reloading services...");
        await stopLanguageServer();
        await bootstrap();
      }
    }),
  );
}

async function deactivate() {
  await stopLanguageServer();
}

module.exports = {
  activate,
  deactivate,
};

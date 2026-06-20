const vscode = require("vscode");
const getBinaryPath = require("./getBinaryPath.js");
const { runVdl } = require("./runVdl.js");

async function initCommand() {
  const activeUri = vscode.window.activeTextEditor?.document?.uri;
  const activeWorkspaceUri = activeUri
    ? vscode.workspace.getWorkspaceFolder(activeUri)?.uri
    : undefined;
  const defaultUri = activeWorkspaceUri || activeUri || vscode.workspace.workspaceFolders?.[0]?.uri;

  const folderUri = await vscode.window.showOpenDialog({
    canSelectFiles: false,
    canSelectFolders: true,
    canSelectMany: false,
    defaultUri,
    openLabel: "Initialize VDL Here",
  });

  if (!folderUri || folderUri.length === 0) {
    return;
  }

  const selectedUri = folderUri[0];
  const folderPath = selectedUri.fsPath;
  const binaryPath = getBinaryPath(selectedUri);

  console.log(`VDL: Initializing at ${folderPath}`);

  try {
    const result = await runVdl(binaryPath, ["init", folderPath]);

    if (result.stdout) {
      console.log(result.stdout);
    }

    vscode.window.showInformationMessage(`VDL initialized successfully in: ${folderPath}`);
  } catch (error) {
    vscode.window.showErrorMessage(`VDL Init Failed: ${error.message}`);
  }
}

module.exports = {
  initCommand,
};

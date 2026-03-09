const vscode = require("vscode");
const { runVdl } = require("./runVdl.js");

async function initCommand(binaryPath) {
  let defaultUri;

  // Open dialog in current project folder
  if (vscode.workspace.workspaceFolders && vscode.workspace.workspaceFolders.length > 0) {
    defaultUri = vscode.workspace.workspaceFolders[0].uri;
  }

  const folderUri = await vscode.window.showOpenDialog({
    canSelectFiles: false,
    canSelectFolders: true,
    canSelectMany: false,
    defaultUri: defaultUri,
    openLabel: "Initialize VDL Here",
  });

  if (!folderUri || folderUri.length === 0) {
    return;
  }

  const folderPath = folderUri[0].fsPath;

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

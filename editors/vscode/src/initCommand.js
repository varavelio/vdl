const vscode = require("vscode");
const execProcess = require("child_process").exec;

async function initCommand(binaryPath) {
  let defaultUri = undefined;

  // Open dialog in current project folder
  if (
    vscode.workspace.workspaceFolders &&
    vscode.workspace.workspaceFolders.length > 0
  ) {
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

  // Quote paths to support spaces (e.g. "Program Files")
  const cmd = `"${binaryPath}" init "${folderPath}"`;

  console.log(`VDL: Initializing at ${folderPath}`);

  execProcess(cmd, (error, stdout, _) => {
    if (error) {
      vscode.window.showErrorMessage(`VDL Init Failed: ${error.message}`);
      return;
    }

    if (stdout) console.log(stdout);

    vscode.window.showInformationMessage(
      `VDL initialized successfully in: ${folderPath}`,
    );
  });
}

module.exports = {
  initCommand,
};

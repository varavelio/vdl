const vscode = require("vscode");
const execProcess = require("child_process").exec;

/**
 * Runs the init command for the VDL binary.
 *
 * @param {string} binaryPath The path to the VDL binary.
 * @returns {Promise<void>} A promise that resolves when the command is complete.
 */
async function initCommand(binaryPath) {
  const folderUri = await vscode.window.showOpenDialog({
    canSelectFiles: false,
    canSelectFolders: true,
    canSelectMany: false,
    openLabel: "Select folder to initialize VDL",
  });

  if (!folderUri || folderUri.length === 0) {
    return;
  }

  const folderPath = folderUri[0].fsPath;
  const initCommand = `${binaryPath} init ${folderPath}`;

  console.log(`VDL: Initializing at ${folderPath}`);
  execProcess(initCommand, (error, stdout, stderr) => {
    if (error) {
      vscode.window.showErrorMessage(
        `VDL: Failed to initialize: ${error.message}`,
      );
      console.error(`VDL: Error initializing schema: ${error.message}`);
      return;
    }

    if (stderr) {
      console.log(`VDL: Init stderr: ${stderr}`);
    }

    if (stdout) {
      console.log(`VDL: Init stdout: ${stdout}`);
    }

    vscode.window.showInformationMessage(`VDL: Initialized at ${folderPath}`);
  });
}

module.exports = {
  initCommand,
};

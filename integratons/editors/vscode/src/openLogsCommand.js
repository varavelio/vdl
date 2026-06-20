const vscode = require("vscode");
const { runVdl } = require("./runVdl.js");

async function openLogsCommand(binaryPath) {
  let logPath;

  try {
    const result = await runVdl(binaryPath, ["lsp", "--log-path"]);
    logPath = result.stdout.trim();
  } catch (error) {
    vscode.window.showErrorMessage(`Failed to get VDL log path: ${error.message}`);
    return;
  }

  if (!logPath) {
    vscode.window.showWarningMessage("VDL returned an empty log path.");
    return;
  }

  try {
    const uri = vscode.Uri.file(logPath);
    const doc = await vscode.workspace.openTextDocument(uri);

    await vscode.window.showTextDocument(doc);
  } catch (error) {
    vscode.window.showErrorMessage(`Failed to open log file at ${logPath}: ${error.message}`);
  }
}

module.exports = {
  openLogsCommand,
};

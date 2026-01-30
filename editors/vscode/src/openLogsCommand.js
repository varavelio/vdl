const vscode = require("vscode");
const cp = require("child_process");

async function openLogsCommand(binaryPath) {
  cp.execFile(binaryPath, ["lsp", "--log-path"], async (error, stdout, _) => {
    if (error) {
      vscode.window.showErrorMessage(
        `Failed to get VDL log path: ${error.message}`,
      );
      return;
    }

    const logPath = stdout.trim();

    if (!logPath) {
      vscode.window.showWarningMessage("VDL returned an empty log path.");
      return;
    }

    try {
      const uri = vscode.Uri.file(logPath);

      const doc = await vscode.workspace.openTextDocument(uri);
      await vscode.window.showTextDocument(doc);
    } catch (err) {
      vscode.window.showErrorMessage(
        `Failed to open log file at ${logPath}: ${err.message}`,
      );
    }
  });
}

module.exports = {
  openLogsCommand,
};

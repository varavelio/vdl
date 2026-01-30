const vscode = require("vscode");
const fs = require("fs");
const path = require("path");
const os = require("os");
const { execSync } = require("child_process");

function getBinaryPath() {
  const config = vscode.workspace.getConfiguration("vdl");
  const configBinaryPath = config.get("binaryPath");

  // 1. Manual Configuration
  if (configBinaryPath && configBinaryPath.trim() !== "") {
    let finalPath = configBinaryPath;
    // Expand '~' for Linux/Mac
    if (finalPath.startsWith("~")) {
      finalPath = path.join(os.homedir(), finalPath.slice(1));
    }

    if (fs.existsSync(finalPath)) {
      return finalPath;
    }
    // Warn instead of throw to allow fallback to PATH
    vscode.window.showWarningMessage(
      `VDL: Configured path not found: ${finalPath}. Searching in PATH...`,
    );
  }

  const isWindows = process.platform === "win32";
  const binaryName = isWindows ? "vdl.exe" : "vdl";

  // 2. GOBIN Environment Variable
  const gobinPath = process.env.GOBIN;
  if (gobinPath) {
    const binaryPath = path.join(gobinPath, binaryName);
    if (fs.existsSync(binaryPath)) {
      return binaryPath;
    }
  }

  // 3. System PATH
  try {
    const command = isWindows ? "where vdl" : "which vdl";
    // stdio: 'pipe' prevents noise in console if it fails
    const stdout = execSync(command, { encoding: "utf8", stdio: "pipe" });

    // Strict cleanup of newlines (Critical for Windows)
    const lines = stdout
      .split(/\r?\n/)
      .map((l) => l.trim())
      .filter((l) => l.length > 0);

    for (const line of lines) {
      if (fs.existsSync(line)) {
        return line;
      }
    }
  } catch (e) {
    // Ignore error if not found in PATH
  }

  let errMsg = "Could not find the vdl binary. ";
  errMsg += "Please ensure it is installed and in your PATH, ";
  errMsg += "or set 'vdl.binaryPath' in VS Code settings.";
  throw new Error(errMsg);
}

module.exports = getBinaryPath;

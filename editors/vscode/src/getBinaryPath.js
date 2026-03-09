const vscode = require("vscode");
const fs = require("node:fs");
const path = require("node:path");
const os = require("node:os");
const { execSync } = require("node:child_process");

function fileExists(filePath) {
  try {
    return Boolean(filePath) && fs.existsSync(filePath);
  } catch (_error) {
    return false;
  }
}

function expandHomeDir(filePath) {
  if (filePath === "~") {
    return os.homedir();
  }

  if (filePath.startsWith("~/") || filePath.startsWith("~\\")) {
    return path.join(os.homedir(), filePath.slice(2));
  }

  return filePath;
}

function getWorkspaceFoldersInPriorityOrder(preferredUri) {
  const workspaceFolders = vscode.workspace.workspaceFolders || [];
  const activeUri = preferredUri || vscode.window.activeTextEditor?.document?.uri;

  if (!activeUri) {
    return workspaceFolders;
  }

  const preferredFolder = vscode.workspace.getWorkspaceFolder(activeUri);

  if (!preferredFolder) {
    return workspaceFolders;
  }

  return [
    preferredFolder,
    ...workspaceFolders.filter(
      (folder) => folder.uri.toString() !== preferredFolder.uri.toString(),
    ),
  ];
}

function getWorkspaceNodeBinCandidates(preferredUri) {
  const isWindows = process.platform === "win32";
  const binaryNames = isWindows ? ["vdl.exe", "vdl.cmd", "vdl.bat", "vdl"] : ["vdl"];
  const candidates = [];

  for (const folder of getWorkspaceFoldersInPriorityOrder(preferredUri)) {
    const binDir = path.join(folder.uri.fsPath, "node_modules", ".bin");

    for (const binaryName of binaryNames) {
      candidates.push(path.join(binDir, binaryName));
    }
  }

  return candidates;
}

function getBinaryPath(preferredUri) {
  const config = vscode.workspace.getConfiguration("vdl");
  const configBinaryPath = config.get("binaryPath");

  // 1. Manual Configuration
  if (configBinaryPath && configBinaryPath.trim() !== "") {
    const finalPath = expandHomeDir(configBinaryPath.trim());

    if (fileExists(finalPath)) {
      return finalPath;
    }

    vscode.window.showWarningMessage(
      `VDL: Configured path not found: ${finalPath}. Searching other locations...`,
    );
  }

  const isWindows = process.platform === "win32";
  const binaryName = isWindows ? "vdl.exe" : "vdl";

  // 2. Workspace-local installation
  for (const candidate of getWorkspaceNodeBinCandidates(preferredUri)) {
    if (fileExists(candidate)) {
      return candidate;
    }
  }

  // 3. GOBIN Environment Variable
  const gobinPath = process.env.GOBIN;
  if (gobinPath) {
    const binaryPath = path.join(gobinPath, binaryName);
    if (fileExists(binaryPath)) {
      return binaryPath;
    }
  }

  // 4. System PATH
  try {
    const command = isWindows ? "where vdl" : "which vdl";
    const stdout = execSync(command, { encoding: "utf8", stdio: "pipe" });

    const lines = stdout
      .split(/\r?\n/)
      .map((l) => l.trim())
      .filter((l) => l.length > 0);

    for (const line of lines) {
      if (fileExists(line)) {
        return line;
      }
    }
  } catch (_e) {
    // Ignore error if not found in PATH
  }

  let errMsg = "Could not find the vdl binary. ";
  errMsg += "Checked 'vdl.binaryPath', workspace 'node_modules/.bin', GOBIN, and PATH. ";
  errMsg += "Please ensure it is installed or set 'vdl.binaryPath' in VS Code settings.";
  throw new Error(errMsg);
}

module.exports = getBinaryPath;

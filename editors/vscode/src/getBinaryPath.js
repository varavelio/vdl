const vscode = require("vscode");
const fs = require("fs");
const path = require("path");
const { execSync } = require("child_process");

/**
 * This function is used to get the binary path of the VDL binary.
 * It checks the following locations in order:
 *
 * 1. The configuration `vdl.binaryPath`
 * 2. The GOBIN environment variable
 * 3. The system PATH
 *
 * The binary to find is `vdl` or `vdl.exe`.
 *
 * If the binary path is not found in any of these locations, it throws an error.
 *
 * @returns {string} The binary path of the VDL binary.
 * @throws {Error} If the binary path is not found in any of the locations.
 */
function getBinaryPath() {
  // 1. Try to get the binary path from the configuration "vdl.binaryPath"
  //    and if found, early return it
  const config = vscode.workspace.getConfiguration("vdl");
  const configBinaryPath = config.get("binaryPath");

  if (configBinaryPath && fs.existsSync(configBinaryPath)) {
    return configBinaryPath;
  }

  const isWindows = process.platform === "win32";
  const binaryName = isWindows ? "vdl.exe" : "vdl";

  // 2. Try to get the binary path from the GOBIN environment variable
  //    and if found, early return it
  const gobinPath = process.env.GOBIN;
  if (gobinPath) {
    const binaryPath = path.join(gobinPath, binaryName);
    if (fs.existsSync(binaryPath)) {
      return binaryPath;
    }
  }

  // 3. Try to get the binary path from the system PATH
  //    and if found, early return it
  try {
    const command = isWindows ? "where vdl" : "which vdl";
    const binaryPath = execSync(command, { encoding: "utf8" }).trim();

    if (binaryPath && fs.existsSync(binaryPath)) {
      return binaryPath;
    }
  } catch {}

  let errMsg = "Could not find the vdl/vdl.exe binary. ";
  errMsg += "Please download it and make sure it's in your PATH. ";
  errMsg +=
    "You can also set a custom binary path in the vdl.binaryPath setting.";
  throw new Error(errMsg);
}

module.exports = getBinaryPath;

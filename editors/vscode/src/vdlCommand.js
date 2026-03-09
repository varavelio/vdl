const path = require("node:path");

function getVdlCommand(binaryPath, args = []) {
  const extension = path.extname(binaryPath).toLowerCase();
  const usesCmdShell =
    process.platform === "win32" && (extension === ".cmd" || extension === ".bat");

  if (usesCmdShell) {
    return {
      command: "cmd.exe",
      args: ["/d", "/c", binaryPath, ...args],
    };
  }

  return {
    command: binaryPath,
    args,
  };
}

module.exports = {
  getVdlCommand,
};

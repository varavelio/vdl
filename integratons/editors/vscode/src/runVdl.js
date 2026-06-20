const cp = require("node:child_process");
const { promisify } = require("node:util");
const { getVdlCommand } = require("./vdlCommand.js");

const execFile = promisify(cp.execFile);

async function runVdl(binaryPath, args, options = {}) {
  const command = getVdlCommand(binaryPath, args);

  return execFile(command.command, command.args, {
    windowsHide: true,
    ...options,
  });
}

module.exports = {
  runVdl,
};

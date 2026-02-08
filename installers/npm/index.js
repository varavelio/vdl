#!/usr/bin/env node

const path = require("node:path");
const fs = require("node:fs");

/**
 * Get the path to the installed VDL binary
 * @returns {string} Absolute path to the vdl binary
 */
function getBinaryPath() {
  const binaryName = process.platform === "win32" ? "vdl.exe" : "vdl";
  const binaryPath = path.join(__dirname, "bin", binaryName);

  if (!fs.existsSync(binaryPath)) {
    throw new Error(
      `VDL binary not found at ${binaryPath}. ` +
        `Installation may have failed. Try reinstalling: npm install @varavel/vdl`,
    );
  }

  return binaryPath;
}

/**
 * Get the version of the installed VDL package (without v prefix)
 * @returns {string} Version string
 */
function getVersion() {
  const packageJson = require("./package.json");
  return packageJson.version;
}

module.exports = {
  getBinaryPath,
  getVersion,
  binaryPath: getBinaryPath(),
};

// If run directly, print the binary path
if (require.main === module) {
  console.log(getBinaryPath());
}

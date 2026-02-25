#!/usr/bin/env node

const { execSync } = require("node:child_process");
const fs = require("node:fs");
const path = require("node:path");
const https = require("node:https");
const crypto = require("node:crypto");

const PLATFORM_MAP = {
  darwin: "darwin",
  linux: "linux",
  win32: "windows",
};

const ARCH_MAP = {
  x64: "amd64",
  arm64: "arm64",
};

/**
 * Gets the platform name in VDL release format.
 * @returns {string} Platform name (darwin, linux, or windows)
 * @throws {Error} If the platform is not supported
 */
function getPlatform() {
  const platform = PLATFORM_MAP[process.platform];
  if (!platform) {
    throw new Error(
      `Unsupported platform: ${process.platform}. VDL supports darwin, linux, and win32.`,
    );
  }
  return platform;
}

/**
 * Gets the architecture name in VDL release format.
 * @returns {string} Architecture name (amd64 or arm64)
 * @throws {Error} If the architecture is not supported
 */
function getArch() {
  const arch = ARCH_MAP[process.arch];
  if (!arch) {
    throw new Error(`Unsupported architecture: ${process.arch}. VDL supports x64 and arm64.`);
  }
  return arch;
}

/**
 * Gets the version from package.json.
 * @returns {string} Version string without 'v' prefix
 */
function getVersion() {
  const packageJson = require("./package.json");
  return packageJson.version.replace(/^v/, "");
}

/**
 * Gets the binary name for the current platform.
 * @returns {string} Binary name (vdl.exe for Windows, vdl for others)
 */
function getBinaryName() {
  return process.platform === "win32" ? "vdl.exe" : "vdl";
}

/**
 * Gets the release filename for the current platform/arch.
 * @returns {string} Filename (e.g., vdl_linux_amd64.tar.gz)
 */
function getReleaseFilename() {
  const platform = getPlatform();
  const arch = getArch();
  const ext = platform === "windows" ? "zip" : "tar.gz";
  return `vdl_${platform}_${arch}.${ext}`;
}

/**
 * Constructs the GitHub release download URL for the current platform.
 * @returns {string} Download URL
 */
function getDownloadURL() {
  const version = getVersion();
  const filename = getReleaseFilename();
  return `https://github.com/varavelio/vdl/releases/download/v${version}/${filename}`;
}

/**
 * Constructs the GitHub release checksums URL.
 * @returns {string} Checksums URL
 */
function getChecksumsURL() {
  const version = getVersion();
  return `https://github.com/varavelio/vdl/releases/download/v${version}/checksums.txt`;
}

/**
 * Verifies the SHA256 checksum of the downloaded buffer.
 * @param {Buffer} binaryBuffer - The downloaded file buffer
 * @param {Buffer} checksumsBuffer - The downloaded checksums.txt buffer
 * @param {string} filename - The expected filename in checksums.txt
 * @throws {Error} If checksum verification fails
 */
function verifyChecksum(binaryBuffer, checksumsBuffer, filename) {
  const checksums = checksumsBuffer.toString("utf8");
  const expectedLine = checksums.split("\n").find((line) => line.trim().endsWith(filename));

  if (!expectedLine) {
    throw new Error(`Checksum for ${filename} not found in checksums.txt`);
  }

  // Checksums file format is: <hash>  <filename>
  const expectedHash = expectedLine.split(/\s+/)[0].trim();

  const calculatedHash = crypto.createHash("sha256").update(binaryBuffer).digest("hex");

  if (expectedHash !== calculatedHash) {
    throw new Error(
      `Checksum verification failed for ${filename}.\nExpected: ${expectedHash}\nCalculated: ${calculatedHash}`,
    );
  }
}

/**
 * Ensures the bin directory exists and returns its path.
 * @returns {string} Absolute path to the bin directory
 */
function ensureBinDir() {
  const binDir = path.join(__dirname, "bin");
  if (!fs.existsSync(binDir)) fs.mkdirSync(binDir, { recursive: true });
  return binDir;
}

/**
 * Extracts a tar.gz archive and locates the VDL binary.
 * @param {Buffer} buffer - Archive buffer
 * @param {string} binDir - Target bin directory
 * @param {string} binaryName - Expected binary name
 * @returns {Promise<void>}
 * @throws {Error} If extraction fails or binary not found
 */
function extractTarGz(buffer, binDir, binaryName) {
  return new Promise((resolve, reject) => {
    const now = Date.now().toString();
    const tempTar = path.join(binDir, `${now}_temp.tar.gz`);
    const tempDir = path.join(binDir, `${now}_temp_extract`);

    try {
      fs.writeFileSync(tempTar, buffer);

      if (!fs.existsSync(tempDir)) {
        fs.mkdirSync(tempDir, { recursive: true });
      }

      execSync(`tar -xzf "${tempTar}" -C "${tempDir}"`, {
        stdio: "pipe",
      });

      let binaryFound = false;
      const searchPaths = [path.join(tempDir, binaryName), path.join(tempDir, "vdl", binaryName)];

      for (const searchPath of searchPaths) {
        if (fs.existsSync(searchPath)) {
          const destPath = path.join(binDir, binaryName);
          fs.copyFileSync(searchPath, destPath);
          fs.chmodSync(destPath, 0o755);
          binaryFound = true;
          break;
        }
      }

      if (!binaryFound) {
        const files = execSync(`find "${tempDir}" -type f`, {
          encoding: "utf8",
        });
        reject(new Error(`Binary ${binaryName} not found in archive.\nFound files:\n${files}`));
      } else {
        resolve();
      }
    } catch (error) {
      reject(new Error(`Failed to extract tar.gz: ${error.message}`));
    } finally {
      try {
        if (fs.existsSync(tempTar)) {
          fs.unlinkSync(tempTar);
        }
        if (fs.existsSync(tempDir)) {
          fs.rmSync(tempDir, { recursive: true, force: true });
        }
      } catch (_) {}
    }
  });
}

/**
 * Extracts a zip archive and locates the VDL binary.
 * @param {Buffer} buffer - Archive buffer
 * @param {string} binDir - Target bin directory
 * @param {string} binaryName - Expected binary name
 * @returns {Promise<void>}
 * @throws {Error} If extraction fails or binary not found
 */
function extractZip(buffer, binDir, binaryName) {
  return new Promise((resolve, reject) => {
    const now = Date.now().toString();
    const tempZip = path.join(binDir, `${now}_temp.zip`);
    const tempDir = path.join(binDir, `${now}_temp_extract`);

    try {
      fs.writeFileSync(tempZip, buffer);

      if (!fs.existsSync(tempDir)) {
        fs.mkdirSync(tempDir, { recursive: true });
      }

      if (process.platform === "win32") {
        execSync(
          `powershell -command "Expand-Archive -Path '${tempZip}' -DestinationPath '${tempDir}' -Force"`,
          { stdio: "pipe" },
        );
      } else {
        execSync(`unzip -q "${tempZip}" -d "${tempDir}"`, { stdio: "pipe" });
      }

      let binaryFound = false;
      const searchPaths = [path.join(tempDir, binaryName), path.join(tempDir, "vdl", binaryName)];

      for (const searchPath of searchPaths) {
        if (fs.existsSync(searchPath)) {
          const destPath = path.join(binDir, binaryName);
          fs.copyFileSync(searchPath, destPath);
          if (process.platform !== "win32") {
            fs.chmodSync(destPath, 0o755);
          }
          binaryFound = true;
          break;
        }
      }

      if (!binaryFound) {
        reject(new Error(`Binary ${binaryName} not found after extraction`));
      } else {
        resolve();
      }
    } catch (error) {
      reject(new Error(`Failed to extract zip: ${error.message}`));
    } finally {
      try {
        if (fs.existsSync(tempZip)) {
          fs.unlinkSync(tempZip);
        }
        if (fs.existsSync(tempDir)) {
          fs.rmSync(tempDir, { recursive: true, force: true });
        }
      } catch (_) {}
    }
  });
}

/**
 * Downloads a file from the given URL with redirect support.
 * @param {string} url - URL to download from
 * @returns {Promise<Buffer>} Downloaded file as buffer
 * @throws {Error} If download fails
 */
function download(url) {
  return new Promise((resolve, reject) => {
    https
      .get(url, (response) => {
        if (
          response.statusCode === 301 ||
          response.statusCode === 302 ||
          response.statusCode === 307 ||
          response.statusCode === 308
        ) {
          return download(response.headers.location).then(resolve).catch(reject);
        }

        if (response.statusCode !== 200) {
          reject(
            new Error(
              `Failed to download: HTTP ${response.statusCode}\n` +
                `URL: ${url}\n` +
                `This usually means the binary for your platform hasn't been released yet.`,
            ),
          );
          return;
        }

        const chunks = [];
        response.on("data", (chunk) => {
          chunks.push(chunk);
        });

        response.on("end", () => {
          resolve(Buffer.concat(chunks));
        });

        response.on("error", reject);
      })
      .on("error", reject);
  });
}

/**
 * Main installation routine.
 * Downloads and installs the VDL binary for the current platform.
 * @returns {Promise<void>}
 */
async function install() {
  try {
    const platform = getPlatform();
    const binaryName = getBinaryName();
    const filename = getReleaseFilename();
    const downloadUrl = getDownloadURL();
    const checksumsUrl = getChecksumsURL();
    const binDir = ensureBinDir();

    console.log(`VDL: Downloading ${filename}...`);

    const [binaryBuffer, checksumsBuffer] = await Promise.all([
      download(downloadUrl),
      download(checksumsUrl),
    ]);

    console.log("VDL: Verifying checksum...");
    verifyChecksum(binaryBuffer, checksumsBuffer, filename);

    console.log("VDL: Extracting...");
    if (platform === "windows") {
      await extractZip(binaryBuffer, binDir, binaryName);
    } else {
      await extractTarGz(binaryBuffer, binDir, binaryName);
    }

    const binaryPath = path.join(binDir, binaryName);
    if (!fs.existsSync(binaryPath)) {
      throw new Error(`Binary not found at ${binaryPath} after installation`);
    }

    if (process.platform !== "win32") {
      fs.chmodSync(binaryPath, 0o755);
    }

    console.log(`VDL: Installation complete!`);
    console.log(`VDL: Run 'vdl --version' to verify.`);
  } catch (error) {
    console.error("VDL: Installation failed");
    console.error(`VDL: ${error.message}`);
    process.exit(1);
  }
}

if (require.main === module) install();
module.exports = { install };

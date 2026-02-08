#!/usr/bin/env node

const { spawn } = require("node:child_process");
const { getBinaryPath } = require("./index.js");
const { install } = require("./install.js");

async function main() {
  let binPath;
  try {
    binPath = getBinaryPath();
  } catch (_) {
    try {
      await install();
      binPath = getBinaryPath();
    } catch (installError) {
      console.error(installError.message);
      process.exit(1);
    }
  }

  const args = process.argv.slice(2);

  // Spawn vdl
  const child = spawn(binPath, args, {
    stdio: ["inherit", "inherit", "inherit"],
  });

  // Propagate OS signals
  const signals = ["SIGINT", "SIGTERM", "SIGHUP"];
  signals.forEach((sig) => {
    process.on(sig, () => {
      if (child.pid) child.kill(sig);
    });
  });

  child.on("exit", (code) => {
    process.exit(code || 0);
  });

  child.on("error", (err) => {
    console.error(err.message);
    process.exit(1);
  });
}

main().catch((err) => {
  console.error(err.message);
  process.exit(1);
});

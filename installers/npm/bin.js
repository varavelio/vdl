#!/usr/bin/env node

const { spawn } = require("node:child_process");
const { getBinaryPath } = require("./index.js");

try {
  const binPath = getBinaryPath();
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
} catch (e) {
  console.error(e.message);
  process.exit(1);
}

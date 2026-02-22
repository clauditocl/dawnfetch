#!/usr/bin/env node
"use strict";

const { spawnSync } = require("node:child_process");

function run(command, args) {
  const res = spawnSync(command, args, { stdio: "inherit", shell: false });
  if (typeof res.status === "number" && res.status === 0) {
    return;
  }
  throw new Error(`command failed: ${command} ${args.join(" ")}`);
}

function installWindows() {
  const cmd =
    'irm https://raw.githubusercontent.com/almightynan/dawnfetch/main/cli/install.ps1 | iex';
  run("powershell", ["-NoProfile", "-Command", cmd]);
}

function installUnix() {
  const cmd =
    "curl -fsSL https://raw.githubusercontent.com/almightynan/dawnfetch/main/cli/install.sh | bash";
  run("bash", ["-lc", cmd]);
}

try {
  if (process.platform === "win32") {
    installWindows();
  } else {
    installUnix();
  }
} catch (err) {
  console.warn("[dawnfetch] npm postinstall failed.");
  console.warn(
    "[dawnfetch] You can install manually with the platform installer from README."
  );
  console.warn(String(err && err.message ? err.message : err));
}


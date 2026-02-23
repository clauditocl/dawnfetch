#!/usr/bin/env node
"use strict";

const { spawnSync } = require("child_process");
const fs = require("fs");
const path = require("path");

const ROOT = path.resolve(__dirname, "..");
const MIN_NODE_MAJOR = 18;

function getNodeMajor() {
  const raw = String(process.versions && process.versions.node ? process.versions.node : "");
  const major = parseInt(raw.split(".")[0], 10);
  return Number.isFinite(major) ? major : 0;
}

function run(command, args) {
  const res = spawnSync(command, args, {
    stdio: "inherit",
    shell: false,
    cwd: ROOT,
    env: process.env
  });
  if (typeof res.status === "number" && res.status === 0) {
    return;
  }
  throw new Error(`command failed: ${command} ${args.join(" ")}`);
}

function installWindows() {
  const script = path.join(ROOT, "cli", "install.ps1");
  if (!fs.existsSync(script)) {
    throw new Error(`missing installer script: ${script}`);
  }
  run("powershell", ["-NoProfile", "-ExecutionPolicy", "Bypass", "-File", script]);
}

function installUnix() {
  const script = path.join(ROOT, "cli", "install.sh");
  if (!fs.existsSync(script)) {
    throw new Error(`missing installer script: ${script}`);
  }
  run("bash", [script]);
}

function printSuccessHint() {
  if (process.platform === "win32") {
    console.log("[dawnfetch] install complete.");
    console.log("[dawnfetch] run now: dawnfetch");
    console.log(
      '[dawnfetch] if command is not recognized in this window, run:'
    );
    console.log(
      '$env:Path=[Environment]::GetEnvironmentVariable("Path","User")+";"+[Environment]::GetEnvironmentVariable("Path","Machine")'
    );
    console.log("[dawnfetch] or close and reopen PowerShell/cmd.");
    return;
  }
  console.log("[dawnfetch] install complete.");
  console.log("[dawnfetch] run now: dawnfetch");
  console.log(
    "[dawnfetch] if command is not found in this shell, run: source ~/.bashrc (or source ~/.zshrc) or restart terminal."
  );
}

try {
  const nodeMajor = getNodeMajor();
  if (nodeMajor > 0 && nodeMajor < MIN_NODE_MAJOR) {
    console.warn(
      `[dawnfetch] Detected Node ${process.versions.node}. npm install helpers target Node ${MIN_NODE_MAJOR}+.` 
    );
    if (process.platform === "win32") {
      console.warn(
        '[dawnfetch] Use direct installer instead: powershell -c "irm https://raw.githubusercontent.com/almightynan/dawnfetch/main/cli/install.ps1 | iex"'
      );
    } else {
      console.warn(
        "[dawnfetch] Use direct installer instead: curl -fsSL https://raw.githubusercontent.com/almightynan/dawnfetch/main/cli/install.sh | bash"
      );
    }
    process.exit(0);
  }

  if (process.platform === "win32") {
    installWindows();
  } else {
    installUnix();
  }
  printSuccessHint();
} catch (err) {
  console.warn("[dawnfetch] npm postinstall failed.");
  console.warn(
    "[dawnfetch] You can install manually with the platform installer from README."
  );
  console.warn(String(err && err.message ? err.message : err));
}

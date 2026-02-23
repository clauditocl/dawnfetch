#!/usr/bin/env node
"use strict";

const { spawnSync } = require("node:child_process");
const fs = require("node:fs");
const path = require("node:path");

const ROOT = path.resolve(__dirname, "..");

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

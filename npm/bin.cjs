#!/usr/bin/env node
"use strict";

const { spawnSync } = require("node:child_process");
const fs = require("node:fs");
const os = require("node:os");
const path = require("node:path");

const SELF_HEAL_ENV = "DAWNFETCH_NPM_SELF_HEAL";

function candidatePaths() {
  if (process.platform === "win32") {
    const local = process.env.LOCALAPPDATA || path.join(os.homedir(), "AppData", "Local");
    return [path.join(local, "dawnfetch", "dawnfetch.exe")];
  }
  return [path.join(os.homedir(), ".local", "bin", "dawnfetch")];
}

function firstExisting(paths) {
  for (const p of paths) {
    try {
      if (fs.existsSync(p)) return p;
    } catch {
      // ignore
    }
  }
  return "";
}

function trySelfHealInstall() {
  const installer = path.join(__dirname, "postinstall.cjs");
  if (!fs.existsSync(installer)) {
    return false;
  }
  const env = { ...process.env, [SELF_HEAL_ENV]: "1" };
  const res = spawnSync(process.execPath, [installer], {
    stdio: "inherit",
    shell: false,
    env
  });
  return typeof res.status === "number" && res.status === 0;
}

let exe = firstExisting(candidatePaths());
if (!exe && process.env[SELF_HEAL_ENV] !== "1") {
  trySelfHealInstall();
  exe = firstExisting(candidatePaths());
}

if (!exe) {
  const lines = [
    "dawnfetch binary was not found in the expected install location.",
    "",
    "This usually means postinstall was blocked (common with bun) or failed.",
    "",
    "For bun, trust package scripts and reinstall:",
    "  bun pm -g untrusted",
    "  bun remove -g dawnfetch && bun add -g dawnfetch",
    "",
    "Try reinstalling with:",
    process.platform === "win32"
      ? '  powershell -c "irm https://raw.githubusercontent.com/almightynan/dawnfetch/main/cli/install.ps1 | iex"'
      : "  curl -fsSL https://raw.githubusercontent.com/almightynan/dawnfetch/main/cli/install.sh | bash"
  ];
  console.error(lines.join("\n"));
  process.exit(1);
}

const res = spawnSync(exe, process.argv.slice(2), { stdio: "inherit" });
if (typeof res.status === "number") {
  process.exit(res.status);
}
process.exit(1);

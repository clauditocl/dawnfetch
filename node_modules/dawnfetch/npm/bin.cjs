#!/usr/bin/env node
"use strict";

const { spawnSync } = require("node:child_process");
const fs = require("node:fs");
const os = require("node:os");
const path = require("node:path");

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

const exe = firstExisting(candidatePaths());
if (!exe) {
  const lines = [
    "dawnfetch binary was not found in the expected install location.",
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


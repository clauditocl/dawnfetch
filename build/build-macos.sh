#!/usr/bin/env bash
# builds macos executables into ./dist for arm64 and amd64.

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DIST_DIR="${ROOT_DIR}/dist"
GO_BIN="${GO_BIN:-go}"
PKG="${PKG:-.}"
LDFLAGS="${LDFLAGS:-}"

mkdir -p "${DIST_DIR}"

build_target() {
  local arch="$1"
  local out="${DIST_DIR}/dawnfetch-macos-${arch}"
  local args=("build" "-trimpath")
  if [[ -n "${LDFLAGS}" ]]; then
    args+=("-ldflags" "${LDFLAGS}")
  fi
  args+=("-o" "${out}" "${PKG}")

  echo "building ${out}"
  CGO_ENABLED=0 GOOS=darwin GOARCH="${arch}" "${GO_BIN}" "${args[@]}"
}

build_target "arm64"
build_target "amd64"

echo "done. macos binaries are in ${DIST_DIR}"

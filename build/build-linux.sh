#!/usr/bin/env bash
# builds linux executables into ./dist for amd64, arm64, and 386.

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DIST_DIR="${ROOT_DIR}/dist"
GO_BIN="${GO_BIN:-go}"
PKG="${PKG:-.}"
LDFLAGS="${LDFLAGS:-}"

mkdir -p "${DIST_DIR}"

build_target() {
  local arch="$1"
  local out="${DIST_DIR}/dawnfetch-linux-${arch}"
  local args=("build" "-trimpath")
  if [[ -n "${LDFLAGS}" ]]; then
    args+=("-ldflags" "${LDFLAGS}")
  fi
  args+=("-o" "${out}" "${PKG}")

  echo "building ${out}"
  CGO_ENABLED=0 GOOS=linux GOARCH="${arch}" "${GO_BIN}" "${args[@]}"
}

build_target "amd64"
build_target "arm64"
build_target "386"

echo "done. linux binaries are in ${DIST_DIR}"

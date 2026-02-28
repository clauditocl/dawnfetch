#!/usr/bin/env bash
# dawnfetch linux/macos installer
# usage:
#   curl -fsSL https://raw.githubusercontent.com/almightynan/dawnfetch/main/cli/install.sh | bash
# optional env vars:
#   DAWNFETCH_VERSION=1.2.3
#   DAWNFETCH_BIN_DIR="$HOME/.local/bin"
#   DAWNFETCH_SHARE_DIR="$HOME/.local/share/dawnfetch"

set -euo pipefail

REPO="almightynan/dawnfetch"
VERSION="${DAWNFETCH_VERSION:-}"
BIN_DIR="${DAWNFETCH_BIN_DIR:-$HOME/.local/bin}"
SHARE_DIR="${DAWNFETCH_SHARE_DIR:-$HOME/.local/share/dawnfetch}"
RELEASE_JSON=""
ASSET_NAME=""
ASSET_URL=""
TMP_DIR=""

info() { printf '[dawnfetch] %s\n' "$*"; }
warn() { printf '[dawnfetch] %s\n' "$*" >&2; }
die() { warn "$*"; exit 1; }

have_cmd() { command -v "$1" >/dev/null 2>&1; }

cleanup() {
  if [[ -n "${TMP_DIR:-}" ]] && [[ -d "${TMP_DIR}" ]]; then
    rm -rf "${TMP_DIR}"
  fi
}

trap cleanup EXIT

fetch_text() {
  local url="$1"
  if have_cmd curl; then
    curl -fsSL "$url"
    return
  fi
  if have_cmd wget; then
    wget -qO- "$url"
    return
  fi
  die "curl or wget is required"
}

try_fetch_text() {
  local url="$1"
  if have_cmd curl; then
    curl -fsSL "$url"
    return
  fi
  if have_cmd wget; then
    wget -qO- "$url"
    return
  fi
  return 1
}

download_file() {
  local url="$1" out="$2"
  if have_cmd curl; then
    curl -fL "$url" -o "$out"
    return
  fi
  if have_cmd wget; then
    wget -O "$out" "$url"
    return
  fi
  die "curl or wget is required"
}

resolve_release() {
  local json=""
  if [[ -n "${VERSION}" ]]; then
    VERSION="${VERSION#v}"
    json="$(try_fetch_text "https://api.github.com/repos/${REPO}/releases/tags/v${VERSION}" || true)"
    if [[ -z "${json}" ]]; then
      json="$(try_fetch_text "https://api.github.com/repos/${REPO}/releases/tags/${VERSION}" || true)"
    fi
    [[ -n "${json}" ]] || die "release v${VERSION} was not found on github"
    RELEASE_JSON="${json}"
    return
  fi

  info "resolving latest release version..."
  json="$(fetch_text "https://api.github.com/repos/${REPO}/releases/latest")"
  RELEASE_JSON="${json}"
  VERSION="$(printf '%s\n' "$RELEASE_JSON" | sed -n 's/.*"tag_name":[[:space:]]*"v\{0,1\}\([^"]*\)".*/\1/p' | head -n1)"
  [[ -n "${VERSION}" ]] || die "failed to resolve latest version"
}

resolve_platform() {
  local uos uarch
  uos="$(uname -s | tr '[:upper:]' '[:lower:]')"
  uarch="$(uname -m | tr '[:upper:]' '[:lower:]')"

  case "$uos" in
    linux) OS="linux" ;;
    darwin) OS="darwin" ;;
    *) die "unsupported os: $uos (linux/macos only)" ;;
  esac

  case "$uarch" in
    x86_64|amd64) ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    i386|i686) ARCH="386" ;;
    armv7l|armv7) ARCH="armv7" ;;
    ppc64le) ARCH="ppc64le" ;;
    s390x) ARCH="s390x" ;;
    *) die "unsupported arch: $uarch" ;;
  esac
}

asset_name_globs() {
  case "${ARCH}" in
    amd64)
      printf '%s\n' "*_${OS}_amd64.tar.gz" "*_${OS}_amd64_v*.tar.gz"
      ;;
    arm64)
      printf '%s\n' "*_${OS}_arm64.tar.gz" "*_${OS}_arm64_v*.tar.gz"
      ;;
    386)
      printf '%s\n' "*_${OS}_386.tar.gz" "*_${OS}_386_*.tar.gz"
      ;;
    armv7)
      printf '%s\n' "*_${OS}_armv7.tar.gz" "*_${OS}_arm_7.tar.gz"
      ;;
    ppc64le)
      printf '%s\n' "*_${OS}_ppc64le.tar.gz" "*_${OS}_ppc64le_*.tar.gz"
      ;;
    s390x)
      printf '%s\n' "*_${OS}_s390x.tar.gz"
      ;;
    *)
      return 1
      ;;
  esac
}

extract_asset_urls() {
  printf '%s' "${RELEASE_JSON}" | tr '\n' ' ' | grep -oE '"browser_download_url"[[:space:]]*:[[:space:]]*"[^"]+"' | sed -E 's/^.*"([^"]+)"$/\1/'
}

resolve_asset_url() {
  local -a urls=()
  local -a globs=()
  local url name glob

  while IFS= read -r url; do
    [[ -n "${url}" ]] || continue
    urls+=("${url}")
  done < <(extract_asset_urls || true)

  while IFS= read -r glob; do
    [[ -n "${glob}" ]] || continue
    globs+=("${glob}")
  done < <(asset_name_globs || true)

  if [[ "${#urls[@]}" -eq 0 ]]; then
    warn "release has no downloadable assets"
    return 1
  fi

  for glob in "${globs[@]}"; do
    for url in "${urls[@]}"; do
      name="${url##*/}"
      if [[ "${name}" == ${glob} ]]; then
        ASSET_NAME="${name}"
        ASSET_URL="${url}"
        return 0
      fi
    done
  done

  warn "no matching archive for ${OS}/${ARCH} in release v${VERSION}"
  warn "available assets:"
  for url in "${urls[@]}"; do
    warn "  - ${url##*/}"
  done
  return 1
}

extract_archive() {
  local archive="$1" dest="$2"
  mkdir -p "$dest"
  if have_cmd tar; then
    tar -xzf "$archive" -C "$dest"
    return
  fi
  die "tar is required"
}

install_files() {
  local root="$1"
  local bin_src themes_src ascii_src
  bin_src="$(find "$root" -type f -name dawnfetch | head -n1 || true)"
  themes_src="$(find "$root" -type f -name themes.json | head -n1 || true)"
  ascii_src="$(find "$root" -type d -name ascii | head -n1 || true)"

  [[ -n "$bin_src" ]] || die "dawnfetch binary not found in archive"
  [[ -n "$themes_src" ]] || die "themes.json not found in archive"
  [[ -n "$ascii_src" ]] || die "ascii directory not found in archive"

  mkdir -p "$BIN_DIR" "$SHARE_DIR"
  install -m 0755 "$bin_src" "$BIN_DIR/dawnfetch"
  install -m 0644 "$themes_src" "$SHARE_DIR/themes.json"
  rm -rf "$SHARE_DIR/ascii"
  cp -R "$ascii_src" "$SHARE_DIR/ascii"
}

print_path_hint() {
  if [[ ":$PATH:" != *":$BIN_DIR:"* ]]; then
    warn "add dawnfetch to PATH:"
    warn "  export PATH=\"$BIN_DIR:\$PATH\""
    warn "if dawnfetch is not found in this shell, reload your shell config:"
    if [[ -n "${ZSH_VERSION:-}" ]]; then
      warn "  source ~/.zshrc"
    elif [[ -n "${BASH_VERSION:-}" ]]; then
      warn "  source ~/.bashrc"
    else
      warn "  source ~/.profile"
    fi
    warn "or restart your terminal."
  fi
}

main() {
  resolve_release
  resolve_platform
  resolve_asset_url || die "failed to resolve release archive for ${OS}/${ARCH}"

  local archive unpack

  info "install version: ${VERSION}"
  info "detected target: ${OS}/${ARCH}"
  info "download: ${ASSET_NAME}"

  TMP_DIR="$(mktemp -d)"
  archive="${TMP_DIR}/${ASSET_NAME}"
  unpack="${TMP_DIR}/unpack"

  download_file "${ASSET_URL}" "$archive"
  extract_archive "$archive" "$unpack"
  install_files "$unpack"

  info "dawnfetch installed successfully."
  info "run now:"
  info "  dawnfetch"
  print_path_hint
}

main "$@"

#!/usr/bin/env bash
# Install rl-stats-daemon.
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/provant-gg/daemon/main/install.sh | bash
#
# Environment overrides:
#   VERSION   Specific tag to install (default: latest)
#   BIN_DIR   Install location (default: /usr/local/bin if writable, else ~/.local/bin)

set -euo pipefail

REPO="provant-gg/daemon"
PROJECT="rl-stats-daemon"
VERSION="${VERSION:-latest}"
BIN_DIR="${BIN_DIR:-}"

c_red=$'\033[31m'; c_green=$'\033[32m'; c_yellow=$'\033[33m'; c_reset=$'\033[0m'
err()  { printf '%serror:%s %s\n' "$c_red"    "$c_reset" "$*" >&2; exit 1; }
info() { printf '%s==>%s %s\n'    "$c_green"  "$c_reset" "$*"; }
warn() { printf '%swarn:%s %s\n'  "$c_yellow" "$c_reset" "$*"; }

detect_os() {
  case "$(uname -s)" in
    Linux)                 echo "Linux"   ;;
    Darwin)                echo "Darwin"  ;;
    FreeBSD)               echo "Freebsd" ;;
    MINGW*|MSYS*|CYGWIN*)  echo "Windows" ;;
    *) err "unsupported OS: $(uname -s)" ;;
  esac
}

detect_arch() {
  case "$(uname -m)" in
    x86_64|amd64)  echo "x86_64" ;;
    aarch64|arm64) echo "arm64"  ;;
    i386|i686)     echo "i386"   ;;
    armv7l)        echo "armv7"  ;;
    armv6l)        echo "armv6"  ;;
    *) err "unsupported architecture: $(uname -m)" ;;
  esac
}

download() {
  local url="$1" out="$2"
  if command -v curl >/dev/null 2>&1; then
    curl -fsSL "$url" -o "$out"
  elif command -v wget >/dev/null 2>&1; then
    wget -q "$url" -O "$out"
  else
    err "need curl or wget"
  fi
}

verify_checksum() {
  local archive="$1" sums="$2" name expected actual
  name=$(basename "$archive")
  expected=$(awk -v f="$name" '$2==f {print $1}' "$sums")
  [ -n "$expected" ] || { warn "no checksum entry for $name, skipping verification"; return 0; }

  if command -v sha256sum >/dev/null 2>&1; then
    actual=$(sha256sum "$archive" | awk '{print $1}')
  elif command -v shasum >/dev/null 2>&1; then
    actual=$(shasum -a 256 "$archive" | awk '{print $1}')
  else
    warn "no sha256sum or shasum tool, skipping verification"
    return 0
  fi

  [ "$expected" = "$actual" ] || err "checksum mismatch: expected $expected, got $actual"
}

tmp=""
cleanup() { [ -n "${tmp:-}" ] && [ -d "${tmp:-}" ] && rm -rf "$tmp"; return 0; }
trap cleanup EXIT

main() {
  command -v tar >/dev/null 2>&1 || err "tar is required"

  local os arch ext archive base_url archive_url sums_url bin
  os=$(detect_os)
  arch=$(detect_arch)
  ext="tar.gz"
  if [ "$os" = "Windows" ]; then
    ext="zip"
    command -v unzip >/dev/null 2>&1 || err "unzip is required on Windows"
  fi
  archive="${PROJECT}_${os}_${arch}.${ext}"

  if [ "$VERSION" = "latest" ]; then
    base_url="https://github.com/${REPO}/releases/latest/download"
  else
    base_url="https://github.com/${REPO}/releases/download/${VERSION}"
  fi
  archive_url="${base_url}/${archive}"
  sums_url="${base_url}/checksums.txt"

  if [ -z "$BIN_DIR" ]; then
    if [ -w "/usr/local/bin" ]; then
      BIN_DIR="/usr/local/bin"
    else
      BIN_DIR="$HOME/.local/bin"
    fi
  fi
  mkdir -p "$BIN_DIR"

  tmp=$(mktemp -d)

  info "detected ${os}/${arch}, version ${VERSION}"
  info "downloading ${archive_url}"
  download "$archive_url" "$tmp/$archive"

  if download "$sums_url" "$tmp/checksums.txt" 2>/dev/null; then
    info "verifying checksum"
    verify_checksum "$tmp/$archive" "$tmp/checksums.txt"
  else
    warn "could not fetch checksums.txt, skipping verification"
  fi

  info "extracting"
  if [ "$ext" = "zip" ]; then
    unzip -q "$tmp/$archive" -d "$tmp"
  else
    tar -xzf "$tmp/$archive" -C "$tmp"
  fi

  bin="$PROJECT"
  [ "$os" = "Windows" ] && bin="${PROJECT}.exe"
  [ -f "$tmp/$bin" ] || err "binary $bin not found in archive"

  install -m 0755 "$tmp/$bin" "$BIN_DIR/$bin"
  info "installed $BIN_DIR/$bin"

  case ":$PATH:" in
    *":$BIN_DIR:"*) ;;
    *) warn "$BIN_DIR is not in PATH; add it to your shell rc to use '$PROJECT' directly" ;;
  esac
}

main "$@"

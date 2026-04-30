#!/usr/bin/env bash
# Called by goreleaser builds[].hooks.post for every binary it produces.
# Only runs makensis for windows/amd64; all other targets are no-op.
#
# Args: GOOS GOARCH BINARY_PATH VERSION PROJECT_NAME

set -euo pipefail

goos="${1:?missing goos}"
goarch="${2:?missing goarch}"
binary="${3:?missing binary path}"
version="${4:?missing version}"
project="${5:?missing project name}"

if [ "$goos" != "windows" ] || [ "$goarch" != "amd64" ]; then
  exit 0
fi

if ! command -v makensis >/dev/null 2>&1; then
  echo "makensis not found in PATH; skipping Windows installer build" >&2
  exit 0
fi

mkdir -p dist
out="dist/${project}_Windows_x86_64_setup.exe"

makensis \
  -V2 \
  -DVERSION="${version}" \
  -DBINARY="$(realpath "$binary")" \
  -DOUTFILE="$(realpath -m "$out")" \
  windows/installer.nsi

echo "built $out"

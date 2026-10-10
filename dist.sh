#!/usr/bin/env bash
# Cross-compile prebuilt binaries used by flake.nix: dist/qwer-<nix system>.
set -euo pipefail
cd "$(dirname "$0")"
declare -A targets=(
  [x86_64-linux]=x86_64-linux-musl
  [aarch64-darwin]=aarch64-macos
)
mkdir -p dist
for system in "${!targets[@]}"; do
  out=$(mktemp -d)
  zig build -Doptimize=ReleaseSmall -Dcpu=baseline -Dtarget="${targets[$system]}" --prefix "$out"
  install -m755 "$out/bin/qwer" "dist/qwer-$system"
  rm -rf "$out"
done
ls -l dist

#!/usr/bin/env bash
# Builds the golangci-lint Go plugin (.so) into a stable, non-tmp location.
set -euo pipefail

dest_dir="${WHENTHEN_PLUGIN_DIR:-$HOME/.local/lib/golangci}"
mkdir -p "$dest_dir"
cd "$(dirname "$0")/.."

go build -buildmode=plugin -o "$dest_dir/whenthen.so" ./pluginmain
echo "$dest_dir/whenthen.so"

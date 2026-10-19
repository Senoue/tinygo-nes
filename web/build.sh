#!/bin/sh
set -eu
cd "$(dirname "$0")/.."
GOOS=js GOARCH=wasm go build -o web/main.wasm .
root=$(go env GOROOT)
runtime="$root/lib/wasm/wasm_exec.js"
if [ ! -f "$runtime" ]; then runtime="$root/misc/wasm/wasm_exec.js"; fi
cp "$runtime" web/wasm_exec.js
printf 'Built web/main.wasm. Serve with: python3 -m http.server 8080 --directory web --bind 127.0.0.1\n'

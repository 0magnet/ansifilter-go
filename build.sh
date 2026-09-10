#!/bin/sh
# Build the demo into docs/, which is what GitHub Pages serves.
#
# The standard Go toolchain, not TinyGo: docs/wasm_exec.js here is Go's, and
# the port leans on reflection and text/template in ways TinyGo does not carry.
set -eu

cd "$(dirname "$0")"
mkdir -p docs
# The demo is its own module (it pulls in the desk, websh and netscrape, which
# the library itself must not depend on), so it is built from inside it.
(cd demo && GOOS=js GOARCH=wasm go build -o ../docs/ansifilter.wasm .)
cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" docs/wasm_exec.js
ls -lh docs/ansifilter.wasm

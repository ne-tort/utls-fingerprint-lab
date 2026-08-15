#!/usr/bin/env bash
# Entry point for the QUIC Initial fingerprint lab.
set -euo pipefail
ROOT="$(cd "$(dirname "$0")" && pwd)"
cd "$ROOT"
CMD="${1:-help}"
shift || true

build() {
  mkdir -p capture/bin
  (cd capture && go build -trimpath -ldflags='-s -w' -o bin/quic-capture .)
  (cd capture && GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o bin/quic-capture-linux .)
  echo "ok → capture/bin/"
}

ensure() {
  [[ -x capture/bin/quic-capture || -f capture/bin/quic-capture ]] || build
}

case "$CMD" in
  help)
    cat <<EOF
quic lab commands:
  build
  capture-listen [listen=:4433] [target=unknown]
  parse <initial.bin>
  list
EOF
    ;;
  build) build ;;
  capture-listen)
    ensure
    LISTEN="${1:-:4433}"
    TARGET="${2:-unknown}"
    mkdir -p captures profiles
    exec ./capture/bin/quic-capture -listen "$LISTEN" -out captures -profiles profiles -default-target "$TARGET" -promote
    ;;
  parse)
    ensure
    [[ $# -ge 1 ]] || { echo "parse requires path"; exit 1; }
    exec ./capture/bin/quic-capture -parse "$1"
    ;;
  list)
    grep -E '^\s+- id:' targets.yaml | sed 's/.*id:[[:space:]]*//'
    ;;
  *) echo "unknown: $CMD"; exit 1 ;;
esac

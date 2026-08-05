#!/usr/bin/env bash
# Unix entry point — mirrors lab.ps1
set -euo pipefail
ROOT="$(cd "$(dirname "$0")" && pwd)"
cd "$ROOT"
export DOCKER_BUILDKIT=1 COMPOSE_DOCKER_CLI_BUILD=1
export COMPOSE_PROFILES=capture,verify,tools
CMD="${1:-help}"
shift || true

compose() { docker compose -f compose.yaml --project-name utls-lab "$@"; }

ensure_labctl() {
  mkdir -p "$ROOT/bin"
  if [[ ! -x "$ROOT/bin/labctl" ]]; then
    (cd "$ROOT/tools" && go get gopkg.in/yaml.v3@v3.0.1 >/dev/null && go build -o "$ROOT/bin/labctl" ./cmd/labctl)
  fi
}

case "$CMD" in
  help)
    cat <<EOF
utls-fingerprint-lab
  ./lab.sh build|list|capture|verify|test|catalog|clean
  See ./lab.ps1 help for flags (-Id / -Group via env ID= GROUP=).
EOF
    ;;
  build)
    python3 scripts/gen-compose.py
    compose build capture tools
    ensure_labctl
    ;;
  list)
    ensure_labctl
    "$ROOT/bin/labctl" -root "$ROOT" list ${STATUS:+-status "$STATUS"} ${GROUP:+-group "$GROUP"}
    ;;
  capture)
    python3 scripts/gen-compose.py
    ensure_labctl
    args=(-root "$ROOT" capture)
    [[ -n "${ID:-}" ]] && args+=(-id "$ID")
    [[ -n "${GROUP:-}" ]] && args+=(-group "$GROUP")
    "$ROOT/bin/labctl" "${args[@]}"
    "$ROOT/bin/labctl" -root "$ROOT" catalog
    ;;
  verify)
    ensure_labctl
    args=(-root "$ROOT" verify)
    [[ -n "${ID:-}" ]] && args+=(-id "$ID")
    "$ROOT/bin/labctl" "${args[@]}"
    ;;
  catalog) ensure_labctl; "$ROOT/bin/labctl" -root "$ROOT" catalog ;;
  test)
    python3 scripts/gen-compose.py
    compose build capture tools
    ensure_labctl
    for tid in openssl3 curl-imp-chrome146 builtin-chrome go-nethttp; do
      "$ROOT/bin/labctl" -root "$ROOT" capture -id "$tid"
      "$ROOT/bin/labctl" -root "$ROOT" verify -id "$tid"
    done
    "$ROOT/bin/labctl" -root "$ROOT" catalog
    echo SMOKE OK
    ;;
  clean)
    rm -rf captures/* || true
    rm -rf profiles/verify-* profiles/*/verify-last.json || true
    echo cleaned
    ;;
  *) echo "unknown: $CMD"; exit 2 ;;
esac

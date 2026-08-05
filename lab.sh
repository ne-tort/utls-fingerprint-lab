#!/usr/bin/env bash
# Unix entry point — mirrors lab.ps1
set -euo pipefail
ROOT="$(cd "$(dirname "$0")" && pwd)"
cd "$ROOT"
export DOCKER_BUILDKIT=1 COMPOSE_DOCKER_CLI_BUILD=1
export COMPOSE_PROFILES=capture,verify,tools
export GOOS="${GOOS:-}" GOARCH="${GOARCH:-}"
CMD="${1:-help}"
shift || true

compose() { docker compose -f compose.yaml --project-name utls-lab "$@"; }

build_host_linux_bins() {
  echo "cross-compiling linux tools (GOCACHE on host)..."
  mkdir -p "$ROOT/bin" "$ROOT/capture/bin" "$ROOT/tools/bin" "$ROOT/clients/go-http/bin"
  export CGO_ENABLED=0 GOOS=linux GOARCH=amd64
  (cd "$ROOT/capture" && go build -trimpath -ldflags='-s -w' -o "$ROOT/capture/bin/capture-linux" .)
  (cd "$ROOT/clients/go-http" && go build -trimpath -ldflags='-s -w' -o "$ROOT/clients/go-http/bin/go-http-linux" .)
  (cd "$ROOT/tools" && go get gopkg.in/yaml.v3@v3.0.1 >/dev/null \
    && go build -trimpath -ldflags='-s -w' -o "$ROOT/tools/bin/verify-linux" ./cmd/verify \
    && go build -trimpath -ldflags='-s -w' -o "$ROOT/tools/bin/emit-builtin-linux" ./cmd/emit-builtin \
    && go build -trimpath -ldflags='-s -w' -o "$ROOT/tools/bin/labctl-linux" ./cmd/labctl \
    && GOOS="$(go env GOHOSTOS)" GOARCH="$(go env GOHOSTARCH)" \
       go build -trimpath -ldflags='-s -w' -o "$ROOT/bin/labctl" ./cmd/labctl)
}

ensure_labctl() {
  if [[ ! -x "$ROOT/bin/labctl" ]]; then
    build_host_linux_bins
  fi
}

case "$CMD" in
  help)
    cat <<EOF
utls-fingerprint-lab
  ./lab.sh build|list|capture|verify|test|catalog|export|clean
  Filters: ID=… GROUP=… STATUS=active
EOF
    ;;
  build)
    python3 scripts/gen-compose.py
    build_host_linux_bins
    compose build capture tools
    echo build ok
    ;;
  list)
    ensure_labctl
    args=(-root "$ROOT" list)
    [[ -n "${STATUS:-}" ]] && args+=(-status "$STATUS")
    [[ -n "${GROUP:-}" ]] && args+=(-group "$GROUP")
    [[ -z "${STATUS:-}" ]] && args+=(-status active)
    "$ROOT/bin/labctl" "${args[@]}"
    ;;
  capture)
    python3 scripts/gen-compose.py
    build_host_linux_bins
    compose build capture tools
    args=(-root "$ROOT" capture)
    [[ -n "${ID:-}" ]] && args+=(-id "$ID")
    [[ -n "${GROUP:-}" ]] && args+=(-group "$GROUP")
    "$ROOT/bin/labctl" "${args[@]}"
    "$ROOT/bin/labctl" -root "$ROOT" catalog
    ;;
  verify)
    build_host_linux_bins
    compose build capture tools
    args=(-root "$ROOT" verify)
    [[ -n "${ID:-}" ]] && args+=(-id "$ID")
    "$ROOT/bin/labctl" "${args[@]}"
    ;;
  catalog) ensure_labctl; "$ROOT/bin/labctl" -root "$ROOT" catalog ;;
  export)
    ensure_labctl
    "$ROOT/bin/labctl" -root "$ROOT" export
    ;;
  test)
    python3 scripts/gen-compose.py
    build_host_linux_bins
    compose build capture tools
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

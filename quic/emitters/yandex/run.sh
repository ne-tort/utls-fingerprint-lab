#!/usr/bin/env bash
# Live Yandex Browser H3 Initial toward capture (dump-only).
set -eu
BIN=$(command -v yandex-browser-stable || command -v yandex-browser || true)
if [ -z "$BIN" ]; then
  echo "yandex-browser binary not found" >&2
  ls -la /usr/bin/yandex* 2>/dev/null || true
  exit 1
fi
# Same force-QUIC pattern as chromium live clients.
timeout 35 "$BIN" --headless=new --no-sandbox --disable-gpu \
  --disable-dev-shm-usage \
  --enable-quic --quic-version=h3 \
  --origin-to-force-quic-on=yandex.fp.lab:4433 \
  --ignore-certificate-errors \
  "https://yandex.fp.lab:4433/" || true

#!/usr/bin/env bash
set -euo pipefail
for i in $(seq 1 30); do
  getent hosts brave.fp.lab.local >/dev/null 2>&1 && break
  sleep 1
done
brave-browser --version || true
brave-browser --headless=new --no-sandbox --disable-gpu --disable-dev-shm-usage \
  --ignore-certificate-errors --allow-insecure-localhost --dump-dom \
  https://brave.fp.lab.local:8443/ || true
echo "brave done"

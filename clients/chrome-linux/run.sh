#!/usr/bin/env bash
set -euo pipefail
for i in $(seq 1 30); do
  getent hosts chrome-linux.fp.lab.local >/dev/null 2>&1 && break
  sleep 1
done
google-chrome --version || true
# Headless fetch; ignore cert so lab capture CA is accepted. CH is saved even if dump-dom fails.
google-chrome --headless=new --no-sandbox --disable-gpu --disable-dev-shm-usage \
  --ignore-certificate-errors --allow-insecure-localhost --dump-dom \
  https://chrome-linux.fp.lab.local:8443/ || true
echo "chrome-linux done"

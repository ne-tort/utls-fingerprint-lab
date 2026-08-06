#!/usr/bin/env bash
set -euo pipefail
export DISPLAY=:99
# xvfb for Electron (no Ozone/headless-new in Electron 33 the same way as Chrome).
Xvfb :99 -screen 0 1280x720x24 -ac +extension GLX +render -noreset >/tmp/xvfb.log 2>&1 &
XVFB_PID=$!
trap 'kill $XVFB_PID 2>/dev/null || true' EXIT
sleep 1
for i in $(seq 1 30); do
  getent hosts electron-sample.fp.lab.local >/dev/null 2>&1 && break
  sleep 1
done
cd /app
./node_modules/.bin/electron --no-sandbox --disable-gpu main.js || true
echo "electron-sample done"

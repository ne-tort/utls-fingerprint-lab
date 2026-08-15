#!/usr/bin/env bash
# Live Firefox H3 Initial toward capture (dump-only server).
set -eu
PROF="${HOME}/ffprof-$$"
mkdir -p "$PROF"
cat >"$PROF/user.js" <<'EOF'
user_pref("network.http.http3.enabled", true);
user_pref("network.http.http3.alt-svc-mapping-for-testing", "firefox.fp.lab;h3=\":4433\"");
user_pref("network.dns.httpssvc.http3_fast_fallback_timeout", 0);
user_pref("network.http.http3.backup_timer_delay", 10000);
user_pref("security.enterprise_roots.enabled", true);
user_pref("network.stricttransportsecurity.preloadlist", false);
EOF
FF=$(command -v firefox || command -v firefox-esr || true)
if [ -z "$FF" ]; then
  echo "firefox binary not found" >&2
  exit 1
fi
timeout 40 "$FF" --headless --profile "$PROF" \
  "https://firefox.fp.lab:4433/" || true

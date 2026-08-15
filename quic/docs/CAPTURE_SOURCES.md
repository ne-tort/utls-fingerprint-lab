# Capture sources — who to fingerprint

Ground truth: UDP Initial → Go `quic-capture` → `profiles/<id>/`.

## Already captured (active)

| Source | How | Persona |
|--------|-----|---------|
| sagernet ChromeParrot | `lab.ps1 matrix` | chromeparrot ≡ hy2 default |
| quic-go plain | matrix | quicgo (no `0x20`) |
| quic-go + datagrams | `lab.ps1 matrix-ext` | quicgodg ≡ hy2plain |
| uquic Chrome 115/146 | matrix | library mimics (no `0x11`) |
| uquic Firefox 116 / A/B/C | matrix + matrix-ext | firefox reference |
| aioquic | live | Python H3 |
| curl+quiche | live | Cloudflare quiche |
| Chromium zenika | live | older chrome-like |
| Chromium chromedp | live `chromiumfresh` | fresher chrome ≡ parrot |
| Yandex Browser | live `yandex` | Chromium fork; `0x11` without `0x3128` |
| Windows Chrome / Edge / Yandex | `lab.ps1 host-browsers` | host install; Chrome≡Edge≡parrot |
| Firefox headless | live | vs uquicff |
| hy2 / tuic outbound | `lab.ps1 hy2` / `tuic` | product stacks |

## Wishlist / next

| Source | Why | Feasibility |
|--------|-----|-------------|
| Microsoft Edge H3 | Chromium fork; expect ≈ chrome unless TP differs | Playwright/Edge image; low priority if ≡ chromiumfresh |
| Safari / Network.framework | macOS-only QUIC | Host capture outside Docker |
| ngtcp2 / nghttp3 client | Interop stack | `ngtcp2/ngtcp2-interop` image experiment |
| Cloudflare quiche rust examples | vs curl-quiche | Build from `_refs` or quiche image |
| mvfst / lsquic / msquic | Diverse stacks | Heavy; optional matrix rows |
| wget2 HTTP/3 | Another libcurl-adjacent | Check distro packages |
| Mobile Chrome/Firefox | Real-device freshness | Manual host capture → same promote |
| quic.tlsfingerprint.io samples | Observatory compare | Diff only; not emit |

## Product stacks (sing-box-lx)

| Outbound | Parrot on | Parrot off |
|----------|-----------|------------|
| hy2 | ≡ chromeparrot | ≡ quicgodg (`0x20`) — [F](../../../../SPECS/TASKS/090-QUIC_FINGERPRINT_PROFILES/modules/F-HY2PLAIN_DATAGRAM_0x20.md) |
| tuic | expect ≡ hy2parrot | expect ≡ hy2plain |
| VLESS+quic / shadowquic | jls parrot variants | separate later |

## Do not confuse

- TCP uTLS / FEATURE 018 / curl-impersonate → **not** QUIC Initial sources.
- Raw `initials/*.bin` → observation only; emit = recipe ([REPLAY_AND_EMIT.md](REPLAY_AND_EMIT.md)).

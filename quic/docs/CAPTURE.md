# Capture plan

## Phase 0 — scaffold (this commit)

- UDP listen + `clienthellod` parse → `captures/` + promote `profiles/`
- Offline `parse` of saved `.bin`
- `targets.yaml` registry (many `wishlist` until compose wired)

## Phase 1 — emitters we control

| Target | How |
|--------|-----|
| `hy2-parrot` | Host/CI runs sing-box-lx hy2 outbound → lab UDP (parrot on) |
| `hy2-plain` | same, `disable_chrome_parrot: true` |
| `tuic-parrot` | optional |

Goal: **fresh ChromeParrot vs live Chrome** diff report.

## Phase 2 — real browsers H3

Docker compose (or host):

- Chromium / Chrome → `https://<id>.fp.lab.local:4433` (H3 only)
- Firefox → same
- Optional: Edge/Brave for alias proof

Capture listens UDP/4433; may need HTTP/3-capable answer or “blackhole accept”
that only needs Initial (clienthellod does not require completing handshake for
fingerprint IDs — confirm per version; if client retries, still OK).

Alternative: point clients at **quic.tlsfingerprint.io** and save reflected
JSON — good for comparison, not for private lab profiles.

## Phase 3 — libraries

| Target | Tooling |
|--------|---------|
| `curl-quiche` | curl+quiche image → H3 |
| `aioquic` | python client |
| `uquic-firefox-116` | reference emit container (uquic), not product |

## Multi-datagram

Chrome with ML-KEM often sends **two** Initials. Capture must:

1. Demux by peer addr
2. `GatherClientInitials` / equivalent until CH complete
3. Store all `initials/NNN.bin` in order

## Diff: parrot freshness

```text
profile chrome-latest  vs  hy2-parrot
  → CH extension/curve/ALPS diff
  → TP id-set / window values diff
  → CID / pad diff
```

Actionable bumps go to `patches/quic-go` ChromeParrot (product), not into TCP uTLS.

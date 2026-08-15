# Ecosystem design (lab-owned)

## Layers

```text
┌─ Capture (clienthellod) ──────────────────────────────┐
│  UDP Initial gather → quic-raw-initial-v1 profiles    │
├─ Emitters (Docker matrix) ────────────────────────────┤
│  chromeparrot = sagernet ChromeParrot ≈ hy2 default   │
│  quicgo-plain = parrot off                            │
│  uquic presets = reference mimics (not product dep)   │
│  (next) live browsers H3 / curl-quiche / aioquic      │
├─ Compare / catalog ───────────────────────────────────┤
│  TP id-set + clienthellod HexID + dg count (+ JA4q)   │
├─ Product (sing-box later) ────────────────────────────┤
│  quic.fingerprint shorts → emit / demux match         │
│  NOT tls.utls                                         │
└───────────────────────────────────────────────────────┘
```

## Source of truth

| Artifact | Truth |
|----------|--------|
| Live capture profile | lab `profiles/` |
| hy2 wire shape today | `emit-chromeparrot` matrix row |
| Firefox shape to port | uquic `QUICFirefox_116*` + live FF later |
| Public observatory | quic.tlsfingerprint.io (compare only) |

## Refs (cloned under `_refs/`)

- `clienthellod` — parser (watch: zero deadline = expired)
- `uquic` — presets including **Chrome_146** (ML-KEM, 2 datagrams)
- `ja4` — JA4/JA4Q docs for later `expected.ja4`

## Next lab steps

1. Commit representative `profiles/*` snapshots for CI compare fixtures.
2. Chromium/Firefox H3 docker clients → live freshness vs chromeparrot.
3. Optional hy2 outbound container (mount prebuilt sing-box) for end-to-end
   parity check vs chromeparrot probe.
4. JA4q computation in capture promote.
5. Export shorts + (later) `lxquicfp` sync — still no sing-box code in this phase.

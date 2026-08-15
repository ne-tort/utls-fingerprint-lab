# Ecosystem design (lab-owned)

## Layers

```text
┌─ Capture (clienthellod) ──────────────────────────────┐
│  UDP Initial gather → quic-raw-initial-v1 profiles    │
├─ Emitters (Docker matrix) ────────────────────────────┤
│  chromeparrot = sagernet ChromeParrot ≈ hy2 default   │
│  quicgo-plain = parrot off                            │
│  uquic presets = reference mimics (not product dep)   │
│  live: aioquic / curl-quiche / chromium (match_only)  │
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
- `aioquic` / `ja4plus` — Python client persona + optional JA4 sidecar
  (see [PYTHON_VS_GO_CAPTURE.md](PYTHON_VS_GO_CAPTURE.md), `compose.live-clients.yaml`)

## Next lab steps

**Done (matrix):** capture promote fills `clienthello.bin`, `tp.json`, `tp_id_set`,
SNI→id; Docker emitters chromeparrot/quicgo/uquic; fixtures + RESULTS_MATRIX;
clienthellod deadline fix.

1. Optional **hy2 outbound** parity service (thin sing-box image + mixed inbound —
   pattern from demux-quic-auth; dump only, no `quic_auth`).
2. ~~Live Chromium~~ — done; ≠ parrot (see RESULTS). Firefox H3 still wishlist.
3. ~~JA4 `q…` fill~~ — `./lab.ps1 ja4` (`compose.ja4.yaml` + ja4plus); store only
   the string in `expected.ja4` (not JA4+ in product).
4. Export shorts + (later) `lxquicfp` sync — still no sing-box code in this phase.
5. Write raw `transport_parameters.bin` if clienthellod exposes extension payload.

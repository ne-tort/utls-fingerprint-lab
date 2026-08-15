# Apply — from profile to wire / config

## Product config (target UX)

Do **not** reuse `tls.utls.fingerprint` for QUIC ([027](../../../../SPECS/TASKS/027-UTLS_OVER_QUIC/SPEC.md),
[090 D](../../../../SPECS/TASKS/090-QUIC_FINGERPRINT_PROFILES/modules/D-UTLS_CONFIG_TRAP.md)).

```json
"quic": { "fingerprint": "chrome" }
```

or protocol sugar (hy2):

```json
"fingerprint": "firefox"
```

replacing boolean `disable_chrome_parrot` over time (`chrome` = parrot on,
`quic-go` = off).

Optional alias name in docs: **`utls_quic`** / package `lxquicfp` — means
“QUIC persona catalog like uTLS shorts”, **not** metacubex uTLS-over-QUIC.

## Runtime mapping (sing-box-lx)

| Short | Emit today | Emit future |
|-------|------------|-------------|
| `chrome` / `chrome-N` | `ChromeParrot=true` | bump from lab |
| `quic-go` | parrot false | same |
| `firefox` / `firefox-N` | — | new path (uquic-inspired + capture) |
| `quiche` / `aioquic` | — | match-only first |

`quic_auth` remains **chrome GREASE-only** until a firefox auth channel exists.

## Sync (future, mirror 018)

```text
quic/dist/export/catalog.json + profiles/<short>/
    → make -f Makefile.lx lx-quic-fp-sync
    → common/quic/lxquicfp/ embed
```

Build-tag TBD (`with_lx_quic_fp` or fold into existing quic tags).

## Verify loop (lab)

1. Capture live → profile
2. Emit candidate (parrot / future spec) → capture again
3. Compare `expected.ja4` + `tp_id_set` + CID lens (tolerance on GREASE)

Blunt “replay initials/*.bin as UDP” is **not** a valid QUIC connection (crypto
needs live TP tied to SCID). Verify = re-emit from structured spec, not raw
datagram replay (unlike TCP HelloCustom blunt mimicry).

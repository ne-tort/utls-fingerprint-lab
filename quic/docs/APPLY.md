# Apply — from profile to wire / config

## Product config (target UX — not wired yet)

Do **not** reuse `tls.utls.fingerprint` for QUIC ([027](../../../../SPECS/TASKS/027-UTLS_OVER_QUIC/SPEC.md),
[090 D](../../../../SPECS/TASKS/090-QUIC_FINGERPRINT_PROFILES/modules/D-UTLS_CONFIG_TRAP.md)).

```json
"quic": { "fingerprint": "chrome" }
```

Lab catalogs shorts via `catalog/utls/` + `dist/export/`. Product package /
`quic.fingerprint` option — **out of lab scope for now**.

## Emit mapping (lab)

| Short | Emit today | Dial in export |
|-------|------------|----------------|
| `chrome` | `ChromeParrot` + datagrams + SCID0 | yes |
| `quic-go` | plain | yes |
| `quic-go-datagram` | plain + `EnableDatagrams` | yes |
| others | match_only / experimental | no |

Lab emitter: `emitters/fromprofile -profile catalog/utls/chrome.json`.

`quic_auth` channel = GREASE TP value; demux match does **not** need fingerprint
short ([I](../../../../SPECS/TASKS/090-QUIC_FINGERPRINT_PROFILES/modules/I-QUIC_AUTH_CHANNEL.md)).

## Export prep

```text
./lab.ps1 export  →  dist/export/
```

See [IMPORT.md](IMPORT.md) · [STATUS.md](STATUS.md).

## Verify loop (lab)

1. Capture → observation
2. Emit (`fromprofile` / parrot) → capture again
3. `./lab.ps1 unify` + `./lab.ps1 exp-stable`

Blunt replay of `initials/*.bin` is not a valid dial path.

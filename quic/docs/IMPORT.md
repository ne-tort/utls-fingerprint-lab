# Importing QUIC lab profiles (prep for future sing-box sync)

## Format

Product-oriented SoT is **`quic-utls-profile-v1`** (identity + emit + random + auth).

| Artifact | Role |
|----------|------|
| `profile.json` | Unified emit recipe (**not** raw `initials/*.bin`) |
| `short.json` | Catalog metadata (`dial`, `emit_kind`, …) |
| `catalog.json` | shorts, aliases, families (`quic-utls-catalog-v1`) |

**Do not** import into `common/tls/lxutls` (TCP).  
**Do not** blunt-replay QUIC Initial UDP as dial.

## Lab export (now)

```powershell
cd lx-test/utls-fingerprint-docker/quic
./lab.ps1 unify          # optional gate
./lab.ps1 export         # → dist/export/
```

Also regenerates `catalog/emit_templates.json` from curated shorts (no hand edit).

## Future product sync (out of lab scope)

When product work resumes: copy `dist/export/` into an embed package (mirror
FEATURE 018 `lx-utls-sync`). **Not implemented in this phase.**

## Dialable shorts (export)

| short | dial | notes |
|-------|------|-------|
| `chrome` | yes | parrot + datagrams + auth-capable |
| `quic-go` | yes | plain |
| `quic-go-datagram` | yes | plain + `0x20` |
| others | no | match / experimental until structured emit |

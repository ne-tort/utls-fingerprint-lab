# Importing lab profiles into sing-box / metacubex uTLS

## Format

`profile.json` field `format`: **`utls-raw-clienthello-v1`**

| File | Use |
|------|-----|
| `clienthello.bin` | Full TLS record containing ClientHello — **source of truth** |
| `profile.json` | Manifest: id, expected JA4, format |
| `meta.json` / `slot.json` | Lab metadata (optional at runtime) |

## Short names (`./lab.ps1 export`)

Export writes `dist/export/`:

| Artifact | Content |
|----------|---------|
| `catalog.json` | families, short names, aliases, JA4, track |
| `profiles/<short>/` | copies of **unique** lab profile dirs (JA4-deduped) |
| `NAMES.md` | human table short ↔ lab id |

Naming rules (per `family` in `targets.yaml`):

1. Prefer `track: latest`, then non-`emit-builtin` by `version` desc, then `captured_at` desc.
2. Append all `emit-builtin` (stock Hello*_Auto / pins) at the **tail**, also by version desc.
3. **Dedup**: within a family, identical JA4 → one short name; other lab IDs → `aliases`.
4. Assign `family`, `family-1`, … only over unique JA4 — **unsuffixed = newest** (usually the latest track).
5. Brave / Edge / Tor / Yandex are **not** in the `chrome` family.

Legacy aliases in export: `chrome_psk`… → `chrome`; empty fingerprint → `chrome`.

Refresh floating channels: `./lab.ps1 refresh-latest` (pull + archive-on-JA4-change + export `--check-dedup`).

## Go (metacubex/utls) — blunt mimicry

```go
fp := &utls.Fingerprinter{AllowBluntMimicry: true}
spec, err := fp.RawClientHello(bin)
uconn := utls.UClient(rawConn, &utls.Config{ServerName: sni}, utls.HelloCustom)
err = uconn.ApplyPreset(spec)
```

## sing-box wiring (FEATURE 018 / SPEC 064)

1. `make -f Makefile.lx lx-utls-sync` embeds export into `common/tls/lxutls/`.
2. Config: `tls.utls.fingerprint: "chrome"` (short name) under build-tag `with_lx_utls`.
3. Runtime: `Lookup` → `RawClientHello` → `HelloCustom` → `ApplyPreset`.
4. Embed stores one blob per unique short name; catalog aliases resolve duplicates.

## Verification before shipping

```powershell
./lab.ps1 verify -Id chromium-stable
./lab.ps1 export
```

JA4 must match `profile.json` → `expected.ja4`. JA3 may drift (GREASE); marked `ja3_hash_unstable`.

## Notes

- Capture advertises ALPN `http/1.1` and `h2` so h2-only ClientHellos (gRPC) complete; verify still speaks HTTP/1.1 for JA4 headers.
- GREASE / extension shuffle: expect JA3 variance; pin on JA4 + structural verify.

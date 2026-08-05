# Importing lab profiles into sing-box / metacubex uTLS

## Format

`profile.json` field `format`: **`utls-raw-clienthello-v1`**

| File | Use |
|------|-----|
| `clienthello.bin` | Full TLS record containing ClientHello — **source of truth** |
| `profile.json` | Manifest: id, expected JA4, format |
| `meta.json` | Lab metadata (optional at runtime) |

## Short names (`./lab.ps1 export`)

Export writes `dist/export/`:

| Artifact | Content |
|----------|---------|
| `catalog.json` | families, short names, aliases, JA4 |
| `profiles/<short>/` | copies of lab profile dirs |
| `NAMES.md` | human table short ↔ lab id |

Naming rules (per `family` in `targets.yaml`):

1. Sort non-`emit-builtin` by `version` desc, then `captured_at` desc.
2. Append all `emit-builtin` (stock Hello*_Auto / pins) at the **tail**, also by version desc.
3. Assign `family`, `family-1`, `family-2`, … — **unsuffixed = newest**.
4. Adding a newer Chrome capture shifts previous `chrome` → `chrome-1`, etc.
5. Brave / Edge / Tor / Yandex are **not** in the `chrome` family.

Legacy aliases in export: `chrome_psk`… → `chrome`; empty fingerprint → `chrome`.

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

## Verification before shipping

```powershell
./lab.ps1 verify -Id chromium-stable
./lab.ps1 export
```

JA4 must match `profile.json` → `expected.ja4`. JA3 may drift (GREASE); marked `ja3_hash_unstable`.

## Notes

- Capture server forces ALPN `http/1.1` so verify can read JA4 response headers. JA4 may show `h1` even when the real client prefers `h2`. Cipher/extension material remains valid for ClientHello mimicry.
- GREASE / extension shuffle: expect JA3 variance; pin on JA4 + structural verify.

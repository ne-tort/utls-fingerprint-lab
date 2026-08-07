# utls-fingerprint-lab

Standalone Docker lab that captures real TLS **ClientHello** bytes, promotes them
to importable **uTLS** profiles (`utls-raw-clienthello-v1`), and verifies replay.

**Product feature:** [FEATURE 018 — UTLS_PROFILES](../../SPECS/FEATURES/018-UTLS_PROFILES/FEATURE.md)
(`with_lx_utls`). This lab is the **source of truth** for catalog short names
consumed by [sing-box-lx](https://github.com/ne-tort/sing-box-lx) as submodule
`lx-test/utls-fingerprint-docker`.

## Quick start

```powershell
./lab.ps1 build
./lab.ps1 capture          # all active targets in targets.yaml
./lab.ps1 refresh-latest   # track=latest only (docker pull / CACHEBUST)
./lab.ps1 verify           # replay every profiles/*/clienthello.bin
./lab.ps1 export           # dist/export: deduped short names + aliases
./lab.ps1 test             # smoke subset
```

Unix: `./lab.sh …` (same commands; filters via `ID=` / `GROUP=`).

sing-box-lx: `make -f Makefile.lx lx-utls-sync` ([FEATURE 018](../../SPECS/FEATURES/018-UTLS_PROFILES/FEATURE.md) / SPEC 064).

## Layout

| Path | Role |
|------|------|
| `lab.ps1` / `lab.sh` | **Single entry point** |
| `targets.yaml` | Registry (`family`, `version`, `track`, `pin`, `image_policy`) |
| `targets.archive.yaml` | Auto-archived pins from latest JA4 changes |
| `compose.yaml` | Generated (`scripts/gen-compose.py`) |
| `capture/` | TLS peek server → `captures/` + `profiles/` |
| `tools/` | `labctl`, `verify`, `emit-builtin` |
| `clients/` | Dedicated client images |
| `profiles/` | Committed fingerprints (+ `slot.json` for latest) |
| `docs/` | EXTENDING, IMPORT, VERSIONING, LINUX_WAVE |
| `WISHLIST.md` | Deferred / host-only targets |

## Profile contract

Each `profiles/<id>/`:

- `clienthello.bin` — full TLS record (source of truth)
- `profile.json` — `utls-raw-clienthello-v1`, expected JA4
- `meta.json` / `slot.json` — lab metadata

Export collapses identical JA4 within a family; see [docs/IMPORT.md](docs/IMPORT.md)
and [docs/VERSIONING.md](docs/VERSIONING.md).

## Build model

1. Host cross-compiles Linux binaries (`capture/bin`, `tools/bin`, …).
2. Thin runtime images `COPY` those binaries.
3. `track: latest` builds use `--pull always` and Dockerfile `CACHEBUST`.

## Requirements

- Docker Desktop / Engine with BuildKit
- Go 1.24+ (host) for cross-compile / `labctl`
- Python 3 (regen `compose.yaml` aliases)

## License

MIT — fingerprints are observations, not browser IP.

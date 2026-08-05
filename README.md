# utls-fingerprint-lab

Standalone Docker lab that captures real TLS **ClientHello** bytes, promotes them
to importable **uTLS** profiles (`utls-raw-clienthello-v1`), and verifies replay.

Designed to live as its own git repository (later: submodule of `sing-box-lx`).

## Quick start

```powershell
./lab.ps1 build
./lab.ps1 capture          # all active targets in targets.yaml
./lab.ps1 verify           # replay every profiles/*/clienthello.bin
./lab.ps1 test             # smoke subset
```

Unix: `./lab.sh …` (same commands; filters via `ID=` / `GROUP=`).

## Layout

| Path | Role |
|------|------|
| `lab.ps1` / `lab.sh` | **Single entry point** |
| `targets.yaml` | Registry of what to capture + how (extension contract) |
| `compose.yaml` | Generated from targets (`scripts/gen-compose.py`) |
| `capture/` | TLS peek server → `captures/` + `profiles/` |
| `tools/` | `labctl`, `verify`, `emit-builtin` |
| `clients/` | Optional dedicated client images |
| `profiles/` | Committed importable fingerprints |
| `catalog/` | Reference indexes (JA4 lists, external DB mirrors) |
| `docs/` | Architecture, extending, sing-box import |
| `WISHLIST.md` | Deferred targets from SPEC 062 |

## Profile contract (for sing-box / metacubex uTLS)

Each `profiles/<id>/`:

- `clienthello.bin` — full TLS record (source of truth)
- `profile.json` — format `utls-raw-clienthello-v1`, expected JA4
- `meta.json` — capture metadata

Import: see [docs/IMPORT.md](docs/IMPORT.md).

## Requirements

- Docker Desktop / Engine with BuildKit
- Go 1.24+ (host) for `bin/labctl`
- Python 3 (regen `compose.yaml` aliases)

## License

MIT — same spirit as upstream tooling; fingerprints are observations, not browser IP.

# Versioning, latest channels, and JA4 dedup

See also [EXTENDING.md](EXTENDING.md) and [IMPORT.md](IMPORT.md).

## Target identity vs fingerprint identity

- **Target** (`targets.yaml` id): how we capture (runner + `track`).
- **Fingerprint** (JA4 / ClientHello): what we ship as a short name.

`track: latest` overwrites `profiles/<id>/` on each refresh. On JA4 change, the previous slot is archived to `profiles/<family>-<major>/` + `targets.archive.yaml`.

## Dedup

Export keeps **one short name per (family, JA4)**. Duplicate lab IDs are catalog `aliases` only — no extra embed blobs.

```text
./lab.ps1 refresh-latest   # pull + capture all latest + export --check-dedup
./lab.ps1 export           # always runs --check-dedup
```

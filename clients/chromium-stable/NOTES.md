# chromium-stable — bootstrap notes

## Approach

- Image: `zenika/alpine-chrome` (Chromium + Alpine).
- Headless + `--ignore-certificate-errors` so self-signed capture cert is OK.
- SNI/Host: `chromium-stable.fp.lab.local` — DNS alias на сервисе `capture`
  (см. `docker-compose.yml` → `networks.default.aliases`).
- Profile: `wave1`.

## Run

```powershell
cd lx-test/utls-fingerprint-docker
docker compose up -d capture
docker compose --profile wave1 run --rm chromium-stable
```

## Expected artifact

`captures/<n>-chromium-stable/meta.json` with JA4 ≈ current Chrome/Chromium,
`utls_fingerprinter_ok: true`.

## Caveats

- Headless may differ slightly from headed desktop Chrome (document if JA4 drifts).
- GREASE / extension shuffle → take ≥2 samples; compare JA4 stability.
- Image tag pin when first green; bump deliberately later (`chromium-beta` target).
- Do **not** write `/etc/hosts` in this image (Permission denied) — use compose aliases.

## Result (2026-08-05)

JA4 stable: `t13d1514h2_acb858a92679_02713d6af862` (2 samples).  
See [WAVE1_CHROMIUM.md](../../../SPECS/TASKS/062-UTLS_FINGERPRINT_LAB/WAVE1_CHROMIUM.md).

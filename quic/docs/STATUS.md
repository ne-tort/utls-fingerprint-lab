# Lab status — QUIC fingerprint (`quic/`)

Скоуп сейчас: **только лаба**. Product embed / CI / `quic.fingerprint` wiring —
вне скоупа (готовим `dist/export` как контракт).

## Сделано

| Область | Артефакт |
|---------|----------|
| Observation | `quic-raw-initial-v1` profiles + capture/emit matrix |
| Unified profile | `schema/quic-utls-profile-v1.schema.json` |
| Curated catalog | `catalog/utls/*.json` (SoT via `sync_utls_catalog.py`) |
| Match gate | `./lab.ps1 unify` → `--strict-all` (24 pinned lab ids) |
| Stable vs random | `./lab.ps1 exp-stable` + `fixtures/stable-random-experiment.json` |
| Auth channel design | GREASE value; demux без fingerprint (docs + module I) |
| Export prep | `./lab.ps1 export` → `dist/export/` (`quic-utls-catalog-v1`) |
| Shared lib | `scripts/quic_utls_lib.py` (norm/match/dial/family/write) |
| Emit from catalog | `emitters/fromprofile` (engine path) |

## Команды лабы

```powershell
./lab.ps1 unify          # sync + extract drafts + match --strict-all
./lab.ps1 exp-stable     # live emit×2 + stable/random asserts
./lab.ps1 export         # dist/export (+ regenerate emit_templates.json)
```

## Ещё нужно в лабе

1. **Docker compose service** для `emit-fromprofile` (сейчас host/Windows path).
2. **Auth overlay live test** в лабе (PSK → GREASE value → capture verify) без product.
3. **Structured emit** для yandex/firefox (сейчас `observation_only` + skeleton).
4. **JA4 annotate** в unify gate (optional / grease-tolerant).
5. **Очистка** `profiles/peer-*` noise; fixtures без `*.log`.
6. **Контракт export ↔ future product** зафиксировать в IMPORT (без реализации sync).
7. Поднять curated shorts при дрейфе live Chrome (diff vs chromeparrot).

## Не делать в лабе / отложено

- `common/quic/lxquicfp`, `lx-quic-fp-sync`, CI check
- `quic.fingerprint` option в sing-box
- Смешивание с TCP `lxutls` / `clienthello.bin` blunt replay

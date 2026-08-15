# quic/ — QUIC Initial fingerprint lab

Отдельный контур внутри [utls-fingerprint-lab](../README.md) для захвата
**QUIC Client Initial** (не TCP ClientHello FEATURE 018).

Связь в sing-box-lx: [SPEC 090](../../../../SPECS/TASKS/090-QUIC_FINGERPRINT_PROFILES/SPEC.md)
· ChromeParrot / `quic_auth` · блокер [027](../../../../SPECS/TASKS/027-UTLS_OVER_QUIC/SPEC.md).

## Зачем отдельно от TCP uTLS

| | TCP lab (корень) | Эта папка `quic/` |
|--|------------------|-------------------|
| Wire | TLS record ClientHello | QUIC Initial UDP (+ CRYPTO → CH) |
| Формат профиля | `utls-raw-clienthello-v1` | **`quic-raw-initial-v1`** |
| Replay / emit | metacubex/utls HelloCustom | sagernet ChromeParrot / будущий catalog |
| Конфиг продукта | `tls.utls.fingerprint` | `quic.fingerprint` / parrot (не `tls.utls`) |

TCP short names **нельзя** импортировать как QUIC-профиль.

## Quick start (Docker-first)

```powershell
cd quic
./lab.ps1 build-emitters
./lab.ps1 unify            # quic-utls catalog match (stable identity)
./lab.ps1 matrix           # parrot / uquic / aioquic → compare
./lab.ps1 roundtrip        # prove emit recipes reproduce structurally
./lab.ps1 compare
```

Docs: [STATUS](docs/STATUS.md) · [UTLS_PROFILE](docs/UTLS_PROFILE.md) · [IMPORT](docs/IMPORT.md) ·
[REPLAY_AND_EMIT](docs/REPLAY_AND_EMIT.md).

## Emit vs observation

Raw `initials/*.bin` = observation (match/diff). Dial needs a **unified**
profile (`quic-utls-profile-v1` in `catalog/utls/`) with `emit.engine` and/or
`emit.structured` — see [UTLS_PROFILE](docs/UTLS_PROFILE.md).
ChromeParrot/hy2 → short `chrome`; plain+datagrams → `quic-go-datagram`;
uquic → lab `uquic_preset` until structured port; aioquic → unmatched / match_only.

## Layout

| Path | Role |
|------|------|
| `lab.ps1` / `lab.sh` | Entry (`unify`, `matrix`, `compare`, …) |
| `catalog/utls/` | **`quic-utls-profile-v1`** shorts (SoT for dial) |
| `compose.yaml` | Docker capture + emitters |
| `scripts/build-emitters.ps1` | linux bins → `bin/` |
| `scripts/match_utls_catalog.py` | Classify profiles → shorts |
| `scripts/compare_profiles.py` | TP id-set table |
| `targets.yaml` | Registry |
| `capture/` | UDP Initial peek (`clienthellod`) |
| `emitters/` | chromeparrot / fromprofile / uquic |
| `captures/` `profiles/` | Raw observation |
| `docs/` | UTLS_PROFILE, CONTRACT, RESULTS_MATRIX |
| `schema/` | JSON Schema |

## Profile contract (кратко)

Каждый `profiles/<id>/`:

- `initials/*.bin` — один или несколько UDP datagram Initial (source of truth)
- `clienthello.bin` — собранный TLS ClientHello (если reassembly OK)
- `transport_parameters.bin` + `tp.json` — TP extension payload / разобранное
- `profile.json` — `format: quic-raw-initial-v1`, ожидаемые JA4 (`q…`), id-set TP, CID lens

Детали: [docs/CONTRACT.md](docs/CONTRACT.md).

## Интеграция (лаба → будущий продукт)

Сейчас готовим только лабу. Export:

```powershell
./lab.ps1 export   # → dist/export/ (quic-utls-catalog-v1)
```

Не смешивать с TCP `lxutls`. Product sync / CI — позже.
Статус: [docs/STATUS.md](docs/STATUS.md) · [docs/IMPORT.md](docs/IMPORT.md).


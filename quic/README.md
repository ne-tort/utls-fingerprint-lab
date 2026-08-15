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
./lab.ps1 matrix           # parrot / uquic / aioquic → compare
./lab.ps1 roundtrip        # prove emit recipes reproduce structurally
./lab.ps1 compare
```

Docs: [REPLAY_AND_EMIT](docs/REPLAY_AND_EMIT.md) · [PYTHON_VS_GO](docs/PYTHON_VS_GO.md) ·
[RESULTS_MATRIX](docs/RESULTS_MATRIX.md).

## Emit vs observation

Raw `initials/*.bin` = observation (match/diff). Dial in sing-box needs
`emit` recipes (`quic-emit-spec-v1`) — see catalog/`emit_templates.json`.
ChromeParrot/hy2 → `sagernet_chrome_parrot`; uquic → lab `uquic_preset` until
ported; aioquic → `match_only`.

## Layout

| Path | Role |
|------|------|
| `lab.ps1` / `lab.sh` | Entry (`matrix`, `compare`, …) |
| `compose.yaml` | Docker capture + emitters |
| `scripts/build-emitters.ps1` | linux bins → `bin/` |
| `scripts/compare_profiles.py` | TP id-set table |
| `targets.yaml` | Registry |
| `capture/` | UDP Initial peek (`clienthellod`) |
| `emitters/` | chromeparrot / uquic |
| `captures/` `profiles/` | Raw + importable |
| `docs/` | Contract, ecosystem, **RESULTS_MATRIX** |
| `schema/` | JSON Schema манифеста |

## Profile contract (кратко)

Каждый `profiles/<id>/`:

- `initials/*.bin` — один или несколько UDP datagram Initial (source of truth)
- `clienthello.bin` — собранный TLS ClientHello (если reassembly OK)
- `transport_parameters.bin` + `tp.json` — TP extension payload / разобранное
- `profile.json` — `format: quic-raw-initial-v1`, ожидаемые JA4 (`q…`), id-set TP, CID lens

Детали: [docs/CONTRACT.md](docs/CONTRACT.md).

## Интеграция в sing-box (целевая)

Не смешивать с `common/tls/lxutls`. Черновик пакета:
`common/quic/lxquicfp` (build-tag рядом с parrot) + sync `make lx-quic-fp-sync`
из `quic/dist/export/`. UX: `quic.fingerprint: "chrome"` — см.
[docs/APPLY.md](docs/APPLY.md).

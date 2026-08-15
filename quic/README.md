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
# once: clone refs (from lab root)
#   git clone --depth 1 https://github.com/gaukas/clienthellod.git _refs/clienthellod
#   git clone --depth 1 https://github.com/refraction-networking/uquic.git _refs/uquic

cd quic
./lab.ps1 build-emitters   # linux bins → bin/
./lab.ps1 matrix           # capture + ChromeParrot / plain / uquic presets → compare
./lab.ps1 compare
```

Host-only peek (dev):

```powershell
./lab.ps1 capture-listen -Target manual
./lab.ps1 parse -Path .\captures\...\initials\000.bin
```

## Emitters vs hy2

| Emitter | What it is |
|---------|------------|
| `emit-chromeparrot` | sagernet `ChromeParrot` (same quic-go as hy2 default) |
| `emit-quicgo-plain` | same binary, parrot off (= hy2 `disable_chrome_parrot`) |
| `emit-uquic-*` | refraction uquic presets (Chrome_146/115, Firefox_116) |

Full hy2 outbound JSON path can be added later (sing-box binary mount); ChromeParrot probe is the intentional wire-level twin of hy2 parrot for fingerprint dumps.

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

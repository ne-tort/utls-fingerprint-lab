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

## Quick start (scaffold)

```powershell
cd quic
./lab.ps1 help
./lab.ps1 build          # cross-compile capture → capture/bin
./lab.ps1 capture-listen # UDP peek на :4433 → captures/ + profiles/
# в другом терминале: клиент H3/QUIC на capture:4433
./lab.ps1 parse -Path ..\captures\...\initial-0.bin
```

Unix: `./lab.sh …`.

Полный docker compose браузеров H3 — следующий этап (см. `docs/CAPTURE.md`);
сейчас — парсер + listen + контракт + targets registry.

## Layout

| Path | Role |
|------|------|
| `lab.ps1` / `lab.sh` | Entry |
| `targets.yaml` | Registry (browsers / parrot / libraries) |
| `capture/` | UDP Initial peek (`gaukas/clienthellod`) |
| `captures/` | Сырые datagram’ы + JSON parse |
| `profiles/` | Импортируемые профили `quic-raw-initial-v1` |
| `docs/` | Что такое FP, экосистема, контракт, apply |
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

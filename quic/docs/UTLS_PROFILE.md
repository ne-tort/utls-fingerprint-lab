# Unified QUIC fingerprint profile (`quic-utls-profile-v1`)

Цель: один документ на persona, из которого после разбора observation
можно **стабильно** снова dial'ить на sagernet/quic-go — по аналогии с
`tls.utls.fingerprint`, без blunt-replay сырых Initial.

## Слои (не смешивать)

| Слой | Где | Роль |
|------|-----|------|
| Observation | `profiles/<id>/` (`quic-raw-initial-v1`) | что сняли |
| Unified profile | `catalog/utls/<short>.json` | identity + emit + **random** + **auth** |
| Engine dial | `emit.engine` → parrot / datagrams / SCID0 | стабильный путь сегодня |
| Structured | `emit.structured` ≈ QUICSpec | будущий apply |
| Auth | `auth.channel = tp_grease_value` | demux `quic.auth` без fingerprint |

```text
quic.fingerprint = "chrome"     # persona / emit
quic_auth = { psk, short_id }   # overlay в GREASE value (optional)

demux rule:  { quic: { auth: true } }   # NO fingerprint required
```

## Стабильные поля identity

Сравниваем / пиним:

- `tp_id_set` с токеном `GREASE` (не значение grease, не id 31N+27)
- markers: `0x11`, `0x3128`, `0x20`, `0x2ab2`, …
- `scid_len` когда фиксирован (parrot / uquic FF = pin)
- `datagram_count_class` (1 vs 2)

Не пиним: сырые `initials/*.bin`, точный JA4 между прогонами, DCID, PN,
значение GREASE, shuffle TP/CH.

## Random (`random.slots`)

Явный список энтропии. `stable_for_identity=false` → **не** match-ключ.

| Типичный slot | Слой | Chrome parrot |
|---------------|------|---------------|
| `tp_grease_id` | TP | random large 31N+27 → token `GREASE` |
| `tp_grease_value` | TP | random 0..15 **или** auth HMAC 8..15 |
| `tp_order` | TP | shuffle (match = set) |
| `tp_0x11_version_grease` | TP | внутри `0x11` |
| `ch_ech_grease` / `ch_ext_shuffle` | CH | да |
| `initial_dcid` / `initial_pn` | Initial | да |

Policy grease в structured: `random_0_15_or_auth_8_15` (chrome) /
`random_0_15` (quic-go).

## Auth (`auth`) — без demux fingerprint

| Поле | Смысл |
|------|--------|
| `capable` | emit может положить HMAC сегодня (`true` только у `chrome`) |
| `channel` | `tp_grease_value` |
| `demux_match` | `{ "quic": { "auth": true } }` |
| `requires_demux_fingerprint` | **всегда false** (предпочтительный дизайн) |
| `emit_hook` | `Config.ChromeGREASEValue` |

Demux verify (`QuicAuthVerifier`) читает **любые** RFC GREASE values — не
требует short `chrome` / `quic.client`. Fingerprint и auth ортогональны.

Детали: SPECS module [I-QUIC_AUTH_CHANNEL](../../../../SPECS/TASKS/090-QUIC_FINGERPRINT_PROFILES/modules/I-QUIC_AUTH_CHANNEL.md).

## Emit status

| `status` | Dial |
|----------|------|
| `emit_ready` | да (`chrome`, `quic-go`, `quic-go-datagram`) |
| `experimental` | lab (`uquic_*`) |
| `observation_only` | match only (+ structured skeleton) |

## Команды

```powershell
./lab.ps1 unify        # sync catalog + match --strict-all
./lab.ps1 exp-stable   # live emit×2 chrome/quic-go/quic-go-datagram + asserts
python scripts/experiment_stable_random.py
```

DoD unify: **все** pinned lab `profiles/*` матчат short (не `exp-*`/`peer-*`).  
DoD exp-stable: stable fields equal; random slots vary **per emit style**.

### Что рандомится (факт лабы)

| Slot | chrome parrot | quic-go plain |
|------|---------------|---------------|
| TP id-set / markers / named windows | stable | stable |
| TP hex_id (clienthellod) | stable | stable |
| CH ext **normalized** / ciphers / groups | stable | stable |
| CH ext **order** / frame chaos | **random** | **fixed** |
| initials SHA / clienthello.bin | random | random (CID/keys/GREASE value) |
| GREASE value (auth overlay) | random or HMAC 8..15 | random 0..15 (no auth emit) |

Фикстура: `fixtures/stable-random-experiment.json`.  
Статус лабы: [STATUS.md](STATUS.md).

Схема: [../schema/quic-utls-profile-v1.schema.json](../schema/quic-utls-profile-v1.schema.json).  
Каталог SoT: [../catalog/utls/](../catalog/utls/) (`sync_utls_catalog.py`).


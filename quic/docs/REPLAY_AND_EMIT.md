# Observation vs emit — can we reproduce a captured fingerprint?

## Short answers

| Question | Answer |
|----------|--------|
| Сняли hy2/ChromeParrot — сможем потом воспроизвести в sing-box/quic-go? | **Да**, но не из сырых `initials/*.bin`. Сегодня emit = флаг `ChromeParrot` / short `chrome`. Capture **проверяет** identity, не кормит blunt replay. |
| Нужна ли ручная работа над каждым отпечатком? | **Да для новых persona** (firefox, live Chrome drift, quiche). Нет для уже поддержанных engine flags. |
| Сможем воспроизвести uquic отпечатки в sing-box? | **В лабе — да** (emit-uquic). **В продукте — нет out-of-box**: нужен порт `QUICSpec` в sagernet/quic-go или structured emit. |
| Как у TCP uTLS (`clienthello.bin` → HelloCustom)? | QUIC **не** так. Raw Initial нельзя «проиграть» как валидный handshake (TP↔SCID, keys, PN). |

## Two artifacts (обязательное разделение)

```text
┌─ Observation  (quic-raw-initial-v1) ─────────────────────┐
│  initials/*.bin, clienthello.bin, tp.json, expected.*    │
│  Назначение: match, diff, CI, demux identity             │
│  НЕ источник blunt emit                                  │
├─ Emit recipe (quic-emit-spec-v1) ────────────────────────┤
│  Как СНОВА построить такой же Initial на dial            │
│  Типы (emit_kind) — см. ниже                             │
└──────────────────────────────────────────────────────────┘
```

Один lab profile **может** содержать оба: observation + `emit` блок.
Export в продукт импортирует **emit recipe** (+ expected для verify), не UDP dumps.

## Emit kinds (унификация)

| `emit_kind` | Когда | Воспроизведение |
|-------------|-------|-----------------|
| `sagernet_chrome_parrot` | hy2/tuic default, matrix `chromeparrot` | sing-box: `ChromeParrot=true` |
| `sagernet_plain` | `disable_chrome_parrot` / `quicgo` | `ChromeParrot=false` |
| `uquic_preset` | lab reference (`chrome-146`, `firefox-116`, …) | lab Docker only; product = port later |
| `structured` | универсальный JSON ≈ uquic `QUICSpec` | будущий lxquicfp → sagernet dial hooks |
| `match_only` | aioquic/quiche/browser пока нет emit | только demux/lab identity |

**Универсальный путь в продукт:** со временем все ценные persona сводятся к
`structured` (или тонкому engine flag, если 1:1 с кодом). `uquic_preset` —
промежуточный ярлык «как в refraction», не module path в go.mod.

Схема: [../schema/quic-emit-spec-v1.schema.json](../schema/quic-emit-spec-v1.schema.json).

## Round-trip (как доказываем, что сняли правильно)

```text
emitter (known emit_kind)
  → capture observation A
  → same emit_kind again
  → capture observation B
  → compare structural identity(A,B)
```

**Structural identity** (stable): `tp_id_set` (GREASE as token), `scid_len`,
`dcid_len`, datagram count class (1 vs 2), optional JA4 `q…` with grease-tolerant
mode.

**Not compared byte-wise:** GREASE values, ECH grease, extension shuffle order,
absolute `initials/*.bin`.

Команда лабы: `./lab.ps1 roundtrip` (chromeparrot + uquic146 + uquicff).

## Mapping к будущему sing-box import

```json
"quic": { "fingerprint": "chrome" }
```

резолвится в catalog entry:

```json
{
  "short": "chrome",
  "emit": { "emit_kind": "sagernet_chrome_parrot" },
  "expected": { "tp_id_set": ["0x11", "0x3128"], "scid_len": 0 }
}
```

```json
"quic": { "fingerprint": "firefox" }
```

пока либо error «unsupported emit», либо (после порта) `structured` /
engine `firefox_parrot`.

Импорт lab → product **отклоняет** профили без `emit` или с
`emit_kind=match_only` для dial (для demux match — OK).

## Practical rule

1. Снимаем observation со **всех** источников (браузеры, libs, parrot).
2. Для dial в sing-box заводим/обновляем **emit recipe** отдельно.
3. ChromeParrot/hy2 уже закрыты recipe’ом `sagernet_chrome_parrot`.
4. uquic/firefox — сначала lab roundtrip через `uquic_preset`, потом порт в
   `structured` / native parrot.
5. Сырой dump без recipe = **только** исследование и demux fingerprints.

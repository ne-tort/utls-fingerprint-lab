# Contract — `quic-raw-initial-v1`

Analog of parent `utls-raw-clienthello-v1`, for **QUIC Initial** personas.

## Directory layout

```text
profiles/<lab-id>/
  profile.json              # manifest (this contract)
  meta.json                 # capture metadata
  initials/
    000.bin                 # first UDP datagram (required)
    001.bin                 # … if multi-datagram (Chrome ML-KEM)
  clienthello.bin           # reassembled TLS ClientHello (optional if parse fail)
  transport_parameters.bin  # raw TP extension payload (optional)
  tp.json                   # decoded TP id → value (best-effort)
  header.json               # version, dcid_len, scid_len, token_len, …
```

## `profile.json` (normative fields)

```json
{
  "format": "quic-raw-initial-v1",
  "id": "chromium-h3-stable",
  "family": "chrome",
  "version": 139,
  "track": "latest",
  "expected": {
    "ja4": "q13d…_…_…",
    "tp_id_set": ["0x11", "0x3128"],
    "dcid_len": 8,
    "scid_len": 0
  },
  "notes": "live Chromium H3; GREASE unstable"
}
```

| Field | Meaning |
|-------|---------|
| `format` | Must be `quic-raw-initial-v1` |
| `family` | Export short-name family (`chrome`, `firefox`, `quic-go`, `quiche`, …) |
| `expected.ja4` | FoxIO-style JA4 with **`q`** prefix when available |
| `expected.tp_id_set` | Sorted hex ids present in client TP (stable chrome markers: `0x11`, `0x3128`) |
| `expected.scid_len` / `dcid_len` | Header fingerprint; Chrome parrot uses SCID 0 |

Dedup within a family: same `(ja4, tp_id_set, scid_len)` → one short; others aliases.
(Exact export rules TBD when `export` lands — mirror parent JA4 dedup.)

## Stability

| Signal | Stable? |
|--------|---------|
| TP id-set (chrome google ids) | high |
| Cipher/extension **structure** | high |
| GREASE values / ECH grease / shuffle order | **unstable** — do not pin bytes |
| Absolute `initials/*.bin` | observation; emit rebuilds from spec |

## Relation to `utls-raw-clienthello-v1`

Different `format` string. Sync tooling must refuse to embed a QUIC profile into
`lxutls` TCP catalog and vice versa.

# Ecosystem — projects for capture / profiles / mimicry

Ground truth for **this** lab is always our capture → `quic/profiles/`.
External projects are references (clone under `_refs/` if needed; gitignored).

## Capture / observatory (parse Initial → JSON / IDs)

| Project | Role | Integration idea |
|---------|------|------------------|
| **[gaukas/clienthellod](https://github.com/gaukas/clienthellod)** | Go lib: `UnmarshalQUICClientInitialPacket`, multi-packet gather, fingerprint IDs (header / CH / TP combo). Powers [quic.tlsfingerprint.io](https://quic.tlsfingerprint.io/) | **Primary parser** in `quic/capture` |
| **[tlsfingerprint.io](https://tlsfingerprint.io/) / QUIC observatory** | Public catalog of observed TLS+QUIC fingerprints | Compare our JA4/TP hashes; not embed |
| **FoxIO JA4 / JA4+** | `q…` JA4 over QUIC CH; TP often harvested **beside** JA4 | Store `expected.ja4`; TP id-set separately |
| **Crank-Git/ja4plus** (and Zeek JA4) | Offline pcap → JA4 including QUIC Initial decrypt | Optional verify tool |
| **Scrapfly H3/QUIC tools** | Documents JA4 + TP + H3 SETTINGS side-by-side | Schema inspiration only |

## Mimicry / emit (apply fingerprint)

| Project | Role | Integration idea |
|---------|------|------------------|
| **sagernet/quic-go `ChromeParrot`** | Live Chrome-shaped emit + GREASE slot for `quic_auth` | Product **chrome** emitter today; refresh from lab diffs |
| **[refraction-networking/uquic](https://github.com/refraction-networking/uquic)** | Hard-fork quic-go: `QUICSpec`, presets `QUICChrome_115`, `QUICFirefox_116A/B/C` | **Reference** for firefox emit shape; **do not** replace module path |
| **enetx/uquic** | Fork of uquic line | Same class; not upstream for us |
| **metacubex/utls Hello\*** | TCP only | **Not** a QUIC emit source ([027](../../../../SPECS/TASKS/027-UTLS_OVER_QUIC/SPEC.md)) |

## TCP-only (parent lab / do not confuse)

| Project | Note |
|---------|------|
| Parent `docs/SOURCES.md` (parroteer, fingerprint-db, curl-impersonate) | TCP ClientHello / JA3/JA4 `t…` |
| FEATURE 018 `lxutls` | `utls-raw-clienthello-v1` — wrong layer for QUIC |

## How pieces map to our contract

```text
Live client / parrot
    → UDP Initial(s)
    → clienthellod parse (+ optional JA4)
    → quic-raw-initial-v1 profile
    → export shorts (chrome, firefox, quic-go, …)
    → sing-box: match (demux) and/or emit (ChromeParrot / future *Parrot)
```

uquic presets = optional **emit fixtures** in CI (“does our firefox path look
like QUICFirefox_116?”), not the catalog source of truth.

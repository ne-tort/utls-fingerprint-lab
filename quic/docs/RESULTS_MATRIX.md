# Matrix results — ChromeParrot vs uquic vs plain

Captured with `./lab.ps1 matrix` (Docker) after fixing clienthellod gather
deadline (`GatherClientInitials()` with zero deadline is immediately expired —
must use `GatherClientInitialsWithDeadline`).

Date: 2026-08-15 · stack: sagernet ChromeParrot emitter + `_refs/uquic` tip
(includes `QUICChrome_146`).

## Table (TP id-set)

| id | family | datagrams | `0x11` | `0x3128` | Distinct TP notes |
|----|--------|-----------|--------|----------|-------------------|
| **chromeparrot** | chrome | **2** | **Y** | **Y** | hy2-equivalent emit; GREASE; google options |
| **quicgo** | quic-go | 2 | n | n | no google ids; has `0xb`,`0xe` |
| **uquic146** | chrome | 2 | **n** | Y | +`0x3127` initial RTT; **no** `0x11` version_information |
| **uquic115** | chrome | 1 | **n** | Y | +`0x4752` google_quic_version; no `0x11` |
| **uquicff** | firefox | 1 | n | n | +`0x2ab2`; different base set |
| **aioquic** | aioquic | 1 | **Y** | **n** | has `0x11` but **no** `0x3128`; SCID len 8; `match_only` |

## Roundtrip

`./lab.ps1 roundtrip` — chromeparrot×2 and uquic146×2 → **structural_match=true**
(GREASE-tolerant tp_id_set + datagram count).

See [REPLAY_AND_EMIT.md](REPLAY_AND_EMIT.md).

## Conclusions

1. **Our ChromeParrot ≠ uquic Chrome_115/146 on TP id-set.** Demux match on
   `0x11`+`0x3128` (SPEC 087) uniquely hits **sagernet parrot**, not uquic
   presets (those lack `0x11`).
2. **Freshness work** should diff live Chromium H3 against **chromeparrot**,
   not blindly import uquic 146 as “current Chrome”.
3. **uquic remains valuable** as Firefox emit reference and as a second chrome
   persona (library mimic), not as drop-in replacement for hy2 parrot.
4. **Plain quic-go** cleanly separates from parrot (no `0x11`/`0x3128`) — good
   negative control / `quic-go` short.
5. Multi-datagram (2) correlates with ML-KEM-sized CH (chromeparrot, uquic146);
   firefox/uquic115/aioquic fit in one UDP datagram in this run.
6. **aioquic ≠ chromeparrot** despite sharing `0x11` — needs `0x3128` for parrot match.

## Reproduce

```powershell
cd lx-test/utls-fingerprint-docker/quic
./lab.ps1 matrix
./lab.ps1 compare
```

Raw profiles: `profiles/{chromeparrot,quicgo,uquic146,uquic115,uquicff}/`.

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
| **curlquiche** | quiche | 1 | n | n | DCID=16 SCID=20; no google ids; `match_only` |
| **chromium** | chrome (live) | **2** | **n** | **n** | `0x4752`+`0xff73db`; **≠ parrot**; zenika alpine image |

## JA4 (`expected.ja4`, ja4plus)

Filled with `./lab.ps1 ja4` (wraps `initials/*.bin` → UDP → ja4plus). Prefix `q` = QUIC.

| id | ja4 |
|----|-----|
| chromeparrot | `q12d039900_55b375c5d22e_5cae79f3dfec` |
| chromium | `q13d0311h3_55b375c5d22e_5a1f323ef56d` |
| uquic146 | `q13d0311h3_55b375c5d22e_653d80c3fe9d` |
| uquic115 | `q13d0310h3_55b375c5d22e_cd85d2d88918` |
| uquicff | `q13d0314h3_55b375c5d22e_2d2a40a25571` |
| quicgo | `q13d0313h3_55b375c5d22e_f902b76752af` |
| aioquic | `q13d0307h3_55b375c5d22e_1cecd519fee8` |
| curlquiche | `q13d0308h3_55b375c5d22e_f0736a66fa6b` |

Notes: chromeparrot shows ALPN `00` in this capture (ja4plus reading); live chromium /
uquic146 share the middle hash (`55b375c5d22e`) but differ on the third. Roundtrip
`uquic146`/`uquic146-b` share identical JA4; chromeparrot×2 can differ (GREASE / CH
noise) — structural compare stays GREASE-tolerant, JA4 does not.

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
7. **curl+quiche** — distinct CID lengths (16/20) and TP set; live via `./lab.ps1 live -Client curl`.
8. **Live Chromium (zenika/alpine-chrome) ≠ chromeparrot.** Same 2-datagram size class,
   but TP has `0x4752`/`0xff73db` and **lacks** `0x11`/`0x3128` (closer to old Chrome /
   uquic115 shape than to sagernet parrot). Freshness ≠ “parrot is current Chrome”;
   parrot remains a **fixed emit persona**. Image age matters — bump Chromium image
   when re-checking freshness.

## Reproduce

```powershell
cd lx-test/utls-fingerprint-docker/quic
./lab.ps1 matrix
./lab.ps1 compare
```

Raw profiles: `profiles/{chromeparrot,quicgo,uquic146,uquic115,uquicff}/`.

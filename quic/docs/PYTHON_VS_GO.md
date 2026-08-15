# Python vs Go for QUIC fingerprint capture

## Recommendation (this lab)

| Task | Prefer | Why |
|------|--------|-----|
| Parse / reassemble Initial → profile | **Go (`clienthellod`)** | Same stack as quic.tlsfingerprint.io; already in `quic/capture` |
| Emit sagernet ChromeParrot / plain | **Go** | Must link `sagernet/quic-go` patches |
| Emit uquic presets | **Go** | `QUICID2Spec` + `UTransport` |
| Emit / sample **aioquic**, many research scripts | **Python** | Native aioquic; fast to add lib personas |
| JA4 `q…` from pcap | **Python + tshark** (FoxIO ja4.py) or Zeek | Mature; BSD JA4 core |
| Headless browser H3 | **Docker browser image** + capture UDP | Language irrelevant; drive Chrome/Firefox |

**Не заменять** Go capture на Python: clienthellod — лучший parser. Python —
дополнительные **emitters** и offline JA4.

## Docker scenarios (lab)

| Scenario | Compose / command | Output identity |
|----------|-------------------|-----------------|
| ChromeParrot ≈ hy2 | `emit-chromeparrot` | `sagernet_chrome_parrot` |
| plain quic-go | `emit-quicgo-plain` | `sagernet_plain` |
| uquic presets | `emit-uquic-*` | `uquic_preset` |
| aioquic | `emit-aioquic` | `match_only` until structured |
| hy2 outbound JSON | `config/hy2-*.json` + sing-box bin | should ≈ parrot/plain |
| Live Chromium/Firefox | wishlist phase 2 | observation → later emit |

```powershell
./lab.ps1 matrix      # includes aioquic
./lab.ps1 roundtrip   # prove recipes reproduce
```

## Why not “one universal raw bin import”?

TCP uTLS: `clienthello.bin` → blunt HelloCustom works on TCP.

QUIC: Initial is bound to connection crypto + live SCID in TP. Catalog import
into sing-box must be **emit recipes** (`quic-emit-spec-v1`), while raw bins
stay for match/diff. See [REPLAY_AND_EMIT.md](REPLAY_AND_EMIT.md).

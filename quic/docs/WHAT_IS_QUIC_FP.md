# What is a QUIC fingerprint?

A **QUIC client fingerprint** is the observable shape of the **first UDP
datagrams** a client sends to open a connection (QUIC **Initial**), including:

1. **Long header** — version, DCID/SCID lengths, token, packet number encoding
2. **Frame layout** — CRYPTO / PADDING / PING “chaos”, coalescing, pad target
3. **TLS ClientHello** inside CRYPTO (ciphers, extensions, ALPN, key shares…)
4. **`quic_transport_parameters`** TLS extension (TP id-set + values + GREASE)

JA4 with prefix `q…` covers mostly the **TLS half**. Transport parameters and
Initial header/pad are **separate signals** — a complete persona must match
both (and stay consistent). See also
[crawlex: H3/QUIC fingerprinting](https://blog.crawlex.net/blog/http3-quic-fingerprinting/).

## vs ChromeParrot

[ChromeParrot](https://github.com/ne-tort/sing-box-lx/blob/lx/patches/quic-go/interface.go)
in sagernet/quic-go is one **emit** implementation of a Chromium-like Initial
(CH + TP + zero SCID + …).

This lab’s job is broader:

- **Capture** unique forms from real browsers / libraries / our parrot
- Store them as **`quic-raw-initial-v1`** profiles
- Later **apply** (emit) selected shorts — ChromeParrot may become the
  `chrome` / `chrome-N` emitter, refreshed from lab dumps; Firefox etc. need
  new emit paths

So: not “must equal ChromeParrot byte-for-byte”, but “stable, named, unique
QUIC personas we can match and optionally parrot”.

## Types of profiles (taxonomy)

| Kind | Example | Typical use |
|------|---------|-------------|
| Browser Chromium | Chrome / Edge / Brave H3 | `chrome` family; often same QTP hash |
| Browser Gecko | Firefox H3 | distinct TP/CID/pad |
| Browser WebKit | Safari | P3; hard |
| Library | quic-go, quiche, aioquic, msquic | match + baseline |
| Tunnel parrot | hy2/tuic ChromeParrot on | regression vs live Chrome |
| Tunnel plain | hy2 `disable_chrome_parrot` | = quic-go persona |
| Research parrot | uquic `QUICFirefox_116*` | reference emitter, not product dep |

## What we do **not** capture here

- TCP ClientHello (parent lab / FEATURE 018)
- HTTP/3 SETTINGS after handshake (useful later; out of v1 contract)
- Application auth (Hy2 password) — orthogonal

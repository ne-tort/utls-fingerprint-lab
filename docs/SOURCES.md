# External sources (reference only)

Clones live under `_refs/` (gitignored). Slim extracts in `catalog/`.

| Source | Use |
|--------|-----|
| [SaamoCha/parroteer](https://github.com/SaamoCha/parroteer) | Real-browser baselines / utls drift model |
| [North-web-dev/fingerprint-db](https://github.com/North-web-dev/fingerprint-db) | Measured uTLS `*_Auto` JA3/JA4/Akamai |
| [lexiforest/curl-impersonate](https://github.com/lexiforest/curl-impersonate) | Live wrappers + YAML signatures |
| [bogdanfinn/tls-client](https://github.com/bogdanfinn/tls-client) | Go HelloID map (not raw bins) |

Ground truth for this lab is always **our capture** → `profiles/`.

## QUIC (see `quic/docs/ECOSYSTEM.md`)

| Source | Use |
|--------|-----|
| [gaukas/clienthellod](https://github.com/gaukas/clienthellod) | Parse QUIC Initial; IDs (powers quic.tlsfingerprint.io) |
| [quic.tlsfingerprint.io](https://quic.tlsfingerprint.io/) | Public QUIC fingerprint observatory |
| [refraction-networking/uquic](https://github.com/refraction-networking/uquic) | Reference parrot presets (Chrome_115 / Firefox_116) |
| FoxIO JA4 (`q…`) | TLS-half hash; TP id-set stored separately |

QUIC profiles use format **`quic-raw-initial-v1`** under `quic/` — not
`utls-raw-clienthello-v1`.

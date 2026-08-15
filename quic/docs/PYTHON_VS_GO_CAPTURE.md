# Python vs Go for QUIC Initial capture (lab note)

Date: 2026-08-15 · refs under `../_refs/`.

## What is already under `_refs/`

| Tree | Role |
|------|------|
| `clienthellod` | **Primary capture/parse** (Go): decrypt Initial, multi-datagram gather, TP + HexID |
| `uquic` | Structured **emit** presets (`QUICSpec` = InitialPacketSpec + ClientHelloSpec) |
| `ja4` | FoxIO JA4+ (Python needs **tshark** JSON; Zeek/Rust paths too) |
| `aioquic` | Python H3 **client library** persona (cloned) |
| `ja4plus` | Pure-Python JA4 with **built-in QUIC Initial decrypt** (cloned; optional sidecar) |

Not cloned (low value for this lab): `quantum-sniffer` (PQ classifier, not profile promote), `QUICkly` (forensic heuristics on header shape only).

## 1. When Python is better vs Go

| Job | Prefer | Why |
|-----|--------|-----|
| Live UDP listen → `quic-raw-initial-v1` promote | **Go** (`quic/capture` + clienthellod) | Already matches contract; multi-datagram gather; HexID; same stack as quic.tlsfingerprint.io |
| Mimic emit (Chrome/Firefox wire shape) | **Go** (uquic / ChromeParrot) | Only mature Initial *control* API in-ecosystem |
| Fill `expected.ja4` (`q…`) from pcap/bins | **Python** | FoxIO `ja4/python` (tshark) or `ja4plus` (scapy decrypt, no tshark) |
| Library persona “what does aioquic look like?” | **Python aioquic** as **emitter** | Not as capturer |
| Browser / curl-quiche live freshness | Container clients → existing Go capture | Capturer stays Go; clients are black boxes |
| Quick scapy/pcap experiments | Python | Fine offline; do not replace promote path |

**Rule of thumb:** capture+promote = Go; JA4 annotation + exotic clients = Python/containers.

## 2. Minimal capture commands (clients → existing capture)

Keep `compose.yaml` `capture` service. From `quic/`:

```powershell
# build capture bin once
./lab.ps1 build-emitters   # or build quic-capture alone

# aioquic
docker compose -f compose.yaml -f compose.live-clients.yaml --profile aioquic up \
  --abort-on-container-exit capture emit-aioquic

# curl+quiche
docker compose -f compose.yaml -f compose.live-clients.yaml --profile curl up \
  --abort-on-container-exit capture emit-curl-quiche

# headless Chromium H3
docker compose -f compose.yaml -f compose.live-clients.yaml --profile chromium up \
  --abort-on-container-exit capture emit-chromium-h3
```

Host-only peek without Docker clients:

```powershell
# terminal A
./lab.ps1 capture-listen -Target manual

# terminal B — aioquic (venv with aioquic installed)
python emitters/aioquic/emit_initial.py --host 127.0.0.1 --port 4433 --sni aioquic.fp.lab

# curl (needs HTTP/3 build)
curl --http3-only -k --connect-timeout 2 https://127.0.0.1:4433/ -H "Host: curlquiche.fp.lab"

# Chromium
chromium --headless=new --enable-quic --quic-version=h3 `
  --origin-to-force-quic-on=127.0.0.1:4433 `
  --ignore-certificate-errors https://127.0.0.1:4433/
```

SNI / Host should match `targets.yaml` promote ids (`aioquic`, `curlquiche`, `chromium`, …).

### Dockerfile sketch (aioquic) — shipped

See `docker/Dockerfile.aioquic` + `emitters/aioquic/emit_initial.py`.

### curl+quiche one-liner

```text
image: alpine/curl-http3
cmd: curl --http3-only -k --connect-timeout 2 --resolve NAME:4433:capture https://NAME:4433/
```

Debian stock `curl` often **lacks** HTTP/3; do not rely on it in CI without checking `curl -V | grep HTTP3`.

### Chromium one-liner

```text
--enable-quic --quic-version=h3 --origin-to-force-quic-on=NAME:4433
--host-resolver-rules=MAP NAME capture
```

Without `--origin-to-force-quic-on`, headless often stays on H2 (empty alt-svc cache).

## 3. Raw Initial bytes → emit / replay?

**No — not as a drop-in wire replay for product emit.**

| Artifact | Use |
|----------|-----|
| `initials/*.bin` | Observation / compare / golden fixtures |
| `clienthello.bin` + `tp.json` / id-set | Derive **structured** expect + future catalog shorts |
| Emit path | Rebuild from **spec**: uquic `QUICSpec` or ChromeParrot knobs |

Why raw UDP cannot drive emit:

1. Initial AEAD keys = HKDF(DCID); replaying ciphertext with a new DCID fails open.
2. CH random, key_shares, GREASE, token tails change per connection.
3. uquic explicitly builds from `InitialPacketSpec` + `ClientHelloSpec` (framing, PN lens, CRYPTO split, TP list) — same model as uTLS `HelloCustom`, not “send these bytes”.

Lab contract already states: absolute bins are observation; emit rebuilds from spec (`docs/CONTRACT.md`).

Optional later: a **spec extractor** (bins → JSON QUICSpec fields) — still structured emit, not raw replay.

## 4. Compose service sketches

Shipped overlay: `compose.live-clients.yaml` (`emit-aioquic`, `emit-curl-quiche`, `emit-chromium-h3`).

JA4 sidecar (shipped): `compose.ja4.yaml` + `scripts/annotate_ja4.py` →
`./lab.ps1 ja4` writes `expected.ja4` (`q…`) into each `profiles/*/profile.json`.

## Recommendation (actionable)

1. **Do not replace** Go `quic-capture` with Python for the matrix.
2. **Add** live profiles via `compose.live-clients.yaml` (aioquic first — smallest; then curl-quiche; then Chromium).
3. **Annotate JA4** with `ja4plus` or FoxIO+tshark sidecar; store only the `q…` string in `profile.json`.
4. **Keep emit** on Go structured specs; use live bins to *diff* chromeparrot vs real Chromium (see RESULTS_MATRIX).

# Linux-obtainable fingerprint targets (wishlist → active)

Classification for SPEC 062 / lab integration. Host-only / emulator → stay wishlist.

## Promote now (Linux Docker)

| ID | Approach | Family | Notes |
|----|----------|--------|-------|
| `okhttp5` | `clients/okhttp5` (OkHttp 5.x) | android | vs okhttp4 / builtin |
| `grpc-go` | `clients/grpc-go` TLS dial | go | h2 + Go crypto/tls |
| `rust-native-tls` | `clients/rust-native-tls` (reqwest+native-tls) | rust | vs rustls |
| `chrome-linux` | `clients/chrome-linux` Dockerfile | chrome | **track: latest** |
| `brave` | `clients/brave` Dockerfile | **brave** | **track: latest** |
| `electron-sample` | `clients/electron-sample` Electron 33 + Xvfb | electron | pinned |
| `python-aiohttp` | aiohttp ClientSession | python | pinned |
| `git-https` | alpine/git:latest | openssl | **track: latest** |
| `firefox-release` | debian firefox-esr | firefox | **track: latest** |

See [EXTENDING.md](EXTENDING.md) for latest/archive/dedup. `./lab.ps1 refresh-latest` pulls floating channels.

## Keep wishlist / later

| ID | Why |
|----|-----|
| `safari-macos-host` | needs macOS host |
| `chrome-android-device` | emulator/device; heavy |
| `edge-linux-deb` | low value (curl-imp + builtin exist) |
| `yandex-linux` | RU repo friction; do after brave/chrome-linux |
| `opera`, `aws-cli`, `grpc-java` | low priority |

## Integration rule

Same contract as [docs/EXTENDING.md](docs/EXTENDING.md): `targets.yaml` → runner in `gen-compose.py` / `clients/` → capture → verify → export.

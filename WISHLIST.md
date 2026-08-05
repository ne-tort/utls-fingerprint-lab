# Wishlist — deferred fingerprint targets

Targets from SPEC 062 / `targets.yaml` (`status: wishlist`) not yet captured in this lab.
All items below are marked `utls_ready: true` unless noted — when implemented they must emit `utls-raw-clienthello-v1`.

| ID | Why | Suggested kind | Priority |
|----|-----|----------------|----------|
| `chromium-beta` | Drift vs stable Chrome | compose (beta image) | medium |
| `brave` | Chromium-fork JA4 check | compose (deb) | medium |
| `edge-linux-deb` | Real Edge vs curl-imp/builtin | compose (deb) | low (proxy exists) |
| `okhttp5` | OkHttp 5 vs 4 / HelloAndroid | compose | medium |
| `rust-native-tls` | Contrast with rustls | compose | low |
| `electron-sample` | Lagging embedded Chromium | compose | medium |
| `grpc-go` | h2 + Go TLS mesh FP | compose | medium |
| `grpc-java` | JSSE/gRPC | compose | low |
| `aws-cli` | botocore mass traffic | compose | low |
| `git-https` | libcurl via git | compose | low |
| `python-aiohttp` | async Python TLS | compose | low |
| `safari-macos-host` | Real Safari (impersonate exists) | host capture | later |
| `chrome-android-device` | Real mobile Chrome | emulator | later |
| `yandex-linux` | RU Chromium-fork | compose (deb) | later |
| `opera` | Chromium-fork | compose | skip/low |

## Explicitly out of scope (skip)

- `vivaldi`, `nginx-proxy`, `wget`, `docker-pull`, `java-minecraft` — low value or fragile installs.

## How to promote

1. Move entry in `targets.yaml` from `wishlist` → `active`, set `kind`.
2. Follow [docs/EXTENDING.md](docs/EXTENDING.md).
3. `./lab.ps1 capture -Id …` && `./lab.ps1 verify -Id …`.
4. Remove from this table when profile is committed.

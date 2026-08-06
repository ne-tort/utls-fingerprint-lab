# Wishlist — deferred fingerprint targets

See also [docs/LINUX_WAVE.md](docs/LINUX_WAVE.md) for Linux-promoted actives.

| ID | Why | Suggested kind | Priority |
|----|-----|----------------|----------|
| `chromium-beta` | Drift vs chrome-linux | compose (google-chrome-beta) | medium |
| `edge-linux-deb` | Real Edge vs curl-imp/builtin | compose (deb) | low |
| `grpc-java` | JSSE/gRPC | compose | low |
| `aws-cli` | botocore mass traffic | compose | low |
| `safari-macos-host` | Real Safari | host capture | later |
| `chrome-android-device` | Real mobile Chrome | emulator | later |
| `yandex-linux` | RU Chromium-fork | compose (deb) | medium |
| `opera` | Chromium-fork | compose | skip/low |

## Promoted to active (this wave)

`okhttp5`, `grpc-go`, `rust-native-tls`, `python-aiohttp`, `chrome-linux`, `brave`, `electron-sample`, `git-https`.

## Explicitly out of scope (skip)

- `vivaldi`, `nginx-proxy`, `wget`, `docker-pull`, `java-minecraft` — low value or fragile installs.

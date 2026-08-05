# Architecture

```
clients / emit-builtin / curl-impersonate
        │  TLS ClientHello (SNI = <id>.fp.lab.local)
        ▼
   capture (:8443)  ──peek──► captures/<n>-<id>/
        │                      profiles/<id>/{clienthello.bin,profile.json,meta.json}
        ▼
   verify (HelloCustom blunt mimicry) ──must match expected.ja4──► OK
```

## Components

- **capture** — terminates TLS enough to read ClientHello, fingerprints JA3/JA4, auto-promotes importable profile.
- **labctl** — reads `targets.yaml`, drives capture/verify/catalog.
- **compose.yaml** — services + DNS aliases for `*.fp.lab.local` on the capture container (regenerated from targets).

## Caching

- BuildKit enabled (`DOCKER_BUILDKIT=1`).
- Go modules / build cache: `--mount=type=cache` in `capture/`, `tools/`, `clients/go-http` Dockerfiles.
- Maven: `--mount=type=cache,target=/root/.m2` for OkHttp.
- Cargo: registry/git cache mounts for rustls client.

## Isolation

This directory is a **standalone git repository**
([ne-tort/utls-fingerprint-lab](https://github.com/ne-tort/utls-fingerprint-lab)).
It does not import sing-box packages. Parent `sing-box-lx` vendors it as submodule
`lx-test/utls-fingerprint-docker`.

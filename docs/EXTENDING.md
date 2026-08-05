# Extending the lab

## Contract

1. Add a row to **`targets.yaml`** (`status: active` or `wishlist`).
2. Implement the runner for its `kind`.
3. Regenerate compose aliases: `python scripts/gen-compose.py` (also done by `lab.ps1 build|capture`).
4. Capture: `./lab.ps1 capture -Id <id>`
5. Verify: `./lab.ps1 verify -Id <id>`
6. Commit `profiles/<id>/` if the JA4 is stable enough to keep.

## Kinds

| kind | How it runs | When to use |
|------|-------------|-------------|
| `compose` | `docker compose run --rm <service>` | Dedicated service in `compose.yaml` / `clients/` |
| `emit-builtin` | `emit-builtin -id <emit_id>` | metacubex/utls Hello* presets |
| `curl-impersonate` | `WRAP` + `TARGET_ID` via `curl-impersonate-one` | Browser-identical BoringSSL curl wrappers |
| `wishlist` | not runnable | Document intent only |

### Adding a `compose` client

1. Create `clients/<id>/` (Dockerfile + code) **or** inline service in `scripts/gen-compose.py`.
2. Ensure the client:
   - dials `capture:8443` (or uses `--connect-to` / custom dialer)
   - sets TLS SNI to `<id>.fp.lab.local`
   - sends `X-Target-Id: <id>` when speaking HTTP
3. Add `targets.yaml` entry with `kind: compose`, `service: <name>`, `needs_dns_alias: true` if the client resolves the SNI hostname via DNS.
4. Prefer BuildKit cache mounts in Dockerfiles (`--mount=type=cache,...`).

### Adding `emit-builtin`

1. Support the Hello ID in `tools/cmd/emit-builtin/main.go` (`mapID`).
2. Add `targets.yaml` with `kind: emit-builtin`, `emit_id: …`.

### Adding `curl-impersonate`

1. Confirm wrapper exists in `lexiforest/curl-impersonate:alpine` (`curl_chrome146`, …).
2. Add `targets.yaml` with `kind: curl-impersonate`, `wrapper: curl_…`.

## Naming

- Profile directory = `id` from `targets.yaml`.
- SNI = `<id>.fp.lab.local` (verify uses `verify-<id>.fp.lab.local`).
- Do not commit `captures/`, `verify-*` dirs, or `bin/`.

## uTLS / sing-box readiness

Every `active` target with `utls_ready: true` must produce `clienthello.bin` loadable by:

```go
utls.Fingerprinter{AllowBluntMimicry: true}.RawClientHello(bin)
```

If blunt mimicry fails, fix capture (prefer full TLS record) before committing the profile.

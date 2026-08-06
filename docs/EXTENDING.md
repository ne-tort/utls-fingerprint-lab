# Extending the lab

## Contract

1. Add a row to **`targets.yaml`** (`status: active` or `wishlist`).
2. Set **`track`**: `pinned` (default) or `latest` (floating channel).
3. Implement the runner for its `kind`.
4. Regenerate compose aliases: `python scripts/gen-compose.py` (also done by `lab.ps1 build|capture`).
5. Capture: `./lab.ps1 capture -Id <id>` (or `./lab.ps1 refresh-latest` for all `track: latest`).
6. Verify: `./lab.ps1 verify -Id <id>`
7. Commit `profiles/<id>/` (and `targets.archive.yaml` if archive ran) when the JA4 is worth keeping.

## Track / versioning

| Field | Meaning |
|-------|---------|
| `track: latest` | Floating channel. Rank forced to head of family (`version` → 9999). Capture uses `image_policy: pull` (docker `--pull always` + `CACHEBUST` on Dockerfiles under `clients/`). |
| `track: pinned` | Reproducible pin (default). Optional `pin:` label (`"146"`, `"5.0.0-alpha.14"`). |
| `image_policy` | `pull` or `cache` (defaults: latest→pull, pinned→cache). |
| `version` | Export rank only (higher = newer). Latest overrides to 9999. |

### Latest refresh + archive

1. `./lab.ps1 refresh-latest` captures every `track: latest` with pull.
2. Each latest slot has `profiles/<id>/slot.json` (`ja4`, `software_major`, …).
3. If JA4 **unchanged** → overwrite profile/meta; no new short name.
4. If JA4 **changed** → previous profile copied to `profiles/<family>-<major>/` and a pinned row appended to **`targets.archive.yaml`** (merged by labctl).

### Dedup on export

`labctl export` (and `lab.ps1 export`) collapses **identical JA4 within a family** to one short name. Duplicate lab IDs become `aliases` in `catalog.json` (no extra `profiles/<short>/` copy). Cross-family equal JA4 (e.g. electron ≈ chrome) stays separate.

`export --check-dedup` fails if two short names in one family share a JA4.

## Kinds

| kind | How it runs | When to use |
|------|-------------|-------------|
| `compose` | `docker compose run --rm <service>` | Dedicated service in `compose.yaml` / `clients/` |
| `emit-builtin` | `emit-builtin -id <emit_id>` | metacubex/utls Hello* presets |
| `curl-impersonate` | `WRAP` + `TARGET_ID` via `curl-impersonate-one` | Browser-identical BoringSSL curl wrappers |
| `wishlist` | not runnable | Document intent / auto-archive placeholders |

### Adding a `compose` client

1. Create `clients/<id>/` (Dockerfile + code) **or** inline service in `scripts/gen-compose.py`.
2. For floating browsers, accept `ARG CACHEBUST` in the install layer.
3. Ensure the client:
   - dials `capture:8443` (or uses `--connect-to` / custom dialer)
   - sets TLS SNI to `<id>.fp.lab.local`
   - sends `X-Target-Id: <id>` when speaking HTTP
4. Add `targets.yaml` entry with `kind: compose`, `service: <name>`, `needs_dns_alias: true` if needed.
5. Prefer BuildKit cache mounts in Dockerfiles (`--mount=type=cache,...`).

### Adding `emit-builtin`

1. Support the Hello ID in `tools/cmd/emit-builtin/main.go` (`mapID`).
2. Add `targets.yaml` with `kind: emit-builtin`, `emit_id: …`, `track: pinned`.

### Adding `curl-impersonate`

1. Confirm wrapper exists in `lexiforest/curl-impersonate:alpine` (`curl_chrome146`, …).
2. Add `targets.yaml` with `kind: curl-impersonate`, `wrapper: curl_…`, `track: pinned`, `pin: "146"`.

## Naming

- Profile directory = `id` from `targets.yaml` / `targets.archive.yaml`.
- SNI = `<id>.fp.lab.local` (verify uses `verify-<id>.fp.lab.local`).
- Do not commit `captures/`, `verify-*` dirs, or `bin/`.

## uTLS / sing-box readiness

Every `active` target with `utls_ready: true` must produce `clienthello.bin` loadable by:

```go
utls.Fingerprinter{AllowBluntMimicry: true}.RawClientHello(bin)
```

If blunt mimicry fails, fix capture (prefer full TLS record) before committing the profile.

# Optional hy2 outbound configs (wishlist)

JSON under this directory mirrors demux-quic-auth client pattern for **Initial
dump only** (no `quic_auth`). Wire shape should ≈ matrix `chromeparrot` /
`quicgo` emitters.

Requires a prebuilt linux `sing-box` (e.g. from
`lx-test/demux-quic-auth-docker/build-bin.ps1`) — not built by `lab.ps1 matrix`.

```text
# capture already up
docker run --rm --network quic-fp-lab_fpnet \
  -v $PWD/config/hy2-parrot.json:/etc/sing-box/config.json:ro \
  -v /path/to/sing-box:/usr/local/bin/sing-box:ro \
  debian:bookworm-slim \
  sing-box run -c /etc/sing-box/config.json
# then: curl -x http://<container>:1080 http://example.com/
```

Handshake to capture may fail (blackhole); Initial dump is enough.
